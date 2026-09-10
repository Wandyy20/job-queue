package postgres

import (
	"context"
	"os"
	"sync"
	"testing"
	"encoding/json"
	"time"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/Wandyy20/job-queue/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	_ = godotenv.Load("../../.env")
	dbURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	return pool
}

func TestClaimWorkers(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	store := NewPostgresJobStore(pool)
	ctx := context.Background()

	job := &models.Job{
		Type:        "test_job",
		Payload:     json.RawMessage(`{}`),
		MaxAttempts: 3,
		RunAt:       time.Now(),
	}
	err := store.Enqueue(ctx, job)
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	numWorkers := 10
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0

	for i := 0; i < numWorkers ; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			claimedJob, err := store.Claim(ctx, fmt.Sprintf("worker-%d", workerID))
			if err != nil {
				t.Errorf("claim error: %v", err)
				return
			}
			if claimedJob != nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if successCount != 1 {
		t.Errorf("expected exactly 1 worker to claim the job, got %d", successCount)
	}
}