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
