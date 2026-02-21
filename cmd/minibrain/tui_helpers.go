package main

import (
	"os"
	"strings"

	"minibrain/internal/userconfig"
)

func currentModel() string {
	v := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
	if v == "" {
		if cfg, err := userconfig.Load(); err == nil {
			v = strings.TrimSpace(cfg.Model)
		}
	}
	if v == "" {
		return "gpt-4.1"
	}
	return v
}
