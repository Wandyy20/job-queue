package handlers

import (
	"context"
	"encoding/json"
	"fmt"
)

type SentimentPayload struct {
	Text string `json:"text"`
}

func HandleClassifySentiment(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input SentimentPayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	sentiment, err := callGemini(ctx, "Classify the sentiment of this text as exactly one word (positive, negative, or neutral): "+input.Text)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{"sentiment": sentiment}
	return json.Marshal(result)
}
