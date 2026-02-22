package agent

import "testing"

func TestParseStructuredOutput(t *testing.T) {
	input := `{
		"read": ["README.md"],
		"patches": [{"path": "README.md", "diff": "@@ -1,1 +1,2 @@\n line\n+added"}],
		"writes": [{"path": "new.txt", "content": "hello"}],
		"deletes": ["old.txt"],
		"message": "done"
	}`

	out, ok := ParseStructuredOutput(input)
	if !ok {
		t.Fatalf("expected ok")
	}
	if len(out.Read) != 1 || out.Read[0] != "README.md" {
		t.Fatalf("unexpected read: %#v", out.Read)
	}
	if len(out.Patches) != 1 || out.Patches[0].Path != "README.md" {
		t.Fatalf("unexpected patches: %#v", out.Patches)
	}
	if len(out.Writes) != 1 || out.Writes[0].Path != "new.txt" {
		t.Fatalf("unexpected writes: %#v", out.Writes)
	}
	if len(out.Deletes) != 1 || out.Deletes[0] != "old.txt" {
		t.Fatalf("unexpected deletes: %#v", out.Deletes)
	}
	if out.Message != "done" {
		t.Fatalf("unexpected message: %q", out.Message)
	}
}

func TestParseStructuredOutputInvalid(t *testing.T) {
	if _, ok := ParseStructuredOutput("not json"); ok {
		t.Fatalf("expected failure")
	}
}
