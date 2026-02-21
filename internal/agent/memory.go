package agent

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func readFileOrEmpty(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

func WritePrefrontalHeader(path, prompt string, mentions []string, refs []FileRef) error {
	now := time.Now().Format(time.RFC3339)
	var b strings.Builder
	b.WriteString("# Session Memory (PREFRONTAL)\n\n")
	b.WriteString("- Started: " + now + "\n")
	b.WriteString("- Prompt: " + prompt + "\n\n")

	b.WriteString("## Mentioned Files\n")
	if len(mentions) == 0 {
		b.WriteString("(none)\n")
	} else {
		for _, m := range mentions {
			b.WriteString("- " + m + "\n")
		}
	}

	b.WriteString("\n## File Load Results\n")
	if len(refs) == 0 {
		b.WriteString("(none)\n")
	} else {
		for _, r := range refs {
			if r.Err != nil {
				b.WriteString("- " + formatMentionPath(r) + ": " + r.Err.Error() + "\n")
			} else {
				b.WriteString("- " + formatMentionPath(r) + ": loaded\n")
			}
		}
	}

	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}

	existing, _ := readFileOrEmpty(path)
	if strings.TrimSpace(existing) != "" {
		b.WriteString("\n---\n\n")
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(b.String())
	return err
}

func formatMentionPath(r FileRef) string {
	if r.Mention != "" && r.Mention != r.Path {
		return r.Mention + " -> " + r.Path
	}
	return r.Path
}

func AppendPrefrontal(path, content string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(content)
}
