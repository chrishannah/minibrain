package agent

import (
	"encoding/json"
	"strings"
)

type StructuredPatch struct {
	Path string `json:"path"`
	Diff string `json:"diff"`
}

type StructuredWrite struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type StructuredResponse struct {
	Read    []string          `json:"read"`
	Patches []StructuredPatch `json:"patches"`
	Writes  []StructuredWrite `json:"writes"`
	Deletes []string          `json:"deletes"`
	Message string            `json:"message"`
}

func ParseStructuredOutput(raw string) (StructuredResponse, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return StructuredResponse{}, false
	}
	var out StructuredResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return StructuredResponse{}, false
	}
	return out, true
}
