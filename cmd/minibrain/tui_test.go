package main

import "testing"

func TestNormalizePermissionResponse(t *testing.T) {
	cases := map[string]string{
		"yes":    "/yes",
		"/yes":   "/yes",
		"always": "/always",
		"no":     "/no",
		"maybe":  "maybe",
		" /no ":  "/no",
	}
	for in, want := range cases {
		if got := normalizePermissionResponse(in); got != want {
			t.Fatalf("%q -> %q, want %q", in, got, want)
		}
	}
}

func TestMentionsReadInProse(t *testing.T) {
	if !mentionsReadInProse("Could I read the file?") {
		t.Fatal("expected prose read detection")
	}
	if mentionsReadInProse("READ cmd/minibrain/tui.go") {
		t.Fatal("did not expect prose read detection")
	}
}
