package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type WebHookPayload struct {
	URL  string          `json:"URL"`
	Data json.RawMessage `json:"data"`
}

func HandleSendWebhook(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input WebHookPayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, input.URL, bytes.NewReader(input.Data))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"sent_url":    input.URL,
	}
	resultJSON, _ := json.Marshal(result)

	return resultJSON, nil
}
