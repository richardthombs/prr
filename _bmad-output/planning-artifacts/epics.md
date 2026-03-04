---
stepsCompleted:
  - step-01-validate-prerequisites
  - step-02-design-epics
  - step-03-create-stories
  - step-04-final-validation
inputDocuments:
  - /Users/richardthombs/dev/prr/_bmad-output/planning-artifacts/prd.md
  - /Users/richardthombs/dev/prr/_bmad-output/planning-artifacts/architecture.md
---

# prr - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for prr, decomposing the requirements from the PRD, UX Design if it exists, and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: Richard can start a review by providing a pull request identifier.
FR2: Richard can run a review without switching away from his current working copy.
FR3: Richard can have PRR resolve repository and remote context for the requested PR.
FR4: Richard can override inferred repository/provider context when defaults are incorrect.
FR5: PRR can maintain a cached mirror per repository for repeat reviews.
FR6: PRR can update cached repository state before review processing.
FR7: PRR can prevent concurrent corruption when up to 5 concurrent reviews target the same repository.
FR8: PRR can fetch a PR merge snapshot into an internal review namespace.
FR9: PRR can fail with an explicit message when required merge snapshot refs are unavailable.
FR10: PRR can create an isolated workspace for each review run.
FR11: PRR can ensure review execution does not modify Richard’s active local working copy.
FR12: Richard can keep an isolated workspace for investigation when requested.
FR13: PRR can remove transient review workspaces automatically when retention is not requested.
FR14: PRR can compute the PR contribution diff using merge-parent comparison semantics.
FR15: PRR can produce a changed-file list for the PR contribution.
FR16: PRR can produce a diff stat summary for the PR contribution.
FR17: PRR can produce a unified patch for the PR contribution.
FR18: PRR can build a review bundle containing required metadata, stat, files, and patch fields.
FR19: PRR can enforce configurable review-input size limits before engine invocation.
FR20: PRR can fail with clear diagnostics when review-input limits are exceeded.
FR21: PRR can submit the generated review bundle to a configured review engine.
FR22: PRR can receive structured review output containing summary, risk, findings, and checklist.
FR23: PRR can include finding identifiers for references within a single review result, with no requirement to correlate findings across reruns.
FR24: PRR can surface review-engine failures with actionable error context.
FR25: Richard can receive a Markdown review report as the default output.
FR26: Richard can request structured JSON output for automation workflows.
FR27: Richard can optionally publish review results back to the pull request.
FR28: PRR can return stable outcome signalling suitable for shell/CI scripting.
FR29: PRR can expose stage-level diagnostics to support troubleshooting.
FR30: Richard can configure default behaviours and override them per run.

### NonFunctional Requirements

NFR1: For typical PRs within configured size limits, PRR should produce a rendered review within 90 seconds on Richard’s normal development machine and network.
NFR2: PRR should provide visible stage progress or clear terminal feedback at each major pipeline stage to avoid perceived hangs.
NFR3: PRR should fail fast (within 5 seconds of detection) when mandatory preconditions are missing (e.g., merge ref unavailable, invalid config).
NFR4: PRR must not persist secrets in logs, review artifacts, or temporary files.
NFR5: PRR must use least-privilege credentials for provider/review-engine operations and rely on externally managed auth mechanisms.
NFR6: PRR must isolate review workspaces so no writes occur in Richard’s active repository working copy.
NFR7: PRR must sanitise error output to avoid leaking tokens, secret URLs, or sensitive headers.
NFR8: PRR must complete or fail with a deterministic terminal state; partial runs must not leave ambiguous review outcomes.
NFR9: PRR must clean transient worktrees by default and support explicit retention only via --keep.
NFR10: PRR must prevent concurrent corruption of shared mirror state via per-repository locking.
NFR11: Re-running the same review command against unchanged source refs should produce functionally equivalent bundle content.
NFR12: PRR must support stable, machine-readable JSON output for automation use cases.
NFR13: PRR must return stable non-zero exit codes by error class (configuration, provider/ref, limit, engine/runtime).
NFR14: PRR must keep stdout/stderr behaviour consistent across versions for script compatibility.
NFR15: PRR must emit stage-level diagnostics sufficient to troubleshoot failures without manual Git forensics in most cases.
NFR16: Configuration validation errors must identify offending fields and expected value format.
NFR17: Internal module boundaries (provider, git workspace, bundle, engine, renderer) must remain separable to enable incremental changes without full rewrites.

