# Story 1.12: Add Cross-OS Build and Smoke Verification for CLI Baseline

Status: done

## Story

As Richard,
I want automated build/smoke verification across macOS, Linux, and Windows,
so that CLI entry and composable command contracts are validated before release.

## Acceptance Criteria

1. Given CI build verification, when pull requests are validated, then `go build ./...` and `go test ./...` run in a matrix for macOS, Linux, and Windows, and failures are reported per OS target.
2. Given the produced CLI binary per OS, when smoke checks run, then `prr --help`, `prr version`, and one `--what-if` composable command path succeed, and stdout/stderr contracts remain stable for scripts.
3. Given local developer workflows, when contributors run documented build/test commands, then instructions use cross-platform Go commands as the source of truth, and Unix-only helper commands are marked optional.

## Tasks / Subtasks

- [x] Add CI workflow matrix for macOS, Linux, and Windows (AC: 1)
- [x] Run build and unit test steps for each OS target (AC: 1)
- [x] Add smoke checks for CLI baseline commands and one composable what-if path (AC: 2)
- [x] Ensure smoke checks validate stdout/stderr contract expectations (AC: 2)
- [x] Update contributor guidance for cross-platform build/test commands (AC: 3)
- [x] Mark Unix-only helper tooling (for example Makefile shell commands) as optional (AC: 3)

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
- 2026-03-05: Added `.github/workflows/cross-os-build-smoke.yml` with a macOS/Linux/Windows matrix that runs `go build ./...` and `go test ./...` for each target.
- 2026-03-05: Added smoke checks for `prr --help`, `prr version`, and `prr checkout <PR_URL> --what-if`, including stdout/stderr contract assertions for script stability.
- 2026-03-05: Updated contributor guidance in `README.md` to make cross-platform Go commands canonical and documented `Makefile` helpers as optional Unix-like conveniences.
- 2026-03-05: Validation run included repository build/test tasks and local smoke verification of `checkout --what-if` command tracing and JSON output.
- 2026-03-05: Code review fixes applied: removed undocumented `diff`/`bundle` command entries from README, made the workflow binary build step shell-safe on Windows, and tightened help/version stdout+stderr contract assertions.

### File List

- .github/workflows/cross-os-build-smoke.yml
- README.md
- Makefile
- _bmad-output/implementation-artifacts/sprint-status.yaml
- _bmad-output/implementation-artifacts/1-12-add-cross-os-build-and-smoke-verification-for-cli-baseline.md

## Change Log

- 2026-03-05: Implemented cross-OS CI build/test + smoke checks and updated cross-platform contributor guidance.
- 2026-03-05: Senior review remediation completed for CI portability and CLI documentation contract alignment.

## Senior Developer Review (AI)

### Review Date

2026-03-05

### Reviewer

GPT-5.3-Codex

### Outcome

Approve

### Summary

Adversarial review identified 2 High and 2 Medium issues. All findings were fixed in this session and revalidated.

### Action Items

- [x] [HIGH] Replace Unix-only `mkdir -p` in CI smoke binary build step with cross-platform PowerShell directory creation (`.github/workflows/cross-os-build-smoke.yml`).
- [x] [HIGH] Remove README command entries for unavailable `prr diff` and `prr bundle` commands (`README.md`).
- [x] [MEDIUM] Ensure story File List includes sprint tracking file changed during workflow execution (`_bmad-output/implementation-artifacts/1-12-add-cross-os-build-and-smoke-verification-for-cli-baseline.md`).
- [x] [MEDIUM] Tighten smoke contract checks to assert stderr behaviour for `prr --help` and `prr version` rather than suppressing stderr (`.github/workflows/cross-os-build-smoke.yml`).
