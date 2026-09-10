package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Wandyy20/job-queue/models"
	"github.com/Wandyy20/job-queue/store"
)

type Pool struct {
	jobStore store.JobStore
	registry *Registry
	numWorkers int 
}

func NewPool(jobStore store.JobStore, registry *Registry, numWorkers int) *Pool{
	return &Pool{
		jobStore: jobStore,
		registry: registry,
		numWorkers: numWorkers,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.numWorkers; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		go p.runWorker(ctx, workerID)
	}
}

func (p *Pool) runWorker(ctx context.Context, workerID string) {
	for {
		select {
		case <- ctx.Done():
			log.Printf("%s shutting down", workerID)
			return
		default:
		}

		job, err := p.jobStore.Claim(ctx, workerID)
		if err != nil {
			log.Printf("%s claim error: %v", workerID, err)
			time.Sleep(1 * time.Second)
			continue
		}
		if job == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		p.processJob(ctx, job)
	}
}

func (p *Pool) processJob(ctx context.Context, job *models.Job) {
	handler, ok := p.registry.Get(job.Type)
	if !ok {
		p.jobStore.Fail(ctx, job.ID, "Unknown job type"+job.Type)
		return
	}

	err := handler(ctx, job.Payload)
	if err != nil {
		p.jobStore.Fail(ctx, job.ID, err.Error())
		return
	}

	p.jobStore.Complete(ctx, job.ID)
}