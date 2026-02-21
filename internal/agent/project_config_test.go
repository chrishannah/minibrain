package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectConfigSaveLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := ProjectConfig{
		AllowReadAlways:  true,
		AllowWriteAlways: true,
		DenyWriteAlways:  false,
	}
	if err := SaveProjectConfig(dir, cfg); err != nil {
		t.Fatalf("save project config: %v", err)
	}

	path := ProjectConfigPath(dir)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}

	loaded := LoadProjectConfig(dir)
	if loaded.AllowReadAlways != cfg.AllowReadAlways {
		t.Fatalf("AllowReadAlways mismatch: got %v want %v", loaded.AllowReadAlways, cfg.AllowReadAlways)
	}
	if loaded.AllowWriteAlways != cfg.AllowWriteAlways {
		t.Fatalf("AllowWriteAlways mismatch: got %v want %v", loaded.AllowWriteAlways, cfg.AllowWriteAlways)
	}
	if loaded.DenyWriteAlways != cfg.DenyWriteAlways {
		t.Fatalf("DenyWriteAlways mismatch: got %v want %v", loaded.DenyWriteAlways, cfg.DenyWriteAlways)
	}
}

func TestProjectConfigPath(t *testing.T) {
	root := "/tmp/example"
	want := filepath.Join(root, ".minibrain", "config.json")
	got := ProjectConfigPath(root)
	if got != want {
		t.Fatalf("path mismatch: got %q want %q", got, want)
	}
}
