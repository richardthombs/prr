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
	"github.com/richardthombs/prr/internal/types"
)

func TestReviewCommandEmitsStructuredJSONAndKeepsDiagnosticsOffStdout(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
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
	reviewEngineFactory = func() engine.ReviewEngine {
		return engine.NewDefaultAdapter()
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"review", "42", "--repo", "https://github.com/acme/repo", "--provider", "github", "--remote", "origin"})

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
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
	})

	runner := stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
		t.Fatalf("expected no external command execution in what-if mode")
		return "", nil
	}}
	service := git.NewServiceWithCacheDir(runner, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }
	reviewEngineFactory = func() engine.ReviewEngine {
		return engine.NewDefaultAdapter()
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
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
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
	reviewEngineFactory = func() engine.ReviewEngine {
		return engine.NewDefaultAdapter()
	}

	stdout := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"review", "https://github.com/acme/repo/pull/42"})

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
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
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
	reviewEngineFactory = func() engine.ReviewEngine {
		return engine.NewDefaultAdapter()
	}

	stdin := bytes.NewBufferString(`{"prId":73,"repoUrl":"https://github.com/acme/repo","provider":"github","remote":"origin"}`)
	stdout := &bytes.Buffer{}
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"review"})

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

func TestReviewCommandClassifiesEngineFailures(t *testing.T) {
	resetReviewFlagState(t)

	originalMirrorFactory := mirrorServiceFactory
	originalEngineFactory := reviewEngineFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalMirrorFactory
		reviewEngineFactory = originalEngineFactory
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

	reviewEngineFactory = func() engine.ReviewEngine {
		return reviewEngineFunc(func(_ context.Context, _ types.BundleV1) (types.Review, error) {
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

func TestRenderCommandProducesDeterministicMarkdown(t *testing.T) {
	resetRenderFlagState(t)

	payload := `{"summary":"Review summary","risk":{"score":0.7,"reasons":["High churn"]},"findings":[{"id":"F002","file":"b.go","line":20,"severity":"important","category":"tests","message":"Missing tests","suggestion":"Add tests"},{"id":"F001","file":"a.go","line":10,"severity":"blocker","category":"security","message":"Input unsanitised","suggestion":"Sanitise input"}],"checklist":["Re-run CI"]}`
	renderOnce := func() (string, string, error) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		rootCmd.SetIn(bytes.NewBufferString(payload))
		rootCmd.SetOut(stdout)
		rootCmd.SetErr(stderr)
		rootCmd.SetArgs([]string{"render"})

		err := Execute()
		return stdout.String(), stderr.String(), err
	}

	text, errText, err := renderOnce()
	if err != nil {
		t.Fatalf("expected render command to succeed, got %v", err)
	}

	for _, expected := range []string{"## Summary", "## Risk", "## Findings", "### Blocker", "### Important", "## Checklist"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected markdown output to include %q, got %q", expected, text)
		}
	}

	firstBlocker := strings.Index(text, "### Blocker")
	firstImportant := strings.Index(text, "### Important")
	if firstBlocker == -1 || firstImportant == -1 || firstBlocker > firstImportant {
		t.Fatalf("expected findings grouped by severity order, got %q", text)
	}

	if strings.TrimSpace(errText) != "" {
		t.Fatalf("expected empty stderr for render without verbose/what-if, got %q", errText)
	}

	textAgain, errTextAgain, err := renderOnce()
	if err != nil {
		t.Fatalf("expected second render command run to succeed, got %v", err)
	}
	if text != textAgain {
		t.Fatalf("expected byte-identical deterministic markdown output, first=%q second=%q", text, textAgain)
	}
	if strings.TrimSpace(errTextAgain) != "" {
		t.Fatalf("expected empty stderr on second render run, got %q", errTextAgain)
	}
}

func TestRenderCommandRejectsMalformedJSON(t *testing.T) {
	resetRenderFlagState(t)

	rootCmd.SetIn(bytes.NewBufferString(`{"summary":`))
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"render"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected render command to fail for malformed JSON")
	}
	if !strings.Contains(err.Error(), "CONFIG_INVALID") {
		t.Fatalf("expected CONFIG_INVALID classification, got %v", err)
	}
}

func TestRenderCommandRejectsMissingRequiredFields(t *testing.T) {
	resetRenderFlagState(t)

	rootCmd.SetIn(bytes.NewBufferString(`{"summary":"","risk":{"score":0.3,"reasons":[]},"findings":[],"checklist":[]}`))
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"render"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected render command to fail for invalid review payload")
	}
	if !strings.Contains(err.Error(), "CONFIG_INVALID") {
		t.Fatalf("expected CONFIG_INVALID classification, got %v", err)
	}
}

