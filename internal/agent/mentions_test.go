package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFileMentions(t *testing.T) {
	in := "please read @foo.txt and @bar/baz.md and @foo.txt"
	out := ExtractFileMentions(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 mentions, got %d", len(out))
	}
	if out[0] != "foo.txt" || out[1] != "bar/baz.md" {
		t.Fatalf("unexpected mentions: %#v", out)
	}
}

func TestSafeRelPath(t *testing.T) {
	if _, err := safeRelPath("/abs/path"); err == nil {
		t.Fatal("expected error for absolute path")
	}
	if _, err := safeRelPath("../x"); err == nil {
		t.Fatal("expected error for traversal path")
	}
	p, err := safeRelPath("ok/path.txt")
	if err != nil || p != filepath.Clean("ok/path.txt") {
		t.Fatalf("unexpected result: %v, %v", p, err)
	}
}

func TestMergeFileRefs(t *testing.T) {
	a := []FileRef{{Path: "a.txt"}, {Path: "b.txt"}}
	b := []FileRef{{Path: "b.txt"}, {Path: "c.txt"}}
	out := MergeFileRefs(a, b)
	if len(out) != 3 {
		t.Fatalf("expected 3 refs, got %d", len(out))
	}
}

func TestLoadMentionedFilesPermission(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "a.txt")
	if err := os.WriteFile(p, []byte("hi"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	refs := LoadMentionedFiles(root, []string{"a.txt"}, false, 0, 0)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if refs[0].Err == nil {
		t.Fatal("expected permission denied error")
	}
}
