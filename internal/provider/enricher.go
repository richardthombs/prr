package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

// CLIRunner executes an external command and returns its stdout.
type CLIRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

// Enricher attempts to populate PR metadata (base branch, base SHA, source branch,
// source SHA) on a PRRef. If enrichment fails the original ref is returned unchanged
// and a warning is emitted — callers must check BaseSHA before relying on it.
type Enricher interface {
	Enrich(ctx context.Context, ref types.PRRef, warnf func(string, ...any)) types.PRRef
}

// EnrichPRRef enriches ref using the provided Enricher.
func EnrichPRRef(ctx context.Context, ref types.PRRef, enricher Enricher, warnf func(format string, args ...any)) types.PRRef {
	return enricher.Enrich(ctx, ref, warnf)
}

// NewDefaultEnricher creates an Enricher that reads mode and credentials from
// environment variables and builds per-provider GitHub and Azure DevOps enrichers.
// The CLI runner is used for CLI-mode enrichment.
//
// Mode is controlled by PRR_ISSUE_PROVIDER_MODE (cli / rest / cli-rest).
// Tokens are read from PRR_GITHUB_TOKEN and PRR_AZURE_DEVOPS_TOKEN.
func NewDefaultEnricher(runner CLIRunner) Enricher {
	mode, _ := parseIssueProviderMode(os.Getenv(issueProviderModeEnv))
	httpClient := &http.Client{Timeout: 30 * time.Second}
	githubToken := strings.TrimSpace(os.Getenv(githubTokenEnv))
	azureToken := strings.TrimSpace(os.Getenv(azureTokenEnv))
	githubAPIBaseURL := firstNonEmptyTrimmed(os.Getenv(githubAPIBaseURLEnv), "https://api.github.com")
	return newEnricherForMode(runner, httpClient, mode, githubToken, azureToken, githubAPIBaseURL)
}

// NewCLIEnricher creates an Enricher that uses only CLI tools (gh / az) for
// both GitHub and Azure DevOps.  Primarily useful for tests and explicit CLI-only
// configurations.
func NewCLIEnricher(runner CLIRunner) Enricher {
	return &perProviderEnricher{
		github:      NewGitHubCLIEnricher(runner),
		azureDevOps: NewAzureDevOpsCLIEnricher(runner),
	}
}

// NewGitHubCLIEnricher creates an Enricher that calls the gh CLI.
func NewGitHubCLIEnricher(runner CLIRunner) Enricher {
	return &githubCLIEnricher{runner: runner}
}

// NewGitHubRESTEnricher creates an Enricher that calls the GitHub REST API
// using token for authentication.
func NewGitHubRESTEnricher(httpClient *http.Client, token, apiBaseURL string) Enricher {
	return &githubRESTEnricher{
		httpClient:       httpClient,
		token:            token,
		githubAPIBaseURL: firstNonEmptyTrimmed(apiBaseURL, "https://api.github.com"),
	}
}

// NewAzureDevOpsCLIEnricher creates an Enricher that calls the az CLI.
func NewAzureDevOpsCLIEnricher(runner CLIRunner) Enricher {
	return &azureDevOpsCLIEnricher{runner: runner}
}

// NewAzureDevOpsRESTEnricher creates an Enricher that calls the Azure DevOps
// REST API using token for authentication.
func NewAzureDevOpsRESTEnricher(httpClient *http.Client, token string) Enricher {
	return &azureDevOpsRESTEnricher{httpClient: httpClient, token: token}
}

// EnrichmentRequiredError is returned when a merge ref is unavailable and
// enrichment either was not attempted or did not produce a usable BaseSHA.
func EnrichmentRequiredError(provider string) error {
	var hint string
	switch provider {
	case "github":
		hint = fmt.Sprintf("ensure 'gh' is installed and authenticated, or set %s", githubTokenEnv)
	case "azure-devops":
		hint = fmt.Sprintf("ensure 'az' is installed and authenticated ('az login'), or set %s", azureTokenEnv)
	default:
		hint = "install and authenticate the appropriate CLI for provider " + provider
	}

	return apperrors.WrapProvider(
		fmt.Sprintf("merge ref is unavailable and PR base branch could not be determined; %s", hint),
		nil,
	)
}

// ---- githubCLIEnricher -------------------------------------------------------

type githubCLIEnricher struct{ runner CLIRunner }

