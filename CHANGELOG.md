# Changelog

All notable changes to the prr project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Features
- Add git progress feedback to stderr during mirror clone/fetch, and print a review commencement message when the review engine starts
- Silent probe for available PR refs using `git ls-remote` before any fetch
- Eliminate spurious `fatal:` messages from git on draft PRs
- Implement `Enricher` interface with separate CLI and REST backends for GitHub and Azure DevOps

### Bug Fixes
- Add `--progress` flag to `git clone` and `git fetch` so progress output is shown on non-TTY stderr (e.g. piped or CI environments)
- Fix `install.ps1`: strip `v` prefix from version string when constructing the archive filename
- Use `git merge-base HEAD <baseSHA>` to derive the diff range on head-ref and source-branch fallback paths, preventing the diff from including unrelated upstream commits; show Copilot quota errors without credential redaction

## v0.3.0 - 2026-03-19

### Features
- Separate issue and PR summaries into distinct AI-generated fields

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
