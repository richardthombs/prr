# Sprint Change Proposal - 2026-03-07 (Release Notes Plan Removal)

## 1. Issue Summary

- **Triggering change:** Remove `docs/release-notes.md` from the release plan.
- **Trigger type:** Process simplification and planning-surface reduction.
- **Context:** Current release planning language requires recording SemVer rationale in `docs/release-notes.md`, and Story 2.1 wording references release-notes/changelog inputs. This makes release notes a planning dependency rather than a pure changelog.
- **Evidence:**
  - `docs/release-process.md` previously instructed recording bump rationale in `docs/release-notes.md`.
  - `_bmad-output/planning-artifacts/epics.md` Story 2.1 acceptance criteria referenced "release notes/changelog inputs".
  - `_bmad-output/implementation-artifacts/2-1-define-release-artifact-matrix-and-naming-contract.md` mirrored the same dependency.
  - `docs/release-notes.md` contained a release planning input template.

## 2. Impact Analysis

### Epic Impact

- **Affected epic:** Epic 2 only.
- **Affected story:** Story 2.1 release contract wording.
- **Story impact:** Clarifies that SemVer decision rationale is required, but storage location is release metadata rather than `release-notes.md`.

### Artifact Conflict Analysis

- **PRD:** No conflict.
- **Architecture:** No conflict.
- **Release contract docs:** Updated to remove hard dependency on `docs/release-notes.md`.
- **Epic planning artifacts:** Updated acceptance-criteria wording for Story 2.1.
- **Release notes document:** Converted back to a changelog-only role by removing planning template section.

### Technical Impact

- **Code/runtime:** No code changes.
- **CI/release workflow:** No workflow execution changes required.
- **Risk level:** Low.

## 3. Recommended Approach

- **Selected path:** Option 1 - Direct Adjustment.
- **Rationale:**
  - This is a documentation and planning contract update.
  - No rollback or MVP scope changes are necessary.
  - Preserves SemVer governance while reducing coupling to a specific file.
- **Effort estimate:** Low.
- **Timeline impact:** Same day.
- **Risk assessment:** Low.

## 4. Detailed Change Proposals

### A) Release Contract (`docs/release-process.md`)

OLD:
- Required recording SemVer bump rationale in `docs/release-notes.md`.

NEW:
- Requires recording SemVer bump rationale in release PR description or tag annotation.

Rationale:
- Keeps auditability while removing `release-notes.md` as a release-plan dependency.

### B) Story Planning (`_bmad-output/planning-artifacts/epics.md`)

OLD:
- Story 2.1 AC3 required rationale in release notes/changelog inputs.

NEW:
- Story 2.1 AC3 requires rationale in release preparation metadata.

Rationale:
- Neutral wording supports process flexibility.

### C) Story Implementation Artifact (`_bmad-output/implementation-artifacts/2-1-define-release-artifact-matrix-and-naming-contract.md`)

OLD:
- AC3 required rationale in release notes/changelog inputs.

NEW:
- AC3 requires rationale in release preparation metadata.

Rationale:
- Keeps implementation record aligned with planning artifact.

### D) Release Notes (`docs/release-notes.md`)

OLD:
- Included a Release Planning Input Template.

NEW:
- Template removed; file remains release changelog content.

Rationale:
- Ensures release notes are informational, not mandatory planning input.

## 5. Implementation Handoff

- **Scope classification:** Minor.
- **Route to:** Development team for direct implementation tracking.
- **Deliverables completed:**
  - Updated release planning contract language.
  - Updated Story 2.1 planning and implementation artifact wording.
  - Removed release planning template from release notes.
- **Success criteria:**
  - `release-notes.md` is not required by release planning contract.
  - SemVer bump rationale requirement remains explicit.

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
- [x] 4.2 Rollback evaluated (not needed).
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
- [x] 6.3 Explicit user approval obtained (user response: ok I approve the plan).
- [N/A] 6.4 `sprint-status.yaml` update (no epic/story lifecycle change).
- [x] 6.5 Next steps and handoff confirmed.

## 7. Approval Prompt

Approved by Richard on 2026-03-07 (`ok I approve the plan`).

## 8. Workflow Completion

- Issue addressed: Remove `docs/release-notes.md` from required release planning inputs.
- Change scope: Minor.
- Artifacts modified: `docs/release-process.md`, `_bmad-output/planning-artifacts/epics.md`, `_bmad-output/implementation-artifacts/2-1-define-release-artifact-matrix-and-naming-contract.md`, `docs/release-notes.md`.
- Routed to: Development team for direct implementation tracking.
