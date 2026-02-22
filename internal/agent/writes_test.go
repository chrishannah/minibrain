package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWriteAndDelete(t *testing.T) {
	in := "WRITE a.txt\n```\nhello\n```\n\nDELETE b.txt\n"
	w := ParseWriteBlocks(in)
	d := ParseDeleteLines(in)
	if len(w) != 1 || w[0].Path != "a.txt" || w[0].Content != "hello" {
		t.Fatalf("unexpected writes: %#v", w)
	}
	if len(d) != 1 || d[0].Path != "b.txt" {
		t.Fatalf("unexpected deletes: %#v", d)
	}
}

func TestParseReadLines(t *testing.T) {
	in := "READ a.txt\nREAD dir/b.md\n"
	r := ParseReadLines(in)
	if len(r) != 2 {
		t.Fatalf("expected 2 read lines, got %d", len(r))
	}
}

func TestApplyWritesDeletes(t *testing.T) {
	root := t.TempDir()
	writes := []WriteOp{{Path: "a.txt", Content: "hi"}}
	applied := ApplyWrites(root, writes)
	if len(applied) != 1 {
		t.Fatalf("expected 1 applied write, got %d", len(applied))
	}
	b, err := os.ReadFile(filepath.Join(root, "a.txt"))
	if err != nil || string(b) != "hi" {
		t.Fatalf("read back failed: %v %s", err, string(b))
	}
	dels := []DeleteOp{{Path: "a.txt"}}
	appliedD := ApplyDeletes(root, dels)
	if len(appliedD) != 1 {
		t.Fatalf("expected 1 applied delete, got %d", len(appliedD))
	}
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err == nil {
		t.Fatal("expected file to be deleted")
	}
}

func TestParseAndApplyPatch(t *testing.T) {
	patch := "PATCH a.txt\n```patch\n@@ -1,2 +1,2 @@\n-hello\n+hello world\n line2\n```\n"
	ops := ParsePatchBlocks(patch)
	if len(ops) != 1 || ops[0].Path != "a.txt" {
		t.Fatalf("unexpected patches: %#v", ops)
	}

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello\nline2\n"), 0644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	applied, _ := ApplyPatches(root, ops)
	if len(applied) != 1 {
		t.Fatalf("expected applied patch, got %d", len(applied))
	}
	b, err := os.ReadFile(filepath.Join(root, "a.txt"))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(b) != "hello world\nline2\n" {
		t.Fatalf("unexpected content: %q", string(b))
	}
}
