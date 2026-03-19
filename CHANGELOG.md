# Changelog

All notable changes to prr are documented here.
This file is updated automatically by GoReleaser from conventional commit messages on each release.

## v0.3.0 - 2026-03-19

### Bug Fixes
- Use canonical Azure DevOps pull request URLs in review output
- Use human-facing Azure DevOps work item URLs instead of REST API URLs in linked issue context

## v0.2.3 - 2026-03-19

### Bug Fixes
- Pass the GitHub repository explicitly to `gh release upload` in bottle jobs so asset uploads do not depend on a local `.git` checkout

## v0.2.2 - 2026-03-19

### Bug Fixes
- Match Homebrew bottle tarball filenames with rebuild suffixes such as `.bottle.1.tar.gz` when uploading release assets

## v0.2.1 - 2026-03-19

### Bug Fixes
- Inject the CLI version with the `v` prefix in GoReleaser ldflags and add version tests to lock the contract
- Remove unsupported Intel macOS Homebrew bottle support and harden bottle asset upload in the release workflow

## v0.2.0 - 2026-03-19

### Features
- Add optional `~/.prr-config.json` support for custom review instructions via `reviewInstructionsFile`
- Enrich review bundles with linked issue and work item context from GitHub and Azure DevOps providers
- Add provider CLI-to-REST fallback for linked issue discovery, with GitHub GraphQL-backed issue resolution
- Revise `prr review` Markdown output to separate PR and issue summaries and group review conclusions by severity
- Extend the release pipeline to build and publish Homebrew bottles
- Add Ubuntu Linux to the Homebrew bottle release matrix

### Bug Fixes
- Fix `scripts/install.sh` archive version naming to match released artifact filenames

### Other
- Add Gas Town workspace hooks

<!-- GoReleaser appends release notes above this line -->

## v0.1.3 - 2026-03-15

### Bug Fixes
- Support reviewing closed GitHub PRs: use `gh`/`az` CLI to obtain base branch SHA when merge ref is unavailable
- Suppress git subprocess stderr bleed; progress output now only appears with `--verbose`

## v0.1.2 - 2026-03-15

### Features
- Add `prr pwd` command to print the PR worktree path
- Increase Copilot default timeout to 120s
- Add goreleaser-based multi-OS release pipeline (Linux, macOS, Windows × x86-64 & ARM64)
- Add Homebrew tap publishing support
- Add `scripts/install.sh` and `scripts/install.ps1` quick-install scripts
- Add MIT license

### Bug Fixes
- Remove tracked `dist/` binaries from git (caused GoReleaser dirty-repo failure)
- Correct smoke test help output regex (`Usage:` not `Use:`)
- Fix release workflow to pass explicit repo to `gh release` commands

### Other
- Pivot to source-first build and install workflow
- `make install` now copies binary to `~/.local/bin`
