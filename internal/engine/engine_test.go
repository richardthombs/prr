package engine

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/richardthombs/prr/internal/types"
)

type fakeRunner struct {
	run func(ctx context.Context, command string, args []string, cwd string, stdinPayload string) (commandResult, error)
}

func (f fakeRunner) Run(ctx context.Context, command string, args []string, cwd string, stdinPayload string) (commandResult, error) {
	if f.run == nil {
		return commandResult{}, nil
	}

	return f.run(ctx, command, args, cwd, stdinPayload)
}

func TestCLIAdapterBuildsCommandWithStdinEnvelopeModelAndWorkDir(t *testing.T) {
	cfg := DefaultAgentConfig()
	capturedCommand := ""
	capturedArgs := []string{}
	capturedCwd := ""
	capturedStdin := ""

	adapter := &CLIAgentAdapter{
		config: cfg,
		runner: fakeRunner{run: func(_ context.Context, command string, args []string, cwd string, stdinPayload string) (commandResult, error) {
			capturedCommand = command
			capturedArgs = append([]string{}, args...)
			capturedCwd = cwd
			capturedStdin = stdinPayload
			return commandResult{Stdout: `{"summary":"ok","risk":{"score":0.1,"reasons":["low"]},"findings":[],"checklist":["run tests"]}`}, nil
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{"a.go"}, Stat: "1 file changed", Patch: "diff --git", PRID: 7},
		WorkDir: "/tmp/work",
		Model:   "gpt-5",
	})
	if err != nil {
		t.Fatalf("expected review to succeed, got %v", err)
	}

	if capturedCommand != "copilot" {
		t.Fatalf("expected command copilot, got %q", capturedCommand)
	}
	joined := strings.Join(capturedArgs, " ")
	if strings.Contains(joined, "-p") {
		t.Fatalf("did not expect -p prompt argument, got %q", joined)
	}
	if !strings.Contains(joined, "--model gpt-5") {
		t.Fatalf("expected model pass-through, got %q", joined)
	}
	if capturedCwd != "/tmp/work" {
		t.Fatalf("expected cwd /tmp/work, got %q", capturedCwd)
	}
	if !strings.Contains(capturedStdin, "INSTRUCTIONS") || !strings.Contains(capturedStdin, "DIFF_BUNDLE_JSON_START") || !strings.Contains(capturedStdin, "DIFF_BUNDLE_JSON_END") {
		t.Fatalf("expected stdin envelope markers, got %q", capturedStdin)
	}
	if !strings.Contains(capturedStdin, `Return ONLY valid JSON`) {
		t.Fatalf("expected schema instructions inside stdin payload, got %q", capturedStdin)
	}
	if !strings.Contains(capturedStdin, "risk.score MUST be a decimal number between 0 and 1 inclusive") {
		t.Fatalf("expected explicit risk score range instructions, got %q", capturedStdin)
	}
	if !strings.Contains(capturedStdin, defaultReviewInstructions) {
		t.Fatalf("expected default review instructions in stdin payload, got %q", capturedStdin)
	}
	if !strings.Contains(capturedStdin, `"version":"v1"`) || !strings.Contains(capturedStdin, `"patch":"diff --git"`) {
		t.Fatalf("expected bundle JSON on stdin, got %q", capturedStdin)
	}
}

func TestCLIAdapterWhatIfSkipsExecutionAndReturnsStructuredReview(t *testing.T) {
	called := false
	logs := []string{}

	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
			called = true
			return commandResult{}, nil
		}},
	}

	review, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok", PRID: 11},
		WorkDir: "/tmp/work",
		WhatIf:  true,
		Logger: func(format string, _ ...any) {
			logs = append(logs, format)
		},
	})
	if err != nil {
		t.Fatalf("expected what-if review to succeed, got %v", err)
	}
	if called {
		t.Fatalf("expected no external execution in what-if mode")
	}
	if review.Summary == "" || review.Risk.Reasons == nil || review.Checklist == nil {
		t.Fatalf("expected structured review response in what-if mode")
	}
	if len(logs) == 0 {
		t.Fatalf("expected diagnostics logs in what-if mode")
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "review engine input envelope") {
		t.Fatalf("expected what-if log to include envelope details, got %q", joined)
	}
}

