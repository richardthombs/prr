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

func TestPRRefFetchWhatIfPrintsPlannedCommandAndSkipsExecution(t *testing.T) {
	resetMirrorPRRefFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	runner := &stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
		t.Fatalf("expected no external command execution in what-if mode")
		return "", nil
	}}
	service := git.NewServiceWithCacheDir(runner, t.TempDir())

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"prref", "fetch", "--pr-id", "202", "--bare-dir", "/tmp/mirror.git", "--what-if"})

	if err := Execute(); err != nil {
		t.Fatalf("expected prref fetch what-if to succeed, got %v", err)
	}

	if !strings.Contains(stderr.String(), "exec: git -C /tmp/mirror.git fetch origin pull/202/merge:refs/prr/pull/202/merge") {
		t.Fatalf("expected planned git fetch command in stderr, got %q", stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["mergeRef"] != "refs/prr/pull/202/merge" {
		t.Fatalf("expected mergeRef in payload, got %#v", payload["mergeRef"])
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

func TestMirrorEnsureUsesRepoFromStdinJSON(t *testing.T) {
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
	stdin := bytes.NewBufferString(`{"prId":777,"repoUrl":"https://example.test/from/stdin","remote":"origin","provider":"github"}`)
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"mirror", "ensure"})

	if err := Execute(); err != nil {
		t.Fatalf("expected mirror ensure to succeed from stdin payload, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["repoUrl"] != "https://example.test/from/stdin" {
		t.Fatalf("expected stdin repoUrl to be used, got %q", payload["repoUrl"])
	}
	if payload["prId"] != float64(777) {
		t.Fatalf("expected stdin prId to be preserved, got %#v", payload["prId"])
	}
}

func TestPRRefFetchUsesFieldsFromStdinJSON(t *testing.T) {
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
		if !strings.Contains(joined, "fetch upstream pull/303/merge:refs/prr/pull/303/merge") {
			t.Fatalf("expected stdin-derived fetch invocation, got %q", joined)
		}

		return "", nil
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	stdout := &bytes.Buffer{}
	stdin := bytes.NewBufferString(`{"prId":303,"repoUrl":"https://example.test/from/stdin","remote":"upstream","provider":"github","bareDir":"/tmp/stdin-mirror.git"}`)
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"prref", "fetch"})

	if err := Execute(); err != nil {
		t.Fatalf("expected prref fetch to succeed from stdin payload, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["prId"] != float64(303) {
		t.Fatalf("expected stdin prId in payload, got %#v", payload["prId"])
	}
	if payload["remote"] != "upstream" {
		t.Fatalf("expected stdin remote in payload, got %#v", payload["remote"])
	}
	if payload["bareDir"] != "/tmp/stdin-mirror.git" {
		t.Fatalf("expected stdin bareDir in payload, got %#v", payload["bareDir"])
	}
}

func TestPRRefFetchFlagOverridesStdin(t *testing.T) {
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
		if !strings.Contains(joined, "fetch origin pull/404/merge:refs/prr/pull/404/merge") {
			t.Fatalf("expected flag-derived fetch invocation, got %q", joined)
		}

		return "", nil
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	stdout := &bytes.Buffer{}
	stdin := bytes.NewBufferString(`{"prId":303,"remote":"upstream","bareDir":"/tmp/stdin-mirror.git"}`)
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"prref", "fetch", "--pr-id", "404", "--remote", "origin", "--bare-dir", "/tmp/flag-mirror.git"})

	if err := Execute(); err != nil {
		t.Fatalf("expected prref fetch to succeed with flag overrides, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["prId"] != float64(404) {
		t.Fatalf("expected flag prId in payload, got %#v", payload["prId"])
	}
	if payload["remote"] != "origin" {
		t.Fatalf("expected flag remote in payload, got %#v", payload["remote"])
	}
	if payload["bareDir"] != "/tmp/flag-mirror.git" {
		t.Fatalf("expected flag bareDir in payload, got %#v", payload["bareDir"])
	}
}