### Additional Requirements

- Starter template requirement: initialise the project with Cobra CLI starter (`spf13/cobra`) and make this the first implementation story, using `go install github.com/spf13/cobra-cli@latest` and `cobra-cli init --pkg-name github.com/richardthombs/prr`.
- Architecture must use a staged pipeline with explicit stage contracts and failure boundaries.
- Primary runtime state must be filesystem-based Git mirror/worktree cache only; no application database in MVP.
- Concurrency-safe mirror updates are required via per-repository locking.
- Provider behaviour must be abstracted behind a `PRProvider` contract (resolve and optional publish).
- Review engine integration must be abstracted behind a `ReviewEngine` contract (bundle-in/review-out).
- Typed error taxonomy and stable non-zero exit code mapping are required (`CONFIG_*`, `PROVIDER_*`, `LIMIT_*`, `ENGINE_*`, `RUNTIME_*`).
- Stage-level structured events and diagnostics must be emitted for observability/troubleshooting.
- Logs and errors must apply redaction for secrets/tokens/headers and avoid secret persistence.
- Configuration precedence must be deterministic (defaults + config file + env + flags) with strict validation.
- Output contract must remain stable: Markdown default, JSON mode for automation, consistent stdout/stderr behaviour.
- Build and release model must target single static CLI binaries per OS/arch, with CI covering lint, unit, integration, and smoke paths.
- Project structure boundaries should separate orchestration, git workspace lifecycle, bundle generation, engine invocation, rendering, publishing, and cross-cutting concerns.

### FR Coverage Map

FR1: Epic 1 - Start a review by PR identifier
FR2: Epic 1 - Run review without leaving current working copy
FR3: Epic 1 - Resolve repository/remote context for requested PR
FR4: Epic 1 - Override inferred provider/repository context
FR5: Epic 2 - Maintain cached mirror per repository
FR6: Epic 2 - Update cached repository state before processing
FR7: Epic 2 - Prevent concurrent mirror corruption with locking
FR8: Epic 2 - Fetch PR merge snapshot into internal namespace
FR9: Epic 2 - Fail clearly when merge snapshot refs are unavailable
FR10: Epic 2 - Create isolated workspace per review run
FR11: Epic 2 - Guarantee no modification of active local working copy
FR12: Epic 2 - Support optional retained workspace for investigation
FR13: Epic 2 - Auto-clean transient workspaces unless retention requested
FR14: Epic 3 - Compute deterministic PR contribution diff from merge parent
FR15: Epic 3 - Produce changed-file list for contribution
FR16: Epic 3 - Produce diff stat summary
FR17: Epic 3 - Produce unified patch output
FR18: Epic 3 - Build complete review bundle payload
FR19: Epic 3 - Enforce configurable input size limits before engine call
FR20: Epic 3 - Fail with actionable diagnostics on limit exceedance
FR21: Epic 4 - Submit review bundle to configured engine
FR22: Epic 4 - Receive structured review output
FR23: Epic 4 - Include finding IDs stable within a single review run
FR24: Epic 4 - Surface review-engine failures with actionable context
FR25: Epic 5 - Render Markdown report as default output
FR26: Epic 5 - Support structured JSON output mode
FR27: Epic 5 - Optionally publish review results to the PR
FR28: Epic 5 - Return stable outcome signalling for shell/CI automation
FR29: Epic 5 - Expose stage-level diagnostics for troubleshooting
FR30: Epic 1 - Configure defaults with per-run override controls

