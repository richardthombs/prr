# Release Notes

## 2026-03-06

### Review command now emits Markdown by default; `--json` flag added; `render` command removed

- `prr review` now emits a formatted Markdown review report to stdout by default (Summary, Risk, Findings, Checklist sections).
- New `--json` flag on `prr review` emits structured JSON (`summary`, `risk`, `findings`, `checklist`) for automation workflows and shell pipelines.
- The `prr render` command has been removed. Rendering is now an internal step of `review`.
- Migration: replace `prr review … | prr render` with `prr review …`; replace `prr review … | some-tool` with `prr review … --json | some-tool`.

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

