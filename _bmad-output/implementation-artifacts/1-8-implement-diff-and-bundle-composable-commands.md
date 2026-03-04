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
3. Given `--verbose` is enabled on `prr diff` and `prr bundle`, when external commands are invoked, then PRR logs each command to stderr before execution.
4. Given `--what-if` is enabled, when `prr diff` or `prr bundle` runs, then PRR prints external commands it would execute and does not execute them.

## Tasks / Subtasks

- [ ] Add command surfaces for `prr diff` and `prr bundle` (AC: 1, 2, 3, 4)
  - [ ] Create `cmd/prr/diff.go` and register command in `cmd/prr/root.go`
  - [ ] Create `cmd/prr/bundle.go` and register command in `cmd/prr/root.go`
  - [ ] Support composable input model used by existing commands (flags first, stdin JSON fallback where appropriate)
  - [ ] Preserve stdout for JSON payloads and stderr for diagnostics only
- [ ] Implement deterministic diff extraction in `internal/git` (AC: 1, 3, 4)
  - [ ] Add `internal/git/diff.go` with APIs to compute files, stat, and patch from merge-parent semantics (`HEAD^1..HEAD`)
  - [ ] Ensure deterministic output ordering for unchanged refs (stable file list order and stable formatting)
  - [ ] Ensure `--verbose` prints each external git command to stderr before execution
  - [ ] Ensure `--what-if` prints commands and performs no external execution
- [ ] Implement bundle assembly and validation in `internal/bundle` (AC: 2)
  - [ ] Add `internal/bundle/bundle.go` for v1 payload assembly (metadata + files + stat + patch)
  - [ ] Add `internal/bundle/schema.go` for required-field and structural validation
  - [ ] Keep external payload keys camelCase and stable for automation consumers
- [ ] Enforce safety limits before review-engine stage (AC: 2)
  - [ ] Add configurable checks for `maxPatchBytes` and `maxChangedFiles`
  - [ ] Return explicit limit diagnostics with stable classified errors via `internal/errors`
  - [ ] Guarantee deterministic failure behaviour (no partial ambiguous bundle contracts on limit failure)
- [ ] Extend shared types and command contracts (AC: 1, 2)
  - [ ] Add/extend `internal/types` contracts for diff result and bundle v1 schema
  - [ ] Ensure contracts interoperate with downstream `review-engine` and `render` stories
- [ ] Add focused tests and regressions (AC: 1, 2, 3, 4)
  - [ ] Command tests in `cmd/prr/*_test.go` for stdout/stderr separation, payload shape, and input handling
  - [ ] Git tests in `internal/git/*_test.go` for diff semantics and what-if/verbose behaviour
  - [ ] Bundle tests in `internal/bundle/*_test.go` for schema validation and limit failures
  - [ ] Determinism regression test for identical refs across reruns

## Dev Notes

- This story is the final prerequisite for Epic 1 composable-stage chain to produce review-engine-ready inputs.
- Current repository structure has moved toward a `checkout` command path; this story intentionally reintroduces explicit `diff` and `bundle` command surfaces required by Epic 1.8 acceptance criteria.
- Keep command handlers thin and delegate all git and bundle logic to internal packages.
- Preserve composability and automation stability: JSON payloads on stdout only, diagnostics on stderr only.
- Avoid incremental-review behaviour (explicit non-goal for v1).

### Technical Requirements

- Required commands:
  - `prr diff`
  - `prr bundle`
- Diff semantics:
  - Must compute PR contribution from merge-parent semantics (`HEAD^1..HEAD`) against isolated review worktree state.
  - Must emit deterministic `files`, `stat`, and `patch` outputs for unchanged refs.
- Bundle semantics:
  - Must assemble v1 bundle payload containing required metadata and diff outputs.
  - Must validate required fields before emission.
- Limits:
  - Must enforce configurable `maxPatchBytes` and `maxChangedFiles` before downstream engine invocation.
  - Must fail with explicit diagnostics and stable error classification.
- Observability:
  - `--verbose` logs external commands to stderr before execution.
  - `--what-if` prints commands and performs no external execution.

### Architecture Compliance

