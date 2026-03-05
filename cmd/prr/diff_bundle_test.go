package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/git"
)

func TestDiffCommandEmitsJSONFromStdinWorkDir(t *testing.T) {
	resetDiffFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	service := git.NewServiceWithCacheDir(stubRunner{runFunc: func(_ context.Context, name string, args ...string) (string, error) {
		if name != "git" {
			t.Fatalf("unexpected command name %q", name)
		}

		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "diff --name-only HEAD^1..HEAD"):
			return "b.txt\na.txt\n", nil
		case strings.Contains(joined, "diff --stat HEAD^1..HEAD"):
			return "2 files changed", nil
		case strings.Contains(joined, "diff --patch --binary HEAD^1..HEAD"):
			return "diff --git a/a.txt b/a.txt", nil
		default:
			return "", nil
		}
	}}, t.TempDir())

	mirrorServiceFactory = func() *git.Service { return service }

	stdin := bytes.NewBufferString(`{"prId":12,"repoUrl":"https://github.com/acme/repo","remote":"origin","provider":"github","bareDir":"/tmp/bare","mergeRef":"refs/prr/pull/12/merge","workDir":"/tmp/work/12"}`)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	diffCmd.SetIn(stdin)
	diffCmd.SetOut(stdout)
	diffCmd.SetErr(stderr)

	if err := diffCmd.RunE(diffCmd, nil); err != nil {
		t.Fatalf("expected diff command to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid diff JSON payload, got %v", err)
	}

	if payload["range"] != "HEAD^1..HEAD" {
		t.Fatalf("expected range in payload, got %#v", payload["range"])
	}
	if payload["workDir"] != "/tmp/work/12" {
		t.Fatalf("expected workDir from stdin, got %#v", payload["workDir"])
	}
	if payload["repoUrl"] != "https://github.com/acme/repo" {
		t.Fatalf("expected repoUrl passthrough, got %#v", payload["repoUrl"])
	}
}

func TestDiffCommandWhatIfLogsAndSkipsExecution(t *testing.T) {
	resetDiffFlagState(t)

	originalFactory := mirrorServiceFactory
	t.Cleanup(func() {
		mirrorServiceFactory = originalFactory
	})

	runner := stubRunner{runFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
		t.Fatalf("expected no external command execution in what-if mode")
		return "", nil
	}}
	service := git.NewServiceWithCacheDir(runner, t.TempDir())
	mirrorServiceFactory = func() *git.Service { return service }

	stdin := bytes.NewBufferString(`{"workDir":"/tmp/work/99"}`)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	diffCmd.SetIn(stdin)
	diffCmd.SetOut(stdout)
	diffCmd.SetErr(stderr)
	if err := diffCmd.Flags().Set("what-if", "true"); err != nil {
		t.Fatalf("failed setting diff --what-if: %v", err)
	}
	if err := diffCmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("failed setting diff --verbose: %v", err)
	}

	if err := diffCmd.RunE(diffCmd, nil); err != nil {
		t.Fatalf("expected diff what-if command to succeed, got %v", err)
	}

	if !strings.Contains(stderr.String(), "exec: git -C /tmp/work/99 diff --name-only HEAD^1..HEAD") {
		t.Fatalf("expected command preview in stderr, got %q", stderr.String())
	}
}

