package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

const (
	issueProviderModeEnv = "PRR_ISSUE_PROVIDER_MODE"
	githubTokenEnv       = "PRR_GITHUB_TOKEN"
	azureTokenEnv        = "PRR_AZURE_DEVOPS_TOKEN"
	githubAPIBaseURLEnv  = "PRR_GITHUB_API_BASE_URL"
	githubGraphQLQuery   = `query($owner: String!, $repo: String!, $number: Int!) { repository(owner: $owner, name: $repo) { pullRequest(number: $number) { closingIssuesReferences(first: 100) { nodes { number url title body state labels(first: 50) { nodes { name } } } } } } }`
)

type issueProviderMode string

const (
	issueProviderModeCLI     issueProviderMode = "cli"
	issueProviderModeREST    issueProviderMode = "rest"
	issueProviderModeCLIREST issueProviderMode = "cli-rest"
)

type defaultProvider struct {
	mode             issueProviderMode
	httpClient       *http.Client
	githubToken      string
	azureDevOpsToken string
	githubAPIBaseURL string
	configErr        error
}

type azureWorkItemResponse struct {
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

func NewDefaultProvider() PRProvider {
	mode, modeErr := parseIssueProviderMode(os.Getenv(issueProviderModeEnv))
	return &defaultProvider{
		mode:             mode,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		githubToken:      strings.TrimSpace(os.Getenv(githubTokenEnv)),
		azureDevOpsToken: strings.TrimSpace(os.Getenv(azureTokenEnv)),
		githubAPIBaseURL: firstNonEmptyTrimmed(os.Getenv(githubAPIBaseURLEnv), "https://api.github.com"),
		configErr:        modeErr,
	}
}

// NewProviderWithValues creates a PRProvider using explicit values rather than
// environment variables. Callers are responsible for supplying resolved config.
func NewProviderWithValues(mode, githubToken, azureDevOpsToken, githubAPIBaseURL string) PRProvider {
	parsedMode, modeErr := parseIssueProviderMode(mode)
	return &defaultProvider{
		mode:             parsedMode,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		githubToken:      strings.TrimSpace(githubToken),
		azureDevOpsToken: strings.TrimSpace(azureDevOpsToken),
		githubAPIBaseURL: firstNonEmptyTrimmed(githubAPIBaseURL, "https://api.github.com"),
		configErr:        modeErr,
	}
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
	if p.configErr != nil {
		return nil, apperrors.WrapConfig("invalid issue provider configuration", p.configErr)
	}

	switch ref.Provider {
	case "github":
		return p.discoverGitHubIssues(ctx, ref, runner)
	case "azure-devops":
		return p.discoverAzureDevOpsIssues(ctx, ref, runner)
	default:
		return nil, apperrors.WrapProvider(fmt.Sprintf("issue discovery is not supported for provider %q", ref.Provider), nil)
	}
}

func parseIssueProviderMode(raw string) (issueProviderMode, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return issueProviderModeCLIREST, nil
	}

	switch issueProviderMode(value) {
	case issueProviderModeCLI, issueProviderModeREST, issueProviderModeCLIREST:
		return issueProviderMode(value), nil
	default:
		return "", fmt.Errorf("%s must be one of: cli, rest, cli-rest", issueProviderModeEnv)
	}
}

func (p *defaultProvider) discoverGitHubIssues(ctx context.Context, ref types.PRRef, runner CLIRunner) ([]types.RelatedIssue, error) {
	switch p.mode {
	case issueProviderModeCLI:
		return discoverGitHubIssuesCLI(ctx, ref, runner)
	case issueProviderModeREST:
		return p.discoverGitHubIssuesREST(ctx, ref)
	case issueProviderModeCLIREST:
		cliIssues, cliErr := discoverGitHubIssuesCLI(ctx, ref, runner)
		if cliErr == nil {
			return cliIssues, nil
		}
		restIssues, restErr := p.discoverGitHubIssuesREST(ctx, ref)
		if restErr == nil {
			return restIssues, nil
		}
		return nil, apperrors.WrapProvider(
			"failed to discover linked GitHub issues using CLI and REST fallback",
			fmt.Errorf("cli error: %w; rest error: %w", cliErr, restErr),
		)
	default:
		return nil, apperrors.WrapConfig("unsupported issue provider mode", fmt.Errorf("mode=%q", p.mode))
	}
}

