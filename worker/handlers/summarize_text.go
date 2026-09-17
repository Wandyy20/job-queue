package handlers

import (
	"context"
	"encoding/json"
	"fmt"
)

type SummarizePayload struct {
	Text string `json:"text"`
}

func HandleSummarizeText(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input SummarizePayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	summary, err := callGemini(ctx, "Summarize this text in 2-3 sentences: "+input.Text)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{"summary": summary}
	return json.Marshal(result)
}
