package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	apperrors "github.com/richardthombs/prr/internal/errors"
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

func TestMirrorAndPRRefCommandsRegistered(t *testing.T) {
	mirrorFound := false
	prrefFound := false

	for _, cmd := range rootCmd.Commands() {
		switch cmd.Name() {
		case "mirror":
			mirrorFound = true
		case "prref":
			prrefFound = true
		}
	}

	if !mirrorFound {
		t.Fatalf("expected mirror command to be registered")
	}

	if !prrefFound {
		t.Fatalf("expected prref command to be registered")
	}
}

func TestMirrorEnsureEmitsJSONWithBareDir(t *testing.T) {
	resetMirrorPRRefFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	cacheRoot := t.TempDir()
	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		if name != "git" {
			t.Fatalf("unexpected command name %q", name)
		}
		if len(args) < 3 || args[0] != "clone" || args[1] != "--mirror" {
			t.Fatalf("expected git clone --mirror invocation, got %v", args)
		}

		return "", nil
	}}, cacheRoot)

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	stdout := &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"mirror", "ensure", "--repo", "https://example.test/org/repo"})

	if err := Execute(); err != nil {
		t.Fatalf("expected mirror ensure to succeed, got %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["repoUrl"] != "https://example.test/org/repo" {
		t.Fatalf("expected repoUrl in payload, got %q", payload["repoUrl"])
	}
	if payload["bareDir"] == "" {
		t.Fatalf("expected bareDir in payload")
	}
}

func TestPRRefFetchEmitsMergeRefAndContext(t *testing.T) {
	resetMirrorPRRefFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		if name != "git" {
			t.Fatalf("unexpected command name %q", name)
		}

		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "fetch origin pull/101/merge:refs/prr/pull/101/merge") {
			t.Fatalf("expected fetch destination to use PRR namespace, got %q", joined)
		}

		return "", nil
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	stdout := &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"prref", "fetch", "--repo", "https://example.test/org/repo", "--pr-id", "101", "--bare-dir", "/tmp/mirror.git"})

	if err := Execute(); err != nil {
		t.Fatalf("expected prref fetch to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["mergeRef"] != "refs/prr/pull/101/merge" {
		t.Fatalf("expected mergeRef in payload, got %#v", payload["mergeRef"])
	}
	if payload["bareDir"] != "/tmp/mirror.git" {
		t.Fatalf("expected bareDir in payload, got %#v", payload["bareDir"])
	}
}

func TestPRRefFetchClassifiesFetchFailureAsProviderError(t *testing.T) {
	resetMirrorPRRefFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
		return "", errors.New("merge ref unavailable")
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"prref", "fetch", "--repo", "https://example.test/org/repo", "--pr-id", "321", "--bare-dir", "/tmp/mirror.git"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected provider error for missing merge ref")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassProvider {
		t.Fatalf("expected provider class error, got %s", appErr.Class)
	}
}

func TestMirrorEnsureFailsWithoutRepo(t *testing.T) {
	resetMirrorPRRefFlagState(t)

	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"mirror", "ensure"})

	err := Execute()
	if err == nil {
		t.Fatalf("expected config error when repo missing")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassConfig {
		t.Fatalf("expected config class error, got %s", appErr.Class)
	}
}

func resetMirrorPRRefFlagState(t *testing.T) {
	t.Helper()

	if err := mirrorEnsureCmd.Flags().Set("repo", ""); err != nil {
		t.Fatalf("failed to reset mirror --repo flag: %v", err)
	}
	if err := mirrorEnsureCmd.Flags().Set("verbose", "false"); err != nil {
		t.Fatalf("failed to reset mirror --verbose flag: %v", err)
	}
	if err := mirrorEnsureCmd.Flags().Set("lock-timeout", "30s"); err != nil {
		t.Fatalf("failed to reset mirror --lock-timeout flag: %v", err)
	}
	if err := mirrorEnsureCmd.Flags().Set("force", "false"); err != nil {
		t.Fatalf("failed to reset mirror --force flag: %v", err)
	}
	if err := mirrorEnsureCmd.Flags().Set("what-if", "false"); err != nil {
		t.Fatalf("failed to reset mirror --what-if flag: %v", err)
	}

	if err := prrefFetchCmd.Flags().Set("pr-id", "0"); err != nil {
		t.Fatalf("failed to reset prref --pr-id flag: %v", err)
	}
	if err := prrefFetchCmd.Flags().Set("repo", ""); err != nil {
		t.Fatalf("failed to reset prref --repo flag: %v", err)
	}
	if err := prrefFetchCmd.Flags().Set("remote", "origin"); err != nil {
		t.Fatalf("failed to reset prref --remote flag: %v", err)
	}
	if err := prrefFetchCmd.Flags().Set("provider", ""); err != nil {
		t.Fatalf("failed to reset prref --provider flag: %v", err)
	}
	if err := prrefFetchCmd.Flags().Set("bare-dir", ""); err != nil {
		t.Fatalf("failed to reset prref --bare-dir flag: %v", err)
	}
}
