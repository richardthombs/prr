# Story 1.10: Replace Unix-Only Mirror Locking with Cross-Platform Lock Strategy

Status: ready-for-dev

## Story

As Richard,
I want mirror-update locking to work on macOS, Linux, and Windows,
so that concurrent review safety is preserved regardless of host OS.

## Acceptance Criteria

1. Given PRR runs on Windows, when mirror locking is compiled and executed, then PRR uses a supported lock implementation (no Unix-only `syscall.Flock` dependency), and lock acquire/release semantics remain equivalent to macOS/Linux behaviour.
2. Given lock contention on any supported OS, when a run waits for lock acquisition, then timeout and `--force` bypass behaviour remain consistent, and lock timeout failures return a stable runtime error class.
3. Given cross-platform unit tests, when lock tests run on macOS, Linux, and Windows, then tests verify lock contention, timeout, and force-bypass behaviour, and do not rely on OS-specific syscall APIs in shared test files.

## Tasks / Subtasks

- [ ] Introduce OS-safe lock abstraction for mirror lock acquisition/release (AC: 1, 2)
- [ ] Replace direct `syscall.Flock` usage in production code with abstraction (AC: 1, 2)
- [ ] Add platform-specific implementations (or equivalent strategy) for Unix and Windows (AC: 1)
- [ ] Preserve existing timeout and `--force` semantics in lock orchestration (AC: 2)
- [ ] Update lock-focused tests to be platform-safe and matrix-ready (AC: 3)
- [ ] Add/adjust CI test expectations for lock behaviour across all target OSes (AC: 3)

## Dev Notes

- Root issue identified in `internal/git/mirror.go` where mirror locking currently uses Unix-specific syscall APIs.
- Maintain current error taxonomy and messaging contracts (`RUNTIME_*`) while changing lock internals.
- Keep command-level behaviour unchanged; this story is an internal portability upgrade.

## References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.10)
- Current lock implementation: `internal/git/mirror.go`
- Related tests: `internal/git/mirror_test.go`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Completion Notes List

- 2026-03-05: Story file created from approved Epic 1 cross-platform additions.

### File List

- _bmad-output/implementation-artifacts/1-10-replace-unix-only-mirror-locking-with-cross-platform-lock-strategy.md
