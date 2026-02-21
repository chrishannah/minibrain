package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProjectConfig struct {
	AllowReadAlways  bool `json:"allow_read_always"`
	AllowWriteAlways bool `json:"allow_write_always"`
	DenyWriteAlways  bool `json:"deny_write_always"`
}

func LoadProjectConfig(root string) ProjectConfig {
	path := ProjectConfigPath(root)
	b, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}
	}
	var cfg ProjectConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return ProjectConfig{}
	}
	return cfg
}

func SaveProjectConfig(root string, cfg ProjectConfig) error {
	path := ProjectConfigPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func ProjectConfigPath(root string) string {
	return filepath.Join(root, ".minibrain", "config.json")
}
