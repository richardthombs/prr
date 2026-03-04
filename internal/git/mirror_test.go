package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("failed to hold lock for timeout test: %v", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

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

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("failed to hold lock for force test: %v", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

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
