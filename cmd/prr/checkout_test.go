package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/git"
)

type stubRunner struct {
	runFunc func(ctx context.Context, name string, args ...string) (string, error)
}

func (r stubRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	if r.runFunc == nil {
		return "", nil
	}

	return r.runFunc(ctx, name, args...)
}

func TestCheckoutCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "checkout" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected checkout command to be registered")
	}
}

func TestCheckoutEmitsPipelineEquivalentPayload(t *testing.T) {
	resetCheckoutFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		if name != "git" {
			t.Fatalf("unexpected command name %q", name)
		}

		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "clone --mirror") &&
			!strings.Contains(joined, "fetch origin pull/987654321/merge:refs/prr/pull/987654321/merge") &&
			!strings.Contains(joined, "worktree add --detach") {
			t.Fatalf("unexpected git invocation %q", joined)
		}

		return "", nil
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	stdout := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"checkout", "https://github.com/checkouttestuser/checkouttestrepo/pull/987654321"})

	if err := Execute(); err != nil {
		t.Fatalf("expected checkout to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON payload, got %v", err)
	}

	if payload["prId"] != float64(987654321) {
		t.Fatalf("expected prId 987654321, got %#v", payload["prId"])
	}
	if payload["repoUrl"] != "https://github.com/checkouttestuser/checkouttestrepo" {
		t.Fatalf("unexpected repoUrl %#v", payload["repoUrl"])
	}
	if payload["provider"] != "github" {
		t.Fatalf("expected provider github, got %#v", payload["provider"])
	}
	if payload["remote"] != "origin" {
		t.Fatalf("expected remote origin, got %#v", payload["remote"])
	}
	if payload["mergeRef"] != "refs/prr/pull/987654321/merge" {
		t.Fatalf("unexpected mergeRef %#v", payload["mergeRef"])
	}
	if payload["workDir"] == "" {
		t.Fatalf("expected workDir in payload")
	}
	if payload["keep"] != false {
		t.Fatalf("expected keep false, got %#v", payload["keep"])
	}
	if payload["cleanup"] != true {
		t.Fatalf("expected cleanup true, got %#v", payload["cleanup"])
	}
}

func TestCheckoutWhatIfSkipsExternalExecution(t *testing.T) {
	resetCheckoutFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	runner := stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
		t.Fatalf("expected no external command execution in what-if mode")
		return "", nil
	}}

	service := git.NewServiceWithCacheDir(runner, t.TempDir())
	mirrorServiceFactory = func() *git.Service {
		return service
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"checkout", "https://github.com/checkouttestuser2/checkouttestrepo2/pull/123456789", "--what-if", "--verbose"})

	if err := Execute(); err != nil {
		t.Fatalf("expected checkout what-if to succeed, got %v", err)
	}

	stderrText := stderr.String()
	if !strings.Contains(stderrText, "exec: git clone --mirror") {
		t.Fatalf("expected mirror command preview in stderr, got %q", stderrText)
	}
	if !strings.Contains(stderrText, "exec: git -C") || !strings.Contains(stderrText, "fetch origin pull/123456789/merge:refs/prr/pull/123456789/merge") {
		t.Fatalf("expected fetch command preview in stderr, got %q", stderrText)
	}
	if !strings.Contains(stderrText, "worktree add --detach") && !strings.Contains(stderrText, "reset --hard") {
		t.Fatalf("expected worktree command preview in stderr, got %q", stderrText)
	}
}

func resetCheckoutFlagState(t *testing.T) {
	t.Helper()

	if err := checkoutCmd.Flags().Set("provider", ""); err != nil {
		t.Fatalf("failed to reset checkout --provider flag: %v", err)
	}
	checkoutCmd.Flags().Lookup("provider").Changed = false
	if err := checkoutCmd.Flags().Set("repo", ""); err != nil {
		t.Fatalf("failed to reset checkout --repo flag: %v", err)
	}
	checkoutCmd.Flags().Lookup("repo").Changed = false
	if err := checkoutCmd.Flags().Set("remote", ""); err != nil {
		t.Fatalf("failed to reset checkout --remote flag: %v", err)
	}
	checkoutCmd.Flags().Lookup("remote").Changed = false
	if err := checkoutCmd.Flags().Set("keep", "false"); err != nil {
		t.Fatalf("failed to reset checkout --keep flag: %v", err)
	}
	checkoutCmd.Flags().Lookup("keep").Changed = false
	if err := checkoutCmd.Flags().Set("verbose", "false"); err != nil {
		t.Fatalf("failed to reset checkout --verbose flag: %v", err)
	}
	checkoutCmd.Flags().Lookup("verbose").Changed = false
	if err := checkoutCmd.Flags().Set("what-if", "false"); err != nil {
		t.Fatalf("failed to reset checkout --what-if flag: %v", err)
	}
	checkoutCmd.Flags().Lookup("what-if").Changed = false
}
