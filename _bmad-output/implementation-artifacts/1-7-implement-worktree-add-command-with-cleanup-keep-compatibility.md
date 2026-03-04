# Story 1.7: Implement Worktree Add Command with Cleanup/Keep Compatibility

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want `prr worktree add` to create isolated review worktrees,
so that workspace lifecycle can be controlled independently and safely.

## Acceptance Criteria

1. Given a valid mirror and merge ref, when I run `prr worktree add`, then PRR creates a detached isolated worktree and emits `workDir`, and no writes are performed in the active working copy.
2. Given default cleanup behaviour and `--keep` override, when this command is used in the review chain, then lifecycle behaviour remains consistent with documented cleanup semantics.
3. Given `--verbose` is enabled, when `prr worktree add` invokes external commands, then PRR logs each command to stderr before execution.
4. Given `--what-if` is enabled, when `prr worktree add` runs, then PRR prints external commands it would execute and performs no external command execution.

## Tasks / Subtasks

- [ ] Add `worktree add` Cobra command surface (AC: 1, 2, 3, 4)
  - [ ] Create `cmd/prr/worktree_add.go` and register under `worktree` parent command
  - [ ] Support equivalent flag inputs for composable use (`--bare-dir`, `--merge-ref`, `--pr-id`, `--repo`, `--keep`, `--verbose`, `--what-if`)
  - [ ] Emit JSON to stdout only (stderr for diagnostics) with deterministic keys including `workDir`
- [ ] Implement worktree lifecycle service in `internal/git` (AC: 1, 2)
  - [ ] Add/create `internal/git/worktree.go` with detached create and remove/prune operations
  - [ ] Use deterministic default path `~/.cache/prr/work/<repoHash>/pr-<PR_ID>/<runId>/`
  - [ ] Use merge ref contract `refs/prr/pull/<PR_ID>/merge` as primary input target
- [ ] Enforce safety and composability constraints (AC: 1, 2)
  - [ ] Ensure implementation never writes to user active working copy paths
  - [ ] Ensure clean up by default and explicit retention only via `--keep`
  - [ ] Return actionable typed errors via central `internal/errors` mapping
- [ ] Implement observability parity with prior commands (AC: 3, 4)
  - [ ] Match `mirror ensure`/`prref fetch` pattern: `exec: ...` lines on stderr before command execution
  - [ ] In `--what-if`, print planned git commands and skip external execution while still returning valid JSON contract
- [ ] Add tests for command and git-layer contracts (AC: 1, 2, 3, 4)
  - [ ] Add command tests in `cmd/prr/*_test.go` for required flags, stdout JSON shape, and stderr verbose/what-if behaviour
  - [ ] Add git service tests in `internal/git/*_test.go` for detached add, remove/prune, and failure classification
  - [ ] Add regression tests proving no external execution in what-if mode

## Dev Notes

- This story is the worktree-stage contract for the composable pipeline between `prref fetch` and `diff`.
- Keep command handlers thin (`cmd/prr/*`) and delegate all git operations to `internal/git` services.
- Maintain strict stdout/stderr channel discipline used by existing command stories:
  - stdout: machine-readable JSON payload
  - stderr: verbose/what-if diagnostics and external command previews
- Ensure behaviour is deterministic and safe for reruns; no hidden mutable process state.

### Technical Requirements

- Command surface: `prr worktree add`.
- Required git operation for create: `git -C <bareDir> worktree add --detach <workDir> refs/prr/pull/<PR_ID>/merge`.
- Required cleanup operation (when not keep): `git -C <bareDir> worktree remove --force <workDir>` then `git -C <bareDir> worktree prune`.
- Must return a JSON-compatible workspace payload containing at least `bareDir`, `workDir`, `mergeRef`.
- Must maintain isolation guarantee: no writes to the active user working copy.
- Must support `--keep` to preserve worktree for inspection and report retained path.
- Must support `--verbose` and `--what-if` with pre-execution command visibility on stderr.

### Architecture Compliance

- `cmd/prr/worktree_add.go` is a boundary adapter only.
- `internal/git` owns lifecycle details (create/remove/prune/path generation).
- `internal/errors` remains single source for class mapping and exit codes.
- No database or metadata persistence added; filesystem mirror/worktree state only.

### Library / Framework Requirements

- Use Cobra command patterns already established in existing command files.
- Reuse `git.Service` + `Runner` execution model and `EnsureOptions`-style observability toggles where practical.
- Avoid introducing new CLI or process execution libraries.

### File Structure Requirements

- Add `cmd/prr/worktree_add.go`.
- Add/extend `internal/git/worktree.go` and relevant tests.
- Update command registration in existing command init flow only; avoid unrelated tree reshaping.

### Testing Requirements

- Validate JSON payload contract fields and casing.
- Validate detached add command arguments and merge-ref target.
- Validate cleanup default path and retained-path behaviour under `--keep`.
- Validate stderr command previews for `--verbose` and no-op execution for `--what-if`.
- Validate failure class mapping (`CONFIG_*`, `RUNTIME_*`, `PROVIDER_*` where applicable).

### Previous Story Intelligence

- Story `1.6` established useful implementation conventions to preserve:
  - command-level `--verbose` and `--what-if` parity
  - stderr `exec: ...` previews before external command execution
  - JSON payloads kept clean on stdout
  - tests proving what-if mode performs no external command execution
- Reuse existing mirror/prref command/test style to avoid introducing divergent command contracts.

### Git Intelligence Summary

- Recent commit history could not be retrieved from the current terminal session; project intelligence is derived from current source and completed Story 1.6 artifact.
- Current codebase already contains mirror/fetch contracts and option-aware git execution paths; this story should extend those patterns rather than reinvent lifecycle abstractions.

### Project Structure Notes

- Keep command boundary explicit (`cmd/prr/*`) and implementation modules under `internal/*`.
- Preserve deterministic naming and path conventions for mirrors, refs, and worktree paths.
- No UX/doc/theme additions are needed; scope is command contract + lifecycle implementation only.

### Latest Technical Information

- No external web research was performed in this run.
- Use currently adopted stack and architecture decisions in repository artifacts (Go + Cobra + system git) for this story.

### References

- `_bmad-output/planning-artifacts/epics.md` (Story 1.7)
- `docs/initial_specification.md` (Workspace Management, CLI command model, Workspace JSON structure)
- `_bmad-output/planning-artifacts/prd.md` (FR10–FR13, NFR6, NFR9, command model and cleanup/keep behaviour)
- `_bmad-output/planning-artifacts/architecture.md` (boundary ownership, command surface, deterministic filesystem state)
- `_bmad-output/implementation-artifacts/1-6-implement-mirror-ensure-and-prref-fetch-commands.md` (prior story patterns and lessons)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow executed from approved sprint change proposal.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story selected explicitly from approved change set: `1-7-implement-worktree-add-command-with-cleanup-keep-compatibility`.
- Story context refreshed with explicit command-level guardrails, deterministic workspace path contract, and prior-story implementation intelligence.

### File List

- _bmad-output/implementation-artifacts/1-7-implement-worktree-add-command-with-cleanup-keep-compatibility.md
