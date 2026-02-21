package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/chrishannah/minibrain/internal/agent"
	"github.com/chrishannah/minibrain/internal/userconfig"
)

type configOptions struct {
	allowRead  bool
	allowWrite bool
	readPaths  []string
}

func buildConfig(root, brainDir string, opts configOptions) agent.Config {
	model := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
	if model == "" {
		if cfg, err := userconfig.Load(); err == nil {
			model = strings.TrimSpace(cfg.Model)
		}
	}
	return agent.Config{
		RootDir:             root,
		BrainDir:            brainDir,
		Model:               model,
		TimeoutSec:          60,
		NeoPath:             "",
		PrefrontalPath:      "",
		StmMaxBytes:         12000,
		StmContextBytes:     4000,
		ConversationBytes:   4000,
		ContextBudgetTokens: 16000,
		ApplyWrites:         opts.allowWrite,
		ReadPaths:           opts.readPaths,
		MaxFilesListed:      2000,
		MaxFileBytes:        512 * 1024,
		MaxTotalReadBytes:   2 * 1024 * 1024,
		AllowReadAll:        opts.allowRead,
	}
}

func baseConfig() (agent.Config, error) {
	root, err := os.Getwd()
	if err != nil {
		return agent.Config{}, fmt.Errorf("failed to get working directory: %w", err)
	}
	brainDir, err := agent.ResolveBrainDir()
	if err != nil {
		return agent.Config{}, fmt.Errorf("failed to resolve brain dir: %w", err)
	}
	if err := agent.EnsureBrainLayout(brainDir, root); err != nil {
		return agent.Config{}, fmt.Errorf("failed to initialize brain dir: %w", err)
	}

	perms := agent.ResolvePermissionState(root, readAllowedFromEnv(), writeAllowedFromEnv())
	cfg := buildConfig(root, brainDir, configOptions{
		allowRead:  perms.AllowRead,
		allowWrite: perms.AllowWrite,
	})
	return cfg, nil
}
