# Story 2.3: Publish Release Artifacts via GitHub Release Workflow

Status: review

## Story

As Richard,
I want release artifacts uploaded automatically,
so that tagged versions are available without manual packaging steps.

## Acceptance Criteria

1. Given a release-triggering tag, when the GitHub Actions release workflow executes, then all generated platform artifacts are attached to the corresponding GitHub Release, and workflow logs provide actionable diagnostics on failure.
2. Given publish failure for any target artifact, when the workflow completes, then the run fails with clear stage-level error reporting, and partial publish outcomes are explicit and non-ambiguous.

## Tasks / Subtasks

- [x] Implement GitHub Release publication stage in `.github/workflows/release.yml` (AC: 1, 2)
  - [x] Replace the current placeholder publish step with a concrete upload implementation.
  - [x] Keep existing gating dependencies (`validate-tag`, `build-artifacts`, `verify-artifact-contract`) unchanged so publish runs only after contract-safe artifacts are verified.
  - [x] Ensure publish stage only runs for valid SemVer release tags accepted by the existing validation gate.
- [x] Attach all contract artifacts to the matching release entity (AC: 1)
  - [x] Download merged artifacts from previous jobs from `dist/release`.
  - [x] Publish the full target matrix outputs using canonical filenames: `prr_<version>_<os>_<arch><ext>`.
  - [x] Ensure release attachment set is complete and deterministic for matrix targets defined in `docs/release-process.md`.
- [x] Implement fail-fast and explicit publish diagnostics (AC: 2)
  - [x] Surface upload failures with actionable messages in workflow logs.
  - [x] Ensure any failed upload causes publish job failure (no silent success on partial uploads).
  - [x] Emit stage-scoped diagnostics indicating which artifact(s) failed.
- [x] Clarify release creation/update behaviour and idempotency (AC: 1, 2)
  - [x] Define whether workflow creates a release if missing or updates an existing release for the same tag.
  - [x] Ensure repeated runs for the same tag do not produce ambiguous duplicates.
  - [x] Keep behavior consistent with GitHub release model and documented process.
- [x] Add verification and documentation updates for publish behaviour (AC: 1, 2)
  - [x] Update `docs/release-process.md` with the final publish stage contract and failure semantics.
  - [x] Add/adjust release notes only if user-visible release process behavior changes.
  - [x] Keep Story 2.4 boundary intact (checksums are not in scope for this story).

## Dev Notes

### Story Foundation and Business Context

- Story 2.3 implements the publication boundary deferred in Story 2.2. Build and artifact contract checks are already in place and must remain the gate before any upload.
- The objective is reliable, fully automated release publication for validated tags, with no ambiguous partial-success state.

### Technical Requirements (Guardrails)

- Reuse and preserve existing release workflow structure:
  - `validate-tag` enforces SemVer release tag contract.
  - `build-artifacts` produces deterministic binaries per target.
  - `verify-artifact-contract` ensures expected filenames and matrix completeness.
- Publish must consume artifacts already verified by `scripts/verify_release_contract.sh`; do not redefine matrix or naming rules in publish logic.
- Keep release metadata source-of-truth as validated tag output from `validate-tag` job.
- Publishing must treat artifact set as atomic from a CI outcome perspective: any failed asset upload fails the job.

### Architecture Compliance Requirements

- Respect deterministic automation contracts and explicit failure boundaries from architecture and PRD.
- Preserve cross-platform release parity by publishing all matrix artifacts, not a subset.
- Keep stage-level diagnostics clear and actionable to match troubleshooting requirements.
- Avoid introducing runtime CLI behavior changes in this story; scope is CI release workflow publication only.

### Library and Framework Requirements

- Existing workflow dependencies in use:
  - `actions/checkout@v4`
  - `actions/setup-go@v5`
  - `actions/upload-artifact@v4`
  - `actions/download-artifact@v4`
- For release publication, use a stable GitHub-native mechanism (for example `gh release upload` or an established release action) and keep implementation explicit.
- Do not introduce unnecessary third-party tooling for tasks already covered by GitHub Actions primitives.

### File Structure Requirements

- Primary files for this story:
  - `.github/workflows/release.yml`
  - `docs/release-process.md`
  - `docs/release-notes.md` (only if behavioural notes are required)
- Keep release publication logic in the release workflow; do not add publish logic inside runtime Go command handlers.
- Do not modify artifact naming script contract in `scripts/verify_release_contract.sh` except when necessary to support publication validation handoff.

### Testing Requirements

- Validate workflow semantics with at least one dry-run or controlled tag-trigger path to confirm:
  - publish stage executes after contract verification gates,
  - all expected artifacts are attached,
  - failure of any upload path fails the publish job.
