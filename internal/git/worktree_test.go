package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	apperrors "github.com/richardthombs/prr/internal/errors"
)

func TestResolveWorktreeDirFromBareDirUsesDeterministicLayout(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	bareDir := filepath.Join(t.TempDir(), "abcdef123456.git")
	workDir, err := service.ResolveWorktreeDirFromBareDir(bareDir, 42)
	if err != nil {
		t.Fatalf("expected worktree path resolution to succeed, got %v", err)
	}

	repoHash := strings.TrimSuffix(filepath.Base(bareDir), ".git")
	if filepath.Base(filepath.Dir(workDir)) != repoHash {
		t.Fatalf("expected worktree repo segment %q, got %q", repoHash, filepath.Base(filepath.Dir(workDir)))
	}
	if filepath.Base(workDir) != "pr-42" {
		t.Fatalf("expected worktree leaf segment pr-42, got %q", filepath.Base(workDir))
	}
}

func TestCreateWorktreeInvokesDetachedGitCommand(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())
	bareDir := filepath.Join(t.TempDir(), "repo.git")
	workDir := filepath.Join(t.TempDir(), "work", "77")

	err := service.CreateWorktree(context.Background(), bareDir, "refs/prr/pull/77/merge", workDir, EnsureOptions{})
	if err != nil {
		t.Fatalf("expected create worktree to succeed, got %v", err)
	}

	if len(runner.commands) != 1 {
		t.Fatalf("expected one command, got %d", len(runner.commands))
	}

	command := strings.Join(runner.commands[0], " ")
	expected := "git -C " + bareDir + " worktree add --detach " + workDir + " refs/prr/pull/77/merge"
	if !strings.Contains(command, expected) {
		t.Fatalf("expected detached worktree add command, got %q", command)
	}
}

func TestCreateWorktreeWhatIfLogsAndSkipsExecution(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())
	bareDir := filepath.Join(t.TempDir(), "repo.git")
	workDir := filepath.Join(t.TempDir(), "work", "88")

	logs := make([]string, 0)
	err := service.CreateWorktree(context.Background(), bareDir, "refs/prr/pull/88/merge", workDir, EnsureOptions{
		Verbose: true,
		WhatIf:  true,
		Logger: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatalf("expected what-if create worktree to succeed, got %v", err)
	}

	if len(logs) == 0 {
		t.Fatalf("expected logged command in what-if mode")
	}
	if !strings.Contains(logs[0], "exec: git -C "+bareDir+" worktree add --detach "+workDir+" refs/prr/pull/88/merge") {
		t.Fatalf("expected logged detached add command, got %q", logs[0])
	}
	if len(runner.commands) != 0 {
		t.Fatalf("expected no external command execution in what-if mode, got %d", len(runner.commands))
	}
}

func TestCreateWorktreeExistingPathResetsToMergeRef(t *testing.T) {
	commands := make([][]string, 0)
	runner := stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		recorded := append([]string{name}, args...)
		commands = append(commands, recorded)

		joined := strings.Join(recorded, " ")
		if strings.Contains(joined, "rev-parse --is-inside-work-tree") {
			return "true", nil
		}

		return "", nil
	}}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	workDir := t.TempDir()
	bareDir := filepath.Join(t.TempDir(), "repo.git")
	err := service.CreateWorktree(context.Background(), bareDir, "refs/prr/pull/188/merge", workDir, EnsureOptions{})
	if err != nil {
		t.Fatalf("expected existing worktree reset to succeed, got %v", err)
	}

	if len(commands) != 2 {
		t.Fatalf("expected rev-parse probe and reset command, got %d", len(commands))
	}

	probeCommand := strings.Join(commands[0], " ")
	if !strings.Contains(probeCommand, "git -C "+workDir+" rev-parse --is-inside-work-tree") {
		t.Fatalf("expected worktree validity probe command first, got %q", probeCommand)
	}

	command := strings.Join(commands[1], " ")
	if !strings.Contains(command, "git -C "+workDir+" reset --hard refs/prr/pull/188/merge") {
		t.Fatalf("expected reset command against existing worktree, got %q", command)
	}
	if strings.Contains(command, "worktree add --detach") {
		t.Fatalf("did not expect worktree add command for existing worktree, got %q", command)
	}
}

