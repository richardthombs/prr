package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/engine"
	"github.com/richardthombs/prr/internal/git"
	"github.com/richardthombs/prr/internal/provider"
	"github.com/richardthombs/prr/internal/types"
)

func TestReviewCommandEmitsStructuredJSONAndKeepsDiagnosticsOffStdout(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		if name != "git" {
			t.Fatalf("unexpected command name %q", name)
		}

		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "a.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "1 file changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt", nil
		default:
			return "", nil
		}
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service { return service }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
			if name != "gh" {
				t.Fatalf("unexpected issue command %q", name)
			}
			if len(args) < 2 || args[0] != "api" {
				t.Fatalf("unexpected issue args %v", args)
			}
			return `[{"number":13,"html_url":"https://github.com/acme/repo/issues/13","title":"Issue title","body":"Issue body","state":"open","labels":[{"name":"bug"}]}]`, nil
		}}
	}
	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, input engine.ReviewInput) (types.Review, error) {
			if len(input.Bundle.Issues) != 1 {
				t.Fatalf("expected issue hydration in bundle, got %+v", input.Bundle.Issues)
			}
			if input.Bundle.Issues[0].ID != "13" {
				t.Fatalf("expected hydrated issue id 13, got %+v", input.Bundle.Issues[0])
			}
			return deterministicReview(), nil
		})
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"review", "42", "--repo", "https://github.com/acme/repo", "--provider", "github", "--remote", "origin", "--json"})

	if err := Execute(); err != nil {
		t.Fatalf("expected review command to succeed, got %v", err)
	}

	if strings.TrimSpace(stderr.String()) != "" {
		t.Fatalf("expected no diagnostics on stderr without verbose/what-if, got %q", stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid review JSON, got %v", err)
	}

	if payload["summary"] == "" {
		t.Fatalf("expected summary field in review JSON")
	}
	if _, ok := payload["risk"].(map[string]any); !ok {
		t.Fatalf("expected risk object in review JSON")
	}
	findings, ok := payload["findings"].([]any)
	if !ok {
		t.Fatalf("expected findings array in review JSON")
	}
	if len(findings) == 0 {
		t.Fatalf("expected default adapter to emit at least one finding")
	}
	firstFinding, ok := findings[0].(map[string]any)
	if !ok {
		t.Fatalf("expected finding object")
	}
	if strings.TrimSpace(firstFinding["id"].(string)) == "" {
		t.Fatalf("expected per-run finding id to be present")
	}
	if _, ok := payload["checklist"].([]any); !ok {
		t.Fatalf("expected checklist array in review JSON")
	}
}

func TestReviewCommandWhatIfVerbosePrintsCommandsToStderr(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	runner := stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
		t.Fatalf("expected no external command execution in what-if mode")
		return "", nil
	}}
	service := git.NewServiceWithCacheDir(runner, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) { return "[]", nil }}
	}
	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, _ engine.ReviewInput) (types.Review, error) {
			return deterministicReview(), nil
		})
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"review", "77", "--repo", "https://github.com/acme/repo", "--provider", "github", "--what-if", "--verbose"})

	if err := Execute(); err != nil {
		t.Fatalf("expected review what-if command to succeed, got %v", err)
	}

	stderrText := stderr.String()
	if !strings.Contains(stderrText, "exec: git clone --mirror") {
		t.Fatalf("expected mirror command preview in stderr, got %q", stderrText)
	}
	if !strings.Contains(stderrText, "diff --name-only HEAD^1..HEAD") {
		t.Fatalf("expected diff command preview in stderr, got %q", stderrText)
	}
}

func TestReviewCommandAcceptsPRURLArgument(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "a.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "1 file changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt", nil
		default:
			return "", nil
		}
	}}, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) { return "[]", nil }}
	}
	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, _ engine.ReviewInput) (types.Review, error) {
			return deterministicReview(), nil
		})
	}

	stdout := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"review", "https://github.com/acme/repo/pull/42", "--json"})

	if err := Execute(); err != nil {
		t.Fatalf("expected review command with PR URL to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid review JSON, got %v", err)
	}
	if payload["summary"] == "" {
		t.Fatalf("expected summary in review JSON")
	}
}

