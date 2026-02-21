package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"minibrain/internal/llm"
)

func GetMemoryStats(brainDir, neoPath, prefrontalPath string) (MemoryStats, error) {
	if brainDir == "" {
		return MemoryStats{}, errors.New("brain dir is required")
	}
	if neoPath == "" {
		neoPath = filepath.Join(brainDir, "cortex", "NEO.md")
	}
	if prefrontalPath == "" {
		prefrontalPath = filepath.Join(brainDir, "cortex", "PREFRONTAL.md")
	}

	neo, _ := readFileOrEmpty(neoPath)
	pre, _ := readFileOrEmpty(prefrontalPath)

	return MemoryStats{
		LtmLines: countNonEmptyLines(neo),
		StmLines: countNonEmptyLines(pre),
		LtmBytes: len(neo),
		StmBytes: len(pre),
	}, nil
}

func ClearShortTerm(cfg Config) error {
	prefrontalPath := cfg.PrefrontalPath
	if prefrontalPath == "" {
		prefrontalPath = filepath.Join(cfg.BrainDir, "cortex", "PREFRONTAL.md")
	}
	if err := ensureDir(filepath.Dir(prefrontalPath)); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Session Memory (PREFRONTAL)\n\n")
	b.WriteString("- Cleared: " + time.Now().Format(time.RFC3339) + "\n")
	return os.WriteFile(prefrontalPath, []byte(b.String()), 0644)
}

func CondenseShortTerm(cfg Config) (string, error) {
	prefrontalPath := cfg.PrefrontalPath
	if prefrontalPath == "" {
		prefrontalPath = filepath.Join(cfg.BrainDir, "cortex", "PREFRONTAL.md")
	}
	content, err := readFileOrEmpty(prefrontalPath)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	model := cfg.Model
	if model == "" {
		model = "gpt-4.1"
	}

	dev := "You condense short-term memory into a compact, future-use summary. Keep it concise, preserve decisions, TODOs, constraints, and file paths. Output plain text only."
	ctx, cancel := contextWithTimeout(cfg.TimeoutSec)
	defer cancel()

	summary, err := llm.CallOpenAI(ctx, model, dev, content)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("# Session Memory (PREFRONTAL)\n\n")
	b.WriteString("- Condensed: " + time.Now().Format(time.RFC3339) + "\n\n")
	b.WriteString(summary)
	if !strings.HasSuffix(summary, "\n") {
		b.WriteString("\n")
	}

	if err := os.WriteFile(prefrontalPath, []byte(b.String()), 0644); err != nil {
		return "", err
	}

	return summary, nil
}

func countNonEmptyLines(s string) int {
	lines := strings.Split(s, "\n")
	count := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			count++
		}
	}
	return count
}

func contextWithTimeout(timeoutSec int) (ctx context.Context, cancel func()) {
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	return context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
}

func AutoCondenseIfNeeded(cfg Config) (bool, error) {
	prefrontalPath := cfg.PrefrontalPath
	if prefrontalPath == "" {
		prefrontalPath = filepath.Join(cfg.BrainDir, "cortex", "PREFRONTAL.md")
	}
	b, err := os.ReadFile(prefrontalPath)
	if err != nil {
		return false, err
	}
	limit := cfg.StmMaxBytes
	if limit <= 0 {
		limit = 12000
	}
	if len(b) <= limit {
		return false, nil
	}
	_, err = CondenseShortTerm(cfg)
	if err != nil {
		return false, err
	}
	return true, nil
}
