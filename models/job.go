package models

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Job struct {
	ID          uuid.UUID
	Type        string
	Payload     json.RawMessage
	Status      string
	Attempts    int
	MaxAttempts int
	RunAt       time.Time
	LockedAt    *time.Time
	LockedBy    *string
	LastError   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Result      *json.RawMessage
}

type JobEvent struct {
	ID        uuid.UUID
	JobID     uuid.UUID
	Event     string
	Detail    *string
	CreatedAt time.Time
}
