package store

import (
	"context"
	"encoding/json"

	"github.com/Wandyy20/job-queue/models"
	"github.com/google/uuid"
)

type JobStore interface {
	Enqueue(ctx context.Context, job *models.Job) error
	Claim(ctx context.Context, workerID string) (*models.Job, error)
	Complete(ctx context.Context, jobID uuid.UUID, result json.RawMessage) error
	Fail(ctx context.Context, jobID uuid.UUID, errMsg string) error
	GetByID(ctx context.Context, jobID uuid.UUID) (*models.Job, error)
	List(ctx context.Context, status string) ([]*models.Job, error)
	Cancel(ctx context.Context, jobID uuid.UUID) error
}

type JobEventStore interface {
	LogEvent(ctx context.Context, event *models.JobEvent) error
	ListEvents(ctx context.Context, jobID uuid.UUID) ([]*models.JobEvent, error)
}
