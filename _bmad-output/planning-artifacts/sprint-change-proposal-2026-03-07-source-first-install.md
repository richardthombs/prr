# Sprint Change Proposal - 2026-03-07 (Source-First Install Pivot)

## 1. Issue Summary

- **Triggering change:** Product direction change away from binary distribution.
- **Trigger type:** Strategic pivot during implementation.
- **Problem statement:** Current Epic 2 planning and documentation prioritise release artifact publishing, but this now adds unnecessary complexity (especially code signing) for a small project.
- **Evidence:**
  - User directive to stop prioritising binary distribution and instead ensure easy source build/install for Windows, Ubuntu Linux, and macOS.
  - Existing Epic 2 stories and release contract documents heavily focused on artifact packaging/publishing.

## 2. Impact Analysis

### Epic Impact

- **Affected epic:** Epic 2.
- **Current state:** Epic 2 centred on release packaging and GitHub Release publication.
- **Required change:** Reframe Epic 2 around source-first build and installation experience.
- **Other epics:** Epic 1 remains valid and unchanged.

### Artifact Conflict Analysis

- **PRD:** No blocker-level conflict with core CLI goals. The pivot narrows distribution method, not core product behaviour.
- **Epics:** Requires replacement of release artifact stories with source build/install stories.
- **Architecture:** Requires updating deployment/distribution assumptions from binary publication to local source builds.
- **UI/UX specs:** N/A (CLI project; no dedicated UI/UX artifact in scope).
- **Secondary artifacts affected:**
  - `Makefile`
  - `README.md`
  - `docs/release-process.md`
  - `docs/release-notes.md`
  - `_bmad-output/implementation-artifacts/sprint-status.yaml`

### Technical Impact

- Runtime behaviour: none.
- Contributor onboarding: significantly improved clarity and lower operational complexity.
- Release automation risk: reduced by de-scoping binary publication from current sprint focus.

## 3. Recommended Approach

- **Selected path:** Option 1 - Direct Adjustment.
- **Rationale:**
  - Change is primarily planning/documentation/tooling scope, not deep runtime architecture.
  - Existing implemented CLI behaviour remains compatible.
  - Fastest route to match current project goals and reduce process overhead.
- **Effort estimate:** Medium.
- **Risk level:** Low to medium (main risk is stale references to binary distribution left in docs/workflows).
- **Timeline impact:** Same-day updates are feasible.

## 4. Detailed Change Proposals

### Stories / Epics (`_bmad-output/planning-artifacts/epics.md`)

OLD:
- Epic 2: Release Packaging and Distribution
- Stories 2.1-2.5 focused on artifact matrix, release build pipeline, publish workflow, checksums, and release packaging docs.

NEW:
- Epic 2: Source Build and Local Installation Experience
- Stories 2.1-2.5 focused on source build/install contract, simplified Makefile workflow, OS-specific install docs, troubleshooting/uninstall guidance, and de-scoping binary publication.

Rationale:
- Align sprint scope with current product intent and reduce non-essential release complexity.

### Architecture (`_bmad-output/planning-artifacts/architecture.md`)

OLD:
- Infrastructure/deployment assumptions framed around multi-OS binary distribution.

NEW:
- Infrastructure/deployment assumptions updated to source-first distribution and local build/install.

Rationale:
- Keep architecture assumptions consistent with implementation and onboarding strategy.

### Makefile (`Makefile`)

OLD:
- Included cross-compilation packaging targets (`build-darwin`, `build-linux`, `build-windows`, `build-all`).

NEW:
- Minimal source-first targets: `build`, `install`, `test`, `clean`.
- `install` uses `go install ./cmd/prr` for platform-neutral installation.

Rationale:
- Reduce complexity and align make targets with contributor-focused source workflow.

### Installation/Contributor Docs (`README.md`, `docs/install.md`, `docs/release-process.md`, `docs/release-notes.md`)

OLD:
- README prominently referenced release contract.
- No dedicated platform install guide for macOS/Ubuntu/Windows.