## Epic List

### Epic 1: CLI Setup, Configuration, and Review Invocation
Richard can initialise PRR, run a review command with the correct PR context, and control defaults/overrides for predictable execution.
**FRs covered:** FR1, FR2, FR3, FR4, FR30

### Epic 2: Safe Repository Snapshot and Isolated Review Workspace
Richard can run reviews against an isolated merge snapshot with mirror/worktree lifecycle safety and zero interference with active local work.
**FRs covered:** FR5, FR6, FR7, FR8, FR9, FR10, FR11, FR12, FR13

### Epic 3: Deterministic Diff and Bundle Preparation with Safety Gates
Richard can reliably generate deterministic PR inputs (files, stat, patch, bundle) with configurable limits and explicit limit-failure behaviour.
**FRs covered:** FR14, FR15, FR16, FR17, FR18, FR19, FR20

### Epic 4: Review Engine Execution and Structured Result Handling
Richard can submit review inputs to a configured engine and receive structured outputs/failures suitable for decision-making.
**FRs covered:** FR21, FR22, FR23, FR24

### Epic 5: Reporting, Publication, and Automation Diagnostics
Richard can consume Markdown/JSON outputs, optionally publish results, and integrate PRR into scripts with stable signalling and diagnostics.
**FRs covered:** FR25, FR26, FR27, FR28, FR29

## Epic 1: CLI Setup, Configuration, and Review Invocation

Richard can initialise PRR, run the primary review flow, and execute MVP composable pipeline commands (`resolve`, `mirror ensure`, `prref fetch`, `worktree add`, `diff`, `bundle`, `review-engine`, `render`, `publish`) with predictable contracts and overrides.

### Story 1.1: Initialise Cobra CLI Project Skeleton

**FRs:** FR1

As Richard,
I want a Cobra-based CLI scaffold initialised for PRR,
So that implementation starts from the architecture-approved starter template.

**Acceptance Criteria:**

**Given** a clean repository
**When** I run the documented Cobra initialisation commands
**Then** the PRR CLI project skeleton is created with a working root command
**And** the scaffold builds and runs a help command successfully.

**Given** the architecture decision requiring Cobra as first implementation story
**When** the scaffold is committed
**Then** command wiring and package layout align with the approved architecture boundaries
**And** no alternate CLI framework is introduced.

### Story 1.2: Implement Review Command Entry and PR Identifier Input

**FRs:** FR1, FR2

As Richard,
I want to run `prr review <PR_ID>` from my current shell location,
So that I can trigger a review quickly without manual setup.

**Acceptance Criteria:**

**Given** a valid PR identifier
**When** I run `prr review <PR_ID>`
**Then** the CLI accepts the input and enters the review pipeline
**And** it does not require changing my active repository working copy.

**Given** missing or malformed command arguments
**When** I run the review command
**Then** I receive actionable usage guidance
**And** the command exits with a stable user-error exit code.

### Story 1.3: Add Context Resolution and Explicit Override Controls

**FRs:** FR3, FR4

As Richard,
I want automatic provider/repository context resolution with explicit override flags,
So that reviews run correctly even when inferred defaults are wrong.

**Acceptance Criteria:**

**Given** resolvable local or configured context
**When** I run a review command without overrides
**Then** PRR resolves provider and repository context correctly
**And** records the resolved context in stage diagnostics.

**Given** incorrect inferred context
**When** I pass explicit override options
**Then** PRR uses the override values for the run
**And** the effective configuration reflects documented precedence.

### Story 1.4: Implement Config Loading, Validation, and Precedence

**FRs:** FR30

As Richard,
I want defaults from config with environment and flag overrides,
So that behaviour is predictable and tunable per run.

**Acceptance Criteria:**

