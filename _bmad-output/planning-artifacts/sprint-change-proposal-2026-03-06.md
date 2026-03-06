# Sprint Change Proposal - 2026-03-06

## 1. Issue Summary

- **Triggering change:** Remove Epic 2, Epic 3, Epic 4, and Epic 5 entirely.
- **Trigger type:** Technical approach changed during implementation.
- **Context:** Epic 1 implementation already introduced internal resolve/mirror/worktree/diff orchestration stories and completed related code paths. Epic 2, Epic 3, Epic 4, and Epic 5 now duplicate capability planning and create planning/tracking noise for a simplified delivery strategy.
- **Evidence:**
  - `sprint-status.yaml` shows Epic 2/3/4/5 and all 2.x/3.x/4.x/5.x stories are still backlog-only (none started).
  - No Epic 2 story implementation artifact files exist under `_bmad-output/implementation-artifacts/2-*.md`.
  - No Epic 3 story implementation artifact files exist under `_bmad-output/implementation-artifacts/3-*.md`.
  - No Epic 4 story implementation artifact files exist under `_bmad-output/implementation-artifacts/4-*.md`.
  - No Epic 5 story implementation artifact files exist under `_bmad-output/implementation-artifacts/5-*.md`.
  - `epics.md` contains full Epic 2, Epic 3, Epic 4, and Epic 5 sections that overlap with the revised technical approach.

## 2. Impact Analysis

### Epic Impact

- **Affected epics:** Epic 2 (remove), Epic 3 (remove), Epic 4 (remove), Epic 5 (remove), Epic 1 (expand FR coverage statement), overall epic list ordering and references.
- **Story impact:**
  - Remove planned stories `2-1` to `2-4`, `3-1` to `3-4`, `4-1` to `4-3`, and `5-1` to `5-4` from active plan/tracking.
  - Remove `epic-2-retrospective`, `epic-3-retrospective`, `epic-4-retrospective`, and `epic-5-retrospective` tracking keys.
- **Dependency impact:** Minimal implementation disruption because Epic 2/3/4/5 items are backlog-only and no active implementation artifacts exist.

### Artifact Conflict Analysis

- **PRD:** No structural conflict required. FR5-FR29 remain valid requirements, but ownership mapping should move away from deleted Epic 2/3/4/5.
- **Epics document:** Requires targeted edits in three places:
  - FR Coverage Map (FR5-FR29 currently mapped to Epic 2/3/4/5)
  - Epic List (remove Epic 2, Epic 3, Epic 4, and Epic 5 entries)
  - Detailed Epic 2, Epic 3, Epic 4, and Epic 5 sections (remove full sections and stories 2.1-2.4, 3.1-3.4, 4.1-4.3, 5.1-5.4)
- **Architecture:** No direct Epic 2/3/4/5 labels detected in the architecture document section reviewed. No mandatory architecture edit required for this change.
- **Sprint tracking:** `sprint-status.yaml` requires removal of Epic 2, Epic 3, Epic 4, and Epic 5 keys.

### Technical Impact

- **Code:** No code changes required.
- **Infrastructure/CI:** No changes required.
- **Risk level:** Low to medium (primarily documentation/tracking consistency risk).

## 3. Recommended Approach

- **Selected path:** Option 1 - Direct Adjustment.
- **Rationale:**
  - Epic 2, Epic 3, Epic 4, and Epic 5 are unstarted and can be removed without rollback.
  - The requested change is a planning and tracking correction, not a runtime refactor.
  - Keeps momentum by avoiding re-plan overhead while restoring artifact consistency.
- **Effort estimate:** Low.
- **Timeline impact:** Same day.
- **Risk:** Low if all references are updated consistently.

## 4. Detailed Change Proposals

### A) Story/Epic Plan Changes (`_bmad-output/planning-artifacts/epics.md`)

#### Proposal A1: FR Coverage Map remap for FR5-FR29

OLD (representative):
- `FR5: Epic 2 - Maintain cached mirror per repository`
- `FR6: Epic 2 - Update cached repository state before processing`
- `FR14: Epic 3 - Compute deterministic PR contribution diff from merge parent`
- `FR21: Epic 4 - Submit review bundle to configured engine`
- `FR25: Epic 5 - Render Markdown report as default output`
- `...`
- `FR29: Epic 5 - Expose stage-level diagnostics for troubleshooting`

NEW:
- `FR5: Epic 1 - Covered by internal mirror/worktree orchestration stories`
- `FR6: Epic 1 - Covered by internal mirror/worktree orchestration stories`
- `FR14: Epic 1 - Covered by internal orchestration stories already completed`
- `FR21: Epic 1 - Covered by internal orchestration stories already completed`
- `FR25: Epic 1 - Covered by internal orchestration stories already completed`
- `...`
- `FR29: Epic 1 - Covered by internal orchestration stories already completed`

Rationale:
- Preserve FRs while removing Epic 2, Epic 3, Epic 4, and Epic 5 as planning containers.

#### Proposal A2: Remove Epic 2, Epic 3, Epic 4, and Epic 5 from Epic List