NEW:
- README points to source install flow and `docs/install.md`.
- Added `docs/install.md` with prerequisites and install steps for macOS, Ubuntu Linux, and Windows.
- `docs/release-process.md` replaced with source distribution/install contract.
- `docs/release-notes.md` updated with pivot entry.

Rationale:
- Make setup practical for users building on their own platform from source.

### Sprint Tracking (`_bmad-output/implementation-artifacts/sprint-status.yaml`)

OLD:
- Epic 2 story keys tracked release artifact publication plan.

NEW:
- Epic 2 story keys track source build/install pivot outcomes.

Rationale:
- Ensure implementation tracking reflects approved sprint direction.

## 5. Implementation Handoff

- **Scope classification:** Moderate.
- **Route to:** Product Owner / Scrum Master + Development.
- **Responsibilities:**
  - PO/SM: Validate sprint scope and story tracking alignment for Epic 2.
  - Development: Maintain docs/tooling consistency and ensure source-install workflow remains verified.
- **Success criteria:**
  - New contributor can install PRR from source on macOS, Ubuntu, or Windows using documented steps.
  - Epic 2 planning and sprint tracking reference source-first goals, not binary publication.
  - Build/test verification remains green across supported OS targets.

## 6. Checklist Status Summary

### Section 1 - Understand Trigger and Context
- [x] 1.1 Triggering story/area identified (Epic 2 distribution scope).
- [x] 1.2 Core problem defined (binary release overhead for a trivial project).
- [x] 1.3 Evidence gathered (explicit user direction and artifact mismatch).

### Section 2 - Epic Impact Assessment
- [x] 2.1 Current epic viability assessed.
- [x] 2.2 Epic-level changes defined.
- [x] 2.3 Remaining epics reviewed (Epic 1 unchanged).
- [x] 2.4 Invalidated future epics identified (binary publication stories de-scoped).
- [x] 2.5 Epic ordering/priority considered (no resequencing required).

### Section 3 - Artifact Conflict and Impact Analysis
- [x] 3.1 PRD conflict check completed.
- [x] 3.2 Architecture conflict check completed.
- [N/A] 3.3 UI/UX conflict check (no UI/UX spec in scope).
- [x] 3.4 Secondary artifact impact documented.

### Section 4 - Path Forward Evaluation
- [x] 4.1 Direct adjustment evaluated (viable).
- [x] 4.2 Rollback evaluated (not needed).
- [x] 4.3 MVP review evaluated (no MVP reduction required).
- [x] 4.4 Recommended path selected (Option 1).

### Section 5 - Sprint Change Proposal Components
- [x] 5.1 Issue summary created.
- [x] 5.2 Impact and artifact changes documented.
- [x] 5.3 Path forward and rationale documented.
- [x] 5.4 MVP impact and high-level action plan documented.
- [x] 5.5 Handoff responsibilities defined.

### Section 6 - Final Review and Handoff
- [x] 6.1 Checklist completion reviewed.
- [x] 6.2 Proposal consistency verified.
- [x] 6.3 User approval interpreted from explicit implementation request.
- [x] 6.4 Sprint status tracking updated for revised epic/story scope.
- [x] 6.5 Next steps and handoff plan confirmed.

## 7. Approval Record

Approved by request context on 2026-03-07 (user requested immediate execution of this pivot).

## 8. Workflow Completion Summary

- **Issue addressed:** Binary distribution complexity no longer justified; source-first install path required.
- **Change scope:** Moderate.
- **Artifacts modified:**
  - `Makefile`
  - `README.md`
  - `docs/install.md`
  - `docs/release-process.md`
  - `docs/release-notes.md`
  - `_bmad-output/planning-artifacts/epics.md`
  - `_bmad-output/planning-artifacts/architecture.md`
  - `_bmad-output/implementation-artifacts/sprint-status.yaml`
- **Routed to:** PO/SM for backlog alignment and Development for execution continuity.
