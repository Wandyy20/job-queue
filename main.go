package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	workerPool := worker.NewPool(jobStore, registry, 3)
	ctx, cancel := context.WithCancel(context.Background())
	workerPool.Start(ctx)

	log.Println("worker pool started")
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	log.Println("shutdown signal received, stopping workers...")
	cancel()

	workerPool.Wait()
	log.Println("all workers stopped, exiting")
}