func discoverGitHubIssuesCLI(ctx context.Context, ref types.PRRef, runner CLIRunner) ([]types.RelatedIssue, error) {
	repoSlug := githubRepoSlug(ref.RepoURL)
	if repoSlug == "" {
		return nil, apperrors.WrapProvider("failed to derive GitHub repository slug for issue discovery", nil)
	}
	owner, repo, ok := splitGitHubRepoSlug(repoSlug)
	if !ok {
		return nil, apperrors.WrapProvider("failed to derive GitHub repository owner/name for issue discovery", nil)
	}

	out, err := runner.Run(
		ctx,
		"gh", "api", "graphql",
		"-f", "query="+githubGraphQLQuery,
		"-F", "owner="+owner,
		"-F", "repo="+repo,
		"-F", fmt.Sprintf("number=%d", ref.PRID),
	)
	if err != nil {
		return nil, apperrors.WrapProvider("failed to discover linked GitHub issues", err)
	}

	return parseGitHubGraphQLIssueList(out)
}

func (p *defaultProvider) discoverGitHubIssuesREST(ctx context.Context, ref types.PRRef) ([]types.RelatedIssue, error) {
	token := strings.TrimSpace(p.githubToken)
	if token == "" {
		return nil, apperrors.WrapProvider(fmt.Sprintf("GitHub REST fallback requires %s", githubTokenEnv), nil)
	}

	repoSlug := githubRepoSlug(ref.RepoURL)
	if repoSlug == "" {
		return nil, apperrors.WrapProvider("failed to derive GitHub repository slug for issue discovery", nil)
	}
	owner, repo, ok := splitGitHubRepoSlug(repoSlug)
	if !ok {
		return nil, apperrors.WrapProvider("failed to derive GitHub repository owner/name for issue discovery", nil)
	}

	baseURL := strings.TrimSuffix(strings.TrimSpace(p.githubAPIBaseURL), "/")
	bodyBytes, err := json.Marshal(map[string]any{
		"query": githubGraphQLQuery,
		"variables": map[string]any{
			"owner":  owner,
			"repo":   repo,
			"number": ref.PRID,
		},
	})
	if err != nil {
		return nil, apperrors.WrapProvider("failed to encode GitHub REST request for issue discovery", err)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/graphql", baseURL),
		strings.NewReader(string(bodyBytes)),
	)
	if err != nil {
		return nil, apperrors.WrapProvider("failed to build GitHub REST request for issue discovery", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	body, err := p.doRequest(req)
	if err != nil {
		return nil, apperrors.WrapProvider("failed to discover linked GitHub issues via REST", err)
	}

	return parseGitHubGraphQLIssueList(body)
}

func parseGitHubGraphQLIssueList(raw string) ([]types.RelatedIssue, error) {
	trimmedRaw := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmedRaw, "[") {
		return parseGitHubIssueListLegacy(trimmedRaw)
	}

	var payload struct {
		Data struct {
			Repository struct {
				PullRequest *struct {
					ClosingIssuesReferences struct {
						Nodes []struct {
							Number int    `json:"number"`
							URL    string `json:"url"`
							Title  string `json:"title"`
							Body   string `json:"body"`
							State  string `json:"state"`
							Labels struct {
								Nodes []struct {
									Name string `json:"name"`
								} `json:"nodes"`
							} `json:"labels"`
						} `json:"nodes"`
					} `json:"closingIssuesReferences"`
				} `json:"pullRequest"`
			} `json:"repository"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, apperrors.WrapProvider("failed to parse GitHub issue discovery response", err)
	}
	if len(payload.Errors) > 0 {
		return nil, apperrors.WrapProvider(
			"failed to parse GitHub issue discovery response",
			fmt.Errorf("graphql error: %s", strings.TrimSpace(payload.Errors[0].Message)),
		)
	}
	if payload.Data.Repository.PullRequest == nil {
		return []types.RelatedIssue{}, nil
	}

	nodes := payload.Data.Repository.PullRequest.ClosingIssuesReferences.Nodes
	issues := make([]types.RelatedIssue, 0, len(nodes))
	for _, item := range nodes {
		if item.Number <= 0 {
			continue
		}
		labels := make([]string, 0, len(item.Labels.Nodes))
		for _, label := range item.Labels.Nodes {
			name := strings.TrimSpace(label.Name)
			if name != "" {
				labels = append(labels, name)
			}
		}
		issues = append(issues, types.RelatedIssue{
			ID:       strconv.Itoa(item.Number),
			Type:     "issue",
			Provider: "github",
			URL:      strings.TrimSpace(item.URL),
			Title:    strings.TrimSpace(item.Title),
			Body:     strings.TrimSpace(item.Body),
			State:    strings.ToLower(strings.TrimSpace(item.State)),
			Labels:   labels,
		})
	}

	return issues, nil
}

func parseGitHubIssueListLegacy(raw string) ([]types.RelatedIssue, error) {
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
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
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
			State:    strings.ToLower(strings.TrimSpace(item.State)),
			Labels:   labels,
		})
	}

	return issues, nil
}

func splitGitHubRepoSlug(slug string) (owner string, repo string, ok bool) {
	parts := strings.Split(strings.TrimSpace(slug), "/")
	if len(parts) != 2 {
		return "", "", false
	}
	owner = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])
	if owner == "" || repo == "" {
		return "", "", false
	}
	return owner, repo, true
}

func (p *defaultProvider) discoverAzureDevOpsIssues(ctx context.Context, ref types.PRRef, runner CLIRunner) ([]types.RelatedIssue, error) {
	switch p.mode {
	case issueProviderModeCLI:
		return discoverAzureDevOpsIssuesCLI(ctx, ref, runner)
	case issueProviderModeREST:
		return p.discoverAzureDevOpsIssuesREST(ctx, ref)
	case issueProviderModeCLIREST:
		cliIssues, cliErr := discoverAzureDevOpsIssuesCLI(ctx, ref, runner)
		if cliErr == nil {
			return cliIssues, nil
		}
		restIssues, restErr := p.discoverAzureDevOpsIssuesREST(ctx, ref)
		if restErr == nil {
			return restIssues, nil
		}
		return nil, apperrors.WrapProvider(
			"failed to discover linked Azure DevOps work items using CLI and REST fallback",
			fmt.Errorf("cli error: %w; rest error: %w", cliErr, restErr),
		)
	default:
		return nil, apperrors.WrapConfig("unsupported issue provider mode", fmt.Errorf("mode=%q", p.mode))
	}
}

func discoverAzureDevOpsIssuesCLI(ctx context.Context, ref types.PRRef, runner CLIRunner) ([]types.RelatedIssue, error) {
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

	fields := "System.Id,System.Title,System.State,System.WorkItemType,System.Tags,System.Description,System.TeamProject"
	items := make([]azureWorkItemResponse, 0, len(ids))
	for _, id := range ids {
		workItemOut, err := runner.Run(ctx, "az", "boards", "work-item", "show",
			"--id", id,
			"--fields", fields,
			"--output", "json",
		)
		if err != nil {
			return nil, apperrors.WrapProvider("failed to fetch Azure DevOps work item details", err)
		}

		var item azureWorkItemResponse
		if err := json.Unmarshal([]byte(strings.TrimSpace(workItemOut)), &item); err != nil {
			return nil, apperrors.WrapProvider("failed to parse Azure DevOps work item details response", err)
		}
		items = append(items, item)
	}

	return buildAzureIssueList(items), nil
}

func (p *defaultProvider) discoverAzureDevOpsIssuesREST(ctx context.Context, ref types.PRRef) ([]types.RelatedIssue, error) {
	token := strings.TrimSpace(p.azureDevOpsToken)
	if token == "" {
		return nil, apperrors.WrapProvider(fmt.Sprintf("Azure DevOps REST fallback requires %s", azureTokenEnv), nil)
	}

	orgProjectBase, repoName, err := azureRepoContext(ref.RepoURL)
	if err != nil {
		return nil, apperrors.WrapProvider("failed to derive Azure DevOps repository context for issue discovery", err)
	}

	workItemsURL := fmt.Sprintf("%s/_apis/git/repositories/%s/pullRequests/%d/workitems?api-version=7.1",
		orgProjectBase,
		url.PathEscape(repoName),
		ref.PRID,
	)
	workItemsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, workItemsURL, nil)
	if err != nil {
		return nil, apperrors.WrapProvider("failed to build Azure DevOps REST request for work-item links", err)
	}
	setAzureAuthHeader(workItemsReq, token)

	workItemsBody, err := p.doRequest(workItemsReq)
	if err != nil {
		return nil, apperrors.WrapProvider("failed to discover linked Azure DevOps work items via REST", err)
	}

	var refsPayload struct {
		Value []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	if err := json.Unmarshal([]byte(workItemsBody), &refsPayload); err != nil {
		return nil, apperrors.WrapProvider("failed to parse Azure DevOps work item discovery response", err)
	}

	ids := make([]string, 0, len(refsPayload.Value))
	for _, item := range refsPayload.Value {
		id := strings.TrimSpace(item.ID)
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return []types.RelatedIssue{}, nil
	}

	fields := "System.Id,System.Title,System.State,System.WorkItemType,System.Tags,System.Description,System.TeamProject"
	items := make([]azureWorkItemResponse, 0, len(ids))
	for _, id := range ids {
		itemURL := fmt.Sprintf("%s/_apis/wit/workitems/%s?api-version=7.1&fields=%s",
			orgProjectBase,
			url.PathEscape(id),
			url.QueryEscape(fields),
		)
		itemReq, err := http.NewRequestWithContext(ctx, http.MethodGet, itemURL, nil)
		if err != nil {
			return nil, apperrors.WrapProvider("failed to build Azure DevOps REST request for work item details", err)
		}
		setAzureAuthHeader(itemReq, token)

		itemBody, err := p.doRequest(itemReq)
		if err != nil {
			return nil, apperrors.WrapProvider("failed to fetch Azure DevOps work item details via REST", err)
		}

		var item azureWorkItemResponse
		if err := json.Unmarshal([]byte(itemBody), &item); err != nil {
			return nil, apperrors.WrapProvider("failed to parse Azure DevOps work item details response", err)
		}
		items = append(items, item)
	}

	return buildAzureIssueList(items), nil
}

func buildAzureIssueList(items []azureWorkItemResponse) []types.RelatedIssue {
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
			URL:      azureWorkItemWebURL(item.URL, item.ID, item.Fields.TeamProject),
			Title:    strings.TrimSpace(item.Fields.Title),
			Body:     strings.TrimSpace(item.Fields.Description),
			State:    strings.TrimSpace(item.Fields.State),
			Labels:   labels,
			Metadata: metadata,
		})
	}

	return issues
}

func (p *defaultProvider) doRequest(req *http.Request) (string, error) {
	return doHTTPRequest(p.httpClient, req)
}

// doHTTPRequest executes req using client, reads the body (up to 2 MiB), and
// returns it as a string. Non-2xx status codes are treated as errors.
func doHTTPRequest(client *http.Client, req *http.Request) (string, error) {
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed reading HTTP response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, req.URL.String(), sanitizeHTTPBody(body))
	}

	return string(body), nil
}

func sanitizeHTTPBody(body []byte) string {
	trimmed := strings.Join(strings.Fields(strings.TrimSpace(string(body))), " ")
	if trimmed == "" {
		return "<empty>"
	}
	const maxLen = 220
	if len(trimmed) > maxLen {
		return trimmed[:maxLen] + "..."
	}
	return trimmed
}

func setAzureAuthHeader(req *http.Request, token string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(":" + strings.TrimSpace(token)))
	req.Header.Set("Authorization", "Basic "+encoded)
}

// azureWorkItemWebURL converts the REST API URL returned by Azure DevOps
// (e.g. https://dev.azure.com/org/project/_apis/wit/workItems/1)
// into the human-facing work item URL
// (e.g. https://dev.azure.com/org/project/_workitems/edit/1).
// When the API URL is org-scoped (no project segment), teamProject is appended.
func azureWorkItemWebURL(apiURL string, id int, teamProject string) string {
	apiURL = strings.TrimSpace(apiURL)
	parsed, err := url.Parse(apiURL)
	if err != nil || parsed.Host == "" {
		return apiURL
	}

	apisIdx := strings.Index(parsed.Path, "/_apis/")
	if apisIdx < 0 {
		return apiURL
	}

	basePath := parsed.Path[:apisIdx]

	// For dev.azure.com the path is /{org}/{project}. If only /{org} is present
	// and we know the project, append it so the link resolves correctly.
	if parsed.Host == "dev.azure.com" {
		segments := strings.Split(strings.Trim(basePath, "/"), "/")
		if len(segments) == 1 && strings.TrimSpace(teamProject) != "" {
			basePath = basePath + "/" + url.PathEscape(strings.TrimSpace(teamProject))
		}
	}

	return fmt.Sprintf("%s://%s%s/_workitems/edit/%d", parsed.Scheme, parsed.Host, basePath, id)
}

func azureRepoContext(repoURL string) (orgProjectBase string, repoName string, err error) {
	parsed, err := url.Parse(strings.TrimSpace(repoURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", "", fmt.Errorf("invalid Azure DevOps repository URL")
	}

	pathSegments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	switch {
	case parsed.Host == "dev.azure.com":
		if len(pathSegments) < 4 || pathSegments[2] != "_git" {
			return "", "", fmt.Errorf("unsupported Azure DevOps repository URL format")
		}
		return fmt.Sprintf("%s://%s/%s/%s", parsed.Scheme, parsed.Host, pathSegments[0], pathSegments[1]), pathSegments[3], nil
	case strings.HasSuffix(parsed.Host, ".visualstudio.com"):
		if len(pathSegments) < 3 || pathSegments[1] != "_git" {
			return "", "", fmt.Errorf("unsupported Azure DevOps repository URL format")
		}
		return fmt.Sprintf("%s://%s/%s", parsed.Scheme, parsed.Host, pathSegments[0]), pathSegments[2], nil
	default:
		return "", "", fmt.Errorf("unsupported Azure DevOps host %q", parsed.Host)
	}
}

func firstNonEmptyTrimmed(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(fallback)
}