func TestCLIAdapterParsesMixedTextAndJSONOutput(t *testing.T) {
	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
			return commandResult{Stdout: "info line\n{\"summary\":\"ok\",\"risk\":{\"score\":0.1,\"reasons\":[\"low\"]},\"findings\":[],\"checklist\":[\"run tests\"]}\ntrailing"}, nil
		}},
	}

	review, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir: "/tmp/work",
	})
	if err != nil {
		t.Fatalf("expected parser to extract JSON object, got %v", err)
	}
	if review.Summary != "ok" {
		t.Fatalf("expected parsed summary 'ok', got %q", review.Summary)
	}
}

func TestCLIAdapterRejectsOutOfRangeRiskScore(t *testing.T) {
	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
			return commandResult{Stdout: `{"summary":"ok","risk":{"score":4,"reasons":["high churn"]},"findings":[{"id":"F001","file":"a.go","line":7,"severity":"important","category":"tests","message":"m","suggestion":"s"}],"checklist":["run tests"]}`}, nil
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir: "/tmp/work",
	})
	if err == nil {
		t.Fatalf("expected risk score validation failure")
	}
	if !strings.Contains(err.Error(), "agent output failed review schema validation") {
		t.Fatalf("expected schema validation classification, got %v", err)
	}
}

func TestDefaultAgentConfigHonorsEnvironmentOverrides(t *testing.T) {
	t.Setenv("PRR_AGENT_COMMAND", "copilot-dev")
	t.Setenv("PRR_AGENT_ARGS", "--foo --bar")
	t.Setenv("PRR_AGENT_MODEL_ARG", "--model-name")
	t.Setenv("PRR_AGENT_INPUT_MODE", "file")
	t.Setenv("PRR_AGENT_OUTPUT_MODE", "json")
	t.Setenv("PRR_AGENT_TIMEOUT_SECONDS", "90")

	cfg := DefaultAgentConfig()
	if cfg.Command != "copilot-dev" {
		t.Fatalf("expected command override, got %q", cfg.Command)
	}
	if strings.Join(cfg.Args, " ") != "--foo --bar" {
		t.Fatalf("expected args override, got %q", strings.Join(cfg.Args, " "))
	}
	if cfg.ModelArg != "--model-name" {
		t.Fatalf("expected model arg override, got %q", cfg.ModelArg)
	}
	if cfg.InputMode != "file" {
		t.Fatalf("expected input mode override, got %q", cfg.InputMode)
	}
	if cfg.OutputMode != "json" {
		t.Fatalf("expected output mode override, got %q", cfg.OutputMode)
	}
	if cfg.TimeoutSeconds != 90 {
		t.Fatalf("expected timeout override, got %d", cfg.TimeoutSeconds)
	}
}

func TestCLIAdapterVerboseLogsCopilotOutput(t *testing.T) {
	logs := []string{}
	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
			return commandResult{
				Stdout: "{\"summary\":\"ok\",\"risk\":{\"score\":0.1,\"reasons\":[\"low\"]},\"findings\":[],\"checklist\":[\"run tests\"]}",
				Stderr: "copilot diagnostic",
			}, nil
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir: "/tmp/work",
		Verbose: true,
		Logger: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "copilot stdout") {
		t.Fatalf("expected verbose stdout log, got %q", joined)
	}
	if !strings.Contains(joined, "copilot stderr") {
		t.Fatalf("expected verbose stderr log, got %q", joined)
	}
	if !strings.Contains(joined, "copilot diagnostic") {
		t.Fatalf("expected stderr content in verbose logs, got %q", joined)
	}
}

func TestCLIAdapterClassifiesMissingBinaryAndTimeout(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{name: "missing binary", err: exec.ErrNotFound, wantMsg: "agent command not found"},
		{name: "timeout", err: context.DeadlineExceeded, wantMsg: "agent command timed out"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := &CLIAgentAdapter{
				config: DefaultAgentConfig(),
				runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
					return commandResult{}, tc.err
				}},
			}

			_, err := adapter.Review(context.Background(), ReviewInput{
				Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
				WorkDir: "/tmp/work",
			})
			if err == nil {
				t.Fatalf("expected failure")
			}
			if !strings.Contains(err.Error(), "ENGINE_FAILURE") || !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("expected engine-classed %q error, got %v", tc.wantMsg, err)
			}
		})
	}
}

