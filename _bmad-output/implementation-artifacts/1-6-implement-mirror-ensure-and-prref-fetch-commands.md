# Story 1.6: Implement Mirror Ensure and PRRef Fetch Commands

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want `prr mirror ensure` and `prr prref fetch` commands,
so that mirror lifecycle and merge-ref acquisition are composable and testable.

## Acceptance Criteria

1. Given repo context input, when I run `prr mirror ensure`, then PRR creates or updates the deterministic mirror location, and emits JSON including `bareDir`.
2. Given valid PR context and mirror state, when I run `prr prref fetch`, then PRR fetches merge ref into `refs/prr/pull/<PR_ID>/merge`, and emits JSON including resolved `mergeRef`.

## Tasks / Subtasks

- [ ] Add `mirror ensure` and `prref fetch` command wiring (AC: 1, 2)
  - [ ] Add `cmd/prr/mirror_ensure.go`
  - [ ] Add `cmd/prr/prref_fetch.go`
  - [ ] Register commands and required flags/input handling
- [ ] Implement mirror ensure command path (AC: 1)
  - [ ] Ensure deterministic bare mirror path creation/reuse
  - [ ] Update mirror state before returning success payload
  - [ ] Return JSON payload containing `bareDir`
- [ ] Implement PR merge ref fetch command path (AC: 2)
  - [ ] Fetch provider merge ref into `refs/prr/pull/<PR_ID>/merge`
  - [ ] Return JSON payload containing `mergeRef` and relevant context
- [ ] Preserve concurrency and safety boundaries (AC: 1, 2)
  - [ ] Use lock-safe mirror update path (no ad-hoc locking in command handler)
  - [ ] Ensure clear behaviour on missing merge ref/provider constraints
- [ ] Add tests and diagnostics checks (AC: 1, 2)
  - [ ] Unit tests for command input/flag handling
  - [ ] Integration-focused tests for mirror path and fetch ref destination
  - [ ] Error classification tests for provider/ref failures

## Dev Notes

- This story operationalises stages covered later by Epic 2 as explicit command surface; keep behaviour aligned with deterministic mirror/worktree architecture.
- Commands are boundary adapters only; mirror/fetch logic remains in `internal/git` and `internal/provider` abstractions.
- Avoid side effects outside mirror/fetch responsibilities.

### Technical Requirements

- Command model:
  - `prr mirror ensure`
  - `prr prref fetch`
- Must emit JSON-compatible outputs for composable piping.
- Merge ref target must follow internal namespace convention.

### Architecture Compliance

- Mirror/worktree state remains filesystem-based only; no database.
- Per-repository locking strategy must be preserved.
- Error and exit mapping centralised; no bespoke command-level mapping.

### Library / Framework Requirements

- Cobra command wiring patterns consistent with scaffold.
- Git operations executed via internal git abstraction and system git commands where designed.

### File Structure Requirements

- New command files expected under `cmd/prr/`.
- Implementation modules expected under `internal/git`, `internal/provider`, `internal/types`, `internal/errors`.

### Testing Requirements

- Verify deterministic mirror path contract.
- Verify ref fetch destination (`refs/prr/pull/<PR_ID>/merge`).
- Validate actionable diagnostics for missing merge ref and runtime failures.

### Previous Story Intelligence

- Story `1.1` provides command wiring baseline and scaffold conventions.
- Story `1.5` should establish composable JSON contract patterns; reuse these conventions.

### Git Intelligence Summary

- Repository currently has foundational CLI only; this story introduces first Git-stateful composable commands and should set pattern for subsequent pipeline commands.

### Project Structure Notes

- Keep command logic in `cmd/prr`, domain logic under `internal/*`.
- Maintain deterministic naming and path conventions from architecture.

### References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.6)
- Mirror/fetch semantics: `docs/initial_specification.md` (Repository Management, PR Snapshot Fetching)
- MVP safety and deterministic constraints: `_bmad-output/planning-artifacts/prd.md`
- Boundary and naming conventions: `_bmad-output/planning-artifacts/architecture.md`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow executed from approved sprint change proposal.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story selected explicitly from approved change set: `1-6-implement-mirror-ensure-and-prref-fetch-commands`.

### File List

- _bmad-output/implementation-artifacts/1-6-implement-mirror-ensure-and-prref-fetch-commands.md
