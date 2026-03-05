# Story 1.9b: Implement Render Command for JSON to Markdown

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want `prr render` to consume review JSON and emit deterministic Markdown,
so that human-readable review output is separated cleanly from `prr review` JSON generation.

## Acceptance Criteria

1. Given a valid review JSON payload on stdin, when I run `prr render`, then PRR prints Markdown to stdout with required sections: Summary, Risk, Findings grouped by severity, and Checklist / Next actions.
2. Given a valid review JSON payload, when I run `prr render`, then output is deterministic for the same input and suitable for automation pipelines.
3. Given invalid or incomplete review JSON, when I run `prr render`, then PRR fails with actionable diagnostics and a stable classed error code.
4. Given `--verbose` is enabled, when `prr render` performs processing, then diagnostics are written to stderr only.
5. Given `--what-if` is enabled, when `prr render` runs, then it reports no external command execution and still validates input/output contract behaviour.

## Tasks / Subtasks

- [x] Add `render` command wiring (AC: 1, 2, 3, 4, 5)
  - [x] Create `cmd/prr/render.go` and register with `rootCmd.AddCommand`.
  - [x] Accept review JSON via stdin (optionally add file input flag if needed, but keep stdin as default).
  - [x] Keep Markdown output on stdout and diagnostics/errors on stderr.

- [x] Implement render-domain contract and validation (AC: 1, 3)
  - [x] Add/confirm `internal/types/review.go` with review schema fields: summary, risk, findings, checklist.
  - [x] Validate required fields and value constraints before rendering.
  - [x] Return classed errors via `internal/errors` for invalid payloads.

- [x] Implement deterministic Markdown rendering (AC: 1, 2)
  - [x] Render sections in fixed order: Summary → Risk → Findings by severity order (`blocker`, `important`, `suggestion`, `nit`) → Checklist.
  - [x] Use stable heading and bullet formatting.
  - [x] Preserve finding field rendering consistency (id, file, line, category, message, suggestion).

- [x] Implement verbose/what-if behaviour for this command path (AC: 4, 5)
  - [x] `--verbose` prints processing diagnostics to stderr.
  - [x] `--what-if` clearly indicates no external commands are executed.

- [x] Add focused tests (AC: 1, 2, 3, 4, 5)
  - [x] `cmd/prr` tests for valid JSON → Markdown shape and section ordering.
  - [x] Determinism test: same input yields byte-identical output.
  - [x] Validation tests for missing required fields and malformed JSON.
  - [x] stdout/stderr separation tests.

## Dev Notes

- This story is split from Story 1.9 under the approved pivot and focuses only on JSON-to-Markdown rendering.
- Current codebase has no `cmd/prr/render.go`; this story introduces the command.
- Keep render implementation independent from provider/git layers; it is a pure transformation stage.

### Technical Requirements

- Input contract is review JSON containing at least:
  - `summary`
  - `risk.score`, `risk.reasons`
  - `findings[]` with `id`, `file`, `line`, `severity`, `category`, `message`, `suggestion`
  - `checklist[]`
- Output contract is Markdown with minimum sections:
  - Summary
  - Risk
  - Findings (grouped by severity)
  - Checklist / Next actions
- `prr render` must be deterministic for identical input.

### Architecture Compliance

- Keep command adapter thin in `cmd/prr/render.go`.
- Place review schema/model in `internal/types`.
- Place reusable formatting logic in an internal render module if introduced.
- Keep `internal/errors` as the only error classification source.

### Library / Framework Requirements

- Cobra command model only.
- No additional external dependency required for Markdown generation; use standard Go formatting.

### File Structure Requirements

- Create: `cmd/prr/render.go`
- Create/Update: `internal/types/review.go`
- Optional create/update: `internal/render/markdown.go` (if extracted from command)
- Add tests under `cmd/prr/*_test.go` and/or `internal/render/*_test.go`

### Testing Requirements

- Valid review payload renders all required sections.
- Findings grouping order is stable and severity buckets are deterministic.
- Invalid JSON/invalid schema returns stable classed errors.
- Output channel behaviour remains automation-safe.

