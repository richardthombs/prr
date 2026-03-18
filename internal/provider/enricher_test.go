package provider

import (
	"context"
	"errors"
	"fmt"
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

// sequentialCLIRunner returns outputs in sequence, one per Run call.
type sequentialCLIRunner struct {
	responses []stubResponse
	index     int
}

type stubResponse struct {
	output string
	err    error
}

func (r *sequentialCLIRunner) Run(_ context.Context, _ string, _ ...string) (string, error) {
	if r.index >= len(r.responses) {
		return "", errors.New("unexpected CLI call")
	}
	resp := r.responses[r.index]
	r.index++
	return resp.output, resp.err
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

func testADORef() types.PRRef {
	return types.PRRef{PRID: 42, Provider: "azure-devops", RepoURL: "https://dev.azure.com/myorg/myproject/_git/myrepo"}
}

func TestFetchLinkedWorkItemsReturnsWorkItemsForADOPR(t *testing.T) {
	listJSON := `[{"id": 101}, {"id": 202}]`
	item101JSON := `{"id": 101, "fields": {"System.Title": "Add login", "System.Description": "Users need to log in", "System.WorkItemType": "User Story", "System.State": "Active"}}`
	item202JSON := `{"id": 202, "fields": {"System.Title": "Fix bug", "System.Description": "", "System.WorkItemType": "Bug", "System.State": "Committed"}}`

	runner := &sequentialCLIRunner{
		responses: []stubResponse{
			{output: listJSON},
			{output: item101JSON},
			{output: item202JSON},
		},
	}
	warnings := make([]string, 0)
	warnf := func(format string, args ...any) { warnings = append(warnings, fmt.Sprintf(format, args...)) }

	items := FetchLinkedWorkItems(context.Background(), testADORef(), runner, warnf)

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 work items, got %d", len(items))
	}
	if items[0].ID != 101 || items[0].Title != "Add login" || items[0].Type != "User Story" || items[0].State != "Active" {
		t.Fatalf("unexpected first work item: %+v", items[0])
	}
	if items[1].ID != 202 || items[1].Title != "Fix bug" || items[1].Type != "Bug" {
		t.Fatalf("unexpected second work item: %+v", items[1])
	}
}

func TestFetchLinkedWorkItemsReturnsNilForNonADOProvider(t *testing.T) {
	runner := &stubCLIRunner{}
	items := FetchLinkedWorkItems(context.Background(), testGitHubRef(), runner, func(string, ...any) {})

	if items != nil {
		t.Fatalf("expected nil for non-ADO provider, got %v", items)
	}
}

func TestFetchLinkedWorkItemsWarnsAndReturnsNilOnListError(t *testing.T) {
	runner := &stubCLIRunner{err: errors.New("az: command not found")}
	warnings := make([]string, 0)
	warnf := func(format string, args ...any) { warnings = append(warnings, fmt.Sprintf(format, args...)) }

	items := FetchLinkedWorkItems(context.Background(), testADORef(), runner, warnf)

	if items != nil {
		t.Fatalf("expected nil on list error, got %v", items)
	}
	if len(warnings) == 0 || !strings.Contains(warnings[0], "work item list unavailable") {
		t.Fatalf("expected 'work item list unavailable' warning, got %v", warnings)
	}
}

func TestFetchLinkedWorkItemsReturnsNilOnEmptyList(t *testing.T) {
	runner := &stubCLIRunner{output: `[]`}
	items := FetchLinkedWorkItems(context.Background(), testADORef(), runner, func(string, ...any) {})

	if items != nil {
		t.Fatalf("expected nil for empty work item list, got %v", items)
	}
}

func TestFetchLinkedWorkItemsSkipsItemOnShowError(t *testing.T) {
	listJSON := `[{"id": 101}, {"id": 202}]`
	item101JSON := `{"id": 101, "fields": {"System.Title": "Add login", "System.WorkItemType": "User Story", "System.State": "Active"}}`

	runner := &sequentialCLIRunner{
		responses: []stubResponse{
			{output: listJSON},
			{output: item101JSON},
			{err: errors.New("az: not found")},
		},
	}
	warnings := make([]string, 0)
	warnf := func(format string, args ...any) { warnings = append(warnings, fmt.Sprintf(format, args...)) }

	items := FetchLinkedWorkItems(context.Background(), testADORef(), runner, warnf)

	if len(items) != 1 {
		t.Fatalf("expected 1 work item after skipping failed fetch, got %d", len(items))
	}
	if items[0].ID != 101 {
		t.Fatalf("expected work item 101, got %d", items[0].ID)
	}
	if len(warnings) == 0 || !strings.Contains(warnings[0], "work item 202 fetch failed") {
		t.Fatalf("expected fetch-failed warning for item 202, got %v", warnings)
	}
}