func TestComposablePipelineResolveMirrorPRRef(t *testing.T) {
	resetResolveFlagState(t)
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
		if strings.Contains(joined, "clone --mirror") {
			return "", nil
		}
		if strings.Contains(joined, " fetch origin pull/2331/merge:refs/prr/pull/2331/merge") {
			return "", nil
		}

		t.Fatalf("unexpected git invocation %q", joined)
		return "", nil
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	resolveOut := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(nil))
	rootCmd.SetOut(resolveOut)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"resolve", "https://github.com/steveyegge/beads/pull/2331"})
	if err := Execute(); err != nil {
		t.Fatalf("resolve should succeed, got %v", err)
	}

	mirrorOut := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(bytes.TrimSpace(resolveOut.Bytes())))
	rootCmd.SetOut(mirrorOut)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"mirror", "ensure"})
	if err := Execute(); err != nil {
		t.Fatalf("mirror ensure should succeed with piped stdin, got %v", err)
	}

	prrefOut := &bytes.Buffer{}
	rootCmd.SetIn(bytes.NewBuffer(bytes.TrimSpace(mirrorOut.Bytes())))
	rootCmd.SetOut(prrefOut)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"prref", "fetch"})
	if err := Execute(); err != nil {
		t.Fatalf("prref fetch should succeed with piped stdin, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(prrefOut.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["prId"] != float64(2331) {
		t.Fatalf("expected prId 2331, got %#v", payload["prId"])
	}
	if payload["mergeRef"] != "refs/prr/pull/2331/merge" {
		t.Fatalf("expected mergeRef for PR 2331, got %#v", payload["mergeRef"])
	}
}

func resetMirrorPRRefFlagState(t *testing.T) {
	t.Helper()

	rootCmd.SetIn(bytes.NewBuffer(nil))

	if err := mirrorEnsureCmd.Flags().Set("repo", ""); err != nil {
		t.Fatalf("failed to reset mirror --repo flag: %v", err)
	}
	mirrorEnsureCmd.Flags().Lookup("repo").Changed = false
	if err := mirrorEnsureCmd.Flags().Set("verbose", "false"); err != nil {
		t.Fatalf("failed to reset mirror --verbose flag: %v", err)
	}
	mirrorEnsureCmd.Flags().Lookup("verbose").Changed = false
	if err := mirrorEnsureCmd.Flags().Set("lock-timeout", "30s"); err != nil {
		t.Fatalf("failed to reset mirror --lock-timeout flag: %v", err)
	}
	mirrorEnsureCmd.Flags().Lookup("lock-timeout").Changed = false
	if err := mirrorEnsureCmd.Flags().Set("force", "false"); err != nil {
		t.Fatalf("failed to reset mirror --force flag: %v", err)
	}
	mirrorEnsureCmd.Flags().Lookup("force").Changed = false
	if err := mirrorEnsureCmd.Flags().Set("what-if", "false"); err != nil {
		t.Fatalf("failed to reset mirror --what-if flag: %v", err)
	}
	mirrorEnsureCmd.Flags().Lookup("what-if").Changed = false

	if err := prrefFetchCmd.Flags().Set("pr-id", "0"); err != nil {
		t.Fatalf("failed to reset prref --pr-id flag: %v", err)
	}
	prrefFetchCmd.Flags().Lookup("pr-id").Changed = false
	if err := prrefFetchCmd.Flags().Set("repo", ""); err != nil {
		t.Fatalf("failed to reset prref --repo flag: %v", err)
	}
	prrefFetchCmd.Flags().Lookup("repo").Changed = false
	if err := prrefFetchCmd.Flags().Set("remote", "origin"); err != nil {
		t.Fatalf("failed to reset prref --remote flag: %v", err)
	}
	prrefFetchCmd.Flags().Lookup("remote").Changed = false
	if err := prrefFetchCmd.Flags().Set("provider", ""); err != nil {
		t.Fatalf("failed to reset prref --provider flag: %v", err)
	}
	prrefFetchCmd.Flags().Lookup("provider").Changed = false
	if err := prrefFetchCmd.Flags().Set("bare-dir", ""); err != nil {
		t.Fatalf("failed to reset prref --bare-dir flag: %v", err)
	}
	prrefFetchCmd.Flags().Lookup("bare-dir").Changed = false
	if err := prrefFetchCmd.Flags().Set("verbose", "false"); err != nil {
		t.Fatalf("failed to reset prref --verbose flag: %v", err)
	}
	prrefFetchCmd.Flags().Lookup("verbose").Changed = false
	if err := prrefFetchCmd.Flags().Set("what-if", "false"); err != nil {
		t.Fatalf("failed to reset prref --what-if flag: %v", err)
	}
	prrefFetchCmd.Flags().Lookup("what-if").Changed = false
}
