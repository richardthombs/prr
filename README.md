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

- `prr resolve <PR_URL>`
  Resolve PR context into stable `PRRef` JSON (`prId`, `repoUrl`, `remote`, `provider`) from a pull-request URL.

- `prr resolve https://dev.azure.com/<org>/<project>/_git/<repo>/pullrequest/<id>`
  Auto-detect Azure DevOps provider context.

- `prr resolve https://github.com/<owner>/<repo>/pull/<id>`
  Auto-detect GitHub provider context.

- `prr review <PR_ID> --max-patch-bytes <bytes> --max-files <count>`  
  Override safety limits for patch size and changed file count.

- `prr --help`  
  Show CLI help and available options.

- `prr --version`  
  Show the installed PRR version.