**Given** defaults, a config file, environment variables, and CLI flags
**When** PRR resolves effective configuration
**Then** precedence is deterministic (flags > env > file > defaults)
**And** the resolved values are available to downstream stages.

**Given** invalid configuration values
**When** I run the command
**Then** PRR fails fast with field-specific validation errors
**And** returns a stable configuration error class exit code.

### Story 1.5: Implement Resolve Command Contract

**FRs:** FR1, FR3, FR4

As Richard,
I want `prr resolve <PR_URL>` to emit deterministic PR reference context,
So that I can script and verify context resolution independently of the full review flow.

**Acceptance Criteria:**

**Given** a valid PR URL and resolvable context
**When** I run `prr resolve <PR_URL>`
**Then** PRR emits a stable JSON `PRRef` payload
**And** supports equivalent override flags and provider auto-detection for supported URL formats.

**Given** missing or invalid context inputs
**When** resolution fails
**Then** PRR returns actionable diagnostics
**And** exits with a stable error-class code.

### Story 1.6: Implement Mirror Ensure and PRRef Fetch Commands

**FRs:** FR5, FR6, FR8, FR9

As Richard,
I want `prr mirror ensure` and `prr prref fetch` commands,
So that mirror lifecycle and merge-ref acquisition are composable and testable.

**Acceptance Criteria:**

**Given** repo context input
**When** I run `prr mirror ensure`
**Then** PRR creates or updates the deterministic mirror location
**And** emits JSON including `bareDir`.

**Given** valid PR context and mirror state
**When** I run `prr prref fetch`
**Then** PRR fetches merge ref into `refs/prr/pull/<PR_ID>/merge`
**And** emits JSON including resolved `mergeRef`.

**Given** `--verbose` is enabled
**When** these commands invoke external commands
**Then** PRR logs each external command to stderr before execution
**And** preserves JSON payloads on stdout.

**Given** `--what-if` is enabled
**When** these commands run
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

### Story 1.7: Implement Worktree Add Command with Cleanup/Keep Compatibility

**FRs:** FR10, FR11, FR12, FR13

As Richard,
I want `prr worktree add` to create isolated review worktrees,
So that workspace lifecycle can be controlled independently and safely.

**Acceptance Criteria:**

**Given** a valid mirror and merge ref
**When** I run `prr worktree add`
**Then** PRR creates a detached isolated worktree and emits `workDir`
**And** no writes are performed in the active working copy.

**Given** default cleanup behaviour and `--keep` override
**When** this command is used in the review chain
**Then** lifecycle behaviour remains consistent with documented cleanup semantics.

**Given** `--verbose` is enabled
**When** `prr worktree add` invokes external commands
**Then** PRR logs each external command to stderr before execution.

**Given** `--what-if` is enabled
**When** `prr worktree add` runs
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

### Story 1.8: Implement Diff and Bundle Composable Commands

**FRs:** FR14, FR15, FR16, FR17, FR18, FR19, FR20

As Richard,
I want `prr diff` and `prr bundle` commands,
So that deterministic review inputs can be generated and validated in composable stages.

**Acceptance Criteria:**

**Given** a valid worktree
**When** I run `prr diff`
**Then** PRR emits deterministic stat/files/patch outputs
**And** output contracts are JSON-compatible.

**Given** valid diff outputs
**When** I run `prr bundle`
**Then** PRR emits a validated v1 bundle payload
**And** enforces configured size limits with explicit failure diagnostics.

**Given** `--verbose` is enabled
**When** `prr diff` or `prr bundle` invokes external commands
**Then** PRR logs each external command to stderr before execution.

**Given** `--what-if` is enabled
**When** `prr diff` or `prr bundle` runs
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

### Story 1.9: Implement Review-Engine, Render, and Publish Composable Commands

**FRs:** FR21, FR22, FR23, FR24, FR25, FR26, FR27, FR28, FR29

