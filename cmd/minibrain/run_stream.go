package main

import (
	"fmt"
	"os"

	"github.com/chrishannah/minibrain/internal/agent"
)

func runAgentStream(prompt string, onDelta func(string)) (agent.Result, error) {
	root, err := os.Getwd()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to get working directory: %w", err)
	}
	brainDir, err := agent.ResolveBrainDir()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to resolve brain dir: %w", err)
	}
	perms := agent.ResolvePermissionState(root, readAllowedFromEnv(), writeAllowedFromEnv())
	cfg := buildConfig(root, brainDir, configOptions{
		allowRead:  perms.AllowRead,
		allowWrite: perms.AllowWrite,
	})
	return agent.RunStream(prompt, cfg, onDelta)
}

func runAgentStreamWithAllow(prompt string, allowRead, allowWrite bool, onDelta func(string)) (agent.Result, error) {
	root, err := os.Getwd()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to get working directory: %w", err)
	}
	brainDir, err := agent.ResolveBrainDir()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to resolve brain dir: %w", err)
	}
	cfg := buildConfig(root, brainDir, configOptions{
		allowRead:  allowRead,
		allowWrite: allowWrite,
	})
	return agent.RunStream(prompt, cfg, onDelta)
}

func runAgentStreamWithAllowAndReads(prompt string, allowRead, allowWrite bool, readPaths []string, onDelta func(string)) (agent.Result, error) {
	root, err := os.Getwd()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to get working directory: %w", err)
	}
	brainDir, err := agent.ResolveBrainDir()
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to resolve brain dir: %w", err)
	}
	cfg := buildConfig(root, brainDir, configOptions{
		allowRead:  allowRead,
		allowWrite: allowWrite,
		readPaths:  readPaths,
	})
	return agent.RunStream(prompt, cfg, onDelta)
}
