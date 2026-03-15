package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type recorderRunner struct {
	commands [][]string
	err      error
}

func (r *recorderRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	record := []string{name}
	record = append(record, args...)
	r.commands = append(r.commands, record)

	if r.err != nil {
		return "", r.err
	}

	return "", nil
}

func TestResolveMirrorDirDeterministic(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	first, err := service.ResolveMirrorDir("https://example.test/org/repo")
	if err != nil {
		t.Fatalf("expected first resolve to succeed, got %v", err)
	}

	second, err := service.ResolveMirrorDir("https://example.test/org/repo")
	if err != nil {
		t.Fatalf("expected second resolve to succeed, got %v", err)
	}

	if first != second {
		t.Fatalf("expected deterministic mirror path, got %q and %q", first, second)
	}
	if !strings.HasSuffix(first, ".git") {
		t.Fatalf("expected deterministic mirror path to end with .git, got %q", first)
	}
	if filepath.Base(first) != "example-org-repo.git" {
		t.Fatalf("expected human-readable mirror path segment, got %q", filepath.Base(first))
	}
}

func TestResolveMirrorDirGitHubUsesProviderOwnerRepoSlug(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	path, err := service.ResolveMirrorDir("https://github.com/steveyegge/beads")
	if err != nil {
		t.Fatalf("expected mirror dir resolution to succeed, got %v", err)
	}

	if filepath.Base(path) != "github-steveyegge-beads.git" {
		t.Fatalf("expected github provider-project-repo naming, got %q", filepath.Base(path))
	}
}

func TestResolveMirrorDirAzureDevOpsUsesProviderProjectRepoSlug(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	// org (ensekltd) is skipped; only project+repo are used to match the visualstudio.com shape
	path, err := service.ResolveMirrorDir("https://dev.azure.com/ensekltd/blackbird/_git/blackbird")
	if err != nil {
		t.Fatalf("expected mirror dir resolution to succeed, got %v", err)
	}

	if filepath.Base(path) != "azure-blackbird-blackbird.git" {
		t.Fatalf("expected azure provider-project-repo naming, got %q", filepath.Base(path))
	}
}

func TestResolveMirrorDirAzureDevOpsCaseSensitiveSlug(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	path, err := service.ResolveMirrorDir("https://dev.azure.com/ensekltd/PayAsYouGo/_git/Payg")
	if err != nil {
		t.Fatalf("expected mirror dir resolution to succeed, got %v", err)
	}

	if filepath.Base(path) != "azure-PayAsYouGo-Payg.git" {
		t.Fatalf("expected case-preserving azure project-repo slug, got %q", filepath.Base(path))
	}
}

func TestResolveMirrorDirVisualStudioUsesAzureProviderSlug(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	path, err := service.ResolveMirrorDir("https://ensekltd.visualstudio.com/blackbird/_git/blackbird")
	if err != nil {
		t.Fatalf("expected mirror dir resolution to succeed, got %v", err)
	}

	if filepath.Base(path) != "azure-blackbird-blackbird.git" {
		t.Fatalf("expected azure provider-project-repo naming, got %q", filepath.Base(path))
	}
}

func TestResolveMirrorDirVisualStudioCaseSensitiveSlug(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	path, err := service.ResolveMirrorDir("https://ensekltd.visualstudio.com/PayAsYouGo/_git/Payg")
	if err != nil {
		t.Fatalf("expected mirror dir resolution to succeed, got %v", err)
	}

	if filepath.Base(path) != "azure-PayAsYouGo-Payg.git" {
		t.Fatalf("expected case-preserving azure project-repo slug, got %q", filepath.Base(path))
	}
}

func TestResolveMirrorDirAzureDevOpsAndVisualStudioProduceSameSlug(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	devAzurePath, err := service.ResolveMirrorDir("https://dev.azure.com/ensekltd/PayAsYouGo/_git/Payg")
	if err != nil {
		t.Fatalf("expected dev.azure.com resolution to succeed, got %v", err)
	}

	vsPath, err := service.ResolveMirrorDir("https://ensekltd.visualstudio.com/PayAsYouGo/_git/Payg")
	if err != nil {
		t.Fatalf("expected visualstudio.com resolution to succeed, got %v", err)
	}

	if filepath.Base(devAzurePath) != filepath.Base(vsPath) {
		t.Fatalf("expected both Azure DevOps URL formats to produce the same slug, got %q and %q",
			filepath.Base(devAzurePath), filepath.Base(vsPath))
	}
}

