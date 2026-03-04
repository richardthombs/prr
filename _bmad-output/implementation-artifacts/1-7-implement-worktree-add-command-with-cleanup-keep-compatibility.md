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
4. Given `--what-if` is enabled, when `prr worktree add` runs, then PRR prints external commands it would execute and does not execute them.

## Tasks / Subtasks

- [ ] Add `worktree add` command wiring (AC: 1, 2)
  - [ ] Create `cmd/prr/worktree_add.go`
  - [ ] Register command and required input flags/JSON payload handling
- [ ] Implement isolated worktree creation path (AC: 1)
  - [ ] Create detached worktree at merge ref from mirror context
  - [ ] Emit JSON output with `workDir` and related workspace context
- [ ] Implement lifecycle compatibility and cleanup semantics (AC: 2)
  - [ ] Ensure command integrates with default cleanup model
  - [ ] Ensure `--keep` behaviour remains explicit and deterministic
- [ ] Enforce isolation guarantees and diagnostics (AC: 1, 2)
  - [ ] Validate no writes occur in active local working copy
  - [ ] Emit actionable diagnostics for workspace creation/cleanup failures
- [ ] Add tests for workspace and keep/cleanup behaviour (AC: 1, 2)
  - [ ] Unit tests for command argument and payload validation
  - [ ] Integration tests for worktree creation path and lifecycle handling
- [ ] Add command observability and dry-run behaviour (AC: 3, 4)
  - [ ] Add `--verbose` command logging for all external commands before execution
  - [ ] Add `--what-if` mode that prints commands without executing
  - [ ] Add tests covering verbose logging and what-if no-execution guarantees

## Dev Notes

- Command must preserve isolation guarantees from architecture and initial specification.
- Worktree location/path semantics should align with deterministic cache/worktree strategy.
- Keep this command composable and script-friendly; output must be stable JSON where applicable.

### Technical Requirements

- Command signature: `prr worktree add`.
- Worktree must be detached and created from merge ref under mirror context.
- Must support compatibility with default cleanup and `--keep` retention behaviour.
- Must support `--verbose` pre-execution command logging for external commands.
- Must support `--what-if` dry-run mode that prints external commands and performs no external mutations.

### Architecture Compliance

- `internal/git` owns worktree lifecycle operations.
- `cmd/prr/worktree_add.go` must remain thin adapter.
- Error handling, logging, and exit mapping remain centralised.

### Library / Framework Requirements

- Cobra command wiring.
- Use existing/system git integration boundary from internal git module.

### File Structure Requirements

- New command file under `cmd/prr/`.
- Worktree logic under `internal/git/worktree.go` (or equivalent module path in architecture).

### Testing Requirements

- Validate detached worktree creation and deterministic path handling.
- Validate non-interference with active local working copy.
- Validate cleanup defaults and `--keep` branch behaviour.

### Previous Story Intelligence

- Story `1.1` established command scaffold and pattern.
- Story `1.6` should establish mirror/ref command contracts used as input to this story.

### Git Intelligence Summary

- This story introduces explicit workspace lifecycle command boundary and should establish a reliable contract for downstream `diff` and `bundle` commands.

### Project Structure Notes

- Keep command boundary explicit (`cmd/prr/*`) and implementation modules under `internal/*`.
- Preserve deterministic naming and path conventions.

### References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.7)
- Worktree management semantics: `docs/initial_specification.md` (Workspace Management)
- Isolation and cleanup constraints: `_bmad-output/planning-artifacts/prd.md`
- Module boundaries and patterns: `_bmad-output/planning-artifacts/architecture.md`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow executed from approved sprint change proposal.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story selected explicitly from approved change set: `1-7-implement-worktree-add-command-with-cleanup-keep-compatibility`.

### File List

- _bmad-output/implementation-artifacts/1-7-implement-worktree-add-command-with-cleanup-keep-compatibility.md
