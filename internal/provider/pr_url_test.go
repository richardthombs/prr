package provider

import "testing"

func TestParsePullRequestURLAzureDevOps(t *testing.T) {
	parsed, err := parsePullRequestURL("https://dev.azure.com/ensekltd/blackbird/_git/blackbird/pullrequest/83438")
	if err != nil {
		t.Fatalf("expected successful URL parsing, got error: %v", err)
	}

	if parsed.PRID != 83438 {
		t.Fatalf("expected PR ID 83438, got %d", parsed.PRID)
	}
	if parsed.RepoURL != "https://dev.azure.com/ensekltd/blackbird/_git/blackbird" {
		t.Fatalf("unexpected repo URL %q", parsed.RepoURL)
	}
	if parsed.Provider != "azure-devops" {
		t.Fatalf("unexpected provider %q", parsed.Provider)
	}
	if parsed.Remote != "origin" {
		t.Fatalf("unexpected remote %q", parsed.Remote)
	}
}

func TestParsePullRequestURLVisualStudio(t *testing.T) {
	parsed, err := parsePullRequestURL("https://ensekltd.visualstudio.com/blackbird/_git/blackbird/pullrequest/84945")
	if err != nil {
		t.Fatalf("expected successful URL parsing, got error: %v", err)
	}

	if parsed.PRID != 84945 {
		t.Fatalf("expected PR ID 84945, got %d", parsed.PRID)
	}
	if parsed.RepoURL != "https://ensekltd.visualstudio.com/blackbird/_git/blackbird" {
		t.Fatalf("unexpected repo URL %q", parsed.RepoURL)
	}
	if parsed.Provider != "azure-devops" {
		t.Fatalf("unexpected provider %q", parsed.Provider)
	}
	if parsed.Remote != "origin" {
		t.Fatalf("unexpected remote %q", parsed.Remote)
	}
}

func TestParsePullRequestURLFailsForUnsupportedProvider(t *testing.T) {
	_, err := parsePullRequestURL("https://gitlab.com/example/repo/-/merge_requests/42")
	if err == nil {
		t.Fatalf("expected unsupported provider error")
	}
}

func TestParsePullRequestURLGitHub(t *testing.T) {
	parsed, err := parsePullRequestURL("https://github.com/steveyegge/beads/pull/2331")
	if err != nil {
		t.Fatalf("expected successful URL parsing, got error: %v", err)
	}

	if parsed.PRID != 2331 {
		t.Fatalf("expected PR ID 2331, got %d", parsed.PRID)
	}
	if parsed.RepoURL != "https://github.com/steveyegge/beads" {
		t.Fatalf("unexpected repo URL %q", parsed.RepoURL)
	}
	if parsed.Provider != "github" {
		t.Fatalf("unexpected provider %q", parsed.Provider)
	}
	if parsed.Remote != "origin" {
		t.Fatalf("unexpected remote %q", parsed.Remote)
	}
}

func TestParsePullRequestURLFailsForMalformedAzurePath(t *testing.T) {
	_, err := parsePullRequestURL("https://dev.azure.com/ensekltd/blackbird/_git/blackbird/pull/83438")
	if err == nil {
		t.Fatalf("expected malformed azure URL error")
	}
}

func TestParsePullRequestURLAzureDevOpsCaseSensitive(t *testing.T) {
	parsed, err := parsePullRequestURL("https://dev.azure.com/ensekltd/PayAsYouGo/_git/Payg/pullrequest/84677")
	if err != nil {
		t.Fatalf("expected successful URL parsing, got error: %v", err)
	}

	if parsed.PRID != 84677 {
		t.Fatalf("expected PR ID 84677, got %d", parsed.PRID)
	}
	if parsed.RepoURL != "https://dev.azure.com/ensekltd/PayAsYouGo/_git/Payg" {
		t.Fatalf("unexpected repo URL %q", parsed.RepoURL)
	}
	if parsed.Provider != "azure-devops" {
		t.Fatalf("unexpected provider %q", parsed.Provider)
	}
}

func TestParsePullRequestURLVisualStudioCaseSensitive(t *testing.T) {
	parsed, err := parsePullRequestURL("https://ensekltd.visualstudio.com/PayAsYouGo/_git/Payg/pullrequest/84677")
	if err != nil {
		t.Fatalf("expected successful URL parsing, got error: %v", err)
	}

	if parsed.PRID != 84677 {
		t.Fatalf("expected PR ID 84677, got %d", parsed.PRID)
	}
	if parsed.RepoURL != "https://ensekltd.visualstudio.com/PayAsYouGo/_git/Payg" {
		t.Fatalf("unexpected repo URL %q", parsed.RepoURL)
	}
	if parsed.Provider != "azure-devops" {
		t.Fatalf("unexpected provider %q", parsed.Provider)
	}
}