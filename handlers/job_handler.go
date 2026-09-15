package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Wandyy20/job-queue/models"
	"github.com/Wandyy20/job-queue/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type JobHandler struct {
	jobStore      store.JobStore
	jobEventStore store.JobEventStore
}

func NewJobHandler(jobStore store.JobStore, jobEventStore store.JobEventStore) *JobHandler {
	return &JobHandler{jobStore: jobStore, jobEventStore: jobEventStore}
}

type CreateJobRequest struct {
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	MaxAttempts int             `json:"maxAttempts"`
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.MaxAttempts == 0 {
		req.MaxAttempts = 5
	}

	job := &models.Job{
		Type:        req.Type,
		Payload:     req.Payload,
		MaxAttempts: req.MaxAttempts,
		RunAt:       time.Now(),
	}

	err := h.jobStore.Enqueue(r.Context(), job)
	if err != nil {
		http.Error(w, "Failed to enqeue job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	job, err := h.jobStore.GetByID(r.Context(), jobID)
	if err != nil {
		http.Error(w, "Failed to get job", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) GetByStatus(w http.ResponseWriter, r *http.Request) {

	status := r.URL.Query().Get("status")
	jobs, err := h.jobStore.List(r.Context(), status)
	if err != nil {
		http.Error(w, "failed to list jobs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (h *JobHandler) GetJobEvents(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	jobEvents, err := h.jobEventStore.ListEvents(r.Context(), jobID)
	if err != nil {
		http.Error(w, "Failed to get job events", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobEvents)
}

func (h *JobHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	err = h.jobStore.Cancel(r.Context(), jobID)

	if err != nil {
		if err.Error() == "job not found or not cancellable" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "failed to cancel job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)

}