func TestCreateWorktreeExistingPathWhatIfLogsResetAndSkipsExecution(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	workDir := t.TempDir()
	bareDir := filepath.Join(t.TempDir(), "repo.git")
	logs := make([]string, 0)
	err := service.CreateWorktree(context.Background(), bareDir, "refs/prr/pull/199/merge", workDir, EnsureOptions{
		Verbose: true,
		WhatIf:  true,
		Logger: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatalf("expected what-if existing worktree reset to succeed, got %v", err)
	}

	if len(logs) == 0 {
		t.Fatalf("expected logged reset command in what-if mode")
	}
	if !strings.Contains(logs[0], "exec: git -C "+workDir+" reset --hard refs/prr/pull/199/merge") {
		t.Fatalf("expected reset command log, got %q", logs[0])
	}
	if len(runner.commands) != 0 {
		t.Fatalf("expected no external command execution in what-if mode, got %d", len(runner.commands))
	}
}

func TestCreateWorktreeExistingInvalidDirectoryFallsBackToAdd(t *testing.T) {
	commands := make([][]string, 0)
	runner := stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		recorded := append([]string{name}, args...)
		commands = append(commands, recorded)

		joined := strings.Join(recorded, " ")
		if strings.Contains(joined, "rev-parse --is-inside-work-tree") {
			return "", errors.New("fatal: not a git repository")
		}

		return "", nil
	}}

	service := NewServiceWithCacheDir(runner, t.TempDir())
	bareDir := filepath.Join(t.TempDir(), "repo.git")

	workDir := filepath.Join(t.TempDir(), "pr-73")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("failed to create existing directory fixture: %v", err)
	}

	err := service.CreateWorktree(context.Background(), bareDir, "refs/prr/pull/73/merge", workDir, EnsureOptions{})
	if err != nil {
		t.Fatalf("expected fallback recreate to succeed, got %v", err)
	}

	if len(commands) < 3 {
		t.Fatalf("expected probe, prune, and add commands, got %d", len(commands))
	}

	foundPrune := false
	foundAdd := false
	for _, commandParts := range commands {
		joined := strings.Join(commandParts, " ")
		if strings.Contains(joined, "git -C "+bareDir+" worktree prune") {
			foundPrune = true
		}
		if strings.Contains(joined, "git -C "+bareDir+" worktree add --detach "+workDir+" refs/prr/pull/73/merge") {
			foundAdd = true
		}
	}

	if !foundPrune {
		t.Fatalf("expected prune command during fallback recreate, got %#v", commands)
	}
	if !foundAdd {
		t.Fatalf("expected detached add command during fallback recreate, got %#v", commands)
	}
}

func TestCreateWorktreeExistingPathNonDirectoryFails(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	nonDirectoryPath := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(nonDirectoryPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create non-directory path fixture: %v", err)
	}

	bareDir := filepath.Join(t.TempDir(), "repo.git")
	err := service.CreateWorktree(context.Background(), bareDir, "refs/prr/pull/200/merge", nonDirectoryPath, EnsureOptions{})
	if err == nil {
		t.Fatalf("expected runtime error when worktree path exists as non-directory")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassRuntime {
		t.Fatalf("expected runtime class error, got %s", appErr.Class)
	}
}

type stubRunner struct {
	runFunc func(ctx context.Context, name string, args ...string) (string, error)
}

func (r stubRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	if r.runFunc == nil {
		return "", nil
	}

	return r.runFunc(ctx, name, args...)
}

func TestCleanupWorktreeInvokesRemoveAndPrune(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())
	bareDir := filepath.Join(t.TempDir(), "repo.git")
	workDir := filepath.Join(t.TempDir(), "work", "99")

	err := service.CleanupWorktree(context.Background(), bareDir, workDir, EnsureOptions{})
	if err != nil {
		t.Fatalf("expected cleanup to succeed, got %v", err)
	}

	if len(runner.commands) != 2 {
		t.Fatalf("expected two cleanup commands, got %d", len(runner.commands))
	}

	first := strings.Join(runner.commands[0], " ")
	if !strings.Contains(first, "git -C "+bareDir+" worktree remove --force "+workDir) {
		t.Fatalf("expected worktree remove command first, got %q", first)
	}

	second := strings.Join(runner.commands[1], " ")
	if !strings.Contains(second, "git -C "+bareDir+" worktree prune") {
		t.Fatalf("expected worktree prune command second, got %q", second)
	}
}

func TestCleanupWorktreeWhatIfLogsAndSkipsExecution(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())
	bareDir := filepath.Join(t.TempDir(), "repo.git")
	workDir := filepath.Join(t.TempDir(), "work", "100")

	logs := make([]string, 0)
	err := service.CleanupWorktree(context.Background(), bareDir, workDir, EnsureOptions{
		Verbose: true,
		WhatIf:  true,
		Logger: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatalf("expected what-if cleanup to succeed, got %v", err)
	}

	if len(logs) < 2 {
		t.Fatalf("expected planned remove+prune logs in what-if mode, got %d log(s)", len(logs))
	}
	if !strings.Contains(logs[0], "exec: git -C "+bareDir+" worktree remove --force "+workDir) {
		t.Fatalf("expected remove command log, got %q", logs[0])
	}
	if !strings.Contains(logs[1], "exec: git -C "+bareDir+" worktree prune") {
		t.Fatalf("expected prune command log, got %q", logs[1])
	}
	if len(runner.commands) != 0 {
		t.Fatalf("expected no external command execution in what-if mode, got %d", len(runner.commands))
	}
}

func TestCleanupWorktreeClassifiesRemoveFailureAsRuntime(t *testing.T) {
	runner := &recorderRunner{err: errors.New("remove failed")}
	service := NewServiceWithCacheDir(runner, t.TempDir())
	bareDir := filepath.Join(t.TempDir(), "repo.git")
	workDir := filepath.Join(t.TempDir(), "work", "101")

	err := service.CleanupWorktree(context.Background(), bareDir, workDir, EnsureOptions{})
	if err == nil {
		t.Fatalf("expected runtime error when remove fails")
	}

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Class != apperrors.ClassRuntime {
		t.Fatalf("expected runtime class error, got %s", appErr.Class)
	}
}
