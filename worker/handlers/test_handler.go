package handlers

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleTestJob(ctx context.Context, payload json.RawMessage) error {
	fmt.Printf("Processing test job with payload: %s\n", payload)
	return  nil
}