package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/types"
)

type stubCLIRunner struct {
	output  string
	err     error
	runFunc func(ctx context.Context, name string, args ...string) (string, error)
}

func (r *stubCLIRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	if r.runFunc != nil {
		return r.runFunc(ctx, name, args...)
	}
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

func TestDiscoverGitHubIssuesReturnsNormalizedIssueData(t *testing.T) {
	runner := &stubCLIRunner{
		runFunc: func(_ context.Context, name string, args ...string) (string, error) {
			if name != "gh" {
				t.Fatalf("expected gh command, got %q", name)
			}
			if len(args) < 2 || args[0] != "api" {
				t.Fatalf("unexpected gh args: %v", args)
			}
			return `[{"number":42,"html_url":"https://github.com/acme/repo/issues/42","title":"Fix race","body":"Details","state":"open","labels":[{"name":"bug"},{"name":"urgent"}]}]`, nil
		},
	}
	provider := NewDefaultProvider()

	issues, err := provider.DiscoverIssues(context.Background(), types.PRRef{
		PRID:     17,
		Provider: "github",
		RepoURL:  "https://github.com/acme/repo",
	}, runner)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	if issues[0].ID != "42" || issues[0].Type != "issue" || issues[0].Provider != "github" {
		t.Fatalf("unexpected normalized issue: %+v", issues[0])
	}
	if issues[0].Title != "Fix race" || issues[0].State != "open" {
		t.Fatalf("unexpected issue details: %+v", issues[0])
	}
	if len(issues[0].Labels) != 2 {
		t.Fatalf("expected two labels, got %+v", issues[0].Labels)
	}
}

func TestDiscoverAzureIssuesReturnsNormalizedWorkItemData(t *testing.T) {
	runner := &stubCLIRunner{
		runFunc: func(_ context.Context, name string, args ...string) (string, error) {
			if name != "az" {
				t.Fatalf("expected az command, got %q", name)
			}
			joined := strings.Join(args, " ")
			switch {
			case strings.Contains(joined, "repos pr work-item list"):
				return `[{"id":1001}]`, nil
			case strings.Contains(joined, "boards work-item show"):
				return `{"id":1001,"url":"https://dev.azure.com/org/project/_apis/wit/workItems/1001","fields":{"System.Title":"Fix release pipeline","System.State":"Active","System.WorkItemType":"Bug","System.Tags":"ops; urgent","System.Description":"Pipeline is flaky","System.TeamProject":"project"}}`, nil
			default:
				t.Fatalf("unexpected az args: %v", args)
				return "", nil
			}
		},
	}
	provider := NewDefaultProvider()

	issues, err := provider.DiscoverIssues(context.Background(), types.PRRef{
		PRID:     55,
		Provider: "azure-devops",
		RepoURL:  "https://dev.azure.com/org/project/_git/repo",
	}, runner)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected one work item, got %d", len(issues))
	}
	item := issues[0]
	if item.ID != "1001" || item.Type != "work-item" || item.Provider != "azure-devops" {
		t.Fatalf("unexpected normalized item: %+v", item)
	}
	if item.Title != "Fix release pipeline" || item.State != "Active" {
		t.Fatalf("unexpected item content: %+v", item)
	}
	if item.Metadata["workItemType"] != "Bug" || item.Metadata["teamProject"] != "project" {
		t.Fatalf("unexpected metadata: %+v", item.Metadata)
	}
}
