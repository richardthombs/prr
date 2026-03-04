package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/git"
)

func TestWorktreeCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "worktree" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected worktree command to be registered")
	}
}

func TestWorktreeAddEmitsJSONWithWorkDir(t *testing.T) {
	resetMirrorPRRefFlagState(t)
	resetWorktreeFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		if name != "git" {
			t.Fatalf("unexpected command name %q", name)
		}

		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "worktree add --detach") {
			t.Fatalf("expected detached worktree add command, got %q", joined)
		}
		if !strings.Contains(joined, "refs/prr/pull/505/merge") {
			t.Fatalf("expected merge ref in command, got %q", joined)
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
	rootCmd.SetArgs([]string{"worktree", "add", "--pr-id", "505", "--bare-dir", "/tmp/mirror.git"})

	if err := Execute(); err != nil {
		t.Fatalf("expected worktree add to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["prId"] != float64(505) {
		t.Fatalf("expected prId 505, got %#v", payload["prId"])
	}
	if payload["mergeRef"] != "refs/prr/pull/505/merge" {
		t.Fatalf("expected default merge ref for PR 505, got %#v", payload["mergeRef"])
	}
	if payload["workDir"] == "" {
		t.Fatalf("expected workDir in output payload")
	}
}

func TestWorktreeAddUsesStdinComposePayload(t *testing.T) {
	resetMirrorPRRefFlagState(t)
	resetWorktreeFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		if name != "git" {
			t.Fatalf("unexpected command name %q", name)
		}

		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "-C /tmp/stdin-mirror.git worktree add --detach") {
			t.Fatalf("expected stdin bare-dir to be used, got %q", joined)
		}
		if !strings.Contains(joined, "refs/prr/pull/606/merge") {
			t.Fatalf("expected stdin merge ref to be used, got %q", joined)
		}

		return "", nil
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service {
		return service
	}

	stdout := &bytes.Buffer{}
	stdin := bytes.NewBufferString(`{"prId":606,"repoUrl":"https://example.test/stdin/repo","remote":"origin","provider":"github","bareDir":"/tmp/stdin-mirror.git","mergeRef":"refs/prr/pull/606/merge"}`)
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs([]string{"worktree", "add"})

	if err := Execute(); err != nil {
		t.Fatalf("expected worktree add from stdin payload to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if payload["prId"] != float64(606) {
		t.Fatalf("expected prId from stdin payload, got %#v", payload["prId"])
	}
	if payload["bareDir"] != "/tmp/stdin-mirror.git" {
		t.Fatalf("expected bareDir from stdin payload, got %#v", payload["bareDir"])
	}
	if payload["mergeRef"] != "refs/prr/pull/606/merge" {
		t.Fatalf("expected mergeRef from stdin payload, got %#v", payload["mergeRef"])
	}
}

func TestWorktreeAddWhatIfLogsAndSkipsExecution(t *testing.T) {
	resetMirrorPRRefFlagState(t)
	resetWorktreeFlagState(t)

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
	rootCmd.SetArgs([]string{"worktree", "add", "--pr-id", "707", "--bare-dir", "/tmp/mirror.git", "--what-if"})

	if err := Execute(); err != nil {
		t.Fatalf("expected worktree add what-if to succeed, got %v", err)
	}

	if !strings.Contains(stderr.String(), "exec: git -C /tmp/mirror.git worktree add --detach") {
		t.Fatalf("expected planned worktree add command in stderr, got %q", stderr.String())
	}
}

func resetWorktreeFlagState(t *testing.T) {
	t.Helper()

	if err := worktreeAddCmd.Flags().Set("pr-id", "0"); err != nil {
		t.Fatalf("failed to reset worktree --pr-id flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("pr-id").Changed = false
	if err := worktreeAddCmd.Flags().Set("repo", ""); err != nil {
		t.Fatalf("failed to reset worktree --repo flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("repo").Changed = false
	if err := worktreeAddCmd.Flags().Set("remote", "origin"); err != nil {
		t.Fatalf("failed to reset worktree --remote flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("remote").Changed = false
	if err := worktreeAddCmd.Flags().Set("provider", ""); err != nil {
		t.Fatalf("failed to reset worktree --provider flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("provider").Changed = false
	if err := worktreeAddCmd.Flags().Set("bare-dir", ""); err != nil {
		t.Fatalf("failed to reset worktree --bare-dir flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("bare-dir").Changed = false
	if err := worktreeAddCmd.Flags().Set("merge-ref", ""); err != nil {
		t.Fatalf("failed to reset worktree --merge-ref flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("merge-ref").Changed = false
	if err := worktreeAddCmd.Flags().Set("keep", "false"); err != nil {
		t.Fatalf("failed to reset worktree --keep flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("keep").Changed = false
	if err := worktreeAddCmd.Flags().Set("verbose", "false"); err != nil {
		t.Fatalf("failed to reset worktree --verbose flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("verbose").Changed = false
	if err := worktreeAddCmd.Flags().Set("what-if", "false"); err != nil {
		t.Fatalf("failed to reset worktree --what-if flag: %v", err)
	}
	worktreeAddCmd.Flags().Lookup("what-if").Changed = false
}
