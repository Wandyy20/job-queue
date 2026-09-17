package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

type PDFReportPayload struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

func HandleGeneratePDFReport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input PDFReportPayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, input.Title)
	pdf.Ln(20)

	pdf.SetFont("Arial", "", 12)
	for _, item := range input.Items {
		pdf.Cell(40, 10, item)
		pdf.Ln(10)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	resultBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	result := map[string]interface{}{
		"pdf_base64": resultBase64,
		"title":      input.Title,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return resultJSON, nil
}