As Richard,
I want `prr review-engine`, `prr render`, and `prr publish` commands,
So that review execution, output rendering, and optional publication are composable for scripting and diagnostics.

**Acceptance Criteria:**

**Given** a valid review bundle
**When** I run `prr review-engine`
**Then** PRR emits structured review JSON with stable per-run finding references
**And** engine failures return actionable, classed errors.

**Given** a valid review JSON payload
**When** I run `prr render`
**Then** PRR outputs Markdown by default and JSON when requested
**And** channel/exit behaviour remains automation-stable.

**Given** publish mode and provider support
**When** I run `prr publish`
**Then** PRR posts rendered review output to the PR
**And** reports publication outcome explicitly.

**Given** `--verbose` is enabled
**When** these commands invoke external commands
**Then** PRR logs each external command to stderr before execution.

**Given** `--what-if` is enabled
**When** these commands run
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

## Epic 2: Safe Repository Snapshot and Isolated Review Workspace

Richard can run reviews against an isolated merge snapshot with mirror/worktree lifecycle safety and zero interference with active local work.

### Story 2.1: Create and Maintain Per-Repository Bare Mirror Cache

**FRs:** FR5, FR6

As Richard,
I want PRR to maintain a reusable mirror cache per repository,
So that repeated reviews are faster and consistent.

**Acceptance Criteria:**

**Given** a repository with no existing cache
**When** a review run starts
**Then** PRR creates a deterministic bare mirror cache path for the repository
**And** subsequent runs reuse that mirror.

**Given** an existing mirror cache
**When** a new review run starts
**Then** PRR updates mirror state before processing
**And** update failures produce actionable diagnostics.

**Given** verbose mode is enabled
**When** mirror lifecycle commands invoke external commands
**Then** PRR logs each external command to stderr before execution.

**Given** what-if mode is enabled
**When** mirror lifecycle commands run
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

### Story 2.2: Add Lock-Safe Mirror Updates for Concurrent Runs

**FRs:** FR7

As Richard,
I want concurrent reviews against the same repository to be safe,
So that mirror state is not corrupted under parallel execution.

**Acceptance Criteria:**

**Given** multiple review invocations targeting the same repository
**When** they attempt mirror updates
**Then** per-repository locking serialises mirror mutation safely
**And** no run leaves the mirror in a corrupted state.

**Given** lock contention
**When** a run waits for lock acquisition
**Then** progress diagnostics indicate lock wait status
**And** timeout or failure is reported with a clear runtime error class.

**Given** force mode is enabled
**When** lock contention exists
**Then** PRR can bypass lock acquisition
**And** diagnostics indicate lock bypass explicitly.

### Story 2.3: Fetch Provider Merge Snapshot with Explicit Missing-Ref Handling

**FRs:** FR8, FR9

As Richard,
I want PRR to fetch the provider merge snapshot into an internal namespace,
So that diff generation uses deterministic merge semantics.

**Acceptance Criteria:**

**Given** a provider that exposes the merge snapshot ref
**When** PRR fetches the PR snapshot
**Then** the merge ref is stored under the internal review namespace
**And** the fetched ref is available to workspace creation.

**Given** a provider or PR where merge snapshot ref is unavailable
**When** fetch is attempted
**Then** PRR fails fast with an explicit actionable message
**And** returns the provider/ref error class exit code.

**Given** verbose mode is enabled
**When** merge-ref fetch invokes external commands
**Then** PRR logs each external command to stderr before execution.

**Given** what-if mode is enabled
**When** merge-ref fetch runs
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

### Story 2.4: Create Isolated Detached Worktree and Enforce Cleanup/Keep Behaviour

**FRs:** FR10, FR11, FR12, FR13

As Richard,
I want each review to run in an isolated detached worktree,
So that my active local working copy remains untouched.

**Acceptance Criteria:**

**Given** a fetched merge snapshot
**When** PRR starts review processing
**Then** it creates a detached isolated worktree for the run
**And** no write operations occur in the active working copy.

