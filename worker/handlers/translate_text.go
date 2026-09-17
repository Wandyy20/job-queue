package handlers

import (
	"context"
	"encoding/json"
	"fmt"
)

type TranslatePayload struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

func HandleTranslateText(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input TranslatePayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	translate, err := callGemini(ctx, "Translate this text into "+input.Language+": "+input.Text)

	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"translated_text": translate,
	}

	return json.Marshal(result)

}
