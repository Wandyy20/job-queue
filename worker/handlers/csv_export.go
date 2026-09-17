package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
)

type CSVExportPayload struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

func HandleCSVExport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input CSVExportPayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write(input.Headers); err != nil {
		return nil, fmt.Errorf("failed to write headers: %w", err)
	}

	for _, row := range input.Rows {
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("csv writer error: %w", err)
	}

	resultBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	result := map[string]interface{}{
		"csv_base64": resultBase64,
		"row_count":  len(input.Rows),
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return resultJSON, nil
}