**Given** default cleanup mode
**When** review processing completes or fails
**Then** transient worktree artifacts are removed automatically
**And** cleanup status is reported in diagnostics.

**Given** `--keep` is set
**When** review processing completes
**Then** the run worktree is retained for investigation
**And** the retained path is printed for user access.

**Given** verbose mode is enabled
**When** worktree lifecycle invokes external commands
**Then** PRR logs each external command to stderr before execution.

**Given** what-if mode is enabled
**When** worktree lifecycle runs
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

## Epic 3: Deterministic Diff and Bundle Preparation with Safety Gates

Richard can reliably generate deterministic PR inputs (files, stat, patch, bundle) with configurable limits and explicit limit-failure behaviour.

### Story 3.1: Compute Deterministic PR Contribution Diff

**FRs:** FR14

As Richard,
I want PRR to compute the PR contribution diff from merge-parent semantics,
So that review inputs are reproducible for unchanged refs.

**Acceptance Criteria:**

**Given** an isolated worktree at the provider merge snapshot
**When** PRR computes contribution diff data
**Then** it uses the documented merge-parent range semantics
**And** generated diff results are functionally equivalent across reruns for unchanged refs.

**Given** diff generation failure
**When** the diff stage runs
**Then** PRR emits clear stage-level diagnostics
**And** returns a stable runtime error classification.

### Story 3.2: Produce Changed File List, Diff Stat, and Unified Patch Outputs

**FRs:** FR15, FR16, FR17

As Richard,
I want PRR to generate files/stat/patch outputs from the contribution diff,
So that all required review inputs are available for bundling.

**Acceptance Criteria:**

**Given** a successful contribution diff computation
**When** output extraction runs
**Then** PRR produces changed-file list, diff stat summary, and unified patch outputs
**And** output generation is deterministic for unchanged source refs.

**Given** very large diffs
**When** output generation runs
**Then** stage diagnostics include size metadata required for limit checks
**And** failures are reported without partial ambiguous state.

### Story 3.3: Build Review Bundle Schema Payload

**FRs:** FR18

As Richard,
I want PRR to package metadata and diff outputs into a single review bundle,
So that the review engine receives a complete, validated input contract.

**Acceptance Criteria:**

**Given** generated files/stat/patch outputs
**When** bundle assembly runs
**Then** PRR creates a bundle containing required metadata and payload fields
**And** the bundle schema is validated before submission.

**Given** missing or invalid bundle fields
**When** validation executes
**Then** PRR fails with actionable contract diagnostics
**And** uses a stable runtime or configuration error class as appropriate.

### Story 3.4: Enforce Configurable Input Size Limits with Clear Failure Modes

**FRs:** FR19, FR20

As Richard,
I want PRR to enforce patch-size and changed-file limits before engine invocation,
So that oversized reviews fail safely and predictably.

**Acceptance Criteria:**

**Given** configured safety limits for patch bytes and changed files
**When** bundle limit checks run
**Then** PRR blocks engine invocation when limits are exceeded
**And** reports explicit actionable limit diagnostics.

**Given** inputs within limits
**When** limit checks complete
**Then** PRR proceeds to engine submission
**And** logs the evaluated limit metrics in stage diagnostics.

## Epic 4: Review Engine Execution and Structured Result Handling

Richard can submit review inputs to a configured engine and receive structured outputs/failures suitable for decision-making.

### Story 4.1: Implement Review Engine Adapter and Bundle Submission

**FRs:** FR21, FR24

As Richard,
I want PRR to submit validated bundles to a configured review engine,
So that review findings are generated from deterministic inputs.

**Acceptance Criteria:**

**Given** a valid review bundle
**When** engine submission is triggered
**Then** PRR sends the bundle via the configured engine adapter contract
**And** handles transport/auth concerns through the configured integration boundary.

