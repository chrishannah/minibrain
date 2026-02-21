package agent

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func buildShortTermContext(prefrontalPath string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	content, err := readFileOrEmpty(prefrontalPath)
	if err != nil {
		return ""
	}
	trim := strings.TrimSpace(content)
	if trim == "" {
		return ""
	}
	if len(content) <= maxBytes {
		return content
	}
	start := len(content) - maxBytes
	if start < 0 {
		start = 0
	}
	chunk := content[start:]
	if idx := strings.Index(chunk, "\n"); idx > 0 && idx < len(chunk)-1 {
		chunk = chunk[idx+1:]
	}
	return strings.TrimSpace(chunk)
}

func loadConversationContext(brainDir string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	path := filepath.Join(brainDir, "cortex", "CONTEXT.md")
	content, err := readFileOrEmpty(path)
	if err != nil {
		return ""
	}
	trim := strings.TrimSpace(content)
	if trim == "" {
		return ""
	}
	if len(content) <= maxBytes {
		return content
	}
	start := len(content) - maxBytes
	if start < 0 {
		start = 0
	}
	chunk := content[start:]
	if idx := strings.Index(chunk, "\n"); idx > 0 && idx < len(chunk)-1 {
		chunk = chunk[idx+1:]
	}
	return strings.TrimSpace(chunk)
}

func appendConversationContext(brainDir, prompt, response string, maxBytes int) {
	if maxBytes <= 0 {
		return
	}
	path := filepath.Join(brainDir, "cortex", "CONTEXT.md")
	_ = ensureDir(filepath.Dir(path))

	summary := strings.TrimSpace(response)
	if len(summary) > 800 {
		summary = summary[:800] + "…"
	}
	entry := "## " + time.Now().UTC().Format(time.RFC3339) + "\n" +
		"Prompt: " + strings.TrimSpace(prompt) + "\n\n" +
		"Response: " + summary + "\n\n"

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(entry)
	_ = f.Close()

	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if len(b) <= maxBytes {
		return
	}
	trimmed := string(b[len(b)-maxBytes:])
	if idx := strings.Index(trimmed, "\n## "); idx > 0 && idx < len(trimmed)-1 {
		trimmed = trimmed[idx+1:]
	}
	_ = os.WriteFile(path, []byte(strings.TrimSpace(trimmed)+"\n"), 0644)
}