- Command boundary remains in `cmd/prr/*`; implementation logic belongs in `internal/*`.
- Diff operations must live under `internal/git` (new `diff.go` and related tests).
- Bundle assembly and validation must live under `internal/bundle`.
- Error taxonomy and exit mapping must continue through centralized `internal/errors`.
- Output contracts must remain stable and automation-safe (machine payloads on stdout, diagnostics on stderr).

### Library / Framework Requirements

- Use Cobra patterns already established in `cmd/prr`.
- Reuse existing process runner approach in `internal/git` for external git execution.
- Keep JSON contract keys camelCase for external payloads.
- Do not introduce new CLI/process libraries unless there is a clear blocker.

### File Structure Requirements

- Add `cmd/prr/diff.go`.
- Add `cmd/prr/bundle.go`.
- Add `internal/git/diff.go` (+ tests).
- Add `internal/bundle/bundle.go`, `internal/bundle/schema.go`, and optional `internal/bundle/limits.go` (+ tests).
- Add/extend types under `internal/types` for diff and bundle payload contracts.
- Update command registration in `cmd/prr/root.go` with minimal change footprint.

### Testing Requirements

- Validate deterministic rerun output equivalence for unchanged refs.
- Validate exact payload shape and required fields for diff and bundle outputs.
- Validate limit-exceeded diagnostics and stable error classification.
- Validate `--verbose` stderr command previews before execution.
- Validate `--what-if` no-execution guarantees and still-valid output contracts.
- Validate command input handling for composable usage patterns.

### Previous Story Intelligence

- Story `1.7` established key guardrails to preserve:
  - strict stdout/stderr separation
  - `--verbose` pre-exec command previews
  - `--what-if` no-execution behaviour covered by tests
  - deterministic JSON contract structure for composability
- Story `1.6` established mirror and merge-ref contract inputs that `prr diff` should consume without re-resolving provider context.
- Keep command design consistent with existing thin command + internal service pattern.

### Git Intelligence Summary

- Recent commit trajectory indicates active iteration around composable commands and workflow shape (including a recent shift to `checkout` command flow).
- For this story, treat Epic 1.8 acceptance criteria as source of truth and implement explicit `diff` and `bundle` commands without breaking current command behaviour.
- Recent work patterns emphasise tests-first guardrails around command contracts and what-if safety; maintain that approach here.

### Project Structure Notes

- Current tree snapshot:
  - `cmd/prr` contains `checkout`, `review`, and `publish` command files
  - `internal/git` currently contains mirror/worktree services, but no dedicated `diff` service file
  - `internal/types` currently contains `prref` only
- This story should introduce missing diff/bundle components while preserving existing architecture boundaries and naming conventions.

### Latest Technical Information

- No external web research was executed in this run.
- Implement using repository-established stack and patterns (Go modules, Cobra, system git invocation).
- If dependency/API uncertainties appear during implementation, prefer local codebase conventions and pinned module versions over introducing new dependencies.

### Project Context Reference

- No `project-context.md` was found in this workspace.
- Story guidance is grounded in:
  - `_bmad-output/planning-artifacts/epics.md`
  - `_bmad-output/planning-artifacts/prd.md`
  - `_bmad-output/planning-artifacts/architecture.md`
  - `_bmad-output/implementation-artifacts/1-7-implement-worktree-add-command-with-cleanup-keep-compatibility.md`
  - `.git/logs/HEAD` (recent commit intelligence)

### References

- `_bmad-output/planning-artifacts/epics.md` (Epic 1, Story 1.8 definition and BDD acceptance criteria)
- `docs/initial_specification.md` (High-Level Flow, deterministic diff, bundle requirements)
- `_bmad-output/planning-artifacts/prd.md` (FR14–FR20, NFR11/NFR12/NFR13/NFR14)
- `_bmad-output/planning-artifacts/architecture.md` (module boundaries, command and contract patterns)
- `_bmad-output/implementation-artifacts/1-7-implement-worktree-add-command-with-cleanup-keep-compatibility.md` (prior story learnings)

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

## Change Log

- 2026-03-04: Story context regenerated for explicit user-selected story `1.8`; added current-repo guardrails, prior-story intelligence, git-history intelligence, and implementation-ready task decomposition while keeping status `ready-for-dev`.
