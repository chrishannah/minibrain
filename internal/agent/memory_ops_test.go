package agent

import "testing"

func TestCountNonEmptyLines(t *testing.T) {
	in := "a\n\n b \n\n"
	if got := countNonEmptyLines(in); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}
