# Sprint Change Proposal

Date: 2026-03-05  
Project: prr  
Prepared for: Richard

## 1) Issue Summary

### Triggering issue
- Trigger story: `1-8-implement-diff-and-bundle-composable-commands` (currently in review)
- Change request: pivot from a composable command set to a simpler command model:
  - one `review` command that performs full Git preparation, diffing, review execution, and emits JSON
  - one additional command that converts review JSON into Markdown

### Problem statement
The current composable-command strategy increases operational and UX complexity for this personal CLI and no longer matches the preferred product direction. The desired product behaviour is now a focused two-command model that preserves deterministic Git processing while simplifying command ergonomics.

### Evidence
- Direct product direction change from user: remove composable command surface and focus on end-to-end `review` + JSON-to-Markdown rendering command.
- Existing planning artefacts currently encode a wider composable command set that conflicts with the new direction.

## 2) Impact Analysis

### Epic impact
- **Epic 1 (in-progress):** materially impacted; stories referencing composable command contracts require rewrite/replacement.
- **Epics 2–5 (backlog):** scope remains useful at capability level (Git prep, diff, engine, rendering), but command-facing framing requires consolidation under the single `review` orchestration path.

### Story impact
- Keep with wording adjustments: `1-2`, `1-3`, `1-4`, `1-10`, `1-11`, `1-12`
- Replace/remove composable-command stories:
  - `1-5-implement-resolve-command-contract` (done) → superseded by internal resolver stage contract
  - `1-6-implement-mirror-ensure-and-prref-fetch-commands` (done) → superseded by internal Git prep stages
  - `1-7-implement-worktree-add-command-with-cleanup-keep-compatibility` (done) → superseded by internal workspace stage
  - `1-8-implement-diff-and-bundle-composable-commands` (review) → superseded by internal review pipeline stage contract
  - `1-9-implement-review-engine-render-and-publish-composable-commands` (ready-for-dev) → split and reframed

### Artifact conflicts
- **PRD conflict:** command structure and MVP command list currently require many composable commands.
- **Epics conflict:** Epic 1 title/description and multiple story definitions explicitly depend on composable commands.
- **Architecture conflict:** command surface in structure examples includes composable commands as first-class CLI commands.
- **UX design docs:** not present, no direct impact.

### Technical impact
- Code paths already implemented for composable operations can be retained internally as pipeline stages.
- CLI surface will contract to:
  - `prr review <PR_ID>` → emits structured JSON review result
  - `prr render <input>` (or stdin) → emits Markdown from JSON
- Optional publish should be deferred or made a later explicit decision to avoid conflicting scope.

## 3) Recommended Approach

### Selected path
**Option 1: Direct Adjustment** with targeted backlog reorganisation (hybrid of Option 1 + selective rollback in planning artefacts only).

### Option evaluation
- **Option 1 (Direct Adjustment):** Viable  
  Effort: Medium  
  Risk: Medium
- **Option 2 (Potential Rollback):** Not viable for code; viable only for planning wording rollback  
  Effort: Medium  
  Risk: Medium
- **Option 3 (PRD MVP Review):** Viable but unnecessary as primary path; MVP remains achievable with narrowed command UX  
  Effort: Low  
  Risk: Low

### Rationale
This preserves useful implementation work while aligning the user-facing product with the new simplified command model. It minimises waste by reframing existing completed stories as internal stage capabilities rather than external CLI contracts.

## 4) Detailed Change Proposals

### A) PRD changes

#### PRD section: MVP command surface

OLD:
- MVP composable commands (must be implemented):
  - `prr resolve <PR_URL>`
  - `prr mirror ensure`
  - `prr prref fetch`
  - `prr worktree add`
  - `prr diff`
  - `prr bundle`
  - `prr review-engine`
  - `prr render`
  - `prr publish`

NEW:
- MVP commands (must be implemented):
  - `prr review <PR_ID>`: performs context resolution, mirror/worktree preparation, deterministic diff, bundle assembly, review-engine execution, and emits structured JSON output.
  - `prr render`: consumes review JSON (file or stdin) and renders Markdown output.
- Internal stages remain modular and testable, but are not exposed as separate user-facing commands in MVP.

Rationale:
Aligns CLI ergonomics with the new product direction while preserving architectural modularity internally.

#### PRD section: Output formats

OLD:
- Default human-readable output: Markdown review report.
- Structured output mode for automation (JSON) for summary/risk/findings/checklist payloads.

NEW:
- `prr review` output: structured JSON review payload for automation and chaining.
- `prr render` output: Markdown report generated from JSON payload.
- Error output remains explicit and actionable with machine-consumable fields.

Rationale:
Matches explicit requirement to separate generation (JSON) from rendering (Markdown).

### B) Epic and story changes

#### Epic 1 title and scope

