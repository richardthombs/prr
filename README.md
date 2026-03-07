# PRR

PRR is a CLI tool that automates pull request review from a single command by creating an isolated merge snapshot, generating a deterministic diff bundle, and sending it to a review engine for actionable output. It is designed to save setup time, keep reviews consistent, and protect your active working copy by running in a separate worktree with clear, script-friendly success and failure signalling.

## Build and test (cross-platform source of truth)

- Build all packages: `go build ./...`
- Run all tests: `go test ./...`
- Build CLI binary only: `go build -o ./prr ./cmd/prr`

These Go commands are the canonical contributor workflow on macOS, Linux, and Windows.
`Makefile` targets are optional convenience helpers for Unix-like environments.

## Install from source (recommended)

- Fast path: `make install` (runs `go install ./cmd/prr`)
- Platform-specific setup and PATH guidance: `docs/install.md`
- Verify install: `prr version`

## Commands

- `prr review <PR_URL>`  
  Run an end-to-end review for a pull request using either numeric ID or full PR URL input. Emits a Markdown review report to stdout by default.

- `prr review <PR_URL> --json`  
  Emit structured JSON output (`summary`, `risk`, `findings`, `checklist`) instead of Markdown — for automation workflows and shell pipelines.

- `prr review <PR_URL> --keep`  
  Keep the isolated review worktree after the run for inspection.

- `prr checkout <PR_URL> | prr review`  
  Pipe checkout JSON directly into review so `prId`, `repoUrl`, `provider`, and `remote` are inferred from stdin.

- `prr checkout <PR_URL> [--verbose] [--what-if]`
  Resolve PR context, ensure/update mirror, fetch merge ref, and prepare/reset the isolated worktree in one command.
  Emits a single JSON payload including `prId`, `repoUrl`, `remote`, `provider`, `bareDir`, `mergeRef`, `workDir`, `keep`, and `cleanup`.
  Supports Azure DevOps and GitHub PR URL formats.

- `prr review [<PR_ID>|<PR_URL>] --max-patch-bytes <bytes> --max-files <count>`  
  Override safety limits for patch size and changed file count.

- `prr review <PR_URL> --model <model_name>`
  Select the Copilot model for this review run; PRR passes this through to Copilot as `--model`.

- `prr --help`  
  Show CLI help and available options.

- `prr version`  
  Show the installed PRR version.

## Checkout example

```bash
# Single-step checkout for PR workspace preparation.

prr checkout "https://github.com/<owner>/<repo>/pull/<id>"
```

The `checkout` output includes workspace fields (for example `bareDir`, `mergeRef`, `workDir`) ready for downstream review pipeline stages.

## Review examples

```bash
# Run review using full PR URL — Markdown output by default
prr review "https://github.com/<owner>/<repo>/pull/<id>"

# Emit structured JSON instead of Markdown (for automation)
prr review "https://github.com/<owner>/<repo>/pull/<id>" --json

# Run review from checkout JSON pipeline
prr checkout "https://github.com/<owner>/<repo>/pull/<id>" | prr review
```