- Run baseline regressions to ensure no unrelated breakage:
  - `go test ./...`
  - `go build ./...`
- Verify logs provide explicit stage-level failure reporting for publication errors.

### Previous Story Intelligence (From 2.2)

- Story 2.2 already implemented deterministic matrix builds and artifact contract verification; do not duplicate those checks in ad hoc form.
- Current `publish-release` job is an intentional placeholder that states deferral to Story 2.3; replace it directly rather than restructuring the full workflow graph.
- Keep release contract boundaries strict:
  - 2.2 = deterministic build/package + verify contract
  - 2.3 = upload/publish
  - 2.4 = checksums and integrity artifacts

### Git Intelligence Summary

- Recent commit `227e8df`: implemented deterministic cross-platform release build pipeline and added artifact contract verifier script.
- Recent commit `f9bf531`: updated release planning contract, release docs, version semantics, and workflow baseline.
- Commit pattern shows release work concentrated in:
  - `.github/workflows/release.yml`
  - `docs/release-process.md`
  - release-story artifacts and sprint status
- No evidence of existing checksum publication yet; that remains intentionally separate for Story 2.4.

### Latest Technical Information

- Repository-local release stack currently uses GitHub Actions workflows with `actions/*` components and Go `1.25.0`.
- External web research was not performed in this run; implement against repository-approved versions and pinned actions already in use.
- Keep compatibility with current validated tag contract:
  - stable `vMAJOR.MINOR.PATCH`
  - pre-release `vMAJOR.MINOR.PATCH-rc.N`

### Project Context Reference

- No `project-context.md` file detected in repository.
- Primary planning and architecture sources used:
  - `_bmad-output/planning-artifacts/epics.md`
  - `_bmad-output/planning-artifacts/architecture.md`
  - `_bmad-output/planning-artifacts/prd.md`
  - `docs/release-process.md`

## References

- `_bmad-output/planning-artifacts/epics.md` (Story 2.3 definition and acceptance criteria)
- `_bmad-output/planning-artifacts/architecture.md` (deployment pipeline, deterministic contracts, diagnostics requirements)
- `_bmad-output/planning-artifacts/prd.md` (automation and failure semantics expectations)
- `_bmad-output/implementation-artifacts/2-2-implement-reproducible-cross-platform-release-build-pipeline.md` (previous story boundaries and implementation learnings)
- `.github/workflows/release.yml` (current publish placeholder and release workflow structure)
- `docs/release-process.md` (release matrix, naming, and story boundary contract)
- `scripts/verify_release_contract.sh` (artifact contract verification gate)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- Create-story workflow execution for user-selected story `2.3`.
- Source analysis included epics, architecture, PRD, Story 2.2 artifact, release workflow, release process docs, and recent git commits.

### Implementation Plan

- Replace `publish-release` placeholder with concrete release publication steps while preserving existing gate dependencies.
- Download and validate the deterministic artifact set from prior jobs before any upload attempt.
- Implement release ensure/create + idempotent upload (`--clobber`) behaviour with fail-fast diagnostics per artifact.
- Update release process documentation to capture final publish contract and explicit failure semantics.

### Completion Notes List

- Created implementation-ready story context for Story 2.3 with explicit publication scope and guardrails.
- Preserved boundary with Story 2.2 (build/verify) and Story 2.4 (checksums).
- Included concrete file touchpoints, acceptance criteria alignment, and failure-handling requirements for publish stage.
- Captured prior-story and git-history intelligence to avoid rework and contract drift.
- Replaced release publish placeholder with concrete publish flow in `.github/workflows/release.yml`.
- Added artifact download and deterministic publish-input validation for all canonical matrix filenames.
- Added release ensure/create behavior using GitHub-native `gh release` commands and idempotent uploads with `--clobber`.
- Added fail-fast upload loop with stage-scoped, artifact-specific error diagnostics to prevent ambiguous partial success.
- Updated `docs/release-process.md` with Story 2.3 publication contract, idempotency model, and failure semantics.
- Validated baseline regressions with `go test ./...` and `go build ./...`.

### File List

- _bmad-output/implementation-artifacts/2-3-publish-release-artifacts-via-github-release-workflow.md
- _bmad-output/implementation-artifacts/sprint-status.yaml
- .github/workflows/release.yml
- docs/release-process.md

### Change Log

- 2026-03-07: Created Story 2.3 implementation artifact with comprehensive release publication guidance and contract guardrails.
- 2026-03-07: Implemented release publication job with deterministic artifact upload, fail-fast diagnostics, and idempotent release handling.
