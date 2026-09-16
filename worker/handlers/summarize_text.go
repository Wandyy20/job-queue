package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type SummarizePayload struct {
	Text string `json:"text"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Contents struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func HandleSummarizeText(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input SummarizePayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: "Summarize this text in 2-3 sentences: " + input.Text},
				},
			},
		},
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.5-flash:generateContent?key=" + apiKey

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
		return nil, fmt.Errorf("gemini api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Contents.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no summary")
	}

	summary := geminiResp.Candidates[0].Contents.Parts[0].Text

	result := map[string]interface{}{
		"summary": summary,
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return resultJSON, err
}
