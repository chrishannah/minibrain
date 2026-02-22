package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/chrishannah/minibrain/internal/userconfig"
)

type responsesRequest struct {
	Model        string        `json:"model"`
	Instructions string        `json:"instructions"`
	Input        string        `json:"input"`
	Stream       bool          `json:"stream,omitempty"`
	Text         *responseText `json:"text,omitempty"`
}

type responseText struct {
	Format *responseFormat `json:"format,omitempty"`
}

type responseFormat struct {
	Type   string          `json:"type"`
	Name   string          `json:"name,omitempty"`
	Strict bool            `json:"strict,omitempty"`
	Schema json.RawMessage `json:"schema,omitempty"`
}

type responsesResponse struct {
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func CallOpenAI(ctx context.Context, model, developerMsg, userMsg string) (string, error) {
	apiKey, err := loadAPIKey()
	if err != nil {
		return "", err
	}
	if model == "" {
		model = "gpt-4.1"
	}

	schema := []byte(`{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "read": { "type": "array", "items": { "type": "string" } },
    "patches": { "type": "array", "items": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "path": { "type": "string" },
        "diff": { "type": "string" }
      },
      "required": ["path", "diff"]
    }},
    "writes": { "type": "array", "items": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "path": { "type": "string" },
        "content": { "type": "string" }
      },
      "required": ["path", "content"]
    }},
    "deletes": { "type": "array", "items": { "type": "string" } },
    "message": { "type": "string" }
  },
  "required": ["read", "patches", "writes", "deletes", "message"]
}`)

	payload := responsesRequest{
		Model:        model,
		Instructions: developerMsg,
		Input:        userMsg,
		Text: &responseText{
			Format: &responseFormat{
				Type:   "json_schema",
				Name:   "minibrain_response",
				Strict: true,
				Schema: schema,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai error: %s", strings.TrimSpace(string(b)))
	}

	var out responsesResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", formatOpenAIError(out.Error.Code, out.Error.Type, out.Error.Message)
	}

	for _, item := range out.Output {
		for _, c := range item.Content {
			if c.Type == "output_text" && strings.TrimSpace(c.Text) != "" {
				return c.Text, nil
			}
		}
	}

	return "", errors.New("no output_text found in response")
}

func CallOpenAIStream(ctx context.Context, model, developerMsg, userMsg string, onDelta func(string)) (string, error) {
	apiKey, err := loadAPIKey()
	if err != nil {
		return "", err
	}
	if model == "" {
		model = "gpt-4.1"
	}

	schema := []byte(`{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "read": { "type": "array", "items": { "type": "string" } },
    "patches": { "type": "array", "items": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "path": { "type": "string" },
        "diff": { "type": "string" }
      },
      "required": ["path", "diff"]
    }},
    "writes": { "type": "array", "items": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "path": { "type": "string" },
        "content": { "type": "string" }
      },
      "required": ["path", "content"]
    }},
    "deletes": { "type": "array", "items": { "type": "string" } },
    "message": { "type": "string" }
  },
  "required": ["read", "patches", "writes", "deletes", "message"]
}`)

	payload := responsesRequest{
		Model:        model,
		Instructions: developerMsg,
		Input:        userMsg,
		Stream:       true,
		Text: &responseText{
			Format: &responseFormat{
				Type:   "json_schema",
				Name:   "minibrain_response",
				Strict: true,
				Schema: schema,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error: %s", strings.TrimSpace(string(b)))
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	var out strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			break
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			continue
		}
		if errMsg := streamError(payload); errMsg != "" {
			return "", errors.New(errMsg)
		}
		if delta := extractStreamDelta(payload); delta != "" {
			out.WriteString(delta)
			if onDelta != nil {
				onDelta(delta)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return out.String(), nil
}

func streamError(payload map[string]any) string {
	if e, ok := payload["error"].(map[string]any); ok {
		if msg, ok := e["message"].(string); ok && msg != "" {
			return msg
		}
		if typ, ok := e["type"].(string); ok && typ != "" {
			return "OpenAI API error: " + typ
		}
	}
	return ""
}

func extractStreamDelta(payload map[string]any) string {
	if v, ok := payload["delta"].(string); ok {
		return v
	}
	if deltaObj, ok := payload["delta"].(map[string]any); ok {
		if text, ok := deltaObj["text"].(string); ok {
			return text
		}
	}
	return ""
}

func loadAPIKey() (string, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey != "" {
		return apiKey, nil
	}
	cfg, err := userconfig.Load()
	if err != nil {
		return "", errors.New("OPENAI_API_KEY is required")
	}
	apiKey = strings.TrimSpace(cfg.OpenAIAPIKey)
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY is required")
	}
	return apiKey, nil
}

func formatOpenAIError(code, typ, msg string) error {
	if code == "insufficient_quota" {
		return errors.New("OpenAI API quota/billing issue: your API key's project has no remaining quota or billing is not enabled")
	}
	if msg != "" {
		return errors.New(msg)
	}
	if code != "" {
		return errors.New("OpenAI API error: " + code)
	}
	if typ != "" {
		return errors.New("OpenAI API error: " + typ)
	}
	return errors.New("OpenAI API error")
}
