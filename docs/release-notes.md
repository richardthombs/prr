# Release Notes

## 2026-03-04

### Resolve command contract update

- Updated `prr resolve` to accept a pull request URL argument: `prr resolve <PR_URL>`.
- Added provider auto-detection from supported PR URL formats:
  - Azure DevOps: `https://dev.azure.com/<org>/<project>/_git/<repo>/pullrequest/<id>`
  - GitHub: `https://github.com/<owner>/<repo>/pull/<id>`
- `resolve` emits stable `PRRef` JSON (`prId`, `repoUrl`, `remote`, `provider`) from URL-decomposed context.
- Override flags remain supported (`--provider`, `--repo`, `--remote`) and take precedence over auto-detected values.

### Tooling

- Updated `Makefile` build target to produce the runnable CLI binary at `./prr` from `./cmd/prr`.
