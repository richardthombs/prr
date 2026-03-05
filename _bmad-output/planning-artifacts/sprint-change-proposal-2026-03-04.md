# Sprint Change Proposal — 2026-03-04

## 1) Issue Summary

### Triggering Story
- Story: `1-1-initialise-cobra-cli-project-skeleton`
- Discovery context: During implementation follow-up and command-model review, the planning artifacts were validated against the source specification.

### Core Problem Statement
The current PRD/Epics/Architecture represent composable pipeline stages primarily as internal capabilities, but they do not explicitly commit all recommended composable commands from `docs/initial_specification.md` as MVP command-surface requirements.

### Evidence
- Source spec explicitly recommends these composable commands for v1 command model support:
  - `prr resolve <PR_URL>`
  - `prr mirror ensure`
  - `prr prref fetch`
  - `prr worktree add`
  - `prr diff`
  - `prr bundle`
  - `prr review-engine`
  - `prr render`
  - `prr publish` (optional)
- PRD and Epics currently prioritise a single primary command (`prr review <PR_ID>`) and capability-level decomposition.

## 2) Impact Analysis

### Epic Impact
- **Epic 1 impacted:** Needs expanded command-surface stories to make composable commands explicit MVP scope.
- **Epics 2–5 impacted:** No functional intent change; implementation remains aligned but now requires command-level exposure/UX and command contracts.
- **Epic ordering impact:** No resequencing required; changes are additive to Epic 1.

### Artifact Conflicts
- **PRD conflict:** MVP framing says “single reliable workflow from one command,” which under-specifies required composable command surface.
- **Epics conflict:** Existing stories cover capabilities but not explicit CLI command endpoints for all recommended commands.
- **Architecture conflict:** Mentions possible subcommands, but not an MVP-required list for complete composable command surface.
- **UI/UX conflict:** N/A (no separate UX spec).

### Technical Impact
- Additional command handlers under `cmd/prr/`.
- Shared pipeline stage interfaces remain valid; command wrappers should map to existing stage contracts.
- Moderate increase in test matrix (per-command CLI contract tests + JSON in/out compatibility checks).

## 3) Recommended Approach

### Selected Path
- **Option 1: Direct Adjustment (selected)**

### Rationale
- The issue is a scope-definition gap, not a fundamental architectural failure.
- Existing FRs and story decomposition already map to the same pipeline capabilities.
- Fastest and lowest-risk correction is to explicitly elevate composable commands to MVP command requirements and add concrete stories.

### Effort / Risk / Timeline
- Effort: **Medium**
- Risk: **Low-Medium** (scope increase but architecturally aligned)
- Timeline impact: **Moderate** (additional Epic 1 implementation stories before/alongside downstream epics)

## 4) Detailed Change Proposals (Old → New)

### A) PRD Updates

#### 1. PRD `Command Structure`

OLD:
- Primary command: `prr review <PR_ID>`.
- Support explicit provider/repo context resolution (automatic where possible, overridable when needed).
- Include `--keep` to retain worktree for inspection; default behaviour is cleanup.
- Use stable exit codes to distinguish user errors, provider constraints, and system/runtime failures.

NEW:
- Primary command: `prr review <PR_ID>`.
- MVP composable commands (must be implemented):
  - `prr resolve <PR_URL>`
  - `prr mirror ensure`
  - `prr prref fetch`
  - `prr worktree add`
  - `prr diff`
  - `prr bundle`
  - `prr review-engine`
  - `prr render`
  - `prr publish` (optional execution path; command must exist)
- Support explicit provider/repo context resolution (automatic where possible, overridable when needed).
- Include `--keep` to retain worktree for inspection; default behaviour is cleanup.
- All commands must support equivalent flags and JSON-compatible stdin/stdout contracts where applicable.
- Use stable exit codes to distinguish user errors, provider constraints, and system/runtime failures.

Rationale: Aligns MVP command surface with source specification and makes composable model testable and explicit.

#### 2. PRD `MVP Feature Set (Phase 1) -> Must-Have Capabilities`

OLD:
- PR context resolution and provider abstraction for merge-ref fetch.
- Bare mirror cache + lock-safe update + isolated detached worktree.
- Deterministic diff generation (`HEAD^1..HEAD`) with stat/files/patch outputs.
- v1 review bundle schema and review engine invocation.
- Markdown default rendering + optional publish integration.
- Configurable safety limits (patch bytes, changed files) and clear errors.
- Cleanup by default, `--keep` override, stable exit-code semantics.

NEW:
- PR context resolution and provider abstraction for merge-ref fetch.
- Bare mirror cache + lock-safe update + isolated detached worktree.
- Deterministic diff generation (`HEAD^1..HEAD`) with stat/files/patch outputs.
- v1 review bundle schema and review engine invocation.
- Markdown default rendering + optional publish integration.
- Configurable safety limits (patch bytes, changed files) and clear errors.
- Cleanup by default, `--keep` override, stable exit-code semantics.
- Explicit MVP command surface includes `review`, `resolve`, `mirror ensure`, `prref fetch`, `worktree add`, `diff`, `bundle`, `review-engine`, `render`, and `publish`.

Rationale: Removes ambiguity that composable commands are “nice-to-have” instead of MVP requirements.

---

### B) Epics Updates

#### 1. Epic 1 narrative and scope

OLD:
Richard can initialise PRR, run a review command with the correct PR context, and control defaults/overrides for predictable execution.

NEW:
Richard can initialise PRR, run the primary review flow, and execute MVP composable pipeline commands (`resolve`, `mirror ensure`, `prref fetch`, `worktree add`, `diff`, `bundle`, `review-engine`, `render`, `publish`) with predictable contracts and overrides.

Rationale: Makes command model explicit in the epic objective.

#### 2. Add new stories in Epic 1

NEW STORIES:
- **Story 1.5:** Implement `prr resolve <PR_URL>` command contract.
- **Story 1.6:** Implement `prr mirror ensure` and `prr prref fetch` commands.
- **Story 1.7:** Implement `prr worktree add` and cleanup/keep compatibility hooks.
- **Story 1.8:** Implement `prr diff` and `prr bundle` commands with JSON outputs.
- **Story 1.9:** Implement `prr review-engine`, `prr render`, and `prr publish` commands.

Rationale: Converts capability-level decomposition into explicit command-deliverable stories.

---

### C) Architecture Updates

#### 1. Command model clarification

OLD:
- Subcommands: verb-first lower-case (`review`, `publish`, `bundle`)

NEW:
- Subcommands: verb-first lower-case.
- MVP command set: `review`, `resolve`, `mirror ensure`, `prref fetch`, `worktree add`, `diff`, `bundle`, `review-engine`, `render`, `publish`.
- These commands are thin command-boundary adapters over existing stage contracts and must preserve deterministic JSON contracts.

Rationale: Keeps architecture aligned with revised MVP command requirements without altering core component boundaries.

## 5) Implementation Handoff

### Scope Classification
- **Moderate** (backlog reorganisation + planning artifact updates + sprint status updates)

### Handoff Recipients
- **Product Owner / Scrum Master:** Accept story additions and sprint plan updates.
- **Development Team:** Implement newly explicit command stories under Epic 1.
- **Architect (advisory):** Validate command-boundary consistency with existing stage contracts.

### Success Criteria
- PRD explicitly states composable command set as MVP requirement.
- Epics include concrete stories for command implementation.
- Architecture command model matches PRD/Epics.
- Sprint status includes added stories in backlog.

---

## Checklist Execution Record

### 1) Understand Trigger and Context
- [x] 1.1 Trigger identified (`1-1`, command-model alignment review)
- [x] 1.2 Core problem precisely defined
- [x] 1.3 Evidence captured from source specification and planning artifacts

### 2) Epic Impact Assessment
- [x] 2.1 Current epic can continue with expanded scope
- [x] 2.2 Epic-level changes identified (Epic 1 scope/story additions)
- [x] 2.3 Future epics reviewed (no structural changes required)
- [x] 2.4 No epic invalidation; additive stories needed
- [x] 2.5 No resequencing required

### 3) Artifact Conflict and Impact Analysis
- [x] 3.1 PRD conflicts identified and change points defined
- [x] 3.2 Architecture command-model updates identified
- [N/A] 3.3 UI/UX spec impacts (no UX artifact)
- [x] 3.4 Secondary artifact updates needed (`sprint-status.yaml`)

### 4) Path Forward Evaluation
- [x] 4.1 Option 1 Direct Adjustment — Viable (Effort Medium, Risk Low-Medium)
- [ ] 4.2 Option 2 Rollback — Not viable
- [ ] 4.3 Option 3 MVP Redefinition — Not viable
- [x] 4.4 Recommended path selected: Option 1

### 5) Proposal Components
- [x] 5.1 Issue summary complete
- [x] 5.2 Epic/artifact impact documented
- [x] 5.3 Recommendation and rationale provided
- [x] 5.4 MVP impact and action plan defined
- [x] 5.5 Handoff plan defined

### 6) Final Review and Handoff
- [x] 6.1 Checklist completeness verified
- [x] 6.2 Proposal consistency verified
- [x] 6.3 Explicit user approval obtained (`yes`)
- [x] 6.4 Sprint status updated with approved Epic 1 story additions
- [x] 6.5 Handoff confirmed (Moderate scope: PO/SM + Dev, Architect advisory)

## Approval & Handoff Log

- Approval decision: **Approved** by Richard on 2026-03-04.
- Scope classification: **Moderate**.
- Routed to:
  - Product Owner / Scrum Master for backlog and sprint planning alignment.
  - Development workflow for implementation of stories 1.5–1.9.
  - Architecture advisory validation for command-boundary consistency.

---

## Addendum — 2026-03-05 (Cross-Platform Story Expansion)

### Summary

Following cross-platform codebase review (macOS, Linux, Windows), Epic 1 is expanded with three additional stories to close portability gaps discovered in locking, path/test contracts, and CI verification.

### Approved Story Additions

- **Story 1.10:** Replace Unix-only mirror locking with a cross-platform lock strategy.
- **Story 1.11:** Normalise cross-platform path and test contracts.
- **Story 1.12:** Add cross-OS build and smoke verification for CLI baseline.

### Impact

- Epic ordering remains unchanged; additions are still within Epic 1.
- Primary affected artefacts: `epics.md`, `prd.md`, `architecture.md`, `implementation-readiness-report-2026-03-04.md`, `validation-report-2026-03-04.md`, and `implementation-artifacts/sprint-status.yaml`.
- Implementation tracking now includes ready-for-dev placeholders for stories 1.10–1.12.

### Handoff Delta

- **Development Team:** Prioritise Story 1.10 before wider mirror-concurrency work to unblock Windows support.
- **QA/Validation:** Re-run cross-platform regression coverage after stories 1.10–1.12 are implemented.
