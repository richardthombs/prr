# Story 1.12: Add Cross-OS Build and Smoke Verification for CLI Baseline

Status: ready-for-dev

## Story

As Richard,
I want automated build/smoke verification across macOS, Linux, and Windows,
so that CLI entry and composable command contracts are validated before release.

## Acceptance Criteria

1. Given CI build verification, when pull requests are validated, then `go build ./...` and `go test ./...` run in a matrix for macOS, Linux, and Windows, and failures are reported per OS target.
2. Given the produced CLI binary per OS, when smoke checks run, then `prr --help`, `prr version`, and one `--what-if` composable command path succeed, and stdout/stderr contracts remain stable for scripts.
3. Given local developer workflows, when contributors run documented build/test commands, then instructions use cross-platform Go commands as the source of truth, and Unix-only helper commands are marked optional.

## Tasks / Subtasks

- [ ] Add CI workflow matrix for macOS, Linux, and Windows (AC: 1)
- [ ] Run build and unit test steps for each OS target (AC: 1)
- [ ] Add smoke checks for CLI baseline commands and one composable what-if path (AC: 2)
- [ ] Ensure smoke checks validate stdout/stderr contract expectations (AC: 2)
- [ ] Update contributor guidance for cross-platform build/test commands (AC: 3)
- [ ] Mark Unix-only helper tooling (for example Makefile shell commands) as optional (AC: 3)

## Dev Notes

- There is currently no `.github/workflows` CI matrix in the repository.
- Existing Makefile helper commands include Unix shell assumptions; this story should avoid making those the canonical cross-platform path.

## References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.12)
- Existing task definitions: workspace tasks for `go build ./...` and `go test ./...`
- Current helper tooling: `Makefile`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Completion Notes List

- 2026-03-05: Story file created from approved Epic 1 cross-platform additions.

### File List

- _bmad-output/implementation-artifacts/1-12-add-cross-os-build-and-smoke-verification-for-cli-baseline.md
