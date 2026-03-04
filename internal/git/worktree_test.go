package git

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestResolveWorktreeDirFromBareDirUsesDeterministicLayout(t *testing.T) {
	service := NewServiceWithCacheDir(&recorderRunner{}, t.TempDir())

	workDir, err := service.ResolveWorktreeDirFromBareDir("/tmp/abcdef123456.git", 42)
	if err != nil {
		t.Fatalf("expected worktree path resolution to succeed, got %v", err)
	}

	if !strings.Contains(workDir, "/prr/work/") {
		t.Fatalf("expected worktree path under prr/work cache root, got %q", workDir)
	}
	if !strings.Contains(workDir, "/abcdef123456/pr-42/") {
		t.Fatalf("expected repo hash and pr id segments in path, got %q", workDir)
	}
}

func TestCreateWorktreeInvokesDetachedGitCommand(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	err := service.CreateWorktree(context.Background(), "/tmp/repo.git", "refs/prr/pull/77/merge", "/tmp/work/77", EnsureOptions{})
	if err != nil {
		t.Fatalf("expected create worktree to succeed, got %v", err)
	}

	if len(runner.commands) != 1 {
		t.Fatalf("expected one command, got %d", len(runner.commands))
	}

	command := strings.Join(runner.commands[0], " ")
	expected := "git -C /tmp/repo.git worktree add --detach /tmp/work/77 refs/prr/pull/77/merge"
	if !strings.Contains(command, expected) {
		t.Fatalf("expected detached worktree add command, got %q", command)
	}
}

func TestCreateWorktreeWhatIfLogsAndSkipsExecution(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	logs := make([]string, 0)
	err := service.CreateWorktree(context.Background(), "/tmp/repo.git", "refs/prr/pull/88/merge", "/tmp/work/88", EnsureOptions{
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
	if !strings.Contains(logs[0], "exec: git -C /tmp/repo.git worktree add --detach /tmp/work/88 refs/prr/pull/88/merge") {
		t.Fatalf("expected logged detached add command, got %q", logs[0])
	}
	if len(runner.commands) != 0 {
		t.Fatalf("expected no external command execution in what-if mode, got %d", len(runner.commands))
	}
}

func TestCleanupWorktreeInvokesRemoveAndPrune(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	err := service.CleanupWorktree(context.Background(), "/tmp/repo.git", "/tmp/work/99", EnsureOptions{})
	if err != nil {
		t.Fatalf("expected cleanup to succeed, got %v", err)
	}

	if len(runner.commands) != 2 {
		t.Fatalf("expected two cleanup commands, got %d", len(runner.commands))
	}

	first := strings.Join(runner.commands[0], " ")
	if !strings.Contains(first, "git -C /tmp/repo.git worktree remove --force /tmp/work/99") {
		t.Fatalf("expected worktree remove command first, got %q", first)
	}

	second := strings.Join(runner.commands[1], " ")
	if !strings.Contains(second, "git -C /tmp/repo.git worktree prune") {
		t.Fatalf("expected worktree prune command second, got %q", second)
	}
}
