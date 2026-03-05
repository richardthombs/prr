package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	apperrors "github.com/richardthombs/prr/internal/errors"
	"github.com/richardthombs/prr/internal/types"
)

type ReviewEngine interface {
	Review(ctx context.Context, input ReviewInput) (types.Review, error)
}

type ReviewInput struct {
	Bundle  types.BundleV1
	WorkDir string
	Model   string
	Verbose bool
	WhatIf  bool
	Logger  func(format string, args ...any)
}

type AgentConfig struct {
	Command        string
	Args           []string
	ModelArg       string
	InputMode      string
	OutputMode     string
	TimeoutSeconds int
}

type commandResult struct {
	Stdout string
	Stderr string
}

type commandRunner interface {
	Run(ctx context.Context, command string, args []string, cwd string, stdinPayload string) (commandResult, error)
}

type CLIAgentAdapter struct {
	config AgentConfig
	runner commandRunner
}

func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Command:        "copilot",
		Args:           []string{"--allow-all-tools"},
		ModelArg:       "--model",
		InputMode:      "stdin",
		OutputMode:     "json-extracted",
		TimeoutSeconds: 60,
	}
}

func NewCLIAdapter(config AgentConfig) ReviewEngine {
	return &CLIAgentAdapter{
		config: config,
		runner: execRunner{},
	}
}

func NewDefaultAdapter() ReviewEngine {
	return NewCLIAdapter(DefaultAgentConfig())
}

func (a *CLIAgentAdapter) Review(ctx context.Context, input ReviewInput) (types.Review, error) {
	if strings.TrimSpace(input.WorkDir) == "" {
		return types.Review{}, apperrors.WrapConfig("review engine requires workDir", nil)
	}

	command, args := buildCommand(a.config, strings.TrimSpace(input.Model))
	if err := validateConfig(a.config, command, args); err != nil {
		return types.Review{}, err
	}

	stdinPayload, err := marshalBundlePayload(input.Bundle)
	if err != nil {
		return types.Review{}, err
	}

	logf := input.Logger
	if logf == nil {
		logf = func(string, ...any) {}
	}

	if input.Verbose || input.WhatIf {
		logf("review engine command: %s", quoteCommand(command, args))
		logf("review engine input mode: %s", a.config.InputMode)
	}

	if input.WhatIf {
		return whatIfReview(input.Bundle), nil
	}

	timeout := time.Duration(a.config.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := a.runner.Run(timeoutCtx, command, args, strings.TrimSpace(input.WorkDir), stdinPayload)
	if input.Verbose {
		logf("copilot stdout:\n%s", previewOutput(result.Stdout))
		if strings.TrimSpace(result.Stderr) != "" {
			logf("copilot stderr:\n%s", previewOutput(result.Stderr))
		}
	}

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return types.Review{}, apperrors.WrapEngine("agent command timed out", nil)
		}
		if errors.Is(err, exec.ErrNotFound) {
			return types.Review{}, apperrors.WrapEngine("agent command not found", nil)
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			message := "agent command failed with non-zero exit"
			if diagnostic := nonZeroExitDiagnostic(result); diagnostic != "" {
				message = message + ": " + diagnostic
			}

			return types.Review{}, apperrors.WrapEngine(message, nil)
		}

		return types.Review{}, apperrors.WrapEngine("agent command execution failed", err)
	}

	if strings.TrimSpace(result.Stdout) == "" {
		return types.Review{}, apperrors.WrapEngine("agent produced empty output", nil)
	}

	review, err := parseReviewOutput(result.Stdout)
	if err != nil {
		return types.Review{}, err
	}

	return review, nil
}

func validateConfig(config AgentConfig, command string, args []string) error {
	if strings.TrimSpace(command) == "" {
		return apperrors.WrapConfig("agent command is required", nil)
	}
	if strings.TrimSpace(config.ModelArg) == "" {
		return apperrors.WrapConfig("agent model_arg is required", nil)
	}
	inputMode := strings.ToLower(strings.TrimSpace(config.InputMode))
	if inputMode != "stdin" && inputMode != "file" {
		return apperrors.WrapConfig("agent input_mode must be stdin or file", nil)
	}
	if len(args) == 0 {
		return apperrors.WrapConfig("agent args must include non-interactive command arguments", nil)
	}

	return nil
}

func buildCommand(config AgentConfig, model string) (string, []string) {
	args := append([]string{}, config.Args...)
	if strings.TrimSpace(model) != "" {
		args = append(args, config.ModelArg, strings.TrimSpace(model))
	}

	return strings.TrimSpace(config.Command), args
}

