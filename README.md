# PRR

PRR is a CLI tool that automates pull request review from a single command by creating an isolated merge snapshot, generating a deterministic diff bundle, and sending it to a review engine for actionable output. It is designed to save setup time, keep reviews consistent, and protect your active working copy by running in a separate worktree with clear, script-friendly success and failure signalling.

## Commands

- `prr review <PR_ID>`  
  Run an end-to-end review for a pull request.

- `prr review <PR_ID> --keep`  
  Keep the isolated review worktree after the run for inspection.

- `prr review <PR_ID> --output-format <markdown|json>`  
  Choose human-readable Markdown (default) or structured JSON output.

- `prr review <PR_ID> --publish`  
  Publish review results back to the pull request when supported.

- `prr review <PR_ID> --provider <provider> --repo <owner/repo>`  
  Override inferred provider/repository context.

- `prr checkout <PR_URL> [--provider <provider>] [--repo <repoUrl>] [--remote <name>] [--keep] [--verbose] [--what-if]`
  Resolve PR context, ensure/update mirror, fetch merge ref, and prepare/reset the isolated worktree in one command.
  Emits a single JSON payload including `prId`, `repoUrl`, `remote`, `provider`, `bareDir`, `mergeRef`, `workDir`, `keep`, and `cleanup`.
  Supports Azure DevOps and GitHub PR URL formats.

- `prr review <PR_ID> --max-patch-bytes <bytes> --max-files <count>`  
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