func TestCLIAdapterClassifiesNonZeroExit(t *testing.T) {
	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
			return commandResult{Stderr: "unknown flag: --json"}, &exec.ExitError{}
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir: "/tmp/work",
	})
	if err == nil {
		t.Fatalf("expected failure")
	}
	if !strings.Contains(err.Error(), "agent command failed with non-zero exit") {
		t.Fatalf("expected non-zero exit classification, got %v", err)
	}
	if !strings.Contains(err.Error(), "unsupported copilot CLI invocation") {
		t.Fatalf("expected actionable invocation diagnostic, got %v", err)
	}
}

func TestCLIAdapterNonZeroExitAuthDiagnostic(t *testing.T) {
	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
			return commandResult{Stderr: "not logged in"}, &exec.ExitError{}
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir: "/tmp/work",
	})
	if err == nil {
		t.Fatalf("expected failure")
	}
	if !strings.Contains(err.Error(), "copilot authentication missing") {
		t.Fatalf("expected auth diagnostic, got %v", err)
	}
}

func TestCLIAdapterRejectsInvalidJSONOutput(t *testing.T) {
	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
			return commandResult{Stdout: "not json"}, nil
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir: "/tmp/work",
	})
	if err == nil {
		t.Fatalf("expected failure")
	}
	if !strings.Contains(err.Error(), "agent output was not valid JSON") {
		t.Fatalf("expected malformed output classification, got %v", err)
	}
}

func TestCLIAdapterWrapsRunnerError(t *testing.T) {
	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, _ string) (commandResult, error) {
			return commandResult{}, errors.New("boom")
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir: "/tmp/work",
	})
	if err == nil {
		t.Fatalf("expected failure")
	}
	if !strings.Contains(err.Error(), "agent command execution failed") {
		t.Fatalf("expected wrapped runner error, got %v", err)
	}
}

func TestCLIAdapterUsesCustomReviewInstructions(t *testing.T) {
	capturedStdin := ""
	customInstructions := "Focus on security vulnerabilities and API contract violations."

	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, stdinPayload string) (commandResult, error) {
			capturedStdin = stdinPayload
			return commandResult{Stdout: `{"summary":"ok","risk":{"score":0.1,"reasons":["low"]},"findings":[],"checklist":["run tests"]}`}, nil
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:             types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir:            "/tmp/work",
		ReviewInstructions: customInstructions,
	})
	if err != nil {
		t.Fatalf("expected success with custom instructions, got %v", err)
	}
	if !strings.Contains(capturedStdin, customInstructions) {
		t.Fatalf("expected custom instructions in stdin payload, got %q", capturedStdin)
	}
	if strings.Contains(capturedStdin, defaultReviewInstructions) {
		t.Fatalf("expected default instructions to be replaced by custom, got %q", capturedStdin)
	}
	if !strings.Contains(capturedStdin, "INSTRUCTIONS") {
		t.Fatalf("expected pipeline INSTRUCTIONS marker in payload, got %q", capturedStdin)
	}
}

func TestCLIAdapterUsesDefaultReviewInstructionsWhenEmpty(t *testing.T) {
	capturedStdin := ""

	adapter := &CLIAgentAdapter{
		config: DefaultAgentConfig(),
		runner: fakeRunner{run: func(_ context.Context, _ string, _ []string, _ string, stdinPayload string) (commandResult, error) {
			capturedStdin = stdinPayload
			return commandResult{Stdout: `{"summary":"ok","risk":{"score":0.1,"reasons":["low"]},"findings":[],"checklist":["run tests"]}`}, nil
		}},
	}

	_, err := adapter.Review(context.Background(), ReviewInput{
		Bundle:  types.BundleV1{Version: "v1", Range: "HEAD^1..HEAD", Files: []string{}, Stat: "ok", Patch: "ok"},
		WorkDir: "/tmp/work",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !strings.Contains(capturedStdin, defaultReviewInstructions) {
		t.Fatalf("expected default instructions in stdin payload, got %q", capturedStdin)
	}
}
