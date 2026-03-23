# Release Notes

## 2026-03-23

### WSL setup documentation added

- Added `docs/wsl-setup.md` with a step-by-step guide for running PRR on Windows 11 via WSL, covering:
  - WSL 2 + Ubuntu installation.
  - Updating Ubuntu and installing Go, git, build-essential, and Make.
  - Installing and configuring the Git Credential Manager (GCM) with the Windows credential store.
  - Installing and configuring the GitHub Copilot CLI (`copilot`, not `gh`), including Node.js setup via nvm.
  - Building and installing PRR from source inside WSL.
  - A troubleshooting reference table for common issues.
- Updated `docs/install.md` to add a WSL callout in the Windows section.
- Updated `README.md` to link to the new WSL guide.

## 2026-03-07

### Distribution strategy pivot: source-first install guidance for macOS, Ubuntu, and Windows

- PRR is now documented as source-first distribution instead of prebuilt binary distribution.
- Added `docs/install.md` with platform-specific prerequisites and install instructions for:
  - macOS
  - Ubuntu Linux
  - Windows
- Simplified `Makefile` to source-focused targets: `build`, `install`, `test`, `clean`.
- Updated `README.md` to point contributors to source install and verification workflow.
- Replaced release-process contract with source distribution and install contract guidance.

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

