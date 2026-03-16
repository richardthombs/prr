package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/types"
)

func TestFetchGitHubContextReturnsTitleAndWorkItems(t *testing.T) {
	runner := &stubCLIRunner{
		output: `{"title":"Fix auth bug","closingIssuesReferences":[{"number":42,"title":"Auth fails for new users","state":"OPEN","url":"https://github.com/acme/repo/issues/42"}]}`,
	}
	ref := types.PRRef{PRID: 7, Provider: "github", RepoURL: "https://github.com/acme/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(string, ...any) {})

	if result.Title != "Fix auth bug" {
		t.Fatalf("expected title 'Fix auth bug', got %q", result.Title)
	}
	if len(result.WorkItems) != 1 {
		t.Fatalf("expected 1 work item, got %d", len(result.WorkItems))
	}
	if result.WorkItems[0].ID != "#42" {
		t.Fatalf("expected work item ID '#42', got %q", result.WorkItems[0].ID)
	}
	if result.WorkItems[0].Title != "Auth fails for new users" {
		t.Fatalf("expected work item title, got %q", result.WorkItems[0].Title)
	}
	if result.WorkItems[0].State != "open" {
		t.Fatalf("expected state 'open', got %q", result.WorkItems[0].State)
	}
	if result.WorkItems[0].URL != "https://github.com/acme/repo/issues/42" {
		t.Fatalf("expected work item URL, got %q", result.WorkItems[0].URL)
	}
	if result.Note != "" {
		t.Fatalf("expected no note on success, got %q", result.Note)
	}
}

func TestFetchGitHubContextReturnsTitleWithNoWorkItems(t *testing.T) {
	runner := &stubCLIRunner{
		output: `{"title":"Refactor config","closingIssuesReferences":[]}`,
	}
	ref := types.PRRef{PRID: 3, Provider: "github", RepoURL: "https://github.com/acme/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(string, ...any) {})

	if result.Title != "Refactor config" {
		t.Fatalf("expected title 'Refactor config', got %q", result.Title)
	}
	if len(result.WorkItems) != 0 {
		t.Fatalf("expected no work items, got %d", len(result.WorkItems))
	}
	if result.Note != "" {
		t.Fatalf("expected no note, got %q", result.Note)
	}
}

func TestFetchGitHubContextSetsNoteOnCLIError(t *testing.T) {
	runner := &stubCLIRunner{err: errors.New("gh: command not found")}
	warnings := []string{}
	ref := types.PRRef{PRID: 1, Provider: "github", RepoURL: "https://github.com/acme/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(f string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(f, args...))
	})

	if result.Note == "" {
		t.Fatalf("expected note on CLI error, got empty")
	}
	if !strings.Contains(result.Note, "unavailable") {
		t.Fatalf("expected note to indicate unavailability, got %q", result.Note)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warning on CLI error")
	}
}

func TestFetchGitHubContextSetsNoteOnMalformedOutput(t *testing.T) {
	runner := &stubCLIRunner{output: "not json"}
	warnings := []string{}
	ref := types.PRRef{PRID: 1, Provider: "github", RepoURL: "https://github.com/acme/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(f string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(f, args...))
	})

	if result.Note == "" {
		t.Fatalf("expected note on malformed output, got empty")
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warning on malformed output")
	}
}

func TestFetchGitHubContextSetsNoteOnBadRepoURL(t *testing.T) {
	runner := &stubCLIRunner{}
	warnings := []string{}
	ref := types.PRRef{PRID: 1, Provider: "github", RepoURL: "https://notgithub.com/acme/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(f string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(f, args...))
	})

	if result.Note == "" {
		t.Fatalf("expected note on bad repo URL, got empty")
	}
}

func TestFetchAzureDevOpsContextReturnsTitleAndWorkItems(t *testing.T) {
	callCount := 0
	runner := &stubCLIRunner{}
	runner.runFuncMulti = func(_ context.Context, name string, args ...string) (string, error) {
		callCount++
		joined := strings.Join(args, " ")
		switch {
		case name == "az" && strings.Contains(joined, "repos pr show"):
			return "Implement login feature\n", nil
		case name == "az" && strings.Contains(joined, "repos pr work-item list"):
			return `[{"id":99}]`, nil
		case name == "az" && strings.Contains(joined, "boards work-item show"):
			return `{"id":99,"fields":{"System.Title":"Login feature request","System.State":"Active"},"_links":{"html":{"href":"https://dev.azure.com/org/project/_workitems/99"}}}`, nil
		default:
			return "", fmt.Errorf("unexpected call: %s %s", name, joined)
		}
	}
	ref := types.PRRef{PRID: 5, Provider: "azure-devops", RepoURL: "https://dev.azure.com/org/project/_git/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(string, ...any) {})

	if result.Title != "Implement login feature" {
		t.Fatalf("expected title 'Implement login feature', got %q", result.Title)
	}
	if len(result.WorkItems) != 1 {
		t.Fatalf("expected 1 work item, got %d", len(result.WorkItems))
	}
	if result.WorkItems[0].ID != "99" {
		t.Fatalf("expected work item ID '99', got %q", result.WorkItems[0].ID)
	}
	if result.WorkItems[0].Title != "Login feature request" {
		t.Fatalf("expected work item title, got %q", result.WorkItems[0].Title)
	}
	if result.WorkItems[0].State != "Active" {
		t.Fatalf("expected state 'Active', got %q", result.WorkItems[0].State)
	}
	if result.Note != "" {
		t.Fatalf("expected no note on success, got %q", result.Note)
	}
}

func TestFetchAzureDevOpsContextSetsNoteOnCLIError(t *testing.T) {
	runner := &stubCLIRunner{err: errors.New("az: command not found")}
	warnings := []string{}
	ref := types.PRRef{PRID: 1, Provider: "azure-devops", RepoURL: "https://dev.azure.com/org/project/_git/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(f string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(f, args...))
	})

	if result.Note == "" {
		t.Fatalf("expected note on CLI error, got empty")
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warning on CLI error")
	}
}

func TestFetchPRContextDispatchesGitHub(t *testing.T) {
	runner := &stubCLIRunner{output: `{"title":"GitHub PR","closingIssuesReferences":[]}`}
	ref := types.PRRef{PRID: 1, Provider: "github", RepoURL: "https://github.com/acme/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(string, ...any) {})

	if result.Title != "GitHub PR" {
		t.Fatalf("expected dispatch to github, got title %q", result.Title)
	}
}

func TestFetchPRContextSetsNoteForUnsupportedProvider(t *testing.T) {
	runner := &stubCLIRunner{}
	warnings := []string{}
	ref := types.PRRef{PRID: 1, Provider: "bitbucket", RepoURL: "https://bitbucket.org/acme/repo"}
	result := FetchPRContext(context.Background(), ref, runner, func(f string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(f, args...))
	})

	if result.Note == "" {
		t.Fatalf("expected note for unsupported provider, got empty")
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warning for unsupported provider")
	}
}
