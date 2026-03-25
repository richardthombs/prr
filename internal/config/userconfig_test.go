package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUserConfigReturnsEmptyWhenFileAbsent(t *testing.T) {
	t.TempDir() // ensure a fresh home-like env
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))

	cfg, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("expected no error for missing config, got %v", err)
	}
	if cfg.ReviewInstructionsFile != "" {
		t.Fatalf("expected empty config, got %+v", cfg)
	}
}

func TestLoadUserConfigParsesReviewInstructionsFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	configData := map[string]string{"reviewInstructionsFile": "/path/to/instructions.md"}
	data, _ := json.Marshal(configData)
	if err := os.WriteFile(filepath.Join(dir, ".prr-config.json"), data, 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ReviewInstructionsFile != "/path/to/instructions.md" {
		t.Fatalf("expected reviewInstructionsFile, got %q", cfg.ReviewInstructionsFile)
	}
}

func TestLoadUserConfigReturnsErrorOnInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	if err := os.WriteFile(filepath.Join(dir, ".prr-config.json"), []byte("not json"), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadUserConfig()
	if err == nil {
		t.Fatalf("expected error for invalid JSON, got nil")
	}
}

func TestResolveReviewInstructionsReturnsDefaultWhenNoFile(t *testing.T) {
	instructions := ResolveReviewInstructions(UserConfig{})
	if instructions != DefaultReviewInstructions {
		t.Fatalf("expected default instructions, got %q", instructions)
	}
}

func TestResolveReviewInstructionsReturnsDefaultWhenFileAbsent(t *testing.T) {
	instructions := ResolveReviewInstructions(UserConfig{ReviewInstructionsFile: "/nonexistent/path/instructions.md"})
	if instructions != DefaultReviewInstructions {
		t.Fatalf("expected default instructions for missing file, got %q", instructions)
	}
}

func TestResolveReviewInstructionsReturnsFileContents(t *testing.T) {
	dir := t.TempDir()
	instructionsPath := filepath.Join(dir, "review.md")
	customInstructions := "Focus on security vulnerabilities and performance bottlenecks."
	if err := os.WriteFile(instructionsPath, []byte(customInstructions), 0600); err != nil {
		t.Fatalf("failed to write instructions file: %v", err)
	}

	instructions := ResolveReviewInstructions(UserConfig{ReviewInstructionsFile: instructionsPath})
	if instructions != customInstructions {
		t.Fatalf("expected custom instructions %q, got %q", customInstructions, instructions)
	}
}

func TestResolveReviewInstructionsReturnsDefaultForEmptyFile(t *testing.T) {
	dir := t.TempDir()
	instructionsPath := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(instructionsPath, []byte("   \n  "), 0600); err != nil {
		t.Fatalf("failed to write empty instructions file: %v", err)
	}

	instructions := ResolveReviewInstructions(UserConfig{ReviewInstructionsFile: instructionsPath})
	if instructions != DefaultReviewInstructions {
		t.Fatalf("expected default instructions for empty file, got %q", instructions)
	}
}