func TestReviewCommandAcceptsPipedCheckoutJSONWithoutArgs(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "a.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "1 file changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt", nil
		default:
			return "", nil
		}
	}}, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) { return "[]", nil }}
	}
	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, _ engine.ReviewInput) (types.Review, error) {
			return deterministicReview(), nil
		})
	}

	stdin := bytes.NewBufferString(`{"prId":73,"repoUrl":"https://github.com/acme/repo","provider":"github","remote":"origin"}`)
	stdout := &bytes.Buffer{}
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"review", "--json"})

	if err := Execute(); err != nil {
		t.Fatalf("expected review command with piped JSON to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid review JSON, got %v", err)
	}
	if payload["summary"] == "" {
		t.Fatalf("expected summary in review JSON")
	}
}

func TestReviewCommandBypassesSetupWithAuthoritativeCheckoutJSON(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	runner := stubRunner{runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "clone --mirror") || strings.Contains(joined, "fetch origin pull/") || strings.Contains(joined, "worktree add --detach") {
			t.Fatalf("expected setup stages to be bypassed, got %q", joined)
		}

		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "a.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "1 file changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt", nil
		default:
			return "", nil
		}
	}}

	service := git.NewServiceWithCacheDir(runner, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) { return "[]", nil }}
	}
	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, _ engine.ReviewInput) (types.Review, error) {
			return deterministicReview(), nil
		})
	}

	stdin := bytes.NewBufferString(`{"prId":73,"repoUrl":"https://github.com/acme/repo","provider":"github","remote":"origin","bareDir":"/tmp/bare","mergeRef":"refs/prr/pull/73/merge","workDir":"/tmp/worktree","cleanup":false}`)
	stdout := &bytes.Buffer{}
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"review", "--json"})

	if err := Execute(); err != nil {
		t.Fatalf("expected review command with full checkout JSON to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid review JSON, got %v", err)
	}
	if payload["summary"] == "" {
		t.Fatalf("expected summary in review JSON")
	}
}

func TestReviewCommandPassesModelFlagToEngine(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "a.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "1 file changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt", nil
		default:
			return "", nil
		}
	}}, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) { return "[]", nil }}
	}

	capturedModel := ""
	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, input engine.ReviewInput) (types.Review, error) {
			capturedModel = input.Model
			return deterministicReview(), nil
		})
	}

	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"review", "42", "--repo", "https://github.com/acme/repo", "--provider", "github", "--model", "gpt-5"})

	if err := Execute(); err != nil {
		t.Fatalf("expected review command to succeed, got %v", err)
	}

	if capturedModel != "gpt-5" {
		t.Fatalf("expected model flag to be passed to engine, got %q", capturedModel)
	}
}

func TestReviewCommandClassifiesEngineFailures(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "a.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "1 file changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt", nil
		default:
			return "", nil
		}
	}}, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) { return "[]", nil }}
	}

	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, _ engine.ReviewInput) (types.Review, error) {
			return types.Review{}, errors.New("adapter failed")
		})
	}

	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"review", "42", "--repo", "https://github.com/acme/repo", "--provider", "github"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected review command to fail when engine fails")
	}
	if !strings.Contains(err.Error(), "ENGINE_FAILURE") {
		t.Fatalf("expected ENGINE_FAILURE classification, got %v", err)
	}
}

