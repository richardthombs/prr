# Story 1.11: Normalise Cross-Platform Path and Test Contracts

Status: ready-for-dev

## Story

As Richard,
I want path handling and command contracts to be OS-agnostic,
so that PRR behaves consistently with Windows path separators and shell environments.

## Acceptance Criteria

1. Given cache, mirror, and worktree paths, when PRR resolves and emits filesystem locations, then paths are created using platform-safe path APIs, and JSON payload fields remain deterministic and valid on macOS, Linux, and Windows.
2. Given unit tests for mirror/worktree path behaviour, when tests run on Windows, then assertions avoid hard-coded `/tmp` and `/` separator assumptions, and use `filepath`-safe expectations to validate deterministic path segments.
3. Given user-facing examples and diagnostics, when commands/logs are reviewed across platforms, then examples avoid Unix-only filesystem assumptions, and command diagnostics remain script-compatible for automation.

## Tasks / Subtasks

- [ ] Audit path construction/output paths in git workspace and command layers (AC: 1)
- [ ] Replace separator-sensitive assertions in mirror/worktree tests (AC: 2)
- [ ] Replace Unix-specific temp path fixtures with `t.TempDir()` or `filepath` contracts (AC: 2)
- [ ] Update docs/examples where Unix-only path assumptions are shown (AC: 3)
- [ ] Verify JSON contracts remain stable while path values are OS-correct (AC: 1, 3)

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

### File List

- _bmad-output/implementation-artifacts/1-11-normalise-cross-platform-path-and-test-contracts.md
