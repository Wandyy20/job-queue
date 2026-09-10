package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/Wandyy20/job-queue/models"
	"github.com/Wandyy20/job-queue/store/postgres"
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

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}

	log.Println("connected to database successfully")

	jobStore := postgres.NewPostgresJobStore(pool)
	jobEventStore := postgres.NewPostgresJobEventStore(pool)
	_ = jobStore
	_ = jobEventStore
}
