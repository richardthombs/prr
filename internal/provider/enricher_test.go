package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/types"
)

type stubCLIRunner struct {
	output string
	err    error
}

func (r *stubCLIRunner) Run(_ context.Context, _ string, _ ...string) (string, error) {
	return r.output, r.err
}

func TestEnrichGitHubPopulatesBaseBranchAndSHA(t *testing.T) {
	runner := &stubCLIRunner{output: "main\nabc1234def5678\n"}
	warnings := make([]string, 0)
	warnf := func(format string, args ...any) { warnings = append(warnings, format) }

	ref := enrichGitHub(context.Background(), testGitHubRef(), runner, warnf)

	if ref.BaseBranch != "main" {
		t.Fatalf("expected BaseBranch 'main', got %q", ref.BaseBranch)
	}
	if ref.BaseSHA != "abc1234def5678" {
		t.Fatalf("expected BaseSHA 'abc1234def5678', got %q", ref.BaseSHA)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
}

func TestEnrichGitHubWarnsAndReturnsUnchangedOnCLIError(t *testing.T) {
	runner := &stubCLIRunner{err: errors.New("gh: command not found")}
	warnings := make([]string, 0)
	warnf := func(format string, args ...any) { warnings = append(warnings, format) }

	original := testGitHubRef()
	ref := enrichGitHub(context.Background(), original, runner, warnf)

	if ref.BaseBranch != "" || ref.BaseSHA != "" {
		t.Fatalf("expected no enrichment on CLI error, got BaseBranch=%q BaseSHA=%q", ref.BaseBranch, ref.BaseSHA)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "enrichment unavailable") {
		t.Fatalf("expected enrichment-unavailable warning, got %v", warnings)
	}
}

func TestEnrichGitHubWarnsOnMalformedOutput(t *testing.T) {
	runner := &stubCLIRunner{output: "main\n"}
	warnings := make([]string, 0)
	warnf := func(format string, args ...any) { warnings = append(warnings, format) }

	ref := enrichGitHub(context.Background(), testGitHubRef(), runner, warnf)

	if ref.BaseSHA != "" {
		t.Fatalf("expected no BaseSHA on malformed output, got %q", ref.BaseSHA)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warning on malformed output")
	}
}

func TestEnrichPRRefDispatchesGitHub(t *testing.T) {
	runner := &stubCLIRunner{output: "main\nabc1234\n"}
	ref := testGitHubRef()
	enriched := EnrichPRRef(context.Background(), ref, runner, func(string, ...any) {})

	if enriched.BaseBranch != "main" {
		t.Fatalf("expected dispatch to github enricher, got BaseBranch=%q", enriched.BaseBranch)
	}
}

func TestEnrichPRRefWarnsForUnknownProvider(t *testing.T) {
	runner := &stubCLIRunner{}
	warnings := make([]string, 0)
	warnf := func(format string, args ...any) { warnings = append(warnings, format) }

	ref := types.PRRef{PRID: 1, Provider: "bitbucket", RepoURL: "https://bitbucket.org/org/repo"}
	EnrichPRRef(context.Background(), ref, runner, warnf)

	if len(warnings) == 0 {
		t.Fatalf("expected warning for unsupported provider")
	}
}

func TestGitHubRepoSlug(t *testing.T) {
	cases := []struct{ url, want string }{
		{"https://github.com/owner/repo", "owner/repo"},
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"https://notgithub.com/owner/repo", ""},
		{"https://github.com/owner", ""},
	}
	for _, c := range cases {
		got := githubRepoSlug(c.url)
		if got != c.want {
			t.Errorf("githubRepoSlug(%q) = %q, want %q", c.url, got, c.want)
		}
	}
}

func testGitHubRef() types.PRRef {
	return types.PRRef{PRID: 3, Provider: "github", RepoURL: "https://github.com/richardthombs/prr"}
}
