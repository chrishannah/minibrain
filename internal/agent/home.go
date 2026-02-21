package agent

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

func ResolveBrainDir() (string, error) {
	if v := os.Getenv("MINIBRAIN_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".minibrain"), nil
}

func EnsureBrainLayout(brainDir, repoRoot string) error {
	if brainDir == "" {
		return errors.New("brain dir is required")
	}
	if err := os.MkdirAll(filepath.Join(brainDir, "cortex"), 0755); err != nil {
		return err
	}

	// Migrate or seed MINIBRAIN.md
	minibrainDst := filepath.Join(brainDir, "MINIBRAIN.md")
	if !fileExists(minibrainDst) {
		copyDefault(repoRoot, "MINIBRAIN.md", minibrainDst, defaultMinibrain())
	}

	// Migrate or seed SOUL.md
	soulDst := filepath.Join(brainDir, "SOUL.md")
	if !fileExists(soulDst) {
		copyDefault(repoRoot, "SOUL.md", soulDst, defaultSoul())
	}

	// Migrate or seed NEO.md
	neoDst := filepath.Join(brainDir, "cortex", "NEO.md")
	if !fileExists(neoDst) {
		copyDefault(repoRoot, filepath.Join("cortex", "NEO.md"), neoDst, defaultNeo())
	}

	// Migrate or seed PREFRONTAL.md
	preDst := filepath.Join(brainDir, "cortex", "PREFRONTAL.md")
	if !fileExists(preDst) {
		copyDefault(repoRoot, filepath.Join("cortex", "PREFRONTAL.md"), preDst, defaultPrefrontal())
	}

	// Migrate or seed CONTEXT.md
	contextDst := filepath.Join(brainDir, "cortex", "CONTEXT.md")
	if !fileExists(contextDst) {
		_ = os.WriteFile(contextDst, []byte(defaultContext()), 0644)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyDefault(repoRoot, repoRel, dst, fallback string) {
	if repoRoot != "" {
		if b, err := os.ReadFile(filepath.Join(repoRoot, repoRel)); err == nil {
			_ = os.WriteFile(dst, b, 0644)
			return
		}
	}
	_ = os.WriteFile(dst, []byte(fallback), 0644)
}

func defaultMinibrain() string {
	return "# MINIBRAIN\n\n" +
		"Core wiring for the agent. Keep this file small and focused on behavior and memory wiring.\n\n" +
		"## Memory Files\n" +
		"- Long-term memory: `cortex/NEO.md`\n" +
		"- Short-term memory: `cortex/PREFRONTAL.md`\n" +
		"- Conversation summary: `cortex/CONTEXT.md`\n" +
		"- Personality: `SOUL.md`\n\n" +
		"## Operating Rules\n" +
		"- Ask before reading file contents unless the user has allowed it.\n" +
		"- Request files using `READ <path>` only (no prose).\n" +
		"- Prefer PATCH for edits; use WRITE/EDIT/DELETE for changes.\n" +
		"- When planning to modify files, include the actual changes in the same response.\n\n" +
		"## Memory Process\n" +
		"- LTM persists across sessions and accumulates durable facts, preferences, and constraints.\n" +
		"- STM is session context that persists across runs and is condensed when large or on request.\n" +
		"- Conversation summary is a compact rolling log of recent prompts and responses.\n\n" +
		"## Promotion Guidance\n" +
		"- Promote durable facts, preferences, or constraints to `NEO.md`.\n" +
		"- Keep `PREFRONTAL.md` focused on current session context and decisions.\n"
}

func defaultSoul() string {
	return "# SOUL\n\n" +
		"Minibrain is a pragmatic, concise assistant focused on getting real work done.\n\n" +
		"Purpose:\n" +
		"- Be useful and help the user achieve their intended results.\n" +
		"- Optimize for correctness, clarity, and momentum.\n\n" +
		"Style:\n" +
		"- Prefer concrete steps over vague guidance.\n" +
		"- Ask one question at a time if clarification is needed.\n" +
		"- Be explicit about assumptions and uncertainty.\n" +
		"- Keep responses short unless the user asks for depth.\n\n" +
		"Behavior:\n" +
		"- Respect file-read permissions; request files with `READ <path>` only.\n" +
		"- Prefer small, reversible changes.\n" +
		"- When editing, favor PATCH over full rewrites.\n" +
		"- Summarize applied changes and call out any risks.\n"
}

func defaultNeo() string {
	return "# Long-Term Memory (NEO)\n\n" +
		"- Project: minibrain\n" +
		"- Purpose: minimal agentic loop in Go with a TUI.\n" +
		"- UX goals: calm, readable UI; clear permission prompts; safe writes.\n" +
		"- Tooling: strict READ-line protocol; prefer PATCH for edits.\n"
}

func defaultPrefrontal() string {
	return "# Short-Term Memory (PREFRONTAL)\n\n" +
		"- Initialized: " + time.Now().UTC().Format(time.RFC3339) + "\n" +
		"- Notes: Session memory persists across runs and is condensed when large or on request.\n"
}

func defaultContext() string {
	return "# Conversation Summary (CONTEXT)\n\n" +
		"- Initialized: " + time.Now().UTC().Format(time.RFC3339) + "\n"
}