func TestBundleCommandBuildsV1Payload(t *testing.T) {
	resetBundleFlagState(t)

	stdin := bytes.NewBufferString(`{"prId":12,"repoUrl":"https://github.com/acme/repo","remote":"origin","provider":"github","mergeRef":"refs/prr/pull/12/merge","range":"HEAD^1..HEAD","files":["a.txt"],"stat":"1 file changed","patch":"diff --git a/a.txt b/a.txt"}`)
	stdout := &bytes.Buffer{}
	bundleCmd.SetIn(stdin)
	bundleCmd.SetOut(stdout)
	bundleCmd.SetErr(&bytes.Buffer{})

	if err := bundleCmd.RunE(bundleCmd, nil); err != nil {
		t.Fatalf("expected bundle command to succeed, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &payload); err != nil {
		t.Fatalf("expected valid bundle JSON payload, got %v", err)
	}

	if payload["version"] != "v1" {
		t.Fatalf("expected v1 bundle version, got %#v", payload["version"])
	}
	if payload["changedFiles"] != float64(1) {
		t.Fatalf("expected changedFiles=1, got %#v", payload["changedFiles"])
	}
}

func TestBundleCommandWhatIfAndVerboseEmitDiagnostics(t *testing.T) {
	resetBundleFlagState(t)

	stdin := bytes.NewBufferString(`{"range":"HEAD^1..HEAD","files":["a.txt"],"stat":"1 file changed","patch":"diff --git"}`)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	bundleCmd.SetIn(stdin)
	bundleCmd.SetOut(stdout)
	bundleCmd.SetErr(stderr)
	if err := bundleCmd.Flags().Set("what-if", "true"); err != nil {
		t.Fatalf("failed setting bundle --what-if: %v", err)
	}
	if err := bundleCmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("failed setting bundle --verbose: %v", err)
	}

	if err := bundleCmd.RunE(bundleCmd, nil); err != nil {
		t.Fatalf("expected bundle what-if command to succeed, got %v", err)
	}

	if !strings.Contains(stderr.String(), "what-if: bundle stage uses no external commands") {
		t.Fatalf("expected what-if diagnostics in stderr, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "bundle: validate diff input and build v1 payload") {
		t.Fatalf("expected verbose diagnostics in stderr, got %q", stderr.String())
	}
}

func TestBundleCommandFailsWhenLimitExceeded(t *testing.T) {
	resetBundleFlagState(t)

	stdin := bytes.NewBufferString(`{"range":"HEAD^1..HEAD","files":["a.txt","b.txt"],"stat":"2 files changed","patch":"diff --git"}`)
	bundleCmd.SetIn(stdin)
	bundleCmd.SetOut(&bytes.Buffer{})
	bundleCmd.SetErr(&bytes.Buffer{})
	if err := bundleCmd.Flags().Set("max-files", "1"); err != nil {
		t.Fatalf("failed setting bundle --max-files: %v", err)
	}

	err := bundleCmd.RunE(bundleCmd, nil)
	if err == nil {
		t.Fatalf("expected bundle command to fail on changed file limit")
	}
	if !strings.Contains(err.Error(), "LIMIT_EXCEEDED") {
		t.Fatalf("expected LIMIT_EXCEEDED diagnostic, got %v", err)
	}
}

func resetDiffFlagState(t *testing.T) {
	t.Helper()

	if err := diffCmd.Flags().Set("work-dir", ""); err != nil {
		t.Fatalf("failed to reset diff --work-dir flag: %v", err)
	}
	diffCmd.Flags().Lookup("work-dir").Changed = false
	if err := diffCmd.Flags().Set("verbose", "false"); err != nil {
		t.Fatalf("failed to reset diff --verbose flag: %v", err)
	}
	diffCmd.Flags().Lookup("verbose").Changed = false
	if err := diffCmd.Flags().Set("what-if", "false"); err != nil {
		t.Fatalf("failed to reset diff --what-if flag: %v", err)
	}
	diffCmd.Flags().Lookup("what-if").Changed = false
}

func resetBundleFlagState(t *testing.T) {
	t.Helper()

	for _, flag := range []struct {
		name  string
		value string
	}{
		{name: "verbose", value: "false"},
		{name: "what-if", value: "false"},
		{name: "max-patch-bytes", value: "0"},
		{name: "max-files", value: "0"},
		{name: "pr-id", value: "0"},
		{name: "repo", value: ""},
		{name: "remote", value: ""},
		{name: "provider", value: ""},
		{name: "merge-ref", value: ""},
	} {
		if err := bundleCmd.Flags().Set(flag.name, flag.value); err != nil {
			t.Fatalf("failed to reset bundle --%s flag: %v", flag.name, err)
		}
		bundleCmd.Flags().Lookup(flag.name).Changed = false
	}
}
