package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	httpHandlers "github.com/Wandyy20/job-queue/handlers"
	"github.com/Wandyy20/job-queue/store/postgres"
	"github.com/Wandyy20/job-queue/worker"
	jobHandlers "github.com/Wandyy20/job-queue/worker/handlers"
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

	registry := worker.NewRegistry()
	registry.Register("test_job", jobHandlers.HandleTestJob)
	registry.Register("send_webhook", jobHandlers.HandleSendWebhook)
	registry.Register("summarize_text", jobHandlers.HandleSummarizeText)
	registry.Register("resize_image", jobHandlers.HandleResizeImage)
	registry.Register("generate_pdf_report", jobHandlers.HandleGeneratePDFReport)
	workerPool := worker.NewPool(jobStore, registry, 3)
	ctx, cancel := context.WithCancel(context.Background())
	workerPool.Start(ctx)

	log.Println("worker pool started")

	jobHandler := httpHandlers.NewJobHandler(jobStore, jobEventStore)
	r := chi.NewRouter()
	r.Post("/jobs", jobHandler.CreateJob)
	r.Get("/jobs", jobHandler.GetByStatus)
	r.Get("/jobs/{id}", jobHandler.GetJob)
	r.Get("/jobs/{id}/events", jobHandler.GetJobEvents)
	r.Delete("/jobs/{id}", jobHandler.CancelJob)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("HTTP server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutdown signal received, stopping workers...")
	cancel()
	workerPool.Wait()
	log.Println("all workers stopped, exiting")
}
