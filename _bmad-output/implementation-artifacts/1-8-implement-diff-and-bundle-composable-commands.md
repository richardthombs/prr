# Story 1.8: Implement Diff and Bundle Composable Commands

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want `prr diff` and `prr bundle` commands,
so that deterministic review inputs can be generated and validated in composable stages.

## Acceptance Criteria

1. Given a valid worktree, when I run `prr diff`, then PRR emits deterministic stat/files/patch outputs, and output contracts are JSON-compatible.
2. Given valid diff outputs, when I run `prr bundle`, then PRR emits a validated v1 bundle payload, and enforces configured size limits with explicit failure diagnostics.

## Tasks / Subtasks

- [ ] Add `diff` and `bundle` command wiring (AC: 1, 2)
  - [ ] Create `cmd/prr/diff.go`
  - [ ] Create `cmd/prr/bundle.go`
  - [ ] Register commands with input/flag handling for composable use
- [ ] Implement deterministic diff command path (AC: 1)
  - [ ] Compute `HEAD^1..HEAD` outputs (stat, files, patch)
  - [ ] Emit JSON-compatible payload suitable for piping
- [ ] Implement bundle assembly command path (AC: 2)
  - [ ] Build v1 bundle with required meta/stat/files/patch fields
  - [ ] Validate schema and return deterministic field shape
- [ ] Enforce input size limits and diagnostics (AC: 2)
  - [ ] Enforce max patch bytes and max changed files limits
  - [ ] Return explicit limit diagnostics and stable exit class on failure
- [ ] Add tests for deterministic output and bundle contract (AC: 1, 2)
  - [ ] Unit tests for diff and bundle payload shape
  - [ ] Tests for limit exceeded failure behaviour
  - [ ] Regression tests for deterministic behaviour on unchanged refs

## Dev Notes

- This story is command-surface realisation of Epic 3 capability boundaries; retain deterministic input-generation guarantees.
- `diff` and `bundle` should remain composable and automation-safe with stable JSON contracts.
- Avoid introducing incremental review logic (explicitly out of v1 scope).

### Technical Requirements

- Commands:
  - `prr diff`
  - `prr bundle`
- Diff semantics must reflect merge-parent range behaviour.
- Bundle must include required v1 fields and pass schema validation.
- Size limits must be configurable and enforced before engine invocation.

### Architecture Compliance

- Diff operations remain in `internal/git/diff.go` (or equivalent).
- Bundle/limit enforcement remains in `internal/bundle/*`.
- Command handlers remain thin and delegate to internal modules.

### Library / Framework Requirements

- Cobra for command surface.
- JSON output contract in camelCase.

### File Structure Requirements

- New command files under `cmd/prr/`.
- Supporting implementation under `internal/git`, `internal/bundle`, `internal/types`, `internal/errors`.

### Testing Requirements

- Validate deterministic rerun output equivalence for unchanged refs.
- Validate schema and limit-failure diagnostics.
- Validate stdout/stderr and exit behaviour for automation stability.

### Previous Story Intelligence

- Stories `1.6` and `1.7` provide mirror/ref/worktree prerequisites.
- Keep composable contract format consistent with prior command stories.

### Git Intelligence Summary

- This story establishes deterministic input payload generation for review-engine stage and should be treated as a contract-hardening milestone.

### Project Structure Notes

- Strictly separate CLI command boundary and internal domain logic.
- Preserve naming and payload conventions from architecture.

### References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.8)
- Diff and bundle v1 requirements: `docs/initial_specification.md` (Diff Generation, Review Bundle)
- MVP limits and deterministic constraints: `_bmad-output/planning-artifacts/prd.md`
- Boundary and contract guidance: `_bmad-output/planning-artifacts/architecture.md`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow executed from approved sprint change proposal.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story selected explicitly from approved change set: `1-8-implement-diff-and-bundle-composable-commands`.

### File List

- _bmad-output/implementation-artifacts/1-8-implement-diff-and-bundle-composable-commands.md