func TestReviewCommandEmitsDeterministicJSONShape(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	runner := stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
		t.Fatalf("expected no external command execution in what-if mode")
		return "", nil
	}}
	mirrorServiceFactory = func() *git.Service { return git.NewServiceWithCacheDir(runner, t.TempDir()) }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) { return "[]", nil }}
	}
	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, _ engine.ReviewInput) (types.Review, error) {
			return types.Review{
				Summary: "Deterministic review",
				Risk: types.Risk{
					Score:   0.25,
					Reasons: []string{"Stable fixture"},
				},
				Findings: []types.Finding{{
					ID:         "F001",
					File:       "a.go",
					Line:       7,
					Severity:   "important",
					Category:   "tests",
					Message:    "Add coverage",
					Suggestion: "Add assertions",
				}},
				Checklist: []string{"Run CI"},
			}, nil
		})
	}

	stdout := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"review", "42", "--repo", "https://github.com/acme/repo", "--provider", "github", "--what-if", "--json"})

	if err := Execute(); err != nil {
		t.Fatalf("expected review command to succeed, got %v", err)
	}

	const expected = "{\"summary\":\"Deterministic review\",\"risk\":{\"score\":0.25,\"reasons\":[\"Stable fixture\"]},\"findings\":[{\"id\":\"F001\",\"file\":\"a.go\",\"line\":7,\"severity\":\"important\",\"category\":\"tests\",\"message\":\"Add coverage\",\"suggestion\":\"Add assertions\"}],\"checklist\":[\"Run CI\"]}\n"
	if stdout.String() != expected {
		t.Fatalf("expected deterministic JSON output.\nwant: %q\n got: %q", expected, stdout.String())
	}
}

func TestReviewCommandEmitsDeterministicMarkdown(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	originalIssueRunnerFactory := issueRunnerFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
		issueRunnerFactory = originalIssueRunnerFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "a.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "1 file changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt", nil
		default:
			return "", nil
		}
	}}, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }
	issueRunnerFactory = func() provider.CLIRunner {
		return stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) { return "[]", nil }}
	}
	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, _ engine.ReviewInput) (types.Review, error) {
			return deterministicReview(), nil
		})
	}

	runReview := func() string {
		resetReviewFlagState(t)
		stdout := &bytes.Buffer{}
		rootCmd.SetIn(bytes.NewBuffer(nil))
		rootCmd.SetOut(stdout)
		rootCmd.SetErr(&bytes.Buffer{})
		rootCmd.SetArgs([]string{"review", "42", "--repo", "https://github.com/acme/repo", "--provider", "github"})
		if err := Execute(); err != nil {
			t.Fatalf("expected review command to succeed, got %v", err)
		}
		return stdout.String()
	}

	first := runReview()
	second := runReview()

	if first != second {
		t.Fatalf("expected deterministic markdown output when running review twice with same inputs")
	}
	for _, expected := range []string{"## Summary", "## Risk", "## Findings", "## Checklist"} {
		if !strings.Contains(first, expected) {
			t.Fatalf("expected rendered markdown to include %q", expected)
		}
	}
}

type reviewEngineFunc func(ctx context.Context, input engine.ReviewInput) (types.Review, error)

func (f reviewEngineFunc) Review(ctx context.Context, input engine.ReviewInput) (types.Review, error) {
	return f(ctx, input)
}

func deterministicReview() types.Review {
	return types.Review{
		Summary: "Deterministic review",
		Risk: types.Risk{
			Score:   0.25,
			Reasons: []string{"Stable fixture"},
		},
		Findings: []types.Finding{{
			ID:         "F001",
			File:       "a.go",
			Line:       7,
			Severity:   "important",
			Category:   "tests",
			Message:    "Add coverage",
			Suggestion: "Add assertions",
		}},
		Checklist: []string{"Run CI"},
	}
}

func resetReviewFlagState(t *testing.T) {
	t.Helper()

	for _, flag := range []struct {
		name  string
		value string
	}{
		{name: "provider", value: ""},
		{name: "repo", value: ""},
		{name: "remote", value: ""},
		{name: "keep", value: "false"},
		{name: "verbose", value: "false"},
		{name: "what-if", value: "false"},
		{name: "max-patch-bytes", value: "0"},
		{name: "max-files", value: "0"},
		{name: "model", value: ""},
		{name: "json", value: "false"},
	} {
		if err := reviewCmd.Flags().Set(flag.name, flag.value); err != nil {
			t.Fatalf("failed to reset review --%s flag: %v", flag.name, err)
		}
		reviewCmd.Flags().Lookup(flag.name).Changed = false
	}
}
