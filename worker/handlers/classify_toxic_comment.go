package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClassifyToxicPayload struct {
	Text string `json:"text"`
}

type classifyModelRequest struct {
	Text string `json:"text"`
}

type classifyModelResponse struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
}

func HandleClassifyToxicComment(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input ClassifyToxicPayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	reqBody := classifyModelRequest{Text: input.Text}
	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := "http://localhost:8000/classify"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBodyJSON))
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("classifier api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var modelResp classifyModelResponse
	if err := json.Unmarshal(bodyBytes, &modelResp); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"label":      modelResp.Label,
		"confidence": modelResp.Confidence,
	}
	return json.Marshal(result)
}