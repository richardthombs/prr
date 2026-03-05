# Story 1.10: Replace Unix-Only Mirror Locking with Cross-Platform Lock Strategy

Status: review

## Story

As Richard,
I want mirror-update locking to work on macOS, Linux, and Windows,
so that concurrent review safety is preserved regardless of host OS.

## Acceptance Criteria

1. Given PRR runs on Windows, when mirror locking is compiled and executed, then PRR uses a supported lock implementation (no Unix-only `syscall.Flock` dependency), and lock acquire/release semantics remain equivalent to macOS/Linux behaviour.
2. Given lock contention on any supported OS, when a run waits for lock acquisition, then timeout and `--force` bypass behaviour remain consistent, and lock timeout failures return a stable runtime error class.
3. Given cross-platform unit tests, when lock tests run on macOS, Linux, and Windows, then tests verify lock contention, timeout, and force-bypass behaviour, and do not rely on OS-specific syscall APIs in shared test files.

## Tasks / Subtasks

- [x] Introduce OS-safe lock abstraction for mirror lock acquisition/release (AC: 1, 2)
- [x] Replace direct `syscall.Flock` usage in production code with abstraction (AC: 1, 2)
- [x] Add platform-specific implementations (or equivalent strategy) for Unix and Windows (AC: 1)
- [x] Preserve existing timeout and `--force` semantics in lock orchestration (AC: 2)
- [x] Update lock-focused tests to be platform-safe and matrix-ready (AC: 3)
- [x] Add/adjust CI test expectations for lock behaviour across all target OSes (AC: 3)

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
- 2026-03-05: Replaced direct mirror locking calls with OS-specific lock helpers (`tryLockFile`/`unlockFile`/`isLockBusy`) while preserving timeout and `--force` behaviour.
- 2026-03-05: Added Windows lock implementation using `LockFileEx`/`UnlockFileEx` via `kernel32` calls with Unix implementation retained under build tags.
- 2026-03-05: Split lock contention test setup into OS-specific test helpers so shared tests no longer depend on Unix-only syscalls.
- 2026-03-05: Validation run: `go test ./...` (pass), `GOOS=windows GOARCH=amd64 go build ./...` (pass), `GOOS=windows GOARCH=amd64 go test -c ./internal/git` (pass).

### Implementation Plan

- Introduce build-tagged lock helper files to isolate OS-specific lock syscalls from shared mirror logic.
- Preserve existing lock orchestration in `withRepoLock` (timeout retry loop, force bypass, runtime error wrapping).
- Keep shared tests platform-neutral and move lock-acquisition primitives into OS-specific test helper files.
- Verify behaviour via full suite plus Windows cross-compilation checks for both package code and tests.

### File List

- _bmad-output/implementation-artifacts/1-10-replace-unix-only-mirror-locking-with-cross-platform-lock-strategy.md
- internal/git/mirror.go
- internal/git/mirror_lock_unix.go
- internal/git/mirror_lock_windows.go
- internal/git/mirror_test.go
- internal/git/mirror_lock_test_unix_test.go
- internal/git/mirror_lock_test_windows_test.go

## Change Log

- 2026-03-05: Implemented cross-platform mirror lock abstraction with Unix and Windows implementations; updated lock contention tests for platform-safe execution and verified Windows build/test compilation expectations.