func (e *githubCLIEnricher) Enrich(ctx context.Context, ref types.PRRef, warnf func(string, ...any)) types.PRRef {
	return enrichGitHub(ctx, ref, e.runner, warnf)
}

// ---- githubRESTEnricher ------------------------------------------------------

type githubRESTEnricher struct {
	httpClient       *http.Client
	token            string
	githubAPIBaseURL string
}

func (e *githubRESTEnricher) Enrich(ctx context.Context, ref types.PRRef, warnf func(string, ...any)) types.PRRef {
	token := strings.TrimSpace(e.token)
	if token == "" {
		warnf("enrichment unavailable (GitHub REST): %s not set", githubTokenEnv)
		return ref
	}

	repoSlug := githubRepoSlug(ref.RepoURL)
	if repoSlug == "" {
		warnf("enrichment skipped (GitHub REST): could not derive repo slug from %q", ref.RepoURL)
		return ref
	}

	baseURL := strings.TrimSuffix(strings.TrimSpace(e.githubAPIBaseURL), "/")
	reqURL := fmt.Sprintf("%s/repos/%s/pulls/%d", baseURL, repoSlug, ref.PRID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		warnf("enrichment skipped (GitHub REST): failed to build request: %v", err)
		return ref
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	body, err := doHTTPRequest(e.httpClient, req)
	if err != nil {
		warnf("enrichment unavailable (GitHub REST): %v", err)
		return ref
	}

	var resp struct {
		Base struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"base"`
		Head struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		warnf("enrichment skipped (GitHub REST): could not parse response: %v", err)
		return ref
	}

	ref.BaseBranch = resp.Base.Ref
	ref.BaseSHA = resp.Base.SHA
	ref.SourceBranch = resp.Head.Ref
	ref.SourceSHA = resp.Head.SHA
	if resp.HTMLURL != "" && ref.PRURL == "" {
		ref.PRURL = resp.HTMLURL
	}
	return ref
}

// ---- azureDevOpsCLIEnricher --------------------------------------------------

type azureDevOpsCLIEnricher struct{ runner CLIRunner }

func (e *azureDevOpsCLIEnricher) Enrich(ctx context.Context, ref types.PRRef, warnf func(string, ...any)) types.PRRef {
	return enrichAzureDevOps(ctx, ref, e.runner, warnf)
}

// ---- azureDevOpsRESTEnricher -------------------------------------------------

type azureDevOpsRESTEnricher struct {
	httpClient *http.Client
	token      string
}

func (e *azureDevOpsRESTEnricher) Enrich(ctx context.Context, ref types.PRRef, warnf func(string, ...any)) types.PRRef {
	token := strings.TrimSpace(e.token)
	if token == "" {
		warnf("enrichment unavailable (Azure DevOps REST): %s not set", azureTokenEnv)
		return ref
	}

	orgProjectBase, repoName, err := azureRepoContext(ref.RepoURL)
	if err != nil {
		warnf("enrichment skipped (Azure DevOps REST): failed to derive repo context: %v", err)
		return ref
	}

	reqURL := fmt.Sprintf("%s/_apis/git/repositories/%s/pullRequests/%d?api-version=7.1",
		orgProjectBase, url.PathEscape(repoName), ref.PRID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		warnf("enrichment skipped (Azure DevOps REST): failed to build request: %v", err)
		return ref
	}
	setAzureAuthHeader(req, token)

	body, err := doHTTPRequest(e.httpClient, req)
	if err != nil {
		warnf("enrichment unavailable (Azure DevOps REST): %v", err)
		return ref
	}

	var resp struct {
		TargetRefName string `json:"targetRefName"`
		SourceRefName string `json:"sourceRefName"`
		LastMergeTargetCommit struct {
			CommitID string `json:"commitId"`
		} `json:"lastMergeTargetCommit"`
		LastMergeSourceCommit struct {
			CommitID string `json:"commitId"`
		} `json:"lastMergeSourceCommit"`
		Links struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
		} `json:"_links"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		warnf("enrichment skipped (Azure DevOps REST): could not parse response: %v", err)
		return ref
	}

	ref.BaseBranch = strings.TrimPrefix(resp.TargetRefName, "refs/heads/")
	ref.BaseSHA = resp.LastMergeTargetCommit.CommitID
	ref.SourceBranch = strings.TrimPrefix(resp.SourceRefName, "refs/heads/")
	ref.SourceSHA = resp.LastMergeSourceCommit.CommitID
	if resp.Links.Web.Href != "" && ref.PRURL == "" {
		ref.PRURL = resp.Links.Web.Href
	}
	return ref
}

// ---- compositeEnricher -------------------------------------------------------

// compositeEnricher runs primary; if BaseSHA is still empty after primary runs,
// it tries fallback. This implements the cli-rest pattern for enrichment.
type compositeEnricher struct {
	primary  Enricher
	fallback Enricher
}

func newCompositeEnricher(primary, fallback Enricher) Enricher {
	return &compositeEnricher{primary: primary, fallback: fallback}
}

func (c *compositeEnricher) Enrich(ctx context.Context, ref types.PRRef, warnf func(string, ...any)) types.PRRef {
	enriched := c.primary.Enrich(ctx, ref, warnf)
	if enriched.BaseSHA != "" {
		return enriched
	}
	return c.fallback.Enrich(ctx, enriched, warnf)
}

// ---- perProviderEnricher -----------------------------------------------------

// perProviderEnricher dispatches to provider-specific Enrichers based on
// ref.Provider.
type perProviderEnricher struct {
	github      Enricher
	azureDevOps Enricher
}

func (p *perProviderEnricher) Enrich(ctx context.Context, ref types.PRRef, warnf func(string, ...any)) types.PRRef {
	switch ref.Provider {
	case "github":
		return p.github.Enrich(ctx, ref, warnf)
	case "azure-devops":
		return p.azureDevOps.Enrich(ctx, ref, warnf)
	default:
		warnf("enrichment skipped: unsupported provider %q", ref.Provider)
		return ref
	}
}

// ---- internal helpers --------------------------------------------------------

// newEnricherForMode builds per-provider enrichers from explicit parameters.
// Primarily used by NewDefaultEnricher and tests.
func newEnricherForMode(runner CLIRunner, httpClient *http.Client, mode issueProviderMode, githubToken, azureToken, githubAPIBaseURL string) Enricher {
	ghCLI := NewGitHubCLIEnricher(runner)
	ghREST := NewGitHubRESTEnricher(httpClient, githubToken, githubAPIBaseURL)
	adoCLI := NewAzureDevOpsCLIEnricher(runner)
	adoREST := NewAzureDevOpsRESTEnricher(httpClient, azureToken)

	var gh, ado Enricher
	switch mode {
	case issueProviderModeCLI:
		gh, ado = ghCLI, adoCLI
	case issueProviderModeREST:
		gh, ado = ghREST, adoREST
	default: // cli-rest
		gh = newCompositeEnricher(ghCLI, ghREST)
		ado = newCompositeEnricher(adoCLI, adoREST)
	}

	return &perProviderEnricher{github: gh, azureDevOps: ado}
}

// enrichGitHub performs GitHub CLI enrichment.
func enrichGitHub(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) types.PRRef {
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

// enrichAzureDevOps performs Azure DevOps CLI enrichment.
func enrichAzureDevOps(ctx context.Context, ref types.PRRef, runner CLIRunner, warnf func(format string, args ...any)) types.PRRef {
	out, err := runner.Run(ctx, "az", "repos", "pr", "show",
		"--id", strconv.Itoa(ref.PRID),
		"--query", "{targetRefName:targetRefName,lastMergeTargetCommit:lastMergeTargetCommit.commitId,sourceRefName:sourceRefName,lastMergeSourceCommit:lastMergeSourceCommit.commitId,webUrl:_links.web.href}",
		"--output", "json",
	)
	if err != nil {
		warnf("enrichment unavailable (az cli): %v", err)
		return ref
	}

	var resp struct {
		TargetRefName         string `json:"targetRefName"`
		LastMergeTargetCommit string `json:"lastMergeTargetCommit"`
		SourceRefName         string `json:"sourceRefName"`
		LastMergeSourceCommit string `json:"lastMergeSourceCommit"`
		WebURL                string `json:"webUrl"`
	}
	if jsonErr := json.Unmarshal([]byte(out), &resp); jsonErr != nil {
		warnf("enrichment skipped: could not parse az output: %v", jsonErr)
		return ref
	}

	ref.BaseBranch = strings.TrimPrefix(resp.TargetRefName, "refs/heads/")
	ref.BaseSHA = resp.LastMergeTargetCommit
	ref.SourceBranch = strings.TrimPrefix(resp.SourceRefName, "refs/heads/")
	ref.SourceSHA = resp.LastMergeSourceCommit
	if resp.WebURL != "" {
		ref.PRURL = resp.WebURL
	}
	return ref
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
