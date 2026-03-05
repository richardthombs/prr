# Story 1.9b: Implement Render Command for JSON to Markdown

Status: ready-for-dev

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

- [ ] Add `render` command wiring (AC: 1, 2, 3, 4, 5)
  - [ ] Create `cmd/prr/render.go` and register with `rootCmd.AddCommand`.
  - [ ] Accept review JSON via stdin (optionally add file input flag if needed, but keep stdin as default).
  - [ ] Keep Markdown output on stdout and diagnostics/errors on stderr.

- [ ] Implement render-domain contract and validation (AC: 1, 3)
  - [ ] Add/confirm `internal/types/review.go` with review schema fields: summary, risk, findings, checklist.
  - [ ] Validate required fields and value constraints before rendering.
  - [ ] Return classed errors via `internal/errors` for invalid payloads.

- [ ] Implement deterministic Markdown rendering (AC: 1, 2)
  - [ ] Render sections in fixed order: Summary → Risk → Findings by severity order (`blocker`, `important`, `suggestion`, `nit`) → Checklist.
  - [ ] Use stable heading and bullet formatting.
  - [ ] Preserve finding field rendering consistency (id, file, line, category, message, suggestion).

- [ ] Implement verbose/what-if behaviour for this command path (AC: 4, 5)
  - [ ] `--verbose` prints processing diagnostics to stderr.
  - [ ] `--what-if` clearly indicates no external commands are executed.

- [ ] Add focused tests (AC: 1, 2, 3, 4, 5)
  - [ ] `cmd/prr` tests for valid JSON → Markdown shape and section ordering.
  - [ ] Determinism test: same input yields byte-identical output.
  - [ ] Validation tests for missing required fields and malformed JSON.
  - [ ] stdout/stderr separation tests.

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

### Completion Notes List

- Story generated from approved change proposal split (`1-9` + `1-9b`).
- Scope narrowed to JSON-to-Markdown rendering only.
- Marked `ready-for-dev` in line with workflow completion.

### File List

- _bmad-output/implementation-artifacts/1-9b-implement-render-command-for-json-to-markdown.md
