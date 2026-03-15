# Changelog

All notable changes to prr are documented here.
This file is updated automatically by GoReleaser from conventional commit messages on each release.

<!-- GoReleaser appends release notes above this line -->

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
