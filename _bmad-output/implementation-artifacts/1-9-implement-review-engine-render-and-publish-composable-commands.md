# Story 1.9: Implement Review-Engine, Render, and Publish Composable Commands

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want `prr review-engine`, `prr render`, and `prr publish` commands,
so that review execution, output rendering, and optional publication are composable for scripting and diagnostics.

## Acceptance Criteria

1. Given a valid review bundle, when I run `prr review-engine`, then PRR emits structured review JSON with stable per-run finding references, and engine failures return actionable, classed errors.
2. Given a valid review JSON payload, when I run `prr render`, then PRR outputs Markdown by default and JSON when requested, and channel/exit behaviour remains automation-stable.
3. Given publish mode and provider support, when I run `prr publish`, then PRR posts rendered review output to the PR, and reports publication outcome explicitly.
4. Given `--verbose` is enabled on these commands, when external commands are invoked, then PRR logs each command to stderr before execution.
5. Given `--what-if` is enabled, when these commands run, then PRR prints external commands it would execute and does not execute them.

## Tasks / Subtasks

- [ ] Add `review-engine`, `render`, and `publish` command wiring (AC: 1, 2, 3)
  - [ ] Create `cmd/prr/review_engine.go`
  - [ ] Create `cmd/prr/render.go`
  - [ ] Extend/align `cmd/prr/publish.go` for composable contract mode
- [ ] Implement review-engine command path (AC: 1)
  - [ ] Accept validated bundle input
  - [ ] Invoke engine abstraction and emit structured review JSON
  - [ ] Surface transport/auth/normalisation failures with stable error classes
- [ ] Implement render command path (AC: 2)
  - [ ] Render Markdown by default
  - [ ] Support JSON output mode when requested
  - [ ] Preserve predictable stdout/stderr usage
- [ ] Implement publish command path (AC: 3)
  - [ ] Publish rendered review to provider via abstraction
  - [ ] Return explicit publication status (performed/skipped/unsupported/failure)
- [ ] Add tests for output/publish/automation contracts (AC: 1, 2, 3)
  - [ ] Unit tests for command modes and payload validation
  - [ ] Tests for review normalisation and finding-id presence
  - [ ] Tests for publish optional path and diagnostics
- [ ] Add command observability and dry-run behaviour (AC: 4, 5)
  - [ ] Add `--verbose` command logging for all external commands before execution
  - [ ] Add `--what-if` mode that prints commands without executing
  - [ ] Add tests covering verbose logging and what-if no-execution guarantees

## Dev Notes

- This story completes explicit composable command surface for MVP and should align tightly with Epic 4/5 capability boundaries.
- Keep provider and review engine swappable through interfaces.
- Ensure diagnostics are useful but sanitised; never leak tokens/secret headers.

### Technical Requirements

- Commands:
  - `prr review-engine`
  - `prr render`
  - `prr publish`
- Structured review output must include summary/risk/findings/checklist fields.
- Rendering must support Markdown default and JSON mode.
- Publish remains optional execution path but command must exist and report explicit status.
- Commands must support `--verbose` pre-execution logging for any external commands they run.
- Commands must support `--what-if` dry-run mode that prints external commands and performs no external mutations.

### Architecture Compliance

- `internal/engine/*` owns engine integration and response normalisation.
- `internal/render/*` owns output formatting.
- `internal/provider/*` (or `internal/publish/*`) owns publish integration.
- Command handlers remain thin adapters in `cmd/prr/*`.

### Library / Framework Requirements

- Cobra command model.
- Structured contracts in camelCase JSON.

### File Structure Requirements

- New command files under `cmd/prr/`.
- Implementation under `internal/engine`, `internal/render`, `internal/provider`/`internal/publish`, `internal/errors`, `internal/logging`.

### Testing Requirements

- Validate review JSON schema and required fields.
- Validate renderer output mode behaviour and channel stability.
- Validate publish outcomes (enabled, disabled, unsupported, failed) with explicit diagnostics.

### Previous Story Intelligence

- Story `1.8` should provide stable bundle output contract consumed by `review-engine`.
- Existing scaffold and command patterns from prior stories should be reused to avoid divergence.

### Git Intelligence Summary

- This story finalises MVP composable command surface and automation-facing output contracts; changes should prioritise contract stability over feature breadth.

### Project Structure Notes

- Preserve command-module boundaries and centralised error/logging abstractions.
- Keep output contracts stable for scripting compatibility.

### References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.9)
- Review output/publish requirements: `docs/initial_specification.md` (Review Output, Rendering, Publishing)
- MVP output/automation constraints: `_bmad-output/planning-artifacts/prd.md`
- Integration boundaries and patterns: `_bmad-output/planning-artifacts/architecture.md`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow executed from approved sprint change proposal.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story selected explicitly from approved change set: `1-9-implement-review-engine-render-and-publish-composable-commands`.

### File List

- _bmad-output/implementation-artifacts/1-9-implement-review-engine-render-and-publish-composable-commands.md
