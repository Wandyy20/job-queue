package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Wandyy20/job-queue/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresJobStore struct {
	db *pgxpool.Pool
}

func NewPostgresJobStore(db *pgxpool.Pool) *PostgresJobStore {
	return &PostgresJobStore{db: db}
}

func (s *PostgresJobStore) Enqueue(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (type, payload, status, max_attempts, run_at)
		VALUES ($1, $2, 'pending', $3, $4)
		RETURNING id, status, attempts, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query,
		job.Type,
		job.Payload,
		job.MaxAttempts,
		job.RunAt,
	).Scan(&job.ID, &job.Status, &job.Attempts, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresJobStore) GetByID(ctx context.Context, jobID uuid.UUID) (*models.Job, error) {
	var job models.Job

	query := `SELECT id, type, payload, status, attempts, max_attempts, run_at, locked_at, locked_by, last_error, result, created_at, updated_at FROM jobs WHERE id = $1`

	err := s.db.QueryRow(ctx, query, jobID).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.RunAt,
		&job.LockedAt,
		&job.LockedBy,
		&job.LastError,
		&job.Result,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *PostgresJobStore) Complete(ctx context.Context, jobID uuid.UUID, result json.RawMessage) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE jobs SET status = 'completed', result = $2, updated_at = now() WHERE id = $1`, jobID, result)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO job_events (job_id, event) VALUES ($1, 'succeeded')`, jobID)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresJobStore) Fail(ctx context.Context, jobID uuid.UUID, errMsg string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var newStatus string
	err = tx.QueryRow(ctx,`
		UPDATE jobs
		SET 
			status = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'pending' END,
			last_error = $2,
			run_at = CASE WHEN attempts >= max_attempts THEN run_at ELSE now() + INTERVAL '5 seconds' * power(2, attempts) END,
			updated_at = now()
		WHERE id = $1
		RETURNING status
	`, jobID, errMsg).Scan(&newStatus)

	if err != nil {
		return err
	}

	event := "retried"
	if newStatus == "dead" {
		event = "dead_lettered"
	}

	_, err = tx.Exec(ctx, `INSERT INTO job_events (job_id, event, detail) VALUES ($1, $2, $3)`, jobID, event, errMsg)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresJobStore) Claim(ctx context.Context, workerID string) (*models.Job, error) {
	var job models.Job

	query := `
		UPDATE jobs
		SET status = 'processing', locked_at = now(), locked_by = $1, attempts = attempts + 1 WHERE id = (
			SELECT id FROM jobs
			WHERE status = 'pending' AND run_at <= now()
			ORDER BY run_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, type, payload, status, attempts, max_attempts, run_at, locked_at, locked_by, last_error, result, created_at, updated_at
	`
	err := s.db.QueryRow(ctx, query, workerID).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.RunAt,
		&job.LockedAt,
		&job.LockedBy,
		&job.LastError,
		&job.Result,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil 
		}
		return nil, err  
	}
	return &job, nil
}

func (s *PostgresJobStore) List(ctx context.Context, status string) ([]*models.Job, error) {
	query := `SELECT id, type, payload, status, attempts, max_attempts, run_at, locked_at, locked_by, last_error, result, created_at, updated_at FROM jobs WHERE status = $1 ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]*models.Job, 0)
	for rows.Next() {
		var job models.Job
		err := rows.Scan(
			&job.ID,
			&job.Type,
			&job.Payload,
			&job.Status,
			&job.Attempts,
			&job.MaxAttempts,
			&job.RunAt,
			&job.LockedAt,
			&job.LockedBy,
			&job.LastError,
			&job.Result,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}
	return jobs, rows.Err()
}

func (s *PostgresJobStore) Cancel(ctx context.Context, jobID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		UPDATE jobs SET status = 'cancelled', updated_at = now()
		WHERE id = $1 AND status = 'pending'
	`, jobID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("job not found or not cancellable")
	}

	_, err = tx.Exec(ctx, `INSERT into job_events (job_id, event) VALUES ($1, 'cancelled')`, jobID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
