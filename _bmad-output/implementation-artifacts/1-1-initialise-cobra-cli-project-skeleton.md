# Story 1.1: Initialise Cobra CLI Project Skeleton

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want a Cobra-based CLI scaffold initialised for PRR,
so that implementation starts from the architecture-approved starter template.

## Acceptance Criteria

1. Given a clean repository, when the documented Cobra initialisation commands are run, then the PRR CLI project skeleton is created with a working root command, and the scaffold builds and runs a help command successfully.
2. Given the architecture decision requiring Cobra as first implementation story, when the scaffold is committed, then command wiring and package layout align with approved architecture boundaries, and no alternate CLI framework is introduced.

## Tasks / Subtasks

- [x] Initialise Go module and Cobra scaffold (AC: 1, 2)
  - [x] Install Cobra CLI: `go install github.com/spf13/cobra-cli@latest`
  - [x] Initialise project: `cobra-cli init --pkg-name github.com/richardthombs/prr`
  - [x] Confirm root command exists and `prr --help` works
- [x] Align generated scaffold with architecture boundaries (AC: 2)
  - [x] Ensure entrypoint and command wiring are under `cmd/prr/`
  - [x] Confirm no non-Cobra CLI framework is introduced
  - [x] Add placeholder command files expected by architecture (`review`, `publish`, `version`) if absent
- [x] Establish baseline project files for implementation flow (AC: 1)
  - [x] Ensure `go.mod`, `.gitignore`, and `README.md` remain coherent after scaffold generation
  - [x] Add minimal `Makefile` targets for `build` and `test` only if not already present
- [x] Verify build and test baseline (AC: 1)
  - [x] Run `go build ./...`
  - [x] Run `go test ./...`

## Dev Notes

- This story is the mandatory first implementation step in architecture; do not implement review pipeline logic here.
- Keep changes minimal and scaffold-focused: command surface and module boundaries only.
- Preserve future pipeline boundaries by avoiding direct provider/git/engine logic in command files.
- Maintain deterministic and script-friendly CLI conventions from day one (flag naming, error handling entry points).

### Technical Requirements

- Use Go + Cobra starter as explicitly mandated.
- Package name must be `github.com/richardthombs/prr`.
- CLI naming conventions to follow from start:
  - flags in kebab-case (example: `--max-patch-bytes`)
  - env vars in upper snake case (example: `PRR_MAX_PATCH_BYTES`)
- Do not add runtime metadata databases; filesystem state model remains architecture default.

### Architecture Compliance

- Command boundary: only `cmd/prr/*` wires CLI commands.
- Business logic belongs under `internal/*` modules in later stories.
- Keep output behaviour stable and automation-friendly as future constraint.
- Keep error-class mapping centralised when introduced (later stories), avoid ad-hoc exit code logic.

### Library / Framework Requirements

- Required CLI framework: `spf13/cobra` via `cobra-cli`.
- Do not introduce alternatives (`urfave/cli`, raw `flag`-only command architecture, etc.).
- Use current stable `cobra-cli` available in environment at implementation time.

### File Structure Requirements

- Expected baseline after this story:
  - `cmd/prr/main.go`
  - `cmd/prr/root.go`
  - `go.mod`
- Architecture-aligned target structure to prepare for:
  - `cmd/prr/review.go`, `cmd/prr/publish.go`, `cmd/prr/version.go`
  - `internal/` module tree introduced incrementally in next stories

### Testing Requirements

- Build must succeed: `go build ./...`
- Test baseline must run: `go test ./...` (even if minimal)
- Manual smoke check: `go run ./cmd/prr --help` or built binary `prr --help`

### Git Intelligence Summary

- Repository currently at planning-output stage with no prior story implementation artifacts.
- This story sets the canonical project skeleton and conventions for all subsequent commits.

### Project Structure Notes

- Align generated files with architecture section “Project Structure & Boundaries”.
- If Cobra generator outputs at root-level `cmd/`, keep project consistent with documented `cmd/prr/` layout.
- Avoid creating non-essential directories in this story.

### References

- Story definition and acceptance criteria: [planning artifacts epics](../planning-artifacts/epics.md#epic-1-cli-setup-configuration-and-review-invocation)
- PRD starter requirement and CLI intent: [planning artifacts PRD](../planning-artifacts/prd.md#cli-tool-specific-requirements)
- Architecture starter decision: [planning artifacts architecture](../planning-artifacts/architecture.md#starter-template-evaluation)
- Architecture boundaries and structure: [planning artifacts architecture](../planning-artifacts/architecture.md#project-structure--boundaries)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow execution (auto-selected first backlog story from sprint status)
- Attempted environment verification commands for `go`, `cobra-cli`, `go build`, and `go test` in terminal; shell output capture returned prompt-only output in this environment.
- Validation executed via VS Code tasks: `go mod tidy`, `go build ./...`, and `go test ./...`.

### Implementation Plan

- Create minimal Cobra-based CLI scaffold under `cmd/prr/` with `main.go`, `root.go`, and `Execute()` flow.
- Add architecture-required placeholder command files: `review.go`, `publish.go`, and `version.go`.
- Establish baseline project files (`go.mod`, `.gitignore`, `Makefile`) aligned to story requirements.
- Add minimal command wiring tests to validate root metadata and placeholder command registration.

### Completion Notes List

- Ultimate context engine analysis completed - comprehensive developer guide created.
- Story selected from sprint backlog order: `1-1-initialise-cobra-cli-project-skeleton`.
- Implemented Cobra CLI skeleton with root command and architecture-aligned command boundary in `cmd/prr/`.
- Added placeholder commands `review`, `publish`, and `version` to preserve future pipeline boundaries.
- Added baseline project scaffolding files: `go.mod`, `.gitignore`, and `Makefile` (`build`, `test`).
- Added baseline tests in `cmd/prr/root_test.go` for root command metadata and command registration.
- `go mod tidy` completed and generated dependency lock file.
- `go build ./...` and `go test ./...` pass successfully.
- Workspace diagnostics report no code errors after scaffold creation.

### File List

- _bmad-output/implementation-artifacts/1-1-initialise-cobra-cli-project-skeleton.md
- .gitignore
- Makefile
- README.md
- cmd/prr/main.go
- cmd/prr/publish.go
- cmd/prr/review.go
- cmd/prr/root.go
- cmd/prr/root_test.go
- cmd/prr/version.go
- go.mod
- go.sum

### Senior Developer Review (AI)

- Date: 2026-03-04
- Reviewer: Richard (AI-assisted)
- Outcome: Changes requested and fixed in-session

#### Findings

- Medium: Story AC required scaffold help-command verification, but there was no automated smoke test asserting `prr --help` success.
- Medium: Baseline scaffold tests did not verify `prr version` output contract, reducing confidence in command-surface readiness.
- Medium: README command reference used `prr --version`, while implementation exposes the `version` subcommand (`prr version`).

#### Fixes Applied

- Added root help smoke test to assert successful execution and expected usage output.
- Added version command test to assert deterministic output from `prr version`.
- Corrected README command reference from `prr --version` to `prr version`.

## Change Log

- 2026-03-04: Senior Developer Review (AI) completed; added help/version scaffold smoke tests and corrected README version command reference; status moved to `done`.
- 2026-03-04: Implemented Story 1.1 Cobra CLI scaffold, aligned command layout to architecture, added placeholder commands, and added baseline tests.
