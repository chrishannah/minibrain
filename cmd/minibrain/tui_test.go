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
	t.Skip("legacy prose read detection removed under strict JSON responses")
}
