package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFilesMax(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(root, "b.txt"), []byte("b"), 0644)
	files, truncated := ListFiles(root, 1)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !truncated {
		t.Fatal("expected truncated to be true")
	}
}

func TestListRelevantFiles(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "alpha.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(root, "beta.txt"), []byte("b"), 0644)
	files, _ := ListRelevantFiles(root, "alpha", 5)
	if len(files) == 0 {
		t.Fatal("expected relevant files")
	}
	if files[0] != "alpha.txt" {
		t.Fatalf("expected alpha.txt first, got %q", files[0])
	}
}
