package git

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDiffContributionUsesExpectedGitCommands(t *testing.T) {
	commands := make([][]string, 0)
	runner := stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		recorded := append([]string{name}, args...)
		commands = append(commands, recorded)

		joined := strings.Join(recorded, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "b.txt\na.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return " a.txt | 1 +\n 1 file changed, 1 insertion(+)", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt\n+hello", nil
		default:
			return "", nil
		}
	}}

	service := NewServiceWithCacheDir(runner, t.TempDir())
	output, err := service.DiffContribution(context.Background(), "/tmp/work/12")
	if err != nil {
		t.Fatalf("expected diff contribution to succeed, got %v", err)
	}

	if output.Range != "HEAD^1..HEAD" {
		t.Fatalf("expected range HEAD^1..HEAD, got %q", output.Range)
	}
	if len(output.Files) != 2 || output.Files[0] != "a.txt" || output.Files[1] != "b.txt" {
		t.Fatalf("expected sorted changed files, got %#v", output.Files)
	}
	if output.WorkDir != "/tmp/work/12" {
		t.Fatalf("expected workDir passthrough, got %q", output.WorkDir)
	}

	if len(commands) != 3 {
		t.Fatalf("expected three git diff commands, got %d", len(commands))
	}
}

func TestDiffContributionWhatIfLogsAndSkipsExecution(t *testing.T) {
	runner := &recorderRunner{}
	service := NewServiceWithCacheDir(runner, t.TempDir())

	logs := make([]string, 0)
	output, err := service.DiffContributionWithOptions(context.Background(), "/tmp/work/20", EnsureOptions{
		Verbose: true,
		WhatIf:  true,
		Logger: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatalf("expected what-if diff contribution to succeed, got %v", err)
	}

	if output.Range != "HEAD^1..HEAD" {
		t.Fatalf("expected range in what-if output, got %q", output.Range)
	}
	if len(logs) != 3 {
		t.Fatalf("expected three logged commands in what-if mode, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "exec: git -C /tmp/work/20 diff --name-only HEAD^1..HEAD") {
		t.Fatalf("expected name-only command log, got %q", logs[0])
	}
	if len(runner.commands) != 0 {
		t.Fatalf("expected no external command execution in what-if mode, got %d", len(runner.commands))
	}
}

func TestDiffContributionDeterministicAcrossReruns(t *testing.T) {
	runner := stubRunner{runFunc: func(_ context.Context, _ string, args ...string) (string, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "b.txt\na.txt\na.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "2 files changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt\n+hello", nil
		default:
			return "", nil
		}
	}}

	service := NewServiceWithCacheDir(runner, t.TempDir())
	first, err := service.DiffContribution(context.Background(), "/tmp/work/12")
	if err != nil {
		t.Fatalf("expected first diff run to succeed, got %v", err)
	}
	second, err := service.DiffContribution(context.Background(), "/tmp/work/12")
	if err != nil {
		t.Fatalf("expected second diff run to succeed, got %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic diff outputs across reruns, first=%#v second=%#v", first, second)
	}
}
