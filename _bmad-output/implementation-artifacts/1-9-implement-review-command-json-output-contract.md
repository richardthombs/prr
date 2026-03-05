# Story 1.9: Implement Review Command JSON Output Contract

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want `prr review <PR_ID>` to emit stable structured JSON review output,
so that one command performs full review orchestration and can be consumed directly by automation and by `prr render`.

## Acceptance Criteria

1. Given a valid review bundle, when I run `prr review <PR_ID>`, then PRR emits structured review JSON with stable per-run finding references, and engine failures return actionable, classed errors.
2. Given a valid review JSON payload, when I run `prr render`, then PRR outputs Markdown deterministically from the JSON payload, and channel/exit behaviour remains automation-stable.
3. Given publish functionality is required later, when scope is revisited post-MVP, then publication is designed as an optional extension and does not expand the MVP command surface.
4. Given `--verbose` is enabled, when these commands invoke external commands, then PRR logs each external command to stderr before execution.
5. Given `--what-if` is enabled, when these commands run, then PRR prints external commands it would execute and does not execute them.

## Tasks / Subtasks

- [ ] Implement `review` command orchestration to produce final JSON review output (AC: 1, 4, 5)
  - [ ] Extend `cmd/prr/review.go` from scaffold command to a `RunE` path with required flags (`--repo`, `--remote`, `--provider`, `--keep`, `--verbose`, `--what-if`).
  - [ ] Orchestrate existing internal stages in sequence: resolve context, ensure/fetch mirror state, create worktree, generate diff, build bundle, run engine, output review JSON.
  - [ ] Reuse existing `internal/git` and `internal/bundle` components; do not duplicate stage logic in command handlers.
  - [ ] Keep JSON payload on stdout and diagnostics on stderr.

- [ ] Add review output contract types and validation (AC: 1)
  - [ ] Add review-domain types under `internal/types` for `Risk`, `Finding`, and `Review` with camelCase JSON tags.
  - [ ] Enforce required fields for summary/risk/findings/checklist before printing final JSON.
  - [ ] Ensure finding IDs are present for each run and need only be stable within the run.

- [ ] Add review engine abstraction and default adapter seam (AC: 1)
  - [ ] Introduce `ReviewEngine` interface and adapter boundary (prefer `internal/engine` package as architecture target).
  - [ ] Return classed errors for engine failures; preserve actionable messages.
  - [ ] Keep implementation swappable (no provider/engine coupling in command files).

- [ ] Implement `render` command for JSON-to-Markdown (AC: 2, 4, 5)
  - [ ] Add `cmd/prr/render.go` command.
  - [ ] Read review JSON from stdin (and optionally file flag if introduced) and render deterministic Markdown sections: Summary, Risk, Findings grouped by severity, Checklist.
  - [ ] Ensure no schema-breaking stdout/stderr behaviour changes.

- [ ] Keep publish out of MVP command surface while preserving extension point (AC: 3)
  - [ ] Do not make publish part of required MVP flow for this story.
  - [ ] Keep provider publish contract as optional extension path only.

- [ ] Add tests for command contracts and error classes (AC: 1, 2, 4, 5)
  - [ ] `cmd/prr` tests for `review` JSON shape and stdout/stderr separation.
  - [ ] `cmd/prr` tests for `render` deterministic Markdown output from fixed JSON fixture.
  - [ ] Tests for `--verbose` and `--what-if` behaviour in new command paths.
  - [ ] Tests for engine failure classification and exit-code mapping.

## Dev Notes

- This story is the first implementation story after the approved command-model pivot.
- The user-facing MVP command set is now `review` and `render`; previously exposed composable commands remain implementation details that can still be reused internally.
- Current codebase already has `diff` and `bundle` command + package logic; prefer reusing their internal logic from `review` rather than duplicating behaviour.
- `cmd/prr/render.go` does not currently exist and should be introduced in this story.
- `cmd/prr/review.go` is currently a stub and is the primary command entrypoint to harden.

### Technical Requirements

- `prr review <PR_ID>` must emit structured JSON review output to stdout.
- JSON output must include at least: `summary`, `risk` (`score`, `reasons`), `findings[]`, and `checklist`.
- Each finding must include: `id`, `file`, `line`, `severity`, `category`, `message`, `suggestion`.
- Severity values: `blocker|important|suggestion|nit`.
- Category values: `correctness|security|performance|readability|api|tests|other`.
- Engine failures must map to stable classed errors and stable non-zero exit codes.
- `prr render` must produce deterministic Markdown from valid review JSON.
- `--what-if` mode must not execute external commands.

### Architecture Compliance

- Command handlers in `cmd/prr/*` remain thin adapters.
- Core logic resides under `internal/*` boundaries (`internal/git`, `internal/bundle`, `internal/provider`, `internal/errors`, and engine/render modules).
- Keep external JSON contracts camelCase.
- Preserve centralized error classification in `internal/errors`.
- Maintain stage-style orchestration boundaries (provider → git → bundle → engine → render).

### Library / Framework Requirements

- Continue using Cobra patterns established in existing commands.
- Continue using existing git invocation/service patterns already used by `diff`/`bundle` flows.
- Do not add new external dependencies unless strictly required.

### File Structure Requirements

- Update: `cmd/prr/review.go`
- Create: `cmd/prr/render.go`
- Create/Update: `internal/types/review.go` (or equivalent in `internal/types`)
- Create/Update: `internal/engine/*` adapter files if absent
- Update tests in `cmd/prr/*_test.go` and relevant internal packages

### Testing Requirements

- Validate `review` returns required JSON fields and deterministic shape for fixed fixtures.
- Validate `render` output includes required sections and stable formatting.
- Validate stdout/stderr separation for automation compatibility.
- Validate classified error → exit code mapping for engine/runtime/config/limit/provider failures.
- Validate `--what-if` and `--verbose` contracts.

### Previous Story Intelligence

- Story `1.8` established reliable diff and bundle behaviour plus guardrails (`--verbose`, `--what-if`, deterministic output checks).
- Reuse those components as internal stages from `review`; avoid implementing parallel paths that diverge from tested behaviour.
- Keep contract discipline: payload on stdout, diagnostics on stderr.

### Git Intelligence Summary

- Recent commit log data was not retrievable from the current tool execution context.
- Use current repository implementation state as source of truth: existing `diff`/`bundle` command and internal package patterns are the baseline for this story.

### Project Structure Notes

- Existing command set in code still includes legacy composable commands (`diff`, `bundle`, `publish`) and no `render` command file yet.
- This story should implement the approved pivot without breaking existing tested internals.

### References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.9)
- Command/output model: `_bmad-output/planning-artifacts/prd.md` (CLI Tool Specific Requirements)
- Boundaries and contracts: `_bmad-output/planning-artifacts/architecture.md` (API boundaries, patterns)
- Previous implementation intelligence: `_bmad-output/implementation-artifacts/1-8-implement-diff-and-bundle-composable-commands.md`
- Original specification details: `docs/initial_specification.md` (Review Output, Rendering, Interfaces)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- create-story workflow executed for approved pivot story key `1-9-implement-review-command-json-output-contract`

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story generated for updated command model (`review` JSON contract + `render` markdown conversion).
- Guidance prioritises reuse of existing internal diff/bundle capabilities from Story 1.8.

### File List

- _bmad-output/implementation-artifacts/1-9-implement-review-command-json-output-contract.md
