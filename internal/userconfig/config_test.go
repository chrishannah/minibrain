package userconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSaveConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MINIBRAIN_HOME", dir)

	cfg := Config{OpenAIAPIKey: "abc", Model: "gpt-4.1"}
	if err := Save(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.OpenAIAPIKey != cfg.OpenAIAPIKey {
		t.Fatalf("api key mismatch: got %q want %q", loaded.OpenAIAPIKey, cfg.OpenAIAPIKey)
	}
	if loaded.Model != cfg.Model {
		t.Fatalf("model mismatch: got %q want %q", loaded.Model, cfg.Model)
	}

	path, err := Path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if path != filepath.Join(dir, "config.json") {
		t.Fatalf("path mismatch: got %q", path)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("json: %v", err)
	}
	if raw["openai_api_key"] != cfg.OpenAIAPIKey {
		t.Fatalf("json api key mismatch")
	}
	if raw["model"] != cfg.Model {
		t.Fatalf("json model mismatch")
	}
}
