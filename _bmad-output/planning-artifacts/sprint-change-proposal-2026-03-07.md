# Sprint Change Proposal - 2026-03-07

## 1. Issue Summary

- **Triggering change:** Remove `darwin/amd64` from the PRR release target matrix.
- **Trigger type:** Scope and platform support adjustment discovered during release-contract review.
- **Context:** The current Story 2.1 artifacts define an initial five-target matrix including `darwin/amd64`. Request is to narrow supported targets to Apple Silicon for macOS while retaining Linux and Windows targets.
- **Evidence:**
  - `docs/release-process.md` listed `darwin/amd64` as an initial target and included an amd64 macOS example artifact.
  - `.github/workflows/release.yml` included a `darwin` + `amd64` build matrix row.
  - `_bmad-output/implementation-artifacts/2-1-define-release-artifact-matrix-and-naming-contract.md` explicitly listed `darwin/amd64` in Story 2.1 tasks.

## 2. Impact Analysis

### Epic Impact

- **Affected epic:** Epic 2 only.
- **Affected story:** Story 2.1 release matrix contract.
- **Story-level impact:**
  - Remove one OS/arch target from the canonical matrix.
  - Keep naming/version semantics unchanged.
  - Keep downstream story boundaries (2.2 build and 2.3 publish) unchanged except they now consume a 4-target matrix.

### Artifact Conflict Analysis

- **PRD:** No conflict. PRD requires cross-platform support across macOS, Linux, and Windows, but does not mandate both macOS architectures.
- **Architecture:** No conflict. Architecture requires cross-platform verification, not a mandatory `darwin/amd64` release target.
- **Release contract docs:** Requires update to remove `darwin/amd64` from ordered matrix and examples.
- **Release workflow:** Requires update to remove `darwin/amd64` matrix entry.
- **Implementation artifact (story record):** Requires update so Story 2.1 text matches the revised matrix.

### Technical Impact

- **Code/runtime:** No production code behavior changes.
- **CI/release:** One less release artifact build target.
- **Risk level:** Low.

## 3. Recommended Approach

- **Selected path:** Option 1 - Direct Adjustment.
- **Rationale:**
  - This is a contained contract change in release planning artifacts and workflow matrix.
  - No rollback or MVP scope reduction is required.
  - Keeps release pipeline deterministic while reducing unnecessary target scope.
- **Effort estimate:** Low.
- **Timeline impact:** Same day.
- **Risk assessment:** Low, provided all release-contract references remain consistent.

## 4. Detailed Change Proposals

### A) Release Contract (`docs/release-process.md`)

OLD:
- Target matrix included `darwin/amd64` as item 1.
- Example list included `prr_v1.4.0_darwin_amd64`.

NEW:
- Target matrix is now:
  1. `darwin/arm64`
  2. `linux/amd64`
  3. `linux/arm64`
  4. `windows/amd64`
- Remove macOS amd64 artifact example.

Rationale:
- Canonical release matrix no longer includes Intel macOS.

### B) Release Workflow (`.github/workflows/release.yml`)

OLD:
- Matrix included `os: darwin`, `arch: amd64`.

NEW:
- Removed `darwin/amd64` row.
- Remaining order aligned with release contract targets.

Rationale:
- Build pipeline must consume Story 2.1 contract directly.

### C) Story Artifact (`_bmad-output/implementation-artifacts/2-1-define-release-artifact-matrix-and-naming-contract.md`)

OLD:
- Story task listed initial targets including `darwin/amd64`.

NEW:
- Story task now lists: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`.

Rationale:
- Keep implementation artifact accurate and auditable.

## 5. Implementation Handoff

- **Scope classification:** Minor.
- **Route to:** Development team for direct implementation.
- **Deliverables completed:**
  - Updated release contract document.
  - Updated release workflow matrix.
  - Updated Story 2.1 implementation artifact.
- **Success criteria:**
  - No remaining `darwin/amd64` references in release target contract/workflow/story artifact.
  - Release artifact naming and SemVer rules remain unchanged.

## 6. Checklist Status Summary

### Section 1 - Trigger and Context
- [x] 1.1 Trigger issue identified.
- [x] 1.2 Core problem defined.
- [x] 1.3 Evidence gathered.

### Section 2 - Epic Impact
- [x] 2.1 Current epic viability assessed.
- [x] 2.2 Epic-level changes determined.
- [x] 2.3 Remaining epics reviewed for impact.
- [x] 2.4 Future epics invalidation/new epic need assessed.
- [x] 2.5 Ordering/priority considered.

### Section 3 - Artifact Conflict Analysis
- [x] 3.1 PRD impact reviewed.
- [x] 3.2 Architecture impact reviewed.
- [N/A] 3.3 UX impact (no UI/UX artifact scope).
- [x] 3.4 Secondary artifact impact reviewed.

### Section 4 - Path Forward Evaluation
- [x] 4.1 Direct adjustment evaluated (viable).
- [x] 4.2 Rollback evaluated (not viable/needed).
- [x] 4.3 PRD MVP review evaluated (not required).
- [x] 4.4 Recommended path selected.

### Section 5 - Proposal Components
- [x] 5.1 Issue summary prepared.
- [x] 5.2 Impact and artifact adjustments documented.
- [x] 5.3 Recommended path and rationale documented.
- [x] 5.4 MVP impact and high-level action plan defined.
- [x] 5.5 Handoff plan defined.

### Section 6 - Final Review and Handoff
- [x] 6.1 Checklist completion reviewed.
- [x] 6.2 Proposal consistency reviewed.
- [x] 6.3 Explicit user approval obtained (user response: yes).
- [N/A] 6.4 `sprint-status.yaml` update (no epic/story lifecycle change).
- [x] 6.5 Next steps and handoff confirmed.

## 7. Approval Prompt

Approved by Richard on 2026-03-07 (`yes`).

## 8. Workflow Completion

- Issue addressed: Remove `darwin/amd64` from release target matrix.
- Change scope: Minor.
- Artifacts modified: `docs/release-process.md`, `.github/workflows/release.yml`, `_bmad-output/implementation-artifacts/2-1-define-release-artifact-matrix-and-naming-contract.md`.
- Routed to: Development team for direct implementation.