func marshalBundlePayload(bundle types.BundleV1) (string, error) {
	bundleJSON, err := json.Marshal(bundle)
	if err != nil {
		return "", apperrors.WrapRuntime("failed to encode review input payload", err)
	}

	instructions := strings.TrimSpace(`INSTRUCTIONS
1) Analyse ONLY the JSON object between DIFF_BUNDLE_JSON_START and DIFF_BUNDLE_JSON_END.
2) Treat that JSON object as the complete review input.
3) Return ONLY valid JSON using this exact schema (no markdown fences or extra prose):
{
	"summary": string,
	"risk": {"score": number, "reasons": string[]},
	"findings": [
		{
			"id": string,
			"file": string,
			"line": number,
			"severity": "blocker"|"important"|"suggestion"|"nit",
			"category": "correctness"|"security"|"performance"|"readability"|"api"|"tests"|"other",
			"message": string,
			"suggestion": string
		}
	],
	"checklist": string[]
}
4) risk.score MUST be a decimal number between 0 and 1 inclusive.
5) Be deterministic and concise.`)

	stdinEnvelope := strings.Join([]string{
		instructions,
		"",
		"DIFF_BUNDLE_JSON_START",
		string(bundleJSON),
		"DIFF_BUNDLE_JSON_END",
		"",
	}, "\n")

	return stdinEnvelope, nil
}

func parseReviewOutput(raw string) (types.Review, error) {
	jsonPayload, err := extractJSONObject(raw)
	if err != nil {
		return types.Review{}, apperrors.WrapEngine("agent output was not valid JSON", nil)
	}

	var review types.Review
	if err := json.Unmarshal([]byte(jsonPayload), &review); err != nil {
		return types.Review{}, apperrors.WrapEngine("agent output could not be parsed as review JSON", nil)
	}

	validated, err := types.NormalizeAndValidateReviewOutput(review)
	if err != nil {
		return types.Review{}, apperrors.WrapEngine("agent output failed review schema validation", err)
	}

	return validated, nil
}

func extractJSONObject(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("empty output")
	}

	if json.Valid([]byte(trimmed)) {
		var parsed any
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			normalized, marshalErr := json.Marshal(parsed)
			if marshalErr == nil {
				return string(normalized), nil
			}
		}
	}

	for i := range trimmed {
		if trimmed[i] != '{' {
			continue
		}

		decoder := json.NewDecoder(strings.NewReader(trimmed[i:]))
		decoder.UseNumber()

		var parsed any
		if err := decoder.Decode(&parsed); err != nil {
			continue
		}

		normalized, err := json.Marshal(parsed)
		if err != nil {
			continue
		}

		return string(normalized), nil
	}

	return "", fmt.Errorf("no json object found")
}

func whatIfReview(bundle types.BundleV1) types.Review {
	summary := "what-if: review agent was not executed"
	if bundle.PRID > 0 {
		summary = fmt.Sprintf("what-if: review agent was not executed for PR #%d", bundle.PRID)
	}

	return types.Review{
		Summary: summary,
		Risk: types.Risk{
			Score:   0,
			Reasons: []string{"What-if mode skips external agent execution."},
		},
		Findings:  []types.Finding{},
		Checklist: []string{"Run without --what-if to execute Copilot review."},
	}
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, command string, args []string, cwd string, stdinPayload string) (commandResult, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = cwd
	cmd.Stdin = strings.NewReader(stdinPayload)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return commandResult{}, context.DeadlineExceeded
		}

		return commandResult{Stdout: stdout.String(), Stderr: stderr.String()}, err
	}

	return commandResult{Stdout: stdout.String(), Stderr: stderr.String()}, nil
}

func nonZeroExitDiagnostic(result commandResult) string {
	raw := strings.TrimSpace(result.Stderr)
	if raw == "" {
		raw = strings.TrimSpace(result.Stdout)
	}
	if raw == "" {
		return ""
	}

	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "unknown command"),
		strings.Contains(lower, "unknown flag"),
		strings.Contains(lower, "invalid option"),
		strings.Contains(lower, "invalid argument"):
		return "unsupported copilot CLI invocation; verify installed copilot version/flags"
	case strings.Contains(lower, "not logged in"),
		strings.Contains(lower, "login required"),
		strings.Contains(lower, "authentication"):
		return "copilot authentication missing; run `copilot auth login`"
	case strings.Contains(lower, "allow-all-tools"),
		strings.Contains(lower, "permission"),
		strings.Contains(lower, "not allowed"):
		return "copilot permission settings blocked non-interactive run; check CLI permissions/options"
	}

	return sanitizeDiagnostic(raw)
}

func sanitizeDiagnostic(raw string) string {
	clean := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if clean == "" {
		return ""
	}

	lower := strings.ToLower(clean)
	if strings.Contains(lower, "token") || strings.Contains(lower, "authorization") || strings.Contains(lower, "bearer ") {
		return "copilot returned an error (details redacted)"
	}

	const maxLen = 220
	if len(clean) > maxLen {
		return clean[:maxLen] + "..."
	}

	return clean
}

func previewOutput(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "(empty)"
	}

	const maxLen = 4000
	if len(trimmed) > maxLen {
		return trimmed[:maxLen] + "\n...[truncated]"
	}

	return trimmed
}

func quoteCommand(command string, args []string) string {
	quoted := make([]string, 0, len(args)+1)
	quoted = append(quoted, shellQuote(command))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}

	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if strings.ContainsAny(value, " \t\n\"'\\$") {
		return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
	}

	return value
}
