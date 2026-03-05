# PRR

PRR is a CLI tool that automates pull request review from a single command by creating an isolated merge snapshot, generating a deterministic diff bundle, and sending it to a review engine for actionable output. It is designed to save setup time, keep reviews consistent, and protect your active working copy by running in a separate worktree with clear, script-friendly success and failure signalling.

## Commands

- `prr review [<PR_ID>|<PR_URL>]`  
  Run an end-to-end review for a pull request using either numeric ID or full PR URL input.

- `prr checkout <PR_URL> | prr review`  
  Pipe checkout JSON directly into review so `prId`, `repoUrl`, `provider`, and `remote` are inferred from stdin.

- `prr review [<PR_ID>|<PR_URL>] --keep`  
  Keep the isolated review worktree after the run for inspection.

- `prr review <PR_ID> --provider <provider> --repo <repoUrl>`  
  Optional explicit overrides when ID-only input is used and automatic context is not sufficient.

- `prr checkout <PR_URL> [--provider <provider>] [--repo <repoUrl>] [--remote <name>] [--keep] [--verbose] [--what-if]`
  Resolve PR context, ensure/update mirror, fetch merge ref, and prepare/reset the isolated worktree in one command.
  Emits a single JSON payload including `prId`, `repoUrl`, `remote`, `provider`, `bareDir`, `mergeRef`, `workDir`, `keep`, and `cleanup`.
  Supports Azure DevOps and GitHub PR URL formats.

- `prr diff [--work-dir <path>] [--verbose] [--what-if]`
  Generate deterministic diff outputs from `HEAD^1..HEAD` in the isolated worktree and emit JSON with `files`, `stat`, and `patch`.
  Accepts stdin JSON (for example from `checkout`) and falls back to `workDir` from that payload when `--work-dir` is omitted.

- `prr bundle [--max-patch-bytes <bytes>] [--max-files <count>] [--verbose] [--what-if]`
  Build and validate a v1 review bundle from `diff` JSON on stdin.
  Enforces optional input-size limits with explicit diagnostics when limits are exceeded.

- `prr review [<PR_ID>|<PR_URL>] --max-patch-bytes <bytes> --max-files <count>`  
  Override safety limits for patch size and changed file count.

- `prr --help`  
  Show CLI help and available options.

- `prr version`  
  Show the installed PRR version.

## Checkout example

```bash
# Single-step checkout for PR workspace preparation.

prr checkout "https://github.com/<owner>/<repo>/pull/<id>"
```

The `checkout` output includes workspace fields (for example `bareDir`, `mergeRef`, `workDir`) ready for downstream `diff` and `bundle` stages.

## Review examples

```bash
# Run review using full PR URL (no --repo/--provider required)
prr review "https://github.com/<owner>/<repo>/pull/<id>"

# Run review from checkout JSON pipeline
prr checkout "https://github.com/<owner>/<repo>/pull/<id>" | prr review
```
