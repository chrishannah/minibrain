package agent

import (
	"os"
	"path/filepath"
)

type UsageStats struct {
	LtmBytes         int
	StmBytes         int
	StmContextBytes  int
	ConvBytes        int
	ConvContextBytes int
	ApproxTokens     int
	BudgetTokens     int
}

func GetUsageStats(cfg Config) (UsageStats, error) {
	brainDir := cfg.BrainDir
	if brainDir == "" {
		var err error
		brainDir, err = ResolveBrainDir()
		if err != nil {
			return UsageStats{}, err
		}
	}
	neoPath := cfg.NeoPath
	if neoPath == "" {
		neoPath = filepath.Join(brainDir, "cortex", "NEO.md")
	}
	prefrontalPath := cfg.PrefrontalPath
	if prefrontalPath == "" {
		prefrontalPath = filepath.Join(brainDir, "cortex", "PREFRONTAL.md")
	}
	contextPath := filepath.Join(brainDir, "cortex", "CONTEXT.md")

	neo, _ := readFileOrEmpty(neoPath)
	pre, _ := readFileOrEmpty(prefrontalPath)
	ctx, _ := readFileOrEmpty(contextPath)

	stmContext := cfg.StmContextBytes
	if stmContext <= 0 {
		stmContext = 4000
	}
	convContext := cfg.ConversationBytes
	if convContext <= 0 {
		convContext = 4000
	}

	stmBytes := len(pre)
	convBytes := len(ctx)

	if stmBytes > stmContext {
		stmBytes = stmContext
	}
	if convBytes > convContext {
		convBytes = convContext
	}

	budget := cfg.ContextBudgetTokens
	if budget <= 0 {
		budget = 16000
	}

	totalBytes := len(neo) + stmBytes + convBytes
	approxTokens := totalBytes / 4

	return UsageStats{
		LtmBytes:         len(neo),
		StmBytes:         len(pre),
		StmContextBytes:  stmBytes,
		ConvBytes:        len(ctx),
		ConvContextBytes: convBytes,
		ApproxTokens:     approxTokens,
		BudgetTokens:     budget,
	}, nil
}

func ContextFileSize(brainDir string) int {
	path := filepath.Join(brainDir, "cortex", "CONTEXT.md")
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return int(info.Size())
}