func TestEnsureMirrorUsesUpdateForExistingMirror(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	bareDir, err := service.EnsureMirror(context.Background(), "https://example.test/org/repo")
	if err != nil {
		t.Fatalf("expected initial mirror ensure to succeed, got %v", err)
	}

	if err := os.MkdirAll(bareDir, 0o755); err != nil {
		t.Fatalf("failed to create simulated mirror dir: %v", err)
	}

	runner.commands = nil
	_, err = service.EnsureMirror(context.Background(), "https://example.test/org/repo")
	if err != nil {
		t.Fatalf("expected second mirror ensure to succeed, got %v", err)
	}

	if len(runner.commands) != 1 {
		t.Fatalf("expected one update command for existing mirror, got %d", len(runner.commands))
	}

	command := strings.Join(runner.commands[0], " ")
	if !strings.Contains(command, "git -C "+bareDir+" remote update --prune") {
		t.Fatalf("expected remote update command, got %q", command)
	}
}

func TestFetchPRMergeRefUsesPRRNamespaceDestination(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	mergeRef, err := service.FetchPRMergeRef(context.Background(), "/tmp/repo.git", "origin", 42)
	if err != nil {
		t.Fatalf("expected fetch to succeed, got %v", err)
	}

	if mergeRef != "refs/prr/pull/42/merge" {
		t.Fatalf("expected merge ref namespace, got %q", mergeRef)
	}

	if len(runner.commands) != 1 {
		t.Fatalf("expected one fetch command, got %d", len(runner.commands))
	}

	command := strings.Join(runner.commands[0], " ")
	if !strings.Contains(command, "fetch origin pull/42/merge:refs/prr/pull/42/merge") {
		t.Fatalf("expected fetch destination to target PRR namespace, got %q", command)
	}
}

func TestFetchPRMergeRefClassifiesErrorsAsProviderFailures(t *testing.T) {
	runner := &recorderRunner{err: errors.New("no merge ref")}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	_, err := service.FetchPRMergeRef(context.Background(), "/tmp/repo.git", "origin", 99)
	if err == nil {
		t.Fatalf("expected provider-classified error")
	}

	if !strings.Contains(err.Error(), "PROVIDER_RESOLUTION") {
		t.Fatalf("expected provider-classified error, got %v", err)
	}
}

func TestEnsureMirrorTimesOutWhenLockHeld(t *testing.T) {
	cacheDir := t.TempDir()
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, cacheDir)

	bareDir, err := service.ResolveMirrorDir("https://example.test/org/repo")
	if err != nil {
		t.Fatalf("expected mirror dir resolution to succeed, got %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(bareDir), 0o755); err != nil {
		t.Fatalf("failed to create cache root: %v", err)
	}

	lockPath := bareDir + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("failed to open lock file: %v", err)
	}
	defer lockFile.Close()

	unlock, err := holdTestLock(lockFile)
	if err != nil {
		t.Fatalf("failed to hold lock for timeout test: %v", err)
	}
	defer unlock()

	_, err = service.EnsureMirrorWithOptions(context.Background(), "https://example.test/org/repo", EnsureOptions{LockTimeout: 300 * time.Millisecond})
	if err == nil {
		t.Fatalf("expected lock timeout error")
	}

	if !strings.Contains(err.Error(), "timed out waiting for mirror lock") {
		t.Fatalf("expected lock timeout diagnostic, got %v", err)
	}
}

func TestEnsureMirrorForceBypassesHeldLock(t *testing.T) {
	cacheDir := t.TempDir()
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, cacheDir)

	bareDir, err := service.ResolveMirrorDir("https://example.test/org/repo")
	if err != nil {
		t.Fatalf("expected mirror dir resolution to succeed, got %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(bareDir), 0o755); err != nil {
		t.Fatalf("failed to create cache root: %v", err)
	}

	lockPath := bareDir + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("failed to open lock file: %v", err)
	}
	defer lockFile.Close()

	unlock, err := holdTestLock(lockFile)
	if err != nil {
		t.Fatalf("failed to hold lock for force test: %v", err)
	}
	defer unlock()

	_, err = service.EnsureMirrorWithOptions(context.Background(), "https://example.test/org/repo", EnsureOptions{LockTimeout: 100 * time.Millisecond, ForceLock: true})
	if err != nil {
		t.Fatalf("expected force lock bypass to succeed, got %v", err)
	}

	if len(runner.commands) != 1 {
		t.Fatalf("expected a git command to run under --force, got %d", len(runner.commands))
	}
}

