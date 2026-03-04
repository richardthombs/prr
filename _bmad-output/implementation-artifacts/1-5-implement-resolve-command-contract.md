# Story 1.5: Implement Resolve Command Contract

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want `prr resolve <PR_ID>` to emit deterministic PR reference context,
so that I can script and verify context resolution independently of the full review flow.

## Acceptance Criteria

1. Given a valid PR identifier and resolvable context, when I run `prr resolve <PR_ID>`, then PRR emits a stable JSON `PRRef` payload, and supports equivalent override flags.
2. Given missing or invalid context inputs, when resolution fails, then PRR returns actionable diagnostics, and exits with a stable error-class code.

## Tasks / Subtasks

- [ ] Add `resolve` command wiring under `cmd/prr/` (AC: 1, 2)
  - [ ] Create `cmd/prr/resolve.go` and register command with root command
  - [ ] Accept `<PR_ID>` argument and validate argument count/type
  - [ ] Add override flags for provider/repo/remote as required by current config model
- [ ] Define and expose `PRRef` contract for command output (AC: 1)
  - [ ] Implement/confirm typed structure for `prId`, `repoUrl`, `remote`, `provider`
  - [ ] Emit deterministic JSON fields in camelCase
- [ ] Implement resolution execution path with provider abstraction (AC: 1, 2)
  - [ ] Call provider resolution boundary only (no direct provider-specific logic in command)
  - [ ] Surface successful payload to stdout in automation-safe format
- [ ] Implement structured failure path and exit mapping (AC: 2)
  - [ ] Classify failures into documented error classes
  - [ ] Ensure actionable user-facing diagnostics without leaking sensitive values
- [ ] Add tests and baseline validation (AC: 1, 2)
  - [ ] Unit tests for arg/flag validation and JSON output shape
  - [ ] Tests for failure diagnostics and non-zero exit behaviour

## Dev Notes

- Keep this command a thin boundary adapter in `cmd/prr/`; orchestration and business logic stays in `internal/*`.
- Output contract must remain stable for shell/CI use and align with composable command intent in specification.
- Do not introduce ad-hoc exit logic in command handlers; use centralised error mapping module.
- Preserve deterministic behaviour: same input context should produce functionally equivalent resolved payload.

### Technical Requirements

- Command signature: `prr resolve <PR_ID>`.
- Support equivalent flags and automation-friendly IO conventions.
- JSON contract must follow camelCase field naming.
- Keep provider logic behind `PRProvider` boundary.

### Architecture Compliance

- CLI boundary: `cmd/prr/*` only wires command and delegates.
- Provider abstraction boundary: `internal/provider/*` handles context resolution details.
- Error class mapping must use central error taxonomy (`CONFIG_*`, `PROVIDER_*`, `RUNTIME_*` as applicable).

### Library / Framework Requirements

- Continue using Cobra (`spf13/cobra`) command model.
- No alternate CLI framework.

### File Structure Requirements

- Expected command file: `cmd/prr/resolve.go`.
- Related contracts/services should live under `internal/types`, `internal/provider`, `internal/errors`, and `internal/config` as needed.

### Testing Requirements

- Unit test command argument validation and JSON output schema shape.
- Test deterministic field output order/shape assumptions where practical.
- Test stable failure classification and non-zero exits.

### Previous Story Intelligence

- Story `1.1` established Cobra scaffold, root command wiring, and placeholder command pattern under `cmd/prr/`.
- Reuse established command registration pattern and minimal command-surface style.

### Git Intelligence Summary

- Current repository has initial CLI scaffold and baseline tests only; this story should establish first substantive composable command behaviour without expanding into unrelated pipeline stages.

### Project Structure Notes

- Maintain strict boundary: command wiring in `cmd/prr`, implementation in `internal/*`.
- Keep naming conventions consistent with architecture (`kebab-case` flags, `UPPER_SNAKE_CASE` env vars, camelCase JSON output fields).

### References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.5)
- Command model and composable commands: `docs/initial_specification.md` (CLI Command Model)
- PRRef schema and provider contract: `docs/initial_specification.md` (Data Structures / Interfaces)
- MVP and automation constraints: `_bmad-output/planning-artifacts/prd.md` (CLI Tool Specific Requirements)
- Command boundary conventions: `_bmad-output/planning-artifacts/architecture.md` (Project Structure & Boundaries)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow executed from approved sprint change proposal.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story selected explicitly from approved change set: `1-5-implement-resolve-command-contract`.

### File List

- _bmad-output/implementation-artifacts/1-5-implement-resolve-command-contract.md