OLD:
- Epic 1: CLI Setup, Configuration, and Review Invocation
- Includes composable command implementation stories.

NEW:
- Epic 1: CLI Setup and Unified Review Orchestration
- Focuses on a single `review` command plus `render` command, with internal staged contracts.

Rationale:
Reflects pivot without discarding implementation value.

#### Story replacement proposals

1) Story replacement for `1-8`:

OLD:
- Story `1-8-implement-diff-and-bundle-composable-commands`

NEW:
- Story `1-8-implement-internal-diff-and-bundle-stages-for-review-pipeline`
- Acceptance criteria updated to validate internal stage contracts invoked by `prr review`.

Rationale:
Retains technical objective, removes external composable command requirement.

2) Story replacement for `1-9`:

OLD:
- Story `1-9-implement-review-engine-render-and-publish-composable-commands`

NEW:
- Story `1-9-implement-review-command-json-output-contract`
- New adjacent story:
  - `1-9b-implement-render-command-for-json-to-markdown`

Rationale:
Makes command responsibilities explicit and aligned to requested UX.

3) Story status reinterpretation for completed composable stories (`1-5`, `1-6`, `1-7`):

OLD:
- Completed as external command contracts.

NEW:
- Mark as completed internal capabilities reused by `prr review` pipeline.
- Add note in story artefacts: “External command contract superseded by unified review command.”

Rationale:
Avoids losing completed engineering value while aligning user-facing scope.

### C) Architecture updates

#### Architecture section: command surface

OLD:
- CLI boundary includes `review`, `resolve`, `mirror ensure`, `prref fetch`, `worktree add`, `diff`, `bundle`, `review-engine`, `render`, `publish`.

NEW:
- CLI boundary includes:
  - `review`
  - `render`
  - `version`/supporting root commands
- Previous composable units remain internal modules/stages under `internal/`.

Rationale:
Preserves modular implementation and testability without exposing broad command surface.

## 5) Implementation Handoff

### Scope classification
**Moderate** (backlog reorganisation + requirement/story reframing; no fundamental product replan required).

### Approval
- User approval: **yes** (2026-03-05)
- Proposal status: **approved for implementation**

### Handoff recipients
- **Product Owner / Scrum Master:** update PRD/epics/story narratives and sequencing.
- **Development:** refactor CLI surface to two commands while preserving internal stage modularity.
- **QA:** re-baseline command-contract tests to `review` JSON and `render` Markdown flow.

### Responsibilities
- PO/SM:
  - apply approved text edits to PRD and epics
  - update sprint statuses for superseded/reframed stories
- Dev:
  - remove/deprecate composable command entrypoints from CLI
  - ensure `review` emits stable JSON contract
  - ensure `render` accepts JSON input and outputs Markdown
- QA:
  - validate deterministic review-input pipeline behaviour remains intact
  - validate stable stdout/stderr and exit code behaviour for the two-command UX

### Success criteria
- CLI user-facing command set is reduced to `review` and `render` for MVP.
- `review` performs all required Git prep + diff + engine orchestration and emits JSON.
- `render` transforms JSON output into Markdown reliably.
- Planning artefacts no longer describe composable commands as MVP UX.

---

## Checklist Status Snapshot

### 1) Understand trigger and context
- 1.1 Trigger identified: [x] Done
- 1.2 Core problem statement: [x] Done
- 1.3 Supporting evidence collected: [x] Done

### 2) Epic impact assessment
- 2.1 Current epic impact assessed: [x] Done
- 2.2 Epic-level changes identified: [x] Done
- 2.3 Future epics reviewed: [x] Done
- 2.4 Obsolescence/new epic need checked: [x] Done
- 2.5 Priority/order changes considered: [x] Done

### 3) Artifact conflict analysis
- 3.1 PRD conflicts: [x] Done
- 3.2 Architecture conflicts: [x] Done
- 3.3 UI/UX conflicts: [N/A] Skip (no UX spec present)
- 3.4 Other artifact impacts: [x] Done

### 4) Path forward evaluation
- 4.1 Direct adjustment: [x] Viable
- 4.2 Potential rollback: [x] Not viable (except planning wording)
- 4.3 PRD MVP review: [x] Viable
- 4.4 Recommended path selected: [x] Done

### 5) Proposal components
- 5.1 Issue summary: [x] Done
- 5.2 Epic/artifact adjustments: [x] Done
- 5.3 Recommended approach and rationale: [x] Done
- 5.4 MVP impact and action plan: [x] Done
- 5.5 Handoff plan: [x] Done

### 6) Final review prep
- 6.1 Checklist completion review: [x] Done
- 6.2 Proposal consistency review: [x] Done
- 6.3 User approval obtained: [x] Done
- 6.4 sprint-status update completed: [x] Done
- 6.5 Next-step and handoff confirmed: [x] Done
