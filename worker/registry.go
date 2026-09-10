package worker

import (
	"context"
	"encoding/json"
)

type HandlerFunc func(ctx context.Context, payload json.RawMessage) error

type Registry struct {
	handlers map[string]HandlerFunc
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]HandlerFunc)}
}

func (r *Registry) Register(jobType string, handler HandlerFunc) {
	r.handlers[jobType] = handler
}

func (r *Registry) Get(jobType string) (HandlerFunc, bool) {
	handler, ok := r.handlers[jobType]
	return handler, ok
}