func clearPRREnvVars(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"PRR_CACHE_DIR", "PRR_REVIEW_INSTRUCTIONS_FILE", "PRR_ISSUE_PROVIDER_MODE",
		"PRR_GITHUB_TOKEN", "PRR_AZURE_DEVOPS_TOKEN", "PRR_GITHUB_API_BASE_URL",
		"PRR_AGENT_COMMAND", "PRR_AGENT_ARGS", "PRR_AGENT_MODEL_ARG", "PRR_AGENT_MODEL_NAME",
		"PRR_AGENT_OUTPUT_MODE", "PRR_AGENT_TIMEOUT_SECONDS",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadUserConfigAllFieldsFromJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	clearPRREnvVars(t)

	cfg := map[string]any{
		"cacheDir":               "/custom/cache",
		"reviewInstructionsFile": "/path/review.md",
		"issueProviderMode":      "rest",
		"githubToken":            "ghtoken",
		"azureDevOpsToken":       "adotoken",
		"githubApiBaseUrl":       "https://ghe.example.com",
		"agentCommand":           "my-agent",
		"agentArgs":              []string{"--flag"},
		"agentModelArg":          "--ai-model",
		"agentModelName":         "gpt-5",
		"agentOutputMode":        "raw",
		"agentTimeoutSeconds":    300,
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, ".prr-config.json"), data, 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loaded, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.CacheDir != "/custom/cache" {
		t.Errorf("cacheDir: got %q", loaded.CacheDir)
	}
	if loaded.ReviewInstructionsFile != "/path/review.md" {
		t.Errorf("reviewInstructionsFile: got %q", loaded.ReviewInstructionsFile)
	}
	if loaded.IssueProviderMode != "rest" {
		t.Errorf("issueProviderMode: got %q", loaded.IssueProviderMode)
	}
	if loaded.GitHubToken != "ghtoken" {
		t.Errorf("githubToken: got %q", loaded.GitHubToken)
	}
	if loaded.AzureDevOpsToken != "adotoken" {
		t.Errorf("azureDevOpsToken: got %q", loaded.AzureDevOpsToken)
	}
	if loaded.GitHubAPIBaseURL != "https://ghe.example.com" {
		t.Errorf("githubApiBaseUrl: got %q", loaded.GitHubAPIBaseURL)
	}
	if loaded.AgentCommand != "my-agent" {
		t.Errorf("agentCommand: got %q", loaded.AgentCommand)
	}
	if len(loaded.AgentArgs) != 1 || loaded.AgentArgs[0] != "--flag" {
		t.Errorf("agentArgs: got %v", loaded.AgentArgs)
	}
	if loaded.AgentModelArg != "--ai-model" {
		t.Errorf("agentModelArg: got %q", loaded.AgentModelArg)
	}
	if loaded.AgentModelName != "gpt-5" {
		t.Errorf("agentModelName: got %q", loaded.AgentModelName)
	}
	if loaded.AgentOutputMode != "raw" {
		t.Errorf("agentOutputMode: got %q", loaded.AgentOutputMode)
	}
	if loaded.AgentTimeoutSeconds != 300 {
		t.Errorf("agentTimeoutSeconds: got %d", loaded.AgentTimeoutSeconds)
	}
}

func TestEnvVarsOverrideConfigFileValues(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	clearPRREnvVars(t)

	// Write config file with one value per field.
	cfg := map[string]any{
		"cacheDir":     "/file/cache",
		"githubToken":  "file-token",
		"agentCommand": "file-agent",
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(dir, ".prr-config.json"), data, 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Env vars override the file values.
	t.Setenv("PRR_CACHE_DIR", "/env/cache")
	t.Setenv("PRR_GITHUB_TOKEN", "env-token")
	// agentCommand not overridden — file value should survive.

	loaded, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.CacheDir != "/env/cache" {
		t.Errorf("expected env PRR_CACHE_DIR to win, got %q", loaded.CacheDir)
	}
	if loaded.GitHubToken != "env-token" {
		t.Errorf("expected env PRR_GITHUB_TOKEN to win, got %q", loaded.GitHubToken)
	}
	if loaded.AgentCommand != "file-agent" {
		t.Errorf("expected file agentCommand to survive when no env var set, got %q", loaded.AgentCommand)
	}
}

func TestEnvVarsPRRReviewInstructionsFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	clearPRREnvVars(t)
	t.Setenv("PRR_REVIEW_INSTRUCTIONS_FILE", "/env/instructions.md")

	loaded, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.ReviewInstructionsFile != "/env/instructions.md" {
		t.Errorf("expected PRR_REVIEW_INSTRUCTIONS_FILE to be applied, got %q", loaded.ReviewInstructionsFile)
	}
}

func TestEnvVarPRRAgentArgsOverridesConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	clearPRREnvVars(t)
	t.Setenv("PRR_AGENT_ARGS", "--yolo --extra")

	loaded, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded.AgentArgs) != 2 || loaded.AgentArgs[0] != "--yolo" || loaded.AgentArgs[1] != "--extra" {
		t.Errorf("expected PRR_AGENT_ARGS to be split into slice, got %v", loaded.AgentArgs)
	}
}

func TestEnvVarPRRAgentTimeoutSeconds(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	clearPRREnvVars(t)
	t.Setenv("PRR_AGENT_TIMEOUT_SECONDS", "45")

	loaded, err := LoadUserConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.AgentTimeoutSeconds != 45 {
		t.Errorf("expected AgentTimeoutSeconds=45, got %d", loaded.AgentTimeoutSeconds)
	}
}
