package store

import (
	"context"
	"github.com/google/uuid"
	"github.com/Wandyy20/job-queue/models"
)

type JobStore interface {
	Enqueue(ctx context.Context, job *models.Job) error
	Claim(ctx context.Context, workerID string) (*models.Job, error)
	Complete(ctx context.Context, jobID uuid.UUID) error
	Fail(ctx context.Context, jobID uuid.UUID, errMsg string) error
	GetByID(ctx context.Context, jobID uuid.UUID) (*models.Job, error)
	List(ctx context.Context, status string) ([]*models.Job, error)
}

type JobEventStore interface {
	LogEvent(ctx context.Context, event *models.JobEvent) error
	ListEvents(ctx context.Context, jobID uuid.UUID) ([]*models.JobEvent, error)
}