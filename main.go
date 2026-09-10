package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/Wandyy20/job-queue/store/postgres"
	"github.com/Wandyy20/job-queue/worker"
	"github.com/Wandyy20/job-queue/worker/handlers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("no .env file found, using system env vars")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}

	log.Println("connected to database successfully")

	jobStore := postgres.NewPostgresJobStore(dbPool)
	jobEventStore := postgres.NewPostgresJobEventStore(dbPool)
	
	_ = jobEventStore

	registry := worker.NewRegistry()
	registry.Register("test_job", handlers.HandleTestJob)
	pool := worker.NewPool(jobStore, registry, 3)
	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)

	log.Println("worker pool started")
	time.Sleep(30 * time.Second)
	cancel()
}
