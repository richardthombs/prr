package config

import (
	"testing"

	apperrors "github.com/richardthombs/prr/internal/errors"
)

func TestResolveAppliesDefaults(t *testing.T) {
	resolved, err := Resolve(ResolveInput{RepoURL: "https://example.test/org/repo"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if resolved.Provider != "azure-devops" {
		t.Fatalf("expected default provider azure-devops, got %q", resolved.Provider)
	}
	if resolved.Remote != "origin" {
		t.Fatalf("expected default remote origin, got %q", resolved.Remote)
	}
}

func TestResolveFailsWhenRepoMissing(t *testing.T) {
	_, err := Resolve(ResolveInput{})
	if err == nil {
		t.Fatalf("expected error for missing repo")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassConfig {
		t.Fatalf("expected config class error, got %s", appErr.Class)
	}
}

func TestResolveHonoursExplicitProvider(t *testing.T) {
	resolved, err := Resolve(ResolveInput{RepoURL: "https://example.test/org/repo", Provider: "github"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resolved.Provider != "github" {
		t.Fatalf("expected explicit provider github, got %q", resolved.Provider)
	}
}

func TestResolveHonoursExplicitRemote(t *testing.T) {
	resolved, err := Resolve(ResolveInput{RepoURL: "https://example.test/org/repo", Remote: "upstream"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resolved.Remote != "upstream" {
		t.Fatalf("expected explicit remote upstream, got %q", resolved.Remote)
	}
}