OLD:
- `### Epic 2: Safe Repository Snapshot and Isolated Review Workspace`
- `### Epic 3: Deterministic Diff and Bundle Preparation with Safety Gates`
- `### Epic 4: Review Engine Execution and Structured Result Handling`
- `### Epic 5: Reporting, Publication, and Automation Diagnostics`

NEW:
- Remove Epic 2, Epic 3, Epic 4, and Epic 5 entries entirely from the Epic List.
- Keep remaining epics with existing labels unless renumbering is explicitly desired.

Rationale:
- Align top-level epic inventory with requested scope.

#### Proposal A3: Remove detailed `## Epic 2`, `## Epic 3`, `## Epic 4`, and `## Epic 5` sections and stories 2.1-2.4 / 3.1-3.4 / 4.1-4.3 / 5.1-5.4

OLD:
- Full section beginning `## Epic 2: Safe Repository Snapshot and Isolated Review Workspace`
- Full section beginning `## Epic 3: Deterministic Diff and Bundle Preparation with Safety Gates`
- Full section beginning `## Epic 4: Review Engine Execution and Structured Result Handling`
- Full section beginning `## Epic 5: Reporting, Publication, and Automation Diagnostics`
- Story blocks `2.1` through `2.4`, `3.1` through `3.4`, `4.1` through `4.3`, and `5.1` through `5.4`.

NEW:
- Remove entire Epic 2, Epic 3, Epic 4, and Epic 5 section blocks.

Rationale:
- Prevent stale backlog plans from appearing active/required.

### B) Sprint Tracking Changes (`_bmad-output/implementation-artifacts/sprint-status.yaml`)

#### Proposal B1: Remove Epic 2, Epic 3, Epic 4, and Epic 5 keys from `development_status`

OLD keys:
- `epic-2`
- `2-1-create-and-maintain-per-repository-bare-mirror-cache`
- `2-2-add-lock-safe-mirror-updates-for-concurrent-runs`
- `2-3-fetch-provider-merge-snapshot-with-explicit-missing-ref-handling`
- `2-4-create-isolated-detached-worktree-and-enforce-cleanup-keep-behaviour`
- `epic-2-retrospective`
- `epic-3`
- `3-1-compute-deterministic-pr-contribution-diff`
- `3-2-produce-changed-file-list-diff-stat-and-unified-patch-outputs`
- `3-3-build-review-bundle-schema-payload`
- `3-4-enforce-configurable-input-size-limits-with-clear-failure-modes`
- `epic-3-retrospective`
- `epic-4`
- `4-1-implement-review-engine-adapter-and-bundle-submission`
- `4-2-normalise-and-validate-structured-review-response`
- `4-3-emit-stage-level-observability-and-sanitised-error-diagnostics`
- `epic-4-retrospective`
- `epic-5`
- `5-1-render-markdown-review-report-as-default-human-output`
- `5-2-provide-stable-json-output-mode-for-automation`
- `5-3-add-optional-publication-of-review-output-back-to-pr`
- `5-4-finalise-stable-exit-codes-and-automation-outcome-contracts`
- `epic-5-retrospective`

NEW:
- Remove all keys above.

Rationale:
- Tracking should only include retained epics/stories.

### C) PRD/Architecture

#### Proposal C1: PRD

- No direct section deletion required.
- Keep FR5-FR29 as product requirements.
- Optional note in PRD change log: implementation ownership consolidated outside Epic 2/3/4/5.

#### Proposal C2: Architecture

- No mandatory edits identified for this change request.

## 5. Implementation Handoff

- **Scope classification:** Moderate.
- **Recipients:** Scrum Master / Product Owner workflow track.
- **Responsibilities:**
  - Apply planning and tracking document edits.
  - Validate no remaining Epic 2/3/4/5 references in planning/tracking artifacts.
  - Confirm sprint routing remains coherent for retained scope after Epic 1.

## 6. Checklist Status Summary

### Section 1 - Trigger and Context
- [x] 1.1 Trigger story/issue identified.
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
- [N/A] 3.3 UX impact (no UX artifact detected for this change scope).
- [x] 3.4 Secondary artifact impact reviewed.

### Section 4 - Path Forward Evaluation
- [x] 4.1 Direct adjustment evaluated (viable).
- [x] 4.2 Rollback evaluated (not needed).
- [x] 4.3 PRD MVP review evaluated (not required for this scope).
- [x] 4.4 Recommended path selected.

### Section 5 - Proposal Components
- [x] 5.1 Issue summary prepared.
- [x] 5.2 Impact and artifact adjustments documented.
- [x] 5.3 Recommended path and rationale documented.
- [x] 5.4 MVP impact and high-level actions defined.
- [x] 5.5 Handoff plan defined.

### Section 6 - Final Review and Handoff
- [x] 6.1 Checklist completion reviewed.
- [x] 6.2 Proposal consistency reviewed.

## 7. Next Action

If approved, apply the document edits in:
- `_bmad-output/planning-artifacts/epics.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

Then run sprint status review via the Scrum Master workflow to confirm clean routing.
