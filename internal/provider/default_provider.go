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

type defaultProvider struct{}

func NewDefaultProvider() PRProvider {
	return &defaultProvider{}
}

func (p *defaultProvider) Resolve(_ context.Context, prID int, opts map[string]string) (types.PRRef, error) {
	repoURL := opts["repoUrl"]
	providerName := opts["provider"]
	remote := opts["remote"]

	if repoURL == "" {
		return types.PRRef{}, fmt.Errorf("missing repository URL")
	}

	return types.PRRef{
		PRID:     prID,
		RepoURL:  repoURL,
		Remote:   remote,
		Provider: providerName,
	}, nil
}

func (p *defaultProvider) DiscoverIssues(ctx context.Context, ref types.PRRef, runner CLIRunner) ([]types.RelatedIssue, error) {
	switch ref.Provider {
	case "github":
		return discoverGitHubIssues(ctx, ref, runner)
	case "azure-devops":
		return discoverAzureDevOpsIssues(ctx, ref, runner)
	default:
		return nil, apperrors.WrapProvider(fmt.Sprintf("issue discovery is not supported for provider %q", ref.Provider), nil)
	}
}

func discoverGitHubIssues(ctx context.Context, ref types.PRRef, runner CLIRunner) ([]types.RelatedIssue, error) {
	repoSlug := githubRepoSlug(ref.RepoURL)
	if repoSlug == "" {
		return nil, apperrors.WrapProvider("failed to derive GitHub repository slug for issue discovery", nil)
	}

	out, err := runner.Run(ctx, "gh", "api", fmt.Sprintf("repos/%s/pulls/%d/issues", repoSlug, ref.PRID))
	if err != nil {
		return nil, apperrors.WrapProvider("failed to discover linked GitHub issues", err)
	}

	var payload []struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		State   string `json:"state"`
		Labels  []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		return nil, apperrors.WrapProvider("failed to parse GitHub issue discovery response", err)
	}

	issues := make([]types.RelatedIssue, 0, len(payload))
	for _, item := range payload {
		if item.Number <= 0 {
			continue
		}
		labels := make([]string, 0, len(item.Labels))
		for _, label := range item.Labels {
			name := strings.TrimSpace(label.Name)
			if name != "" {
				labels = append(labels, name)
			}
		}
		issues = append(issues, types.RelatedIssue{
			ID:       strconv.Itoa(item.Number),
			Type:     "issue",
			Provider: "github",
			URL:      strings.TrimSpace(item.HTMLURL),
			Title:    strings.TrimSpace(item.Title),
			Body:     strings.TrimSpace(item.Body),
			State:    strings.TrimSpace(item.State),
			Labels:   labels,
		})
	}

	return issues, nil
}

func discoverAzureDevOpsIssues(ctx context.Context, ref types.PRRef, runner CLIRunner) ([]types.RelatedIssue, error) {
	out, err := runner.Run(ctx, "az", "repos", "pr", "work-item", "list", "--id", strconv.Itoa(ref.PRID), "--output", "json")
	if err != nil {
		return nil, apperrors.WrapProvider("failed to discover linked Azure DevOps work items", err)
	}

	var refs []struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &refs); err != nil {
		return nil, apperrors.WrapProvider("failed to parse Azure DevOps work item discovery response", err)
	}

	if len(refs) == 0 {
		return []types.RelatedIssue{}, nil
	}

	ids := make([]string, 0, len(refs))
	for _, workItem := range refs {
		if workItem.ID > 0 {
			ids = append(ids, strconv.Itoa(workItem.ID))
		}
	}

	if len(ids) == 0 {
		return []types.RelatedIssue{}, nil
	}

	type workItemResponse struct {
		ID     int `json:"id"`
		Fields struct {
			Title        string `json:"System.Title"`
			State        string `json:"System.State"`
			WorkItemType string `json:"System.WorkItemType"`
			Tags         string `json:"System.Tags"`
			Description  string `json:"System.Description"`
			TeamProject  string `json:"System.TeamProject"`
		} `json:"fields"`
		URL string `json:"url"`
	}

	fields := "System.Id,System.Title,System.State,System.WorkItemType,System.Tags,System.Description,System.TeamProject"
	items := make([]workItemResponse, 0, len(ids))
	for _, id := range ids {
		workItemOut, err := runner.Run(ctx, "az", "boards", "work-item", "show",
			"--id", id,
			"--fields", fields,
			"--output", "json",
		)
		if err != nil {
			return nil, apperrors.WrapProvider("failed to fetch Azure DevOps work item details", err)
		}

		var item workItemResponse
		if err := json.Unmarshal([]byte(strings.TrimSpace(workItemOut)), &item); err != nil {
			return nil, apperrors.WrapProvider("failed to parse Azure DevOps work item details response", err)
		}
		items = append(items, item)
	}

	issues := make([]types.RelatedIssue, 0, len(items))
	for _, item := range items {
		if item.ID <= 0 {
			continue
		}

		var labels []string
		for _, tag := range strings.Split(item.Fields.Tags, ";") {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				labels = append(labels, trimmed)
			}
		}

		metadata := map[string]string{}
		if strings.TrimSpace(item.Fields.WorkItemType) != "" {
			metadata["workItemType"] = strings.TrimSpace(item.Fields.WorkItemType)
		}
		if strings.TrimSpace(item.Fields.TeamProject) != "" {
			metadata["teamProject"] = strings.TrimSpace(item.Fields.TeamProject)
		}

		issues = append(issues, types.RelatedIssue{
			ID:       strconv.Itoa(item.ID),
			Type:     "work-item",
			Provider: "azure-devops",
			URL:      strings.TrimSpace(item.URL),
			Title:    strings.TrimSpace(item.Fields.Title),
			Body:     strings.TrimSpace(item.Fields.Description),
			State:    strings.TrimSpace(item.Fields.State),
			Labels:   labels,
			Metadata: metadata,
		})
	}

	return issues, nil
}
