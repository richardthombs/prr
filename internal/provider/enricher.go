package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

// CLIRunner executes an external command and returns its stdout.
type CLIRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

// EnrichPRRef attempts to populate BaseBranch and BaseSHA on the given PRRef
// using the provider-appropriate CLI (gh for GitHub, az for Azure DevOps).
//
// If the CLI is not available or the call fails, a warning is written to
// warnf and the original ref is returned unchanged — callers must check
// whether enrichment succeeded by inspecting BaseSHA before relying on it.
func EnrichPRRef(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) types.PRRef {
	switch ref.Provider {
	case "github":
		return enrichGitHub(ctx, ref, runner, warnf)
	case "azure-devops":
		return enrichAzureDevOps(ctx, ref, runner, warnf)
	default:
		warnf("enrichment skipped: unsupported provider %q", ref.Provider)
		return ref
	}
}

// EnrichmentRequiredError is returned when a merge ref is unavailable and
// enrichment either was not attempted or did not produce a usable BaseSHA.
func EnrichmentRequiredError(provider string) error {
	var hint string
	switch provider {
	case "github":
		hint = "ensure 'gh' is installed and authenticated"
	case "azure-devops":
		hint = "ensure 'az' is installed and authenticated ('az login')"
	default:
		hint = "install and authenticate the appropriate CLI for provider " + provider
	}

	return apperrors.WrapProvider(
		fmt.Sprintf("merge ref is unavailable and PR base branch could not be determined; %s", hint),
		nil,
	)
}

type githubPRResponse struct {
	BaseRefName string `json:"base.ref"`
	BaseRefOid  string `json:"base.sha"`
}

func enrichGitHub(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) types.PRRef {
	// Derive "owner/repo" from the repo URL.
	repoSlug := githubRepoSlug(ref.RepoURL)
	if repoSlug == "" {
		warnf("enrichment skipped: could not derive repo slug from %q", ref.RepoURL)
		return ref
	}

	out, err := runner.Run(ctx, "gh", "api",
		fmt.Sprintf("repos/%s/pulls/%d", repoSlug, ref.PRID),
		"--jq", ".base.ref,.base.sha",
	)
	if err != nil {
		warnf("enrichment unavailable (gh cli): %v", err)
		return ref
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) == "" || strings.TrimSpace(lines[1]) == "" {
		warnf("enrichment skipped: unexpected gh output %q", out)
		return ref
	}

	ref.BaseBranch = strings.TrimSpace(lines[0])
	ref.BaseSHA = strings.TrimSpace(lines[1])
	return ref
}

type azurePRResponse struct {
	TargetRefName string `json:"targetRefName"`
}

func enrichAzureDevOps(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) types.PRRef {
	out, err := runner.Run(ctx, "az", "repos", "pr", "show",
		"--id", strconv.Itoa(ref.PRID),
		"--query", "{targetRefName:targetRefName,lastMergeTargetCommit:lastMergeTargetCommit.commitId}",
		"--output", "json",
	)
	if err != nil {
		warnf("enrichment unavailable (az cli): %v", err)
		return ref
	}

	var resp struct {
		TargetRefName          string `json:"targetRefName"`
		LastMergeTargetCommit  string `json:"lastMergeTargetCommit"`
	}
	if jsonErr := json.Unmarshal([]byte(out), &resp); jsonErr != nil {
		warnf("enrichment skipped: could not parse az output: %v", jsonErr)
		return ref
	}

	// ADO returns refs/heads/main — strip the prefix for a clean branch name.
	ref.BaseBranch = strings.TrimPrefix(resp.TargetRefName, "refs/heads/")
	ref.BaseSHA = resp.LastMergeTargetCommit
	return ref
}

// FetchLinkedWorkItems retrieves the ADO work items linked to a pull request and
// returns their details. Only supported for the "azure-devops" provider; any
// other provider returns nil. CLI errors are reported via warnf and a nil slice
// is returned so that missing work items never block the review.
func FetchLinkedWorkItems(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) []types.WorkItem {
	if ref.Provider != "azure-devops" {
		return nil
	}

	// List the work items linked to the PR.
	listOut, err := runner.Run(ctx, "az", "repos", "pr", "work-items", "list",
		"--id", strconv.Itoa(ref.PRID),
		"--output", "json",
	)
	if err != nil {
		warnf("work item list unavailable (az cli): %v", err)
		return nil
	}

	var listed []struct {
		ID int `json:"id"`
	}
	if jsonErr := json.Unmarshal([]byte(listOut), &listed); jsonErr != nil {
		warnf("work item list skipped: could not parse az output: %v", jsonErr)
		return nil
	}

	if len(listed) == 0 {
		return nil
	}

	// Fetch details for each linked work item.
	workItems := make([]types.WorkItem, 0, len(listed))
	for _, item := range listed {
		wi, fetchErr := fetchWorkItemDetails(ctx, item.ID, runner)
		if fetchErr != nil {
			warnf("work item %d fetch failed: %v", item.ID, fetchErr)
			continue
		}
		workItems = append(workItems, wi)
	}

	return workItems
}

func fetchWorkItemDetails(ctx context.Context, id int, runner CLIRunner) (types.WorkItem, error) {
	out, err := runner.Run(ctx, "az", "boards", "work-item", "show",
		"--id", strconv.Itoa(id),
		"--output", "json",
	)
	if err != nil {
		return types.WorkItem{}, err
	}

	var resp struct {
		ID     int `json:"id"`
		Fields struct {
			Title        string `json:"System.Title"`
			Description  string `json:"System.Description"`
			WorkItemType string `json:"System.WorkItemType"`
			State        string `json:"System.State"`
		} `json:"fields"`
	}
	if jsonErr := json.Unmarshal([]byte(out), &resp); jsonErr != nil {
		return types.WorkItem{}, fmt.Errorf("could not parse work item JSON: %w", jsonErr)
	}

	return types.WorkItem{
		ID:          resp.ID,
		Type:        resp.Fields.WorkItemType,
		Title:       resp.Fields.Title,
		Description: resp.Fields.Description,
		State:       resp.Fields.State,
	}, nil
}

// githubRepoSlug extracts "owner/repo" from a GitHub HTTPS URL.
func githubRepoSlug(repoURL string) string {
	trimmed := strings.TrimSpace(repoURL)
	trimmed = strings.TrimSuffix(trimmed, ".git")
	const prefix = "https://github.com/"
	if !strings.HasPrefix(trimmed, prefix) {
		return ""
	}
	slug := strings.TrimPrefix(trimmed, prefix)
	parts := strings.SplitN(slug, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	return parts[0] + "/" + parts[1]
}
