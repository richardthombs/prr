package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultReviewInstructions = "Perform a code review of the included changes"

// UserConfig represents the configuration loaded from ~/.prr-config.json.
// Environment variables override file values; see LoadUserConfig.
type UserConfig struct {
	// CacheDir is the base directory for prr's local cache (repos and worktrees).
	// Env: PRR_CACHE_DIR. Default: ~/.cache/prr.
	CacheDir string `json:"cacheDir"`

	// ReviewInstructionsFile is the path to a file containing custom review instructions.
	// Env: PRR_REVIEW_INSTRUCTIONS_FILE.
	ReviewInstructionsFile string `json:"reviewInstructionsFile"`

	// IssueProviderMode controls how linked issues are fetched: "cli", "rest", or "cli-rest".
	// Env: PRR_ISSUE_PROVIDER_MODE. Default: "cli-rest".
	IssueProviderMode string `json:"issueProviderMode"`

	// GitHubToken is the personal access token for GitHub REST/GraphQL API calls.
	// Env: PRR_GITHUB_TOKEN.
	GitHubToken string `json:"githubToken"`

	// AzureDevOpsToken is the personal access token for Azure DevOps REST API calls.
	// Env: PRR_AZURE_DEVOPS_TOKEN.
	AzureDevOpsToken string `json:"azureDevOpsToken"`

	// GitHubAPIBaseURL overrides the GitHub API base URL (for GitHub Enterprise Server).
	// Env: PRR_GITHUB_API_BASE_URL. Default: "https://api.github.com".
	GitHubAPIBaseURL string `json:"githubApiBaseUrl"`

	// AgentCommand is the CLI command used to invoke the review agent.
	// Env: PRR_AGENT_COMMAND. Default: "copilot".
	AgentCommand string `json:"agentCommand"`

	// AgentArgs are the arguments passed to the review agent command.
	// Env: PRR_AGENT_ARGS (space-separated). Default: ["--allow-all-tools"].
	AgentArgs []string `json:"agentArgs"`

	// AgentModelArg is the flag name used to specify the AI model.
	// Env: PRR_AGENT_MODEL_ARG. Default: "--model".
	AgentModelArg string `json:"agentModelArg"`

	// AgentModelName is the default AI model name passed to the agent.
	// Overridden per-run by the --model flag.
	// Env: PRR_AGENT_MODEL_NAME.
	AgentModelName string `json:"agentModelName"`

	// AgentOutputMode controls how the agent's output is parsed.
	// Env: PRR_AGENT_OUTPUT_MODE. Default: "json-extracted".
	AgentOutputMode string `json:"agentOutputMode"`

	// AgentTimeoutSeconds is the maximum time in seconds to wait for the review agent.
	// Env: PRR_AGENT_TIMEOUT_SECONDS. Default: 120.
	AgentTimeoutSeconds int `json:"agentTimeoutSeconds"`
}

// LoadUserConfig loads the user configuration from ~/.prr-config.json.
// Environment variables override file values; if the file is absent the config
// is initialised from environment variables only.
func LoadUserConfig() (UserConfig, error) {
	var cfg UserConfig

	homeDir, err := os.UserHomeDir()
	if err == nil {
		data, readErr := os.ReadFile(filepath.Join(homeDir, ".prr-config.json"))
		if readErr != nil && !os.IsNotExist(readErr) {
			return UserConfig{}, readErr
		}
		if readErr == nil {
			if err := json.Unmarshal(data, &cfg); err != nil {
				return UserConfig{}, err
			}
		}
	}

	applyEnvOverrides(&cfg)
	return cfg, nil
}

// applyEnvOverrides applies environment variable values on top of cfg,
// with environment variables taking precedence over config file values.
func applyEnvOverrides(cfg *UserConfig) {
	if v := strings.TrimSpace(os.Getenv("PRR_CACHE_DIR")); v != "" {
		cfg.CacheDir = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_REVIEW_INSTRUCTIONS_FILE")); v != "" {
		cfg.ReviewInstructionsFile = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_ISSUE_PROVIDER_MODE")); v != "" {
		cfg.IssueProviderMode = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_GITHUB_TOKEN")); v != "" {
		cfg.GitHubToken = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_AZURE_DEVOPS_TOKEN")); v != "" {
		cfg.AzureDevOpsToken = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_GITHUB_API_BASE_URL")); v != "" {
		cfg.GitHubAPIBaseURL = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_AGENT_COMMAND")); v != "" {
		cfg.AgentCommand = v
	}
	if v := os.Getenv("PRR_AGENT_ARGS"); strings.TrimSpace(v) != "" {
		cfg.AgentArgs = strings.Fields(strings.TrimSpace(v))
	}
	if v := strings.TrimSpace(os.Getenv("PRR_AGENT_MODEL_ARG")); v != "" {
		cfg.AgentModelArg = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_AGENT_MODEL_NAME")); v != "" {
		cfg.AgentModelName = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_AGENT_OUTPUT_MODE")); v != "" {
		cfg.AgentOutputMode = v
	}
	if v := strings.TrimSpace(os.Getenv("PRR_AGENT_TIMEOUT_SECONDS")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cfg.AgentTimeoutSeconds = parsed
		}
	}
}

// ResolveReviewInstructions returns the review instructions from the file
// specified in cfg.ReviewInstructionsFile, or the default if no file is
// specified or the file cannot be read.
func ResolveReviewInstructions(cfg UserConfig) string {
	if path := strings.TrimSpace(cfg.ReviewInstructionsFile); path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if instructions := strings.TrimSpace(string(data)); instructions != "" {
				return instructions
			}
		}
	}

	return DefaultReviewInstructions
}
