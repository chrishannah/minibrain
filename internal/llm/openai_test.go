package llm

import (
	"testing"

	"minibrain/internal/userconfig"
)

func TestLoadAPIKeyFromEnv(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "env-key")
	key, err := loadAPIKey()
	if err != nil {
		t.Fatalf("expected key, got error: %v", err)
	}
	if key != "env-key" {
		t.Fatalf("expected env key, got %q", key)
	}
}

func TestLoadAPIKeyFromConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MINIBRAIN_HOME", dir)
	t.Setenv("OPENAI_API_KEY", "")
	if err := userconfig.Save(userconfig.Config{OpenAIAPIKey: "file-key"}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	key, err := loadAPIKey()
	if err != nil {
		t.Fatalf("expected key, got error: %v", err)
	}
	if key != "file-key" {
		t.Fatalf("expected file key, got %q", key)
	}
}

func TestLoadAPIKeyMissing(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	temp := t.TempDir()
	t.Setenv("MINIBRAIN_HOME", temp)
	t.Setenv("HOME", temp)
	_, err := loadAPIKey()
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}
