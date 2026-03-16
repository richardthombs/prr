package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/richardthombs/prr/internal/types"
)

// PRContextResult holds the PR title and linked work items fetched from a provider CLI.
type PRContextResult struct {
	Title     string
	WorkItems []types.WorkItem
	Note      string // populated when context could not be fully fetched
}

// FetchPRContext fetches the PR title and linked work items or issues using the
// provider-appropriate CLI tool (gh for GitHub, az for Azure DevOps).
//
// If the CLI is unavailable or the call fails, Note is populated with a
// human-readable explanation and the other fields may be empty or partial.
func FetchPRContext(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) PRContextResult {
	switch ref.Provider {
	case "github":
		return fetchGitHubContext(ctx, ref, runner, warnf)
	case "azure-devops":
		return fetchAzureDevOpsContext(ctx, ref, runner, warnf)
	default:
		note := fmt.Sprintf("work item context unavailable: unsupported provider %q", ref.Provider)
		warnf("%s", note)
		return PRContextResult{Note: note}
	}
}

func fetchGitHubContext(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) PRContextResult {
	repoSlug := githubRepoSlug(ref.RepoURL)
	if repoSlug == "" {
		note := "work item context unavailable: could not determine GitHub repository from URL"
		warnf("%s", note)
		return PRContextResult{Note: note}
	}

	out, err := runner.Run(ctx, "gh", "pr", "view",
		fmt.Sprintf("%d", ref.PRID),
		"--repo", repoSlug,
		"--json", "title,closingIssuesReferences",
	)
	if err != nil {
		note := "work item context unavailable: gh CLI not available or not authenticated"
		warnf("pr context fetch failed (gh cli): %v", err)
		return PRContextResult{Note: note}
	}

	var response struct {
		Title                   string `json:"title"`
		ClosingIssuesReferences []struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			State  string `json:"state"`
			URL    string `json:"url"`
		} `json:"closingIssuesReferences"`
	}
	if err := json.Unmarshal([]byte(out), &response); err != nil {
		note := "work item context unavailable: could not parse gh CLI output"
		warnf("pr context parse failed: %v", err)
		return PRContextResult{Note: note}
	}

	result := PRContextResult{Title: response.Title}
	for _, issue := range response.ClosingIssuesReferences {
		result.WorkItems = append(result.WorkItems, types.WorkItem{
			ID:    fmt.Sprintf("#%d", issue.Number),
			Title: issue.Title,
			State: strings.ToLower(issue.State),
			URL:   issue.URL,
		})
	}
	return result
}

const maxWorkItemsToFetch = 10

func fetchAzureDevOpsContext(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) PRContextResult {
	// Fetch PR title.
	titleOut, err := runner.Run(ctx, "az", "repos", "pr", "show",
		"--id", fmt.Sprintf("%d", ref.PRID),
		"--query", "title",
		"--output", "tsv",
	)
	if err != nil {
		note := "work item context unavailable: az CLI not available or not authenticated"
		warnf("pr context fetch failed (az cli): %v", err)
		return PRContextResult{Note: note}
	}
	title := strings.TrimSpace(titleOut)

	// Fetch linked work item IDs.
	wiOut, err := runner.Run(ctx, "az", "repos", "pr", "work-item", "list",
		"--id", fmt.Sprintf("%d", ref.PRID),
		"--output", "json",
	)
	if err != nil {
		warnf("work item list unavailable: %v", err)
		return PRContextResult{Title: title, Note: "linked work items could not be fetched"}
	}

	var workItemRefs []struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal([]byte(wiOut), &workItemRefs); err != nil {
		warnf("work item list parse failed: %v", err)
		return PRContextResult{Title: title, Note: "linked work items could not be parsed"}
	}

	var workItems []types.WorkItem
	for i, wi := range workItemRefs {
		if i >= maxWorkItemsToFetch {
			break
		}
		wiDetails, err := runner.Run(ctx, "az", "boards", "work-item", "show",
			"--id", fmt.Sprintf("%d", wi.ID),
			"--output", "json",
		)
		if err != nil {
			warnf("work item %d details unavailable: %v", wi.ID, err)
			workItems = append(workItems, types.WorkItem{
				ID:    fmt.Sprintf("%d", wi.ID),
				Title: "(unavailable)",
			})
			continue
		}

		var wiDetail struct {
			ID     int `json:"id"`
			Fields struct {
				Title string `json:"System.Title"`
				State string `json:"System.State"`
			} `json:"fields"`
			Links struct {
				HTML struct {
					Href string `json:"href"`
				} `json:"html"`
			} `json:"_links"`
		}
		if err := json.Unmarshal([]byte(wiDetails), &wiDetail); err != nil {
			warnf("work item %d parse failed: %v", wi.ID, err)
			workItems = append(workItems, types.WorkItem{
				ID:    fmt.Sprintf("%d", wi.ID),
				Title: "(unavailable)",
			})
			continue
		}

		workItems = append(workItems, types.WorkItem{
			ID:    fmt.Sprintf("%d", wiDetail.ID),
			Title: wiDetail.Fields.Title,
			State: wiDetail.Fields.State,
			URL:   wiDetail.Links.HTML.Href,
		})
	}

	return PRContextResult{Title: title, WorkItems: workItems}
}
