# Story 1.11: Normalise Cross-Platform Path and Test Contracts

Status: done

## Story

As Richard,
I want path handling and command contracts to be OS-agnostic,
so that PRR behaves consistently with Windows path separators and shell environments.

## Acceptance Criteria

1. Given cache, mirror, and worktree paths, when PRR resolves and emits filesystem locations, then paths are created using platform-safe path APIs, and JSON payload fields remain deterministic and valid on macOS, Linux, and Windows.
2. Given unit tests for mirror/worktree path behaviour, when tests run on Windows, then assertions avoid hard-coded `/tmp` and `/` separator assumptions, and use `filepath`-safe expectations to validate deterministic path segments.
3. Given user-facing examples and diagnostics, when commands/logs are reviewed across platforms, then examples avoid Unix-only filesystem assumptions, and command diagnostics remain script-compatible for automation.

## Tasks / Subtasks

- [x] Audit path construction/output paths in git workspace and command layers (AC: 1)
- [x] Replace separator-sensitive assertions in mirror/worktree tests (AC: 2)
- [x] Replace Unix-specific temp path fixtures with `t.TempDir()` or `filepath` contracts (AC: 2)
- [x] Update docs/examples where Unix-only path assumptions are shown (AC: 3)
- [x] Verify JSON contracts remain stable while path values are OS-correct (AC: 1, 3)

## Dev Notes

- Current tests in `internal/git/mirror_test.go` and `internal/git/worktree_test.go` use hard-coded `/tmp` and `/` checks that are likely to fail on Windows.
- Preserve deterministic naming semantics while removing separator assumptions.

## References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.11)
- Relevant code: `internal/git/worktree.go`, `internal/git/mirror.go`
- Relevant tests: `internal/git/worktree_test.go`, `internal/git/mirror_test.go`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Completion Notes List

- 2026-03-05: Story file created from approved Epic 1 cross-platform additions.
- 2026-03-05: Audited mirror and worktree path construction; production path handling already used `filepath` APIs and deterministic naming.
- 2026-03-05: Updated `internal/git/worktree_test.go` to remove Unix-only `/tmp` fixtures and slash-based expectations, replacing them with `t.TempDir()` and `filepath`-safe assertions.
- 2026-03-05: Updated `internal/git/mirror_test.go` path assertions to use `filepath.Base` for OS-agnostic deterministic checks.
- 2026-03-05: Updated documentation examples in `docs/initial_specification.md` from `~/.cache/...` to `<user-cache-dir>/...` for cross-platform guidance.
- 2026-03-05: Validation run completed with green regression (`go test ./...`) and successful build verification (`go build ./...`).

### File List

- _bmad-output/implementation-artifacts/1-11-normalise-cross-platform-path-and-test-contracts.md
- _bmad-output/implementation-artifacts/sprint-status.yaml
- docs/initial_specification.md
- internal/git/mirror_test.go
- internal/git/worktree_test.go

## Change Log

- 2026-03-05: Completed Story 1.11 by normalising path-related test contracts for cross-platform execution and updating path examples to platform-neutral cache-dir notation.