func TestEnsureMirrorVerboseLogsCommandBeforeExecution(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	logs := make([]string, 0)
	_, err := service.EnsureMirrorWithOptions(context.Background(), "https://example.test/org/repo", EnsureOptions{
		LockTimeout: 5 * time.Second,
		Verbose:     true,
		Logger: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatalf("expected ensure mirror to succeed, got %v", err)
	}

	if len(logs) == 0 {
		t.Fatalf("expected command logs when verbose enabled")
	}
	if !strings.Contains(logs[0], "exec: git clone --mirror") {
		t.Fatalf("expected pre-execution git clone log, got %q", logs[0])
	}
}

func TestEnsureMirrorWhatIfLogsWithoutExecutingCommands(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	logs := make([]string, 0)
	bareDir, err := service.EnsureMirrorWithOptions(context.Background(), "https://example.test/org/repo", EnsureOptions{
		WhatIf: true,
		Logger: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
		Verbose: true,
	})
	if err != nil {
		t.Fatalf("expected what-if ensure mirror to succeed, got %v", err)
	}

	if bareDir == "" {
		t.Fatalf("expected bareDir to be resolved in what-if mode")
	}
	if len(logs) == 0 {
		t.Fatalf("expected command logs in what-if mode")
	}
	if !strings.Contains(logs[0], "exec: git clone --mirror") {
		t.Fatalf("expected logged clone command in what-if mode, got %q", logs[0])
	}
	if len(runner.commands) != 0 {
		t.Fatalf("expected zero external commands executed in what-if mode, got %d", len(runner.commands))
	}
}

func TestFetchPRHeadRefUsesPRRNamespaceDestination(t *testing.T) {
runner := &recorderRunner{}
service := NewServiceWithCacheDir(runner, t.TempDir())

headRef, err := service.FetchPRHeadRef(context.Background(), "/tmp/repo.git", "origin", 42, EnsureOptions{})
if err != nil {
t.Fatalf("expected head ref fetch to succeed, got %v", err)
}

if headRef != "refs/prr/pull/42/head" {
t.Fatalf("expected head ref namespace, got %q", headRef)
}

if len(runner.commands) != 1 {
t.Fatalf("expected one fetch command, got %d", len(runner.commands))
}

command := strings.Join(runner.commands[0], " ")
if !strings.Contains(command, "fetch origin pull/42/head:refs/prr/pull/42/head") {
t.Fatalf("expected fetch destination to target PRR namespace, got %q", command)
}
}

func TestFetchPRHeadRefClassifiesErrorsAsProviderFailures(t *testing.T) {
runner := &recorderRunner{err: errors.New("no head ref")}
service := NewServiceWithCacheDir(runner, t.TempDir())

_, err := service.FetchPRHeadRef(context.Background(), "/tmp/repo.git", "origin", 99, EnsureOptions{})
if err == nil {
t.Fatalf("expected provider-classified error")
}

if !strings.Contains(err.Error(), "PROVIDER_RESOLUTION") {
t.Fatalf("expected provider-classified error, got %v", err)
}
}

func TestResolveMergeBaseIssuesMergeBaseCommand(t *testing.T) {
runner := stubRunner{runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
joined := strings.Join(args, " ")
if strings.Contains(joined, "merge-base") {
return "abc1234def5678\n", nil
}
return "", nil
}}

service := NewServiceWithCacheDir(runner, t.TempDir())

base, err := service.ResolveMergeBase(context.Background(), "/tmp/repo.git", "refs/prr/pull/5/head", "HEAD", EnsureOptions{})
if err != nil {
t.Fatalf("expected merge base resolution to succeed, got %v", err)
}

if base != "abc1234def5678" {
t.Fatalf("expected trimmed merge base SHA, got %q", base)
}
}
