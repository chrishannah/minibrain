package agent

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/chrishannah/minibrain/internal/llm"
)

func RunStream(prompt string, cfg Config, onDelta func(string)) (Result, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return Result{}, errors.New("prompt is required")
	}

	root := cfg.RootDir
	if root == "" {
		return Result{}, errors.New("root dir is required")
	}

	brainDir := cfg.BrainDir
	if brainDir == "" {
		var err error
		brainDir, err = ResolveBrainDir()
		if err != nil {
			return Result{}, fmt.Errorf("failed to resolve brain dir: %w", err)
		}
	}
	if err := EnsureBrainLayout(brainDir, root); err != nil {
		return Result{}, fmt.Errorf("failed to initialize brain dir: %w", err)
	}

	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 60
	}

	neoPath := cfg.NeoPath
	if neoPath == "" {
		neoPath = filepath.Join(brainDir, "cortex", "NEO.md")
	}

	prefrontalPath := cfg.PrefrontalPath
	if prefrontalPath == "" {
		prefrontalPath = filepath.Join(brainDir, "cortex", "PREFRONTAL.md")
	}

	neo, err := readFileOrEmpty(neoPath)
	if err != nil {
		return Result{}, fmt.Errorf("failed to read NEO.md: %w", err)
	}

	agentConfig, _ := readFileOrEmpty(filepath.Join(brainDir, "MINIBRAIN.md"))
	soul, _ := readFileOrEmpty(filepath.Join(brainDir, "SOUL.md"))

	mentions := ExtractFileMentions(prompt)
	fileRefs := LoadMentionedFiles(root, mentions, cfg.AllowReadAll, cfg.MaxFileBytes, cfg.MaxTotalReadBytes)
	if len(cfg.ReadPaths) > 0 {
		extra := LoadMentionedFiles(root, cfg.ReadPaths, true, cfg.MaxFileBytes, cfg.MaxTotalReadBytes)
		fileRefs = MergeFileRefs(fileRefs, extra)
	}
	maxFiles := cfg.MaxFilesListed
	if maxFiles <= 0 {
		maxFiles = 2000
	}
	fileList, truncated := ListRelevantFiles(root, prompt, maxFiles)

	if err := WritePrefrontalHeader(prefrontalPath, prompt, mentions, fileRefs); err != nil {
		return Result{}, fmt.Errorf("failed to write PREFRONTAL.md: %w", err)
	}

	stmContext := buildShortTermContext(prefrontalPath, cfg.StmContextBytes)
	convContext := loadConversationContext(brainDir, cfg.ConversationBytes)
	devMsg := BuildDeveloperMessage(agentConfig, soul, neo, stmContext, convContext, prompt, fileRefs, fileList, truncated)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSec)*time.Second)
	defer cancel()

	var out strings.Builder
	llmOut, err := llm.CallOpenAIStream(ctx, cfg.Model, devMsg, prompt, func(delta string) {
		if delta == "" {
			return
		}
		out.WriteString(delta)
		if onDelta != nil {
			onDelta(delta)
		}
	})
	if err != nil {
		AppendPrefrontal(prefrontalPath, "\n## LLM Error\n"+err.Error()+"\n")
		return Result{PrefrontalPath: prefrontalPath}, err
	}
	if llmOut == "" {
		llmOut = out.String()
	}

	proposedWrites := ParseWriteBlocks(llmOut)
	proposedDeletes := ParseDeleteLines(llmOut)
	proposedPatches := ParsePatchBlocks(llmOut)
	var appliedWrites []WriteOp
	var appliedDeletes []DeleteOp
	var appliedPatches []PatchOp
	applied := false
	if cfg.ApplyWrites {
		appliedWrites = ApplyWrites(root, proposedWrites)
		appliedDeletes = ApplyDeletes(root, proposedDeletes)
		appliedPatches, _ = ApplyPatches(root, proposedPatches)
		applied = true
	}

	AppendPrefrontal(prefrontalPath, "\n## LLM Output\n"+llmOut+"\n")
	if applied {
		AppendPrefrontal(prefrontalPath, FormatWritesSummary(appliedWrites))
		AppendPrefrontal(prefrontalPath, FormatDeletesSummary(appliedDeletes))
		AppendPrefrontal(prefrontalPath, FormatPatchesSummary(appliedPatches))
	} else {
		AppendPrefrontal(prefrontalPath, FormatWritesSummaryWithTitle("Proposed Writes", proposedWrites))
		AppendPrefrontal(prefrontalPath, FormatDeletesSummaryWithTitle("Proposed Deletes", proposedDeletes))
		AppendPrefrontal(prefrontalPath, FormatPatchesSummaryWithTitle("Proposed Patches", proposedPatches))
	}

	condensed, err := AutoCondenseIfNeeded(cfg)
	if err != nil {
		AppendPrefrontal(prefrontalPath, "\n## Condense Error\n"+err.Error()+"\n")
	}

	appendConversationContext(brainDir, prompt, llmOut, cfg.ConversationBytes)

	stats, _ := GetMemoryStats(brainDir, neoPath, prefrontalPath)

	return Result{
		LLMOutput:         llmOut,
		ProposedWrites:    proposedWrites,
		ProposedDeletes:   proposedDeletes,
		ProposedPatches:   proposedPatches,
		AppliedWrites:     appliedWrites,
		AppliedDeletes:    appliedDeletes,
		AppliedPatches:    appliedPatches,
		Applied:           applied,
		PrefrontalPath:    prefrontalPath,
		Mentions:          mentions,
		FileRefs:          fileRefs,
		FileList:          fileList,
		FileListTruncated: truncated,
		Memory:            stats,
		Condensed:         condensed,
	}, nil
}
