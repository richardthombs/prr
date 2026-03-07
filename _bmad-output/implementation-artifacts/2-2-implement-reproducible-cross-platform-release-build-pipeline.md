# Story 2.2: Implement Reproducible Cross-Platform Release Build Pipeline

Status: review

## Story

As Richard,
I want deterministic build packaging for each target platform,
so that release binaries are reproducible and ready for distribution.

## Acceptance Criteria

1. Given a valid release tag, when the release build pipeline runs, then static binaries are produced for each configured OS/arch target, and build metadata (`version`, `commit`, `buildDate`) is embedded consistently.
2. Given unchanged source and build inputs, when release packaging is re-run, then produced artifact structure remains functionally equivalent, and any variance is limited to explicitly documented metadata fields.

## Tasks / Subtasks

- [x] Implement deterministic cross-platform build matrix in release workflow (AC: 1, 2)
  - [x] Preserve target order from release contract: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`.
  - [x] Produce one artifact per target with canonical filename: `prr_<version>_<os>_<arch><ext>`.
  - [x] Ensure Windows target emits `.exe` suffix and non-Windows targets do not.
- [x] Embed build metadata at compile time for all matrix builds (AC: 1)
  - [x] Inject `version`, `commit`, and `buildDate` into `cmd/prr/version.go` link-time vars via `-ldflags`.
  - [x] Source `version` from validated tag output produced by `validate-tag` job.
  - [x] Use a deterministic UTC timestamp format for `buildDate` (RFC3339).
- [x] Align binary production with static/reproducible expectations (AC: 1, 2)
  - [x] Build from repository root with explicit `GOOS`/`GOARCH`, `CGO_ENABLED=0`, and stable build flags (for example `-trimpath`).
  - [x] Ensure workflow packaging does not inject ad hoc fields or naming variants.
  - [x] Document any unavoidable non-functional byte variance fields, while preserving functional equivalence.
- [x] Add verification gates for artifact contract conformance (AC: 2)
  - [x] Validate expected filenames exist for all matrix entries before publish stage.
  - [x] Verify each built binary reports expected version via `prr version` smoke check.
  - [x] Fail fast with actionable diagnostics when expected artifacts are missing/misnamed.
- [x] Add or extend tests where feasible for build metadata behaviour (AC: 1, 2)
  - [x] Keep `cmd/prr/version_test.go` aligned with release-tag vs dev-build resolution rules.
  - [x] Add script/workflow checks for deterministic artifact naming and matrix ordering.

## Dev Notes

### Story Foundation and Business Context

- This story operationalises the release contract from Story 2.1 into an executable, deterministic build stage.
- Scope is build and packaging reproducibility only; upload/publication belongs to Story 2.3.
- The highest risk is contract drift (naming, ordering, metadata) between docs and workflow implementation.

### Technical Requirements (Guardrails)

- Canonical release tag formats are already enforced in `.github/workflows/release.yml` by `validate-tag`; Story 2.2 must consume that output and not re-interpret SemVer.
- Artifact naming and matrix order are contract-bound by `docs/release-process.md` and must stay stable for script compatibility.
- Build metadata must map directly to existing variables in `cmd/prr/version.go`:
  - `version` (release tag for tagged builds)
  - `commit` (full SHA)
  - `buildDate` (RFC3339 UTC)
- `resolvedVersion()` in `cmd/prr/version.go` is already release/dev aware. Build pipeline should supply release metadata so release artifacts print exact release tags.

### Architecture Compliance Requirements

- Keep command/runtime behaviour compatible across macOS, Linux, and Windows as required by architecture and PRD.
- Preserve deterministic, script-friendly contracts (stable filenames and no ad hoc output channel changes).
- Keep boundaries intact:
  - Story 2.2: build/package reproducibly
  - Story 2.3: publish to GitHub release
  - Story 2.4: checksum/integrity artifacts

### Library and Framework Requirements

- Go toolchain from `go.mod`: `go 1.25.0`.
- CLI framework remains `github.com/spf13/cobra v1.10.1`; this story must not introduce alternative CLI frameworks.
- GitHub Actions actions currently used in workflow:
  - `actions/checkout@v4`
  - `actions/setup-go@v5`

### File Structure Requirements

- Primary implementation files:
  - `.github/workflows/release.yml`
  - `docs/release-process.md` (contract sync only if needed)
  - `cmd/prr/version.go` (only if metadata handling changes are required)
  - `cmd/prr/version_test.go` (regression safety for version semantics)
- Do not move release logic into unrelated CLI command handlers.
- Keep release-specific validation and packaging logic within workflow/scripts, not runtime `review` pipeline modules.

### Testing Requirements

- Run canonical verification commands:
  - `go test ./...`
  - `go build ./...`
- Add workflow-level assertions for:
  - Matrix completeness and deterministic ordering
  - Expected artifact filenames
  - `prr version` result for built artifacts
- Preserve existing command contract tests; avoid regressions in unrelated review workflows.

### Previous Story Intelligence (From 2.1)

- 2.1 already defined the contract and introduced release-tag validation gates; do not duplicate or fork contract definitions.
- 2.1 aligned `cmd/prr/version.go` with release/dev semantics and added tests. 2.2 should extend this path rather than replacing it.
- 2.1 explicitly deferred publish implementation to 2.3; keep that boundary to avoid scope creep.

### Git Intelligence Summary

- Recent commit `f9bf531` updated release planning artifacts, `docs/release-process.md`, `.github/workflows/release.yml`, and version tests. This is the baseline for 2.2.
- Recent release-related changes were concentrated in docs + workflow + version command tests, indicating intended implementation surface for this story.
- No evidence of introducing new packaging libraries; pattern is to use native Go build plus Actions workflow orchestration.

### Latest Technical Information

- Repository-local current versions:
  - Go: `1.25.0` (from `go.mod`)
  - Cobra: `v1.10.1` (from `go.mod`)
  - Actions setup: `actions/checkout@v4`, `actions/setup-go@v5` (from `.github/workflows/release.yml`)
- No external dependency upgrade is required by this story definition; prioritise deterministic pipeline behaviour over dependency churn.

### Project Context Reference

- No `project-context.md` file detected in the repository.
- Primary planning sources used for this story:
  - `_bmad-output/planning-artifacts/epics.md`
  - `_bmad-output/planning-artifacts/architecture.md`
  - `_bmad-output/planning-artifacts/prd.md`
  - `docs/release-process.md`

## References

- `_bmad-output/planning-artifacts/epics.md` (Story 2.2 definition and acceptance criteria)
- `_bmad-output/planning-artifacts/architecture.md` (cross-platform runtime/build expectations; deterministic contracts)
- `_bmad-output/planning-artifacts/prd.md` (automation, output stability, and cross-platform requirements)
- `_bmad-output/implementation-artifacts/2-1-define-release-artifact-matrix-and-naming-contract.md` (previous story learnings and boundaries)
- `docs/release-process.md` (release target matrix, naming contract, SemVer and metadata contract)
- `.github/workflows/release.yml` (current validate/build/publish staging baseline)
- `cmd/prr/version.go` (metadata and release/dev version resolution behavior)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow execution for user-selected story `2.2`.
- Implemented deterministic matrix builds and artifact verification gates in `.github/workflows/release.yml`.
- Added reusable release contract verification script at `scripts/verify_release_contract.sh`.

### Completion Notes List

- Implemented cross-platform build matrix entries with contract order and canonical file naming.
- Added deterministic build metadata injection (`version`, `commit`, `buildDate`) via `-ldflags` from validated release tag and commit metadata.
- Enabled reproducible build controls (`CGO_ENABLED=0`, `-trimpath`, `-buildid=`) and retained static build expectations.
- Added per-target `prr version` smoke checks to ensure embedded release version correctness.
- Added pre-publish artifact contract verification gate for expected names and no naming drift.
- Documented allowed non-functional variance fields in release process contract documentation.
- Verified no regressions with `go test ./...` and `go build ./...`.

### File List

- _bmad-output/implementation-artifacts/2-2-implement-reproducible-cross-platform-release-build-pipeline.md
- _bmad-output/implementation-artifacts/sprint-status.yaml
- .github/workflows/release.yml
- docs/release-process.md
- scripts/verify_release_contract.sh

## Change Log

- 2026-03-07: Implemented deterministic cross-platform release build pipeline with compile-time metadata injection, per-target version smoke checks, and pre-publish artifact contract verification.