### Previous Story Intelligence

- Story 1.9 now owns `review` JSON orchestration; this story consumes that JSON contract and should not duplicate review-engine orchestration.
- Existing diff/bundle work (Story 1.8) is unrelated to render internals and should remain untouched.

### Git Intelligence Summary

- Recent commit history was not available via terminal tooling in this run.
- Use current repository conventions from existing command files and tests as implementation baseline.

### Project Structure Notes

- Existing command files include `review.go`, `diff.go`, `bundle.go`, and `publish.go`; `render.go` is absent and should be added.
- Maintain compatibility with current root command registration pattern (`init()` + `rootCmd.AddCommand`).

### References

- Pivot approval and split rationale: `_bmad-output/planning-artifacts/sprint-change-proposal-2026-03-05.md`
- Story umbrella context: `_bmad-output/implementation-artifacts/1-9-implement-review-command-json-output-contract.md`
- CLI/output requirements: `_bmad-output/planning-artifacts/prd.md`
- Rendering requirements and review schema: `docs/initial_specification.md`
- Architecture boundaries: `_bmad-output/planning-artifacts/architecture.md`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- create-story-style generation for split story key `1-9b-implement-render-command-for-json-to-markdown`
- Verified render implementation in `cmd/prr/render.go` and contract validation in `internal/types/review.go`.
- Added focused render tests for malformed JSON, missing required fields, and byte-identical output determinism.
- Executed targeted and full test validation (`go test ./...`) and build verification (`go build ./...`).
- Code-review workflow run identified and fixed validation/classification gaps in render input handling.

### Completion Notes List

- Story generated from approved change proposal split (`1-9` + `1-9b`).
- Scope narrowed to JSON-to-Markdown rendering only.
- Confirmed `render` command wiring and root command registration are present and working.
- Confirmed deterministic markdown section ordering and stable findings grouping.
- Added explicit invalid-input tests to enforce classed errors for malformed and incomplete payloads.
- Split review validation into strict render input validation and review output normalisation.
- Enforced positive finding line validation for render input payloads.
- Enforced required finding ID for render input payloads (no silent ID synthesis).
- Story implementation and post-review fixes validated and marked `done`.

### File List

- _bmad-output/implementation-artifacts/1-9b-implement-render-command-for-json-to-markdown.md
- _bmad-output/implementation-artifacts/sprint-status.yaml
- cmd/prr/render.go
- cmd/prr/review.go
- cmd/prr/review_render_test.go
- cmd/prr/root_test.go
- internal/errors/errors.go
- internal/errors/errors_test.go
- internal/types/review.go

## Senior Developer Review (AI)

### Review Date

2026-03-05

### Outcome

Approve

### Summary

- Adversarial review identified 5 issues (3 High, 2 Medium).
- All High and Medium issues were fixed in this pass.
- All acceptance criteria remain satisfied after fixes.

### Findings Resolved

- [x] [High] Enforce required finding IDs for render input validation (`internal/types/review.go`).
- [x] [High] Classify invalid/incomplete `prr render` payload validation as input/config failures (`internal/types/review.go`, `cmd/prr/render.go`, `cmd/prr/review_render_test.go`).
- [x] [High] Enforce positive finding line constraint for render input (`internal/types/review.go`, `cmd/prr/review_render_test.go`).
- [x] [Medium] Reconcile story File List with actual changed source files (`_bmad-output/implementation-artifacts/1-9b-implement-render-command-for-json-to-markdown.md`).
- [x] [Medium] Reconcile additional tracked source-file diffs in working tree (`cmd/prr/root_test.go`, `internal/errors/errors.go`, `internal/errors/errors_test.go`).

### Notes

- Local `prr` binary diff was excluded from review scope as a generated non-source artifact.

## Change Log

- 2026-03-05: Completed dev implementation for Story 1.9b and moved status to review.
- 2026-03-05: Completed adversarial code-review remediation, updated story status to done, and reconciled sprint tracking.
