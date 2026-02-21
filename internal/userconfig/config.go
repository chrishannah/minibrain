package userconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	OpenAIAPIKey string `json:"openai_api_key"`
	Model        string `json:"model"`
}

func Load() (Config, error) {
	path, err := pathForUser()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(cfg Config) error {
	path, err := pathForUser()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func Path() (string, error) {
	return pathForUser()
}

func pathForUser() (string, error) {
	if v := strings.TrimSpace(os.Getenv("MINIBRAIN_HOME")); v != "" {
		return filepath.Join(v, "config.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".minibrain", "config.json"), nil
}
