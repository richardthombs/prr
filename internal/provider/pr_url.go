package provider

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type pullRequestContext struct {
	PRID     int
	RepoURL  string
	Provider string
	Remote   string
}

func parsePullRequestURL(rawURL string) (pullRequestContext, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return pullRequestContext{}, fmt.Errorf("invalid pull request URL")
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return pullRequestContext{}, fmt.Errorf("invalid pull request URL")
	}

	switch parsedURL.Host {
	case "dev.azure.com":
		return parseAzureDevOpsPullRequestURL(parsedURL)
	case "github.com":
		return parseGitHubPullRequestURL(parsedURL)
	default:
		return pullRequestContext{}, fmt.Errorf("unsupported pull request URL provider")
	}
}


func parseAzureDevOpsPullRequestURL(parsedURL *url.URL) (pullRequestContext, error) {
	pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathSegments) < 6 {
		return pullRequestContext{}, fmt.Errorf("invalid Azure DevOps pull request URL format")
	}

	if pathSegments[2] != "_git" || pathSegments[4] != "pullrequest" {
		return pullRequestContext{}, fmt.Errorf("invalid Azure DevOps pull request URL format")
	}

	prID, err := strconv.Atoi(pathSegments[5])
	if err != nil || prID <= 0 {
		return pullRequestContext{}, fmt.Errorf("invalid pull request identifier in URL")
	}

	repoURL := fmt.Sprintf("%s://%s/%s", parsedURL.Scheme, parsedURL.Host, strings.Join(pathSegments[:4], "/"))

	return pullRequestContext{
		PRID:     prID,
		RepoURL:  repoURL,
		Provider: "azure-devops",
		Remote:   "origin",
	}, nil
}

func parseGitHubPullRequestURL(parsedURL *url.URL) (pullRequestContext, error) {
	pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathSegments) < 4 {
		return pullRequestContext{}, fmt.Errorf("invalid GitHub pull request URL format")
	}

	if pathSegments[2] != "pull" {
		return pullRequestContext{}, fmt.Errorf("invalid GitHub pull request URL format")
	}

	prID, err := strconv.Atoi(pathSegments[3])
	if err != nil || prID <= 0 {
		return pullRequestContext{}, fmt.Errorf("invalid pull request identifier in URL")
	}

	repoURL := fmt.Sprintf("%s://%s/%s/%s", parsedURL.Scheme, parsedURL.Host, pathSegments[0], pathSegments[1])

	return pullRequestContext{
		PRID:     prID,
		RepoURL:  repoURL,
		Provider: "github",
		Remote:   "origin",
	}, nil
}