package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/disintegration/imaging"
)

type ResizeImagePayload struct {
	ImageBase64 string `json:"image_base64"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

func HandleResizeImage(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	var input ResizeImagePayload
	if err := json.Unmarshal(payload, &input); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	imgBytes, err := base64.StdEncoding.DecodeString(input.ImageBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 image: %w", err)
	}

	img, err := imaging.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	resized := imaging.Resize(img, input.Width, input.Height, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, resized, imaging.PNG); err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	resultBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	result := map[string]interface{}{
		"resized_image_base64": resultBase64,
		"width":                input.Width,
		"height":               input.Height,
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return resultJSON, nil
}
