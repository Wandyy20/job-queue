package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
	registry.Register("csv_export", jobHandlers.HandleCSVExport)
	registry.Register("classify_sentiment", jobHandlers.HandleClassifySentiment)
	registry.Register("translate_text", jobHandlers.HandleTranslateText)
	registry.Register("classify_toxic_comment", jobHandlers.HandleClassifyToxicComment)
	workerPool := worker.NewPool(jobStore, registry, 5)
	ctx, cancel := context.WithCancel(context.Background())
	workerPool.Start(ctx)

	log.Println("worker pool started")

	jobHandler := httpHandlers.NewJobHandler(jobStore, jobEventStore)
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
	}))

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
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				recovered, err := jobStore.RecoverStaleJobs(ctx, 5*time.Minute)
				if err != nil {
					log.Printf("stale job recovery error: %v", err)
				} else if recovered > 0 {
					log.Printf("recovered %d stale job(s)", recovered)
				}
			}
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
