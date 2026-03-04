package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	apperrors "github.com/richardthombs/prr/internal/errors"
)

func TestResolveCommandEmitsDeterministicPRRefJSON(t *testing.T) {
	resetResolveFlagState(t)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"resolve", "https://dev.azure.com/ensekltd/blackbird/_git/blackbird/pullrequest/83438"})

	err := Execute()
	if err != nil {
		t.Fatalf("expected resolve command to succeed, got error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	expected := `{"prId":83438,"repoUrl":"https://dev.azure.com/ensekltd/blackbird/_git/blackbird","remote":"origin","provider":"azure-devops"}`
	if output != expected {
		t.Fatalf("unexpected JSON output. expected %q, got %q", expected, output)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("expected valid json payload, got: %v", err)
	}

	if _, ok := payload["prId"]; !ok {
		t.Fatalf("missing prId field in payload")
	}
	if _, ok := payload["repoUrl"]; !ok {
		t.Fatalf("missing repoUrl field in payload")
	}
	if _, ok := payload["remote"]; !ok {
		t.Fatalf("missing remote field in payload")
	}
	if _, ok := payload["provider"]; !ok {
		t.Fatalf("missing provider field in payload")
	}
	if _, hasSnakeCase := payload["repo_url"]; hasSnakeCase {
		t.Fatalf("unexpected snake_case field detected in payload")
	}
}

func TestResolveCommandFailsForInvalidPRURL(t *testing.T) {
	resetResolveFlagState(t)

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"resolve", "not-a-url"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected error for invalid PR URL")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassConfig {
		t.Fatalf("expected config class error, got %s", appErr.Class)
	}
	if !strings.Contains(appErr.Error(), "invalid pull request URL") {
		t.Fatalf("expected actionable invalid URL diagnostic, got %q", appErr.Error())
	}
}

func TestResolveCommandAllowsRepoOverride(t *testing.T) {
	resetResolveFlagState(t)

	stdout := &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{
		"resolve",
		"https://dev.azure.com/ensekltd/blackbird/_git/blackbird/pullrequest/83438",
		"--repo",
		"https://dev.azure.com/ensekltd/override/_git/override",
	})

	err := Execute()
	if err != nil {
		t.Fatalf("expected success with repo override, got error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.Contains(output, `"repoUrl":"https://dev.azure.com/ensekltd/override/_git/override"`) {
		t.Fatalf("expected output to use repo override, got %q", output)
	}
}

func TestResolveCommandDetectsGitHubProviderFromURL(t *testing.T) {
	resetResolveFlagState(t)

	stdout := &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"resolve", "https://github.com/steveyegge/beads/pull/2331"})

	err := Execute()
	if err != nil {
		t.Fatalf("expected success for GitHub URL, got error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.Contains(output, `"provider":"github"`) {
		t.Fatalf("expected provider github, got %q", output)
	}
	if !strings.Contains(output, `"repoUrl":"https://github.com/steveyegge/beads"`) {
		t.Fatalf("expected github repo URL, got %q", output)
	}
	if !strings.Contains(output, `"prId":2331`) {
		t.Fatalf("expected PR id 2331, got %q", output)
	}
}

func TestResolveCommandAllowsProviderAndRemoteOverrides(t *testing.T) {
	resetResolveFlagState(t)

	stdout := &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{
		"resolve",
		"https://github.com/steveyegge/beads/pull/2331",
		"--provider",
		"github-enterprise",
		"--remote",
		"upstream",
	})

	err := Execute()
	if err != nil {
		t.Fatalf("expected success with provider/remote overrides, got error: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if !strings.Contains(output, `"provider":"github-enterprise"`) {
		t.Fatalf("expected provider override in output, got %q", output)
	}
	if !strings.Contains(output, `"remote":"upstream"`) {
		t.Fatalf("expected remote override in output, got %q", output)
	}
}

func TestResolveCommandInvalidInputMapsToStableNonZeroExitCode(t *testing.T) {
	resetResolveFlagState(t)

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"resolve", "not-a-url"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected error for invalid PR URL")
	}

	if got := apperrors.ExitCode(err); got != 2 {
		t.Fatalf("expected stable config exit code 2, got %d", got)
	}
}

func TestResolveCommandFailsForInvalidArgumentCount(t *testing.T) {
	resetResolveFlagState(t)

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"resolve"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected error for invalid argument count")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassConfig {
		t.Fatalf("expected config class error, got %s", appErr.Class)
	}
}

func resetResolveFlagState(t *testing.T) {
	t.Helper()

	if err := resolveCmd.Flags().Set("provider", ""); err != nil {
		t.Fatalf("failed to reset provider flag: %v", err)
	}
	if err := resolveCmd.Flags().Set("repo", ""); err != nil {
		t.Fatalf("failed to reset repo flag: %v", err)
	}
	if err := resolveCmd.Flags().Set("remote", ""); err != nil {
		t.Fatalf("failed to reset remote flag: %v", err)
	}
}