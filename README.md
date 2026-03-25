# PRR

PRR is a CLI tool that automates pull request review from a single command by creating an isolated merge snapshot, generating a deterministic diff bundle, augmenting it with information from linked issues, and sending it to a review engine for actionable output. It is designed to save setup time, keep reviews consistent, and protect your active working copy by running in a separate worktree with clear, script-friendly success and failure signalling.

## Installation

**Quick Install (macOS/Linux):**
```bash
curl -fsSL https://raw.githubusercontent.com/richardthombs/prr/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/richardthombs/prr/main/scripts/install.ps1 | iex
```

Verify installation:
```bash
prr version
```

## Commands

- `prr review <PR_URL> [--json] [--verbose] [--keep] [--model <model_name>]`  
   Run an end-to-end review for a pull request using either numeric ID or full PR URL input. Emits a Markdown review report to stdout by default

- `prr check <PR_URL>`  
   Clone and checkout the relvant branch, but don't perform a review. _(Useful for conveniently getting a worktree of the PR's branch without affecting your own clones)_

- `prr pwd <PR_URL>`
   Returns the path of the worktree for the specified PR.

- `prr version`  
   Show the installed PRR version.

### Parameters
- `PR_URL`
   The URL to an Azure DevOps PR or a GitHub PR

- `model_name`
   The name of the model to use when reviewing the PR. This should make sense to the agent you are invoking (Copilot is the only agent invoked currently)


## Configuration

PRR is configured via `~/.prr-config.json` and/or environment variables. Environment variables always take precedence over the config file. If neither is set, the default value applies.

### General

| JSON key | Environment variable | Default | Description |
|---|---|---|---|
| `cacheDir` | `PRR_CACHE_DIR` | `~/.cache/prr` | Base directory for the local git mirror cache. Bare repos are stored under `<cacheDir>/repos` and worktrees under `<cacheDir>/work`. |
| `reviewInstructionsFile` | `PRR_REVIEW_INSTRUCTIONS_FILE` | *(built-in prompt)* | Path to a Markdown file whose contents replace the default review prompt. Must be an absolute path. If the file is absent or empty, the default is used. |

### Issue discovery

PRR can enrich reviews with linked issue context from GitHub or Azure DevOps.

| JSON key | Environment variable | Default | Description |
|---|---|---|---|
| `issueProviderMode` | `PRR_ISSUE_PROVIDER_MODE` | `cli-rest` | How linked issues are fetched: `cli` (CLI only), `rest` (REST only), or `cli-rest` (CLI with automatic REST fallback). |
| `githubToken` | `PRR_GITHUB_TOKEN` | — | Personal access token for GitHub REST and GraphQL API calls. Required for REST mode with GitHub. |
| `azureDevOpsToken` | `PRR_AZURE_DEVOPS_TOKEN` | — | Personal access token for Azure DevOps REST API calls. Required for REST mode with Azure DevOps. |
| `githubApiBaseUrl` | `PRR_GITHUB_API_BASE_URL` | `https://api.github.com` | Override the GitHub API base URL. Useful for GitHub Enterprise Server. |

CLI mode uses `gh` for GitHub and `az` for Azure DevOps. If both CLI and REST fail in `cli-rest` mode, PRR returns a provider error with both failure paths for diagnosis.

### Review agent

| JSON key | Environment variable | Default | Description |
|---|---|---|---|
| `agentCommand` | `PRR_AGENT_COMMAND` | `copilot` | The CLI command used to invoke the review agent. |
| `agentArgs` | `PRR_AGENT_ARGS` | `--allow-all-tools` | Arguments passed to the agent command (space-separated in env var form). |
| `agentModelName` | `PRR_AGENT_MODEL_NAME` | — | Default AI model name to use. Overridden per-run by the `--model` flag. |
| `agentModelArg` | `PRR_AGENT_MODEL_ARG` | `--model` | The flag name the agent uses to accept a model name (e.g. change to `--ai-model` if your agent binary differs). |
| `agentOutputMode` | `PRR_AGENT_OUTPUT_MODE` | `json-extracted` | How the agent's output is parsed. `json-extracted` scans stdout for the first valid JSON object. |
| `agentTimeoutSeconds` | `PRR_AGENT_TIMEOUT_SECONDS` | `120` | Maximum time in seconds to wait for the review agent before aborting. |

### Example `~/.prr-config.json`

```json
{
  "reviewInstructionsFile": "~/.config/prr/review-prompt.md",
  "agentModelName": "gpt-4.5",
  "issueProviderMode": "cli-rest",
  "githubToken": "ghp_...",
  "azureDevOpsToken": "..."
}
```
