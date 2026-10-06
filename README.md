# Job Queue Service

A concurrency-safe background job queue built with Go and PostgreSQL, with a React dashboard for submitting jobs and monitoring them in real time. Workers claim jobs using row-level locking (`SELECT ... FOR UPDATE SKIP LOCKED`), guaranteeing exactly-once processing — proven with an automated concurrency test where 10 goroutines race for the same job.

**Live demo:** https://job-queue-eta.vercel.app/
**API:** https://job-queue-production-5e86.up.railway.app

## Features

- Submit background jobs via REST API or dashboard, with arbitrary JSON payloads
- Concurrency-safe job claiming (tested with concurrent goroutines)
- Automatic retry with exponential backoff, dead-letter handling for jobs that exceed max attempts
- Stale job recovery — jobs stuck in `processing` (e.g. from a crashed worker) are automatically reset and retried
- Full job history via `job_events`
- Job handler registry — add new job types without touching worker internals
- Graceful shutdown — in-flight jobs finish before exit
- Job cancellation for pending jobs
- Dashboard: live status overview, filterable job list, per-job detail view with event timeline, result preview/download (PDF, CSV, images)

## Job types

| Type | Description |
|---|---|
| `test_job` | No-op, used for testing the queue itself |
| `send_webhook` | POSTs a payload to a given URL, retries on failure |
| `summarize_text` | Summarizes text via the Gemini API |
| `classify_sentiment` | Classifies text sentiment via the Gemini API |
| `translate_text` | Translates text via the Gemini API |
| `resize_image` | Resizes a base64-encoded image |
| `generate_pdf_report` | Generates a PDF from structured input |
| `csv_export` | Generates a CSV from structured input |
| `classify_toxic_comment` | Classifies Indonesian text as hate speech / not, using a self-fine-tuned IndoBERT model ([repo](https://github.com/Wandyy20/indonesian-hate-speech-classifier)) |

> **Note:** `classify_toxic_comment` calls a separate FastAPI model-serving service that is **not deployed** alongside this demo — it requires its own inference infrastructure (model weights, GPU/CPU resources) that's out of scope for this deployment. It runs and is fully tested locally; see the [model repo](https://github.com/Wandyy20/indonesian-hate-speech-classifier) for details. In production, this job type would point at a dedicated model-serving deployment.

## Tech stack

Go (`net/http` + [chi](https://github.com/go-chi/chi)) · PostgreSQL (Neon, `pgx/v5`) · React + Vite + Tailwind (frontend) · Railway (API) · Vercel (frontend)

## Architecture

```
├── models/          # Job, JobEvent
├── store/           # Storage interfaces + Postgres implementation
├── worker/          # Worker pool + job handler registry
├── handlers/        # HTTP handlers
├── migrations/      # SQL schema
├── frontend/         # React dashboard
├── Dockerfile
└── main.go
```

Storage is defined entirely behind Go interfaces (`store.JobStore`, `store.JobEventStore`). Handlers and the worker pool depend only on these interfaces, never on the concrete Postgres implementation.

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

`SKIP LOCKED` lets concurrent workers skip rows already locked by another worker instead of waiting — this is what makes it safe for multiple workers to pull from the same queue without double-processing a job.

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | `/jobs` | Submit a new job |
| GET | `/jobs/{id}` | Get a job by ID |
| GET | `/jobs/{id}/events` | Get a job's event history |
| GET | `/jobs?status=pending` | List jobs by status (omit for all) |
| DELETE | `/jobs/{id}` | Cancel a pending job |

### Example: submit a job

```
POST /jobs
Content-Type: application/json

{
  "type": "summarize_text",
  "payload": { "text": "..." }
}
```

## Running locally

### Prerequisites
- Go 1.23+
- Node.js 18+
- A PostgreSQL database (local via Docker, or a hosted instance like Neon)

### Backend

```bash
git clone https://github.com/Wandyy20/job-queue.git
cd job-queue

# run the migrations against your database
psql "<your-database-url>" -f migrations/001_init_schema.sql
psql "<your-database-url>" -f migrations/002_add_result_column.sql
```

Create a `.env` file in the project root:
```
DATABASE_URL=postgres://user:password@host/dbname?sslmode=require
GEMINI_API_KEY=your-gemini-api-key
```

```bash
go mod tidy
go run main.go
```

API runs on `http://localhost:8080`.

### Frontend

```bash
cd frontend
npm install
```

Create `frontend/.env`:
```
VITE_API_URL=http://localhost:8080
```

```bash
npm run dev
```

Dashboard runs on `http://localhost:5173`.

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
func(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)
```

## Scope

Built around one core problem: **safe concurrent job processing**. Auth is deliberately out of scope, keeping the focus on queueing, concurrency correctness, and reliability — with a handful of real job types (webhooks, AI calls, file generation, a self-trained ML model) to prove the design generalizes beyond a toy example.
