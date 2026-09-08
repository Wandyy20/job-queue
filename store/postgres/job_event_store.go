package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Wandyy20/job-queue/models"
	"github.com/google/uuid"
)

type PostgresJobEventStore struct {
	db *pgxpool.Pool
}

func NewPostgresJobEventStore(db *pgxpool.Pool) *PostgresJobEventStore {
	return &PostgresJobEventStore{db: db}
}

func (s *PostgresJobEventStore) LogEvent(ctx context.Context, event *models.JobEvent) error {
	query := `
		INSERT INTO job_events (job_id, event, detail) VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := s.db.QueryRow(ctx, query, event.JobID, event.Event, event.Detail).Scan(&event.ID, &event.CreatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresJobEventStore) ListEvents(ctx context.Context, jobID uuid.UUID) ([]*models.JobEvent, error) {
	query := `
		SELECT id, job_id, event, detail, created_at FROM job_events
		WHERE job_id = $1 ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobEvents []*models.JobEvent
	for rows.Next() {
		var jobEvent models.JobEvent
		err := rows.Scan(
			&jobEvent.ID,
			&jobEvent.JobID,
			&jobEvent.Event,
			&jobEvent.Detail,
			&jobEvent.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobEvents = append(jobEvents, &jobEvent)
	}
	return jobEvents, rows.Err()
}