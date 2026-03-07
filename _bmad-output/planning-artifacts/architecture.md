---
stepsCompleted:
  - 1
  - 2
  - 3
  - 4
  - 5
  - 6
  - 7
  - 8
inputDocuments:
  - /Users/richardthombs/dev/prr/_bmad-output/planning-artifacts/prd.md
  - /Users/richardthombs/dev/prr/docs/initial_specification.md
workflowType: 'architecture'
lastStep: 8
status: 'complete'
completedAt: '2026-03-04'
project_name: 'prr'
user_name: 'Richard'
date: '2026-03-04'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**
The PRD defines a single-command review workflow with five capability clusters:
1) PR identification and context resolution,
2) repository snapshot and mirror/worktree lifecycle,
3) deterministic diff and bundle generation,
4) review-engine invocation and result identity stability,
5) rendering/publication plus automation-friendly behaviour.
Architecturally, this implies a staged pipeline with clear stage contracts and failure boundaries.

**Non-Functional Requirements:**
Architecture-shaping NFRs are:
- deterministic terminal states and rerun equivalence for unchanged refs,
- strict non-interference with active local working copy,
- concurrency-safe shared mirror updates,
- explicit and stable machine-readable integration surfaces (JSON + exit codes),
- security controls for secret handling and sanitised diagnostics,
- operability via stage-level troubleshooting signals.
These require strong separation of concerns and robust boundary contracts.

**Scale & Complexity:**
Focused v1 product scope with moderate implementation complexity due to Git/provider edge cases and determinism requirements.

- Primary domain: cli_tool / developer tooling
- Complexity level: medium
- Estimated architectural components: 8

### Technical Constraints & Dependencies

- Requires provider merge snapshot refs; missing merge refs must fail fast with actionable diagnostics.
- Uses cached bare mirrors plus detached worktrees; architecture must enforce lock-safe concurrent access.
- Runtime and command behaviour must support macOS, Linux, and Windows with equivalent contracts.
- Must apply configurable safety limits (patch bytes, changed files) before review-engine invocation.
- Must preserve predictable output contracts (Markdown default + structured JSON mode).
- Must support optional publish step without coupling core review flow to a single provider implementation.

### Cross-Cutting Concerns Identified

- Determinism and reproducibility across runs.
- Isolation and workspace safety guarantees.
- Concurrency and lock management around repository mirrors.
- Error taxonomy and actionable diagnostics.
- Observability/troubleshooting at each pipeline stage.
- Security hygiene for credentials and logs.
- Configuration precedence and validation.
- Automation compatibility (stable schema, stdout/stderr behaviour, exit semantics).

## Starter Template Evaluation

### Primary Technology Domain

CLI tool (Go) based on project requirements analysis.

### Starter Options Considered

1. Cobra CLI starter (`spf13/cobra`)
  - Strong command/subcommand model, mature ecosystem, common for production Go CLIs.
2. urfave/cli
  - Lightweight and flexible, good for straightforward command apps.
3. No heavy framework + standard `flag` package
  - Minimal dependencies, maximum control, but more scaffolding effort.

### Selected Starter: Cobra CLI starter

**Rationale for Selection:**
Cobra best fits a command-oriented tool with a focused MVP subcommand set (`review`). It balances structure, maintainability, and ecosystem maturity for an intermediate solo developer workflow while keeping complex Git and engine orchestration as internal stages.

**Initialization Command:**

```bash
go install github.com/spf13/cobra-cli@latest
cobra-cli init --pkg-name github.com/richardthombs/prr
```

**Architectural Decisions Provided by Starter:**

**Language & Runtime:**
Go modules, idiomatic package layout, compiled single-binary CLI runtime.

**Styling Solution:**
N/A for terminal-first CLI; output formatting handled by renderer modules.

**Build Tooling:**
Standard Go build/test toolchain; easy CI integration.

**Testing Framework:**
Go `testing` package baseline; command handlers can be unit-tested by package boundaries.

**Code Organization:**
`cmd/` entrypoints + internal packages for provider, git workspace, bundle, engine, renderer.

**Development Experience:**
Fast iteration with `go test`, `go run`, and straightforward cross-platform source builds/install flows.

**Note:** Project initialization using this command should be the first implementation story.

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
- State model limited to repository mirrors and detached worktrees only (no metadata database).
- Configuration hierarchy and validation contract.
- Provider/review-engine interface boundaries.
- Error taxonomy and exit-code contract.
- Secrets/auth handling model.

**Important Decisions (Shape Architecture):**
- Logging and observability structure.
- Source distribution and installation strategy.
- CI quality gates and test layers.

**Deferred Decisions (Post-MVP):**
- Optional metadata persistence layer (only if future requirements demand it).
- Multi-engine orchestration.
- Team-scale policy/profile management.

### Data Architecture

- **Primary state:** Filesystem-based Git mirror/worktree cache only.
- **Metadata model:** Stateless execution; no application database for run history or publish state.
- **Validation boundary:** Strict schema validation for inbound/outbound bundle/review payloads.
- **Caching:** Filesystem-based Git mirror/worktree cache as first-class operational storage.

### Authentication & Security

- **Credential source:** Environment variables / host-native credential helpers (no secret persistence in PRR-owned files).
- **Authorization model:** Provider tokens scoped to least privilege.
- **Redaction policy:** Structured redaction for logs/errors (tokens, headers, secret URLs).
- **At-rest protection:** Keep sensitive transient artifacts out of persisted logs by default.

### API & Communication Patterns

- **Internal architecture style:** Pipeline orchestration with explicit stage contracts.
- **Provider abstraction:** `PRProvider` boundary for resolve + publish.
- **Review engine abstraction:** `ReviewEngine` boundary for bundle-in / review-out.
- **Error contract:** Typed error classes mapped to stable non-zero exit codes.
- **I/O contract:** Markdown default, JSON mode for automation, stable stdout/stderr channeling and schema semantics.

### Frontend Architecture

Not applicable (CLI product). Output rendering uses Markdown formatter + JSON serializer modules.

### Infrastructure & Deployment

- **Build/distribution:** Source-first distribution; users build from repository checkout on their platform.
- **CI baseline:** lint + unit tests + integration tests (Git/provider contract focus) + smoke CLI run.
- **Cross-platform verification:** run build/test/smoke matrix across macOS, Linux, and Windows.
- **Runtime target:** Local developer machine execution (no mandatory hosted runtime for MVP).
- **Monitoring/diagnostics:** Stage-level event logs and debug mode with per-run trace IDs.

### Decision Impact Analysis

**Implementation Sequence:**
1. Core contracts (`PRRef`, workspace, bundle, review, error taxonomy)
2. Git workspace/mirror subsystem with locking
3. Provider adapter + review-engine adapter
4. Bundle/render pipeline + JSON/Markdown outputs
5. Publish adapter + integration tests + source-install documentation and onboarding

**Cross-Component Dependencies:**
- Error taxonomy affects all command handlers and renderer outputs.
- Config model affects provider, engine, and safety-limit enforcement.
- Review-output variability is accepted; only review-input generation is required to be deterministic.
- Logging/redaction policy applies across all stage executors.

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Critical Conflict Points Identified:**
11 areas where AI agents could make different choices and introduce inconsistency.

### Naming Patterns

**Filesystem State Naming Conventions:**
- Mirror paths: deterministic repo-hash directories (`~/.cache/prr/repos/<repoHash>.git`)
- Worktree paths: scoped by repo hash + PR + run id (`~/.cache/prr/work/<repoHash>/pr-<PR_ID>/<runId>/`)
- Internal refs: `refs/prr/pull/<PR_ID>/merge`
- Lock files: one lock per repo hash adjacent to mirror path

**API Naming Conventions:**
- Internal JSON contracts: `camelCase` for bundle/review payloads exposed to external engines.
- CLI flags: kebab-case (`--output-format`, `--max-patch-bytes`)
- Env vars: upper snake case (`PRR_PROVIDER`, `PRR_MAX_FILES`)
- Subcommands: verb-first lower-case (`review`, `version`)

**Code Naming Conventions:**
- Go packages: short lowercase (`provider`, `workspace`, `renderer`)
- Exported identifiers: `PascalCase`; unexported: `camelCase`
- Files: lowercase with underscores only where clarity improves (`review_engine.go`, `error_codes.go`)
- Interfaces: capability-oriented names (`PRProvider`, `ReviewEngine`, `GitWorkspace`)

### Structure Patterns

**Project Organization:**
- `cmd/prr/` for CLI entrypoint and command wiring.
- `internal/` for implementation modules (provider, git, bundle, engine, render, config, logging).
- `pkg/` only if a true public library surface emerges (not default for MVP).
- Tests:
  - unit tests co-located (`*_test.go`)
  - integration tests under `test/integration/`

**File Structure Patterns:**
- Configuration:
  - defaults in code
  - optional user config file under standard OS config dir
  - env + flags override precedence
- Docs in `docs/`, architecture/planning artifacts in `_bmad-output/planning-artifacts/`

### Format Patterns

**API Response Formats:**
- Internal result envelopes for machine output:
  - success: `{ "status": "ok", "data": ... }`
  - error: `{ "status": "error", "error": { "code": "...", "message": "...", "details": ... } }`
- Markdown remains default human output.

**Data Exchange Formats:**
- External bundle/review JSON fields remain `camelCase` to match PRD structures.
- Timestamps:
  - serialized as UTC ISO-8601 strings where emitted in JSON/log output
- Nulls: explicit `null` in JSON; avoid sentinel values.

### Communication Patterns

**Event System Patterns:**
- Stage event names: `stage.<name>.<state>` (e.g., `stage.fetch.started`, `stage.fetch.completed`)
- Structured payload keys: `runId`, `stage`, `provider`, `prId`, `durationMs`, `result`
- Logging levels: `debug`, `info`, `warn`, `error`; no ad-hoc levels.

**State Management Patterns:**
- Pipeline context object passed explicitly across stages (no hidden global mutable state).
- Stage outputs are immutable value objects.
- Retries handled by explicit retry wrappers, not inline ad-hoc loops.

### Process Patterns

**Error Handling Patterns:**
- All errors map to typed domain errors with stable code classes:
  - `CONFIG_*`, `PROVIDER_*`, `LIMIT_*`, `ENGINE_*`, `RUNTIME_*`
- CLI exit code mapping is centralized in one module.
- User-facing errors are actionable; debug detail only in debug mode.

**Loading State Patterns:**
- Every stage emits `started` and terminal (`completed`/`failed`) events.
- Long-running operations include periodic progress log points.
- No silent waits beyond a short threshold.

### Enforcement Guidelines

**All AI Agents MUST:**
- Follow naming/format conventions exactly (no mixed snake/camel within the same boundary).
- Use centralized error and logging abstractions instead of custom per-module patterns.
- Preserve stage contract inputs/outputs without side effects on unrelated modules.

**Pattern Enforcement:**
- PR checklist includes naming/format/error-contract checks.
- CI runs:
  - formatting/lint
  - unit + integration tests
  - contract fixture validation for JSON payloads
- Pattern exceptions must be documented in architecture notes before merge.

### Pattern Examples

**Good Examples:**
- CLI flag: `--max-patch-bytes`
- Env var: `PRR_MAX_PATCH_BYTES`
- Worktree path: `~/.cache/prr/work/<repoHash>/pr-12345/<runId>/`
- Event: `stage.bundle.completed`
- Error code: `PROVIDER_MERGE_REF_MISSING`

**Anti-Patterns:**
- Mixed naming in same boundary (`max_patch_bytes` and `maxPatchBytes` together in one JSON contract)
- Ad-hoc exit codes inside command handlers
- Direct token logging in provider/engine adapters
- Hidden mutable package-level state for pipeline context

## Project Structure & Boundaries

### Complete Project Directory Structure

```text
prr/
├── README.md
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
├── .editorconfig
├── .env.example
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── cmd/
│   └── prr/
│       ├── main.go
│       ├── root.go
│       ├── render_helpers.go
│       ├── review.go
│       └── version.go
├── internal/
│   ├── app/
│   │   ├── pipeline.go
│   │   ├── context.go
│   │   └── stage_runner.go
│   ├── config/
│   │   ├── config.go
│   │   ├── load.go
│   │   ├── validate.go
│   │   └── precedence.go
│   ├── provider/
│   │   ├── provider.go
│   │   ├── github/
│   │   │   ├── resolve.go
│   │   │   └── publish.go
│   │   └── azure/
│   │       ├── resolve.go
│   │       └── publish.go
│   ├── git/
│   │   ├── mirror.go
│   │   ├── fetch.go
│   │   ├── worktree.go
│   │   ├── diff.go
│   │   └── lock.go
│   ├── bundle/
│   │   ├── bundle.go
│   │   ├── schema.go
│   │   └── limits.go
│   ├── engine/
│   │   ├── engine.go
│   │   ├── client.go
│   │   └── normalize.go
│   ├── render/
│   │   ├── markdown.go
│   │   ├── json.go
│   │   └── sections.go
│   ├── publish/
│   │   └── publish.go
│   ├── errors/
│   │   ├── codes.go
│   │   ├── classify.go
│   │   └── exit_codes.go
│   ├── logging/
│   │   ├── logger.go
│   │   ├── events.go
│   │   └── redact.go
│   └── types/
│       ├── prref.go
│       ├── workspace.go
│       ├── bundle.go
│       └── review.go
├── test/
│   ├── integration/
│   │   ├── review_flow_test.go
│   │   ├── merge_ref_missing_test.go
│   │   └── limits_test.go
│   ├── fixtures/
│   │   ├── bundle/
│   │   └── review/
│   └── helpers/
│       └── gitrepo.go
├── docs/
│   ├── initial_specification.md
│   ├── architecture-notes.md
│   └── operational-runbook.md
└── _bmad-output/
  ├── planning-artifacts/
  └── implementation-artifacts/
```

### Architectural Boundaries

**API Boundaries:**
- CLI boundary at `cmd/prr/*` only.
- MVP command surface at CLI boundary is `review` (Markdown default, `--json` for automation).
- Provider boundary at `internal/provider/*` (`resolve`, optional `publish`).
- Review engine boundary at `internal/engine/*` (bundle in, review out).

**Component Boundaries:**
- `internal/app` orchestrates stages; no direct provider/git logic in command files.
- `internal/git` owns mirror/fetch/worktree/diff lifecycle.
- `internal/bundle` owns payload construction + limit enforcement.
- `internal/render` owns output formatting only.

**Service Boundaries:**
- `provider` and `engine` communicate via typed contracts in `internal/types`.
- `publish` uses provider abstraction, never bypasses it.
- No runtime metadata store is required for MVP; state remains in mirrors/worktrees only.

**Data Boundaries:**
- External contracts: `camelCase` JSON payloads.
- Internal state: mirror/worktree filesystem artifacts only.
- Secrets never written to logs, output payloads, or temporary artifacts.

### Requirements to Structure Mapping

**Feature/FR Mapping:**
- FR1–FR4 (PR identification/context): `cmd/prr/review.go`, `internal/provider/*`, `internal/config/*`
- FR5–FR13 (mirror/worktree isolation): `internal/git/*`
- FR14–FR20 (diff/bundle/limits): `internal/git/diff.go`, `internal/bundle/*`
- FR21–FR24 (review engine): `internal/engine/*`, `internal/errors/*`
- FR25–FR30 (output/publish/automation): `internal/render/*`, `internal/publish/*`, `cmd/prr/*`, `internal/logging/*`

**Cross-Cutting Concerns:**
- Error taxonomy + exits: `internal/errors/*`
- Redaction + stage events: `internal/logging/*`
- Config precedence/validation: `internal/config/*`
- Contracts and stable types: `internal/types/*`

### Integration Points

**Internal Communication:**
- `cmd` → `app` orchestration → stage modules (`provider` → `git` → `bundle` → `engine` → `render`/`publish`)
- Shared typed objects passed through explicit pipeline context.

**External Integrations:**
- Git remotes via system git commands in `internal/git`.
- PR providers via `internal/provider/github` and `internal/provider/azure`.
- Review service through `internal/engine/client.go`.

**Data Flow:**
1. Resolve PR reference
2. Ensure/update mirror + fetch merge ref
3. Create detached worktree
4. Generate stat/files/patch
5. Build bundle + enforce limits
6. Submit to engine
7. Render output and optionally publish
8. Cleanup or keep worktree

### File Organization Patterns

**Configuration Files:**
- Runtime config from defaults + file + env + flags, merged in `internal/config/precedence.go`.

**Source Organization:**
- Domain modules under `internal/`, command wiring under `cmd/`, contracts under `internal/types`.

**Test Organization:**
- Unit tests co-located; integration flow tests under `test/integration`; fixtures under `test/fixtures`.

**Asset Organization:**
- No frontend assets required; docs and templates under `docs/` and `_bmad-output/`.

### Development Workflow Integration

**Development Server Structure:**
- N/A (CLI). Local execution via `go run ./cmd/prr`.

**Build Process Structure:**
- Source build/test/install targets in `Makefile`; CI workflows in `.github/workflows`.

**Deployment Structure:**
- Repository source checkout + local build/install on target OS; no mandatory hosted runtime.

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:**
Core decisions are compatible: Go + Cobra command model, staged pipeline orchestration, provider/review-engine abstractions, filesystem state model (mirrors/worktrees only), and centralized error/logging contracts all align without architectural contradiction.

**Pattern Consistency:**
Naming, structure, communication, and process patterns reinforce the architecture choices and reduce agent-level variance (especially around error contracts, JSON boundaries, and stage events).

**Structure Alignment:**
The project tree supports all major decisions, with clear boundaries for `cmd`, orchestration, adapters, filesystem state handling, rendering, and cross-cutting concerns.

### Requirements Coverage Validation ✅

**Epic/Feature Coverage:**
No epic document provided; FR-category mapping is complete and concrete.

**Functional Requirements Coverage:**
All FR groups are architecturally mapped:
- PR resolution/context → provider/config/cmd
- mirror/worktree isolation → git subsystem
- diff/bundle/limits → git+bundle
- review orchestration → engine+errors
- output/publish/automation → render/publish/cmd/logging

**Non-Functional Requirements Coverage:**
- Determinism: explicit staged contracts for review-input generation, typed boundaries
- Security: redaction + secret-handling policy
- Reliability: stable exit code taxonomy + lifecycle ownership
- Integration: machine-readable JSON contracts + output conventions
- Maintainability: modular boundaries and test layering

### Implementation Readiness Validation ✅

**Decision Completeness:**
Critical implementation-blocking decisions are documented; deferred items are explicitly marked post-MVP.

**Structure Completeness:**
Directory/file structure is specific enough to begin story implementation and testing.

**Pattern Completeness:**
Conflict-prone areas (naming, errors, events, flags/env/config precedence) are covered with enforceable rules and examples.

### Gap Analysis Results

**Critical Gaps:** None identified.

**Important Gaps:**
- Live web verification for current dependency versions could not be confirmed in this environment; versions should be verified at implementation kickoff.
- Engine contract fixture examples should be added early to lock schema expectations.

**Nice-to-Have Gaps:**
- Add an ADR index in `docs/architecture-notes.md` as decisions evolve.
- Add a short contribution conventions doc referencing pattern rules.

### Validation Issues Addressed

- Requirement-to-module mappings were made explicit to avoid implementation ambiguity.
- Cross-cutting concerns were centralized (errors/logging/config contracts) to prevent drift.

### Architecture Completeness Checklist

**✅ Requirements Analysis**
- [x] Project context thoroughly analyzed
- [x] Scale and complexity assessed
- [x] Technical constraints identified
- [x] Cross-cutting concerns mapped

**✅ Architectural Decisions**
- [x] Critical decisions documented with versions strategy
- [x] Technology stack fully specified
- [x] Integration patterns defined
- [x] Performance/operability considerations addressed

**✅ Implementation Patterns**
- [x] Naming conventions established
- [x] Structure patterns defined
- [x] Communication patterns specified
- [x] Process patterns documented

**✅ Project Structure**
- [x] Complete directory structure defined
- [x] Component boundaries established
- [x] Integration points mapped
- [x] Requirements-to-structure mapping complete

### Architecture Readiness Assessment

**Overall Status:** READY FOR IMPLEMENTATION

**Confidence Level:** High

**Key Strengths:**
- Strong boundary clarity for multi-agent implementation
- Deterministic review-input pipeline aligned with PRD intent
- Explicit cross-cutting standards for errors/logging/config

**Areas for Future Enhancement:**
- Incremental review architecture extensions
- richer publish integrations and policy packs
- optional multi-engine orchestration strategy

### Implementation Handoff

**AI Agent Guidelines:**
- Follow architectural decisions exactly as documented.
- Apply implementation patterns uniformly across modules.
- Respect module boundaries and stage contracts.
- Use centralized error/logging/config abstractions.

**First Implementation Priority:**
Initialize the Go/Cobra project scaffold, then implement core contracts and stage orchestration shell.