**Given** submission or transport failures
**When** engine interaction fails
**Then** PRR reports actionable engine failure diagnostics
**And** returns a stable engine error class exit code.

### Story 4.2: Normalise and Validate Structured Review Response

**FRs:** FR22, FR23

As Richard,
I want engine responses normalised into a consistent internal result structure,
So that rendering and publication are reliable across providers/engines.

**Acceptance Criteria:**

**Given** a successful engine response
**When** response normalisation runs
**Then** PRR produces structured review output with summary, risk, findings, and checklist
**And** each finding includes a per-run identifier usable within the result.

**Given** malformed or incomplete engine payloads
**When** normalisation/validation runs
**Then** PRR fails with actionable diagnostics
**And** does not emit ambiguous partial review output.

### Story 4.3: Emit Stage-Level Observability and Sanitised Error Diagnostics

**FRs:** FR29

As Richard,
I want stage-level diagnostics that are detailed but safe,
So that I can troubleshoot failures without secret leakage.

**Acceptance Criteria:**

**Given** a review run through all stages
**When** diagnostics are emitted
**Then** started/completed/failed events are recorded with stage context and run identifiers
**And** output remains consistent for automation parsing.

**Given** logs or user-facing errors containing sensitive fields
**When** diagnostics are rendered
**Then** tokens, secret headers, and secret URLs are redacted
**And** no secrets are persisted in logs or artifacts.

## Epic 5: Reporting, Publication, and Automation Diagnostics

Richard can consume Markdown/JSON outputs, optionally publish results, and integrate PRR into scripts with stable signalling and diagnostics.

### Story 5.1: Render Markdown Review Report as Default Human Output

**FRs:** FR25

As Richard,
I want a clear Markdown report by default,
So that I can quickly understand summary, risk, findings, and checklist.

**Acceptance Criteria:**

**Given** a valid structured review result
**When** default rendering runs
**Then** PRR outputs a Markdown report with summary, risk, findings, and checklist sections
**And** formatting is stable and readable for immediate action.

**Given** rendering errors
**When** report generation fails
**Then** PRR returns actionable diagnostics
**And** preserves stable error classification semantics.

### Story 5.2: Provide Stable JSON Output Mode for Automation

**FRs:** FR26

As Richard,
I want a structured JSON output mode,
So that PRR can be integrated into shell and CI workflows.

**Acceptance Criteria:**

**Given** JSON output mode is requested
**When** review output is emitted
**Then** PRR returns stable machine-readable JSON with documented fields
**And** avoids schema-breaking channel inconsistencies.

**Given** automation consumption in scripts
**When** PRR exits
**Then** stdout/stderr usage and exit signalling remain consistent
**And** non-zero codes map to the documented error classes.

### Story 5.3: Add Optional Publication of Review Output Back to PR

**FRs:** FR27

As Richard,
I want to optionally publish generated review output to the pull request,
So that review insights are shared directly in the collaboration workflow.

**Acceptance Criteria:**

**Given** publish mode is enabled and provider supports publication
**When** a review completes successfully
**Then** PRR posts the rendered result back to the PR
**And** reports publication outcome in diagnostics.

**Given** publication is disabled or unsupported
**When** a review completes
**Then** PRR skips publish without affecting core review completion
**And** provides explicit status indicating publication was not performed.

### Story 5.4: Finalise Stable Exit Codes and Automation Outcome Contracts

**FRs:** FR28, FR29

As Richard,
I want deterministic completion signals and error classes,
So that automation can reliably react to review outcomes.

**Acceptance Criteria:**

**Given** each major failure class (config, provider/ref, limit, engine, runtime)
**When** failures occur
**Then** PRR exits with the documented stable non-zero code for that class
**And** error payloads contain actionable machine-consumable context.

**Given** successful runs
**When** PRR exits
**Then** it returns success status with deterministic terminal state
**And** run diagnostics indicate completion and cleanup/keep disposition.
