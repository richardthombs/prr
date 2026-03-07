# Story 2.1: Define Release Artifact Matrix and Naming Contract

Status: review

## Story

As Richard,
I want a clear release artifact contract,
so that every release produces predictable files across OS/architecture targets.

## Acceptance Criteria

1. Given supported target platforms, when I review release packaging configuration, then target OS/arch combinations are explicitly defined, and artifact filenames follow a stable semantic-version naming scheme.
2. Given pre-release and stable tags, when release packages are generated, then artifact metadata reflects the version/tag correctly, and no ad hoc naming is introduced.
3. Given a proposed next release, when the team classifies completed changes, then the selected SemVer bump (`major`, `minor`, `patch`) follows the documented decision matrix, and the rationale is recorded in release preparation metadata.
4. Given an invalid or non-SemVer release tag, when the packaging workflow starts, then the workflow fails fast with actionable validation diagnostics, and no release artifacts are published.
5. Given a release build and a non-release/dev build, when `prr version` is executed for each, then release output reports the exact SemVer tag, and dev output reports `v0.0.0-dev+<shortsha>` with commit metadata.

## Tasks / Subtasks

- [x] Define and document the release target matrix and artifact naming schema (AC: 1, 2)
  - [x] Add a canonical matrix for initial targets (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`) in release documentation or workflow config.
  - [x] Define filename template and extension rules by platform, including `.exe` for Windows binaries.
  - [x] Include deterministic ordering rules for artifact lists to preserve script compatibility.
- [x] Define SemVer source-of-truth and bump decision matrix (AC: 2, 3)
  - [x] Set canonical release source to Git tags `vMAJOR.MINOR.PATCH`.
  - [x] Document pre-release conventions (`-rc.N`) and non-release/dev format (`v0.0.0-dev+<shortsha>`).
  - [x] Add explicit bump rules with examples for MAJOR, MINOR, PATCH triggers.
- [x] Add release tag validation and fail-fast behavior in release workflow entry (AC: 4)
  - [x] Validate incoming tags against SemVer before build/publish stages execute.
  - [x] Emit actionable diagnostics when validation fails.
  - [x] Ensure workflow exits before artifact publication on invalid tags.
- [x] Define metadata contract consumed by `prr version` and packaging (AC: 2, 5)
  - [x] Confirm fields: `version`, `commit`, `buildDate` and clarify release vs dev expectations.
  - [x] Align link-time variable injection contract with existing `cmd/prr/version.go` output behavior.
  - [x] Add or update tests for release-tag and dev-build output paths.
- [x] Document implementation boundaries and downstream impact for follow-on stories (AC: 1-5)
  - [x] Capture what Story 2.2 (build pipeline) and 2.3 (publish workflow) will consume from this contract.
  - [x] Ensure docs references are explicit so implementation does not reinvent artifact rules.

## Dev Notes

### Business and Technical Context

- This story defines the contract that all release automation will follow. It should avoid implementing full packaging/publishing mechanics; those belong to Stories 2.2 and 2.3.
- The main outputs here are deterministic naming/versioning rules and validation behavior that prevent ambiguity in CI/release flows.

### Architecture Constraints

- Build and release model must target single static CLI binaries per OS/arch with CI verification.
- Cross-platform contract parity is required for macOS, Linux, and Windows.
- Stable stdout/stderr and machine-readable behavior should be preserved for scripts.

### Implementation Guardrails

- Reuse existing project command and output conventions; do not introduce alternate version output formats.
- Keep naming contract deterministic and parseable by scripts.
- Prefer explicit mapping tables and validation over inferred behavior.
- Fail fast on invalid SemVer input and avoid partial publish behavior.

### Suggested File Touchpoints

- `.github/workflows/release.yml` (tag validation gate and use of naming/version variables)
- `cmd/prr/version.go` (ensure output contract aligns with release/dev version semantics)
- `README.md` and/or `docs/release-process.md` (artifact naming and SemVer policy)
- Optional helper module for release metadata derivation if needed (keep scope constrained)

### Testing Requirements

- Add or update tests for SemVer parsing/validation logic.
- Add coverage for version output behavior for release-tag vs dev-build mode.
- If release workflow scripts are updated, include a workflow-level validation step that can fail independently before build/publish.

### Risks and Anti-Patterns to Avoid

- Do not allow ad hoc file naming patterns to accumulate across scripts/workflows.
- Do not accept non-SemVer tags for release publishing.
- Do not couple this contract to one-off local assumptions that break cross-platform packaging.

## References

- `_bmad-output/planning-artifacts/epics.md` (Story 2.1 section)
- `_bmad-output/planning-artifacts/architecture.md` (Infrastructure and deployment, pattern enforcement, project boundaries)
- `_bmad-output/planning-artifacts/prd.md` (CLI output and automation consistency expectations)
- `docs/release-process.md` (release process baseline)
- `.github/workflows/release.yml` (current release workflow entrypoint)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow execution for user-selected story `2.1`.

### Completion Notes List

- Added `docs/release-process.md` as the canonical contract for release target matrix, deterministic artifact naming, SemVer source and bump rules, and downstream boundaries for Stories 2.2/2.3.
- Added `.github/workflows/release.yml` with explicit release-tag validation (`vMAJOR.MINOR.PATCH` and `vMAJOR.MINOR.PATCH-rc.N`) that fails before build/publish stages.
- Updated `cmd/prr/version.go` to support link-time metadata fields (`version`, `commit`, `buildDate`) and deterministic release/dev output behaviour via `resolvedVersion()`.
- Added `cmd/prr/version_test.go` and updated `cmd/prr/root_test.go` to cover release and dev version output paths.
- Verified regression safety with `go test ./...` and `go build ./...`.

### File List

- .github/workflows/release.yml
- README.md
- cmd/prr/root_test.go
- cmd/prr/version.go
- cmd/prr/version_test.go
- docs/release-process.md
- _bmad-output/implementation-artifacts/sprint-status.yaml
- _bmad-output/implementation-artifacts/2-1-define-release-artifact-matrix-and-naming-contract.md

### Change Log

- 2026-03-06: Defined release artifact/versioning contract, added release tag validation workflow gate, aligned `prr version` release/dev metadata behaviour, and added tests for version output paths.
