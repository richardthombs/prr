package main

import (
	"testing"
)

func TestResolvedVersionWithReleaseBuild(t *testing.T) {
	previous := version
	t.Cleanup(func() { version = previous })

	version = "v0.2.0"
	got := resolvedVersion()
	if got != "v0.2.0" {
		t.Fatalf("expected %q, got %q", "v0.2.0", got)
	}
}

func TestResolvedVersionWithDevBuild(t *testing.T) {
	previousVersion, previousCommit := version, commit
	t.Cleanup(func() {
		version = previousVersion
		commit = previousCommit
	})

	version = "v0.0.0-dev"
	commit = "abcdef1234567"
	got := resolvedVersion()
	if got != "v0.0.0-dev+abcdef1" {
		t.Fatalf("expected %q, got %q", "v0.0.0-dev+abcdef1", got)
	}
}

func TestResolvedVersionWithReleaseCandidate(t *testing.T) {
	previous := version
	t.Cleanup(func() { version = previous })

	version = "v1.0.0-rc.1"
	got := resolvedVersion()
	if got != "v1.0.0-rc.1" {
		t.Fatalf("expected %q, got %q", "v1.0.0-rc.1", got)
	}
}

func TestShortCommit(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "unknown"},
		{"literal unknown", "unknown", "unknown"},
		{"short commit", "abcdef1", "abcdef1"},
		{"long commit", "abcdef1234567", "abcdef1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shortCommit(tc.input)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
