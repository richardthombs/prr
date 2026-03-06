package provider

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

type stubProvider struct {
	resolveFunc func(ctx context.Context, prID int, opts map[string]string) (types.PRRef, error)
}

func (s stubProvider) Resolve(ctx context.Context, prID int, opts map[string]string) (types.PRRef, error) {
	return s.resolveFunc(ctx, prID, opts)
}

func TestResolverDelegatesToProvider(t *testing.T) {
	provider := stubProvider{
		resolveFunc: func(_ context.Context, prID int, opts map[string]string) (types.PRRef, error) {
			if prID != 10 {
				t.Fatalf("unexpected prID: %d", prID)
			}
			if opts["repoUrl"] != "https://example.test/org/repo" {
				t.Fatalf("unexpected repoUrl option: %q", opts["repoUrl"])
			}
			return types.PRRef{PRID: 10, RepoURL: opts["repoUrl"], Remote: opts["remote"], Provider: opts["provider"]}, nil
		},
	}

	resolver := NewResolver(provider)

	result, err := resolver.Resolve(context.Background(), 10, ResolveOptions{RepoURL: "https://example.test/org/repo"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.Remote != "origin" {
		t.Fatalf("expected default remote origin, got %q", result.Remote)
	}
	if result.Provider != "azure-devops" {
		t.Fatalf("expected default provider azure-devops, got %q", result.Provider)
	}
}

func TestResolverClassifiesProviderFailure(t *testing.T) {
	provider := stubProvider{
		resolveFunc: func(_ context.Context, _ int, _ map[string]string) (types.PRRef, error) {
			return types.PRRef{}, errors.New("provider boom")
		},
	}

	resolver := NewResolver(provider)

	_, err := resolver.Resolve(context.Background(), 10, ResolveOptions{RepoURL: "https://example.test/org/repo"})
	if err == nil {
		t.Fatalf("expected provider error")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassProvider {
		t.Fatalf("expected provider class, got %s", appErr.Class)
	}
}

func TestResolveFromPullRequestURLAzureDevOps(t *testing.T) {
	provider := stubProvider{
		resolveFunc: func(_ context.Context, prID int, opts map[string]string) (types.PRRef, error) {
			return types.PRRef{PRID: prID, RepoURL: opts["repoUrl"], Remote: opts["remote"], Provider: opts["provider"]}, nil
		},
	}
	resolver := NewResolver(provider)

	result, err := resolver.ResolveFromPullRequestURL(
		context.Background(),
		"https://dev.azure.com/ensekltd/blackbird/_git/blackbird/pullrequest/83438",
		ResolveOptions{},
	)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.PRID != 83438 {
		t.Fatalf("expected PR ID 83438, got %d", result.PRID)
	}
	if result.Provider != "azure-devops" {
		t.Fatalf("expected provider azure-devops, got %q", result.Provider)
	}
	if result.RepoURL != "https://dev.azure.com/ensekltd/blackbird/_git/blackbird" {
		t.Fatalf("expected repo URL from parsed URL, got %q", result.RepoURL)
	}
}

func TestResolveFromPullRequestURLGitHub(t *testing.T) {
	provider := stubProvider{
		resolveFunc: func(_ context.Context, prID int, opts map[string]string) (types.PRRef, error) {
			return types.PRRef{PRID: prID, RepoURL: opts["repoUrl"], Remote: opts["remote"], Provider: opts["provider"]}, nil
		},
	}
	resolver := NewResolver(provider)

	result, err := resolver.ResolveFromPullRequestURL(
		context.Background(),
		"https://github.com/steveyegge/beads/pull/2331",
		ResolveOptions{},
	)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.PRID != 2331 {
		t.Fatalf("expected PR ID 2331, got %d", result.PRID)
	}
	if result.Provider != "github" {
		t.Fatalf("expected provider github, got %q", result.Provider)
	}
}

func TestResolveFromPullRequestURLHonoursRepoOverride(t *testing.T) {
	provider := stubProvider{
		resolveFunc: func(_ context.Context, prID int, opts map[string]string) (types.PRRef, error) {
			return types.PRRef{PRID: prID, RepoURL: opts["repoUrl"], Remote: opts["remote"], Provider: opts["provider"]}, nil
		},
	}
	resolver := NewResolver(provider)

	result, err := resolver.ResolveFromPullRequestURL(
		context.Background(),
		"https://github.com/steveyegge/beads/pull/2331",
		ResolveOptions{RepoURL: "https://github.com/override/repo"},
	)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.RepoURL != "https://github.com/override/repo" {
		t.Fatalf("expected overridden repo URL, got %q", result.RepoURL)
	}
}

func TestResolveFromPullRequestURLHonoursProviderOverride(t *testing.T) {
	provider := stubProvider{
		resolveFunc: func(_ context.Context, prID int, opts map[string]string) (types.PRRef, error) {
			return types.PRRef{PRID: prID, RepoURL: opts["repoUrl"], Remote: opts["remote"], Provider: opts["provider"]}, nil
		},
	}
	resolver := NewResolver(provider)

	result, err := resolver.ResolveFromPullRequestURL(
		context.Background(),
		"https://github.com/steveyegge/beads/pull/2331",
		ResolveOptions{Provider: "custom"},
	)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.Provider != "custom" {
		t.Fatalf("expected overridden provider custom, got %q", result.Provider)
	}
}

func TestResolveFromPullRequestURLRejectsInvalidURL(t *testing.T) {
	provider := stubProvider{
		resolveFunc: func(_ context.Context, _ int, _ map[string]string) (types.PRRef, error) {
			t.Fatalf("unexpected provider call for invalid URL")
			return types.PRRef{}, nil
		},
	}
	resolver := NewResolver(provider)

	_, err := resolver.ResolveFromPullRequestURL(context.Background(), "not-a-url", ResolveOptions{})
	if err == nil {
		t.Fatalf("expected error for invalid URL")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassConfig {
		t.Fatalf("expected config class error, got %s", appErr.Class)
	}
}