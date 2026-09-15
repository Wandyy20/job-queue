# Job Queue Service

A concurrency-safe background job queue built with Go and PostgreSQL. Workers claim jobs using row-level locking (`SELECT ... FOR UPDATE SKIP LOCKED`), guaranteeing exactly-once processing — proven with an automated concurrency test where 10 goroutines race for the same job.

## Features

- Submit background jobs via REST API with arbitrary JSON payloads
- Concurrency-safe job claiming (tested with concurrent goroutines)
- Automatic retry with exponential backoff, dead-letter handling for jobs that exceed max attempts
- Full job history via `job_events`
- Job handler registry — add new job types without touching worker internals
- Graceful shutdown — in-flight jobs finish before exit
- Job cancellation for pending jobs

## Tech stack

Go (`net/http` + [chi](https://github.com/go-chi/chi)) · PostgreSQL (`pgx/v5`) · React + Vite (frontend)

## Architecture

```
├── models/          # Job, JobEvent
├── store/           # Storage interfaces + Postgres implementation
├── worker/          # Worker pool + job handler registry
├── handlers/        # HTTP handlers
├── migrations/      # SQL schema
├── frontend/         # React dashboard
└── main.go
```

## How job claiming works

```sql
UPDATE jobs
SET status = 'processing', locked_at = now(), locked_by = $1, attempts = attempts + 1
WHERE id = (
    SELECT id FROM jobs
    WHERE status = 'pending' AND run_at <= now()
    ORDER BY run_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;
```

`SKIP LOCKED` lets concurrent workers skip rows already locked by another worker instead of waiting — this is what makes it safe for multiple workers to pull from the same queue without double-processing.

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/jobs` | Submit a new job |
| GET | `/jobs/{id}` | Get a job by ID |
| GET | `/jobs/{id}/events` | Get a job's event history |
| GET | `/jobs?status=pending` | List jobs by status |
| DELETE | `/jobs/{id}` | Cancel a pending job |

## Getting started

### Prerequisites
- Go 1.23+
- Docker (for PostgreSQL) — or a local PostgreSQL install

### Setup

```bash
docker run --name jobqueue-pg -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=job_queue -p 5432:5432 -d postgres:16
docker exec -i jobqueue-pg psql -U postgres -d job_queue < migrations/001_init_schema.sql
```

Create a `.env` file in the project root:
```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/job_queue?sslmode=disable
```

Install dependencies and run:
```bash
go mod tidy
go run main.go
```

Server starts on `http://localhost:8080`.

## Testing

```bash
go test ./store/postgres/... -v
```

Includes an automated concurrency test: 10 goroutines attempt to claim the same job simultaneously, asserting exactly one succeeds — proving the `SKIP LOCKED` claim query is safe under concurrent load.

## Adding a new job type

Job types are registered against a handler function — no changes to worker internals required:

```go
registry.Register("send_email", handlers.HandleSendEmail)
```

A handler just needs to match this signature:

```go
func(ctx context.Context, payload json.RawMessage) error
```

## Scope

Intentionally scoped around one core problem: **safe concurrent job processing**. Auth and job-type-specific business logic are kept minimal, keeping the focus on queueing, concurrency correctness, and reliability.