func TestRenderCommandRejectsMissingFindingID(t *testing.T) {
	resetRenderFlagState(t)

	rootCmd.SetIn(bytes.NewBufferString(`{"summary":"Review summary","risk":{"score":0.2,"reasons":["Low risk"]},"findings":[{"id":"","file":"a.go","line":11,"severity":"important","category":"tests","message":"Missing tests","suggestion":"Add tests"}],"checklist":["Run CI"]}`))
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"render"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected render command to fail for missing finding id")
	}
	if !strings.Contains(err.Error(), "CONFIG_INVALID") {
		t.Fatalf("expected CONFIG_INVALID classification, got %v", err)
	}
}

func TestRenderCommandRejectsNonPositiveFindingLine(t *testing.T) {
	resetRenderFlagState(t)

	rootCmd.SetIn(bytes.NewBufferString(`{"summary":"Review summary","risk":{"score":0.2,"reasons":["Low risk"]},"findings":[{"id":"F001","file":"a.go","line":0,"severity":"important","category":"tests","message":"Missing tests","suggestion":"Add tests"}],"checklist":["Run CI"]}`))
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"render"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected render command to fail for non-positive finding line")
	}
	if !strings.Contains(err.Error(), "CONFIG_INVALID") {
		t.Fatalf("expected CONFIG_INVALID classification, got %v", err)
	}
}

func TestRenderCommandVerboseWhatIfDiagnostics(t *testing.T) {
	resetRenderFlagState(t)

	stdin := bytes.NewBufferString(`{"summary":"ok","risk":{"score":0.1,"reasons":[]},"findings":[],"checklist":[]}`)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"render", "--verbose", "--what-if"})

	if err := Execute(); err != nil {
		t.Fatalf("expected render command to succeed, got %v", err)
	}

	if !strings.Contains(stderr.String(), "render: transform review JSON to markdown") {
		t.Fatalf("expected verbose diagnostics in stderr, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "what-if: render stage uses no external commands") {
		t.Fatalf("expected what-if diagnostics in stderr, got %q", stderr.String())
	}
}

type reviewEngineFunc func(ctx context.Context, bundle types.BundleV1) (types.Review, error)

func (f reviewEngineFunc) Review(ctx context.Context, bundle types.BundleV1) (types.Review, error) {
	return f(ctx, bundle)
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
	} {
		if err := reviewCmd.Flags().Set(flag.name, flag.value); err != nil {
			t.Fatalf("failed to reset review --%s flag: %v", flag.name, err)
		}
		reviewCmd.Flags().Lookup(flag.name).Changed = false
	}
}

func resetRenderFlagState(t *testing.T) {
	t.Helper()

	for _, flag := range []struct {
		name  string
		value string
	}{
		{name: "verbose", value: "false"},
		{name: "what-if", value: "false"},
	} {
		if err := renderCmd.Flags().Set(flag.name, flag.value); err != nil {
			t.Fatalf("failed to reset render --%s flag: %v", flag.name, err)
		}
		renderCmd.Flags().Lookup(flag.name).Changed = false
	}
}
