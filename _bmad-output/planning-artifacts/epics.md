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
FR5: Epic 1 - Covered by internal orchestration stories
FR6: Epic 1 - Covered by internal orchestration stories
FR7: Epic 1 - Covered by internal orchestration stories
FR8: Epic 1 - Covered by internal orchestration stories
FR9: Epic 1 - Covered by internal orchestration stories
FR10: Epic 1 - Covered by internal orchestration stories
FR11: Epic 1 - Covered by internal orchestration stories
FR12: Epic 1 - Covered by internal orchestration stories
FR13: Epic 1 - Covered by internal orchestration stories
FR14: Epic 1 - Covered by internal orchestration stories
FR15: Epic 1 - Covered by internal orchestration stories
FR16: Epic 1 - Covered by internal orchestration stories
FR17: Epic 1 - Covered by internal orchestration stories
FR18: Epic 1 - Covered by internal orchestration stories
FR19: Epic 1 - Covered by internal orchestration stories
FR20: Epic 1 - Covered by internal orchestration stories
FR21: Epic 1 - Covered by internal orchestration stories
FR22: Epic 1 - Covered by internal orchestration stories
FR23: Epic 1 - Covered by internal orchestration stories
FR24: Epic 1 - Covered by internal orchestration stories
FR25: Epic 1 - Covered by internal orchestration stories
FR26: Epic 1 - Covered by internal orchestration stories
FR27: Epic 1 - Covered by internal orchestration stories
FR28: Epic 1 - Covered by internal orchestration stories
FR29: Epic 1 - Covered by internal orchestration stories
FR30: Epic 1 - Configure defaults with per-run override controls

## Epic List

### Epic 1: CLI Setup, Configuration, and Unified Review Orchestration
Richard can initialise PRR, run a review command with the correct PR context, and control defaults/overrides for predictable execution.
**FRs covered:** FR1, FR2, FR3, FR4, FR30

## Epic 1: CLI Setup, Configuration, and Unified Review Orchestration

Richard can initialise PRR, run the primary review flow with a single `review` command that orchestrates internal stages, and use `render` to convert review JSON into Markdown with predictable contracts and overrides.

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

### Story 1.5: Implement Internal Resolve Stage Contract

**FRs:** FR1, FR3, FR4

As Richard,
I want deterministic PR reference resolution as an internal review stage,
So that `prr review` can reliably prepare context without requiring a separate user-facing resolve command.

**Acceptance Criteria:**

**Given** a valid PR identifier and resolvable context
**When** I run `prr review <PR_ID>`
**Then** PRR resolves and validates a stable internal `PRRef` payload
**And** supports equivalent override flags and provider auto-detection for supported URL formats.

**Given** missing or invalid context inputs
**When** resolution fails
**Then** PRR returns actionable diagnostics
**And** exits with a stable error-class code.

### Story 1.6: Implement Internal Mirror Ensure and PRRef Fetch Stages

**FRs:** FR5, FR6, FR8, FR9

As Richard,
I want mirror lifecycle and merge-ref acquisition implemented as internal stages,
So that `prr review` can execute these steps deterministically without separate user-facing commands.

**Acceptance Criteria:**

**Given** repo context input
**When** I run `prr review <PR_ID>`
**Then** PRR creates or updates the deterministic mirror location
**And** carries `bareDir` through internal stage context.

**Given** valid PR context and mirror state
**When** the fetch stage executes inside `prr review`
**Then** PRR fetches merge ref into `refs/prr/pull/<PR_ID>/merge`
**And** carries resolved `mergeRef` through internal stage context.

**Given** `--verbose` is enabled
**When** these commands invoke external commands
**Then** PRR logs each external command to stderr before execution
**And** preserves JSON payloads on stdout.

**Given** `--what-if` is enabled
**When** these commands run
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

### Story 1.7: Implement Internal Worktree Stage with Cleanup/Keep Compatibility

**FRs:** FR10, FR11, FR12, FR13

As Richard,
I want `prr review` to create isolated review worktrees via an internal stage,
So that workspace lifecycle remains safe while keeping the command surface minimal.

**Acceptance Criteria:**

**Given** a valid mirror and merge ref
**When** I run `prr review <PR_ID>`
**Then** PRR creates a detached isolated worktree and records `workDir` in stage context
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

### Story 1.8: Implement Internal Diff and Bundle Stages

**FRs:** FR14, FR15, FR16, FR17, FR18, FR19, FR20

As Richard,
I want deterministic diff and bundle preparation stages to run within `prr review`,
So that review inputs are generated and validated end-to-end from one command.

**Acceptance Criteria:**

**Given** a valid worktree
**When** `prr review` executes diff processing
**Then** PRR produces deterministic stat/files/patch outputs
**And** output contracts are JSON-compatible.

**Given** valid diff outputs
**When** `prr review` executes bundle preparation
**Then** PRR produces a validated v1 bundle payload
**And** enforces configured size limits with explicit failure diagnostics.

**Given** `--verbose` is enabled
**When** `prr diff` or `prr bundle` invokes external commands
**Then** PRR logs each external command to stderr before execution.

**Given** `--what-if` is enabled
**When** `prr diff` or `prr bundle` runs
**Then** PRR prints external commands it would execute
**And** performs no external command execution.

### Story 1.9: Rework Review Command to Invoke Agent CLI and Emit Renderer-Compatible JSON

**FRs:** FR21, FR22, FR23, FR24, FR25, FR26, FR27, FR28, FR29

As Richard,
I want `prr review <PR_URL>` to pass deterministic diff JSON plus deterministic review instructions to an agent CLI,
So that PRR emits structured review JSON that `prr render` can consume directly.

**Acceptance Criteria:**

**Given** a valid PR URL input (or equivalent checkout JSON piped from stdin)
**When** I run `prr review`
**Then** PRR prepares deterministic review input internally and invokes the GitHub Copilot agent CLI (`copilot`) in non-interactive mode with deterministic prompt + input payload
**And** deterministic instructions plus payload framing are provided via stdin envelope
**And** PRR emits structured review JSON with stable per-run finding references.

**Given** checkout JSON is piped from `prr checkout <PR_URL>`
**When** I run `prr review` without positional args
**Then** PRR reads PR context from stdin and skips checkout-stage setup (`resolve`, mirror ensure/fetch, worktree creation)
**And** proceeds directly with review stages from supplied context.

**Given** malformed, partial, or non-JSON agent output
**When** `prr review` parses the response
**Then** PRR fails with actionable diagnostics and stable classed non-zero exit codes
**And** does not emit ambiguous partial review JSON.

**Given** a successful agent CLI response
**When** PRR normalises the response
**Then** output JSON matches renderer requirements (`summary`, `risk`, `findings`, `checklist`)
**And** channel/exit behaviour remains automation-stable.

**Given** agent invocation fails (CLI missing, auth/config missing, timeout, non-zero exit)
**When** `prr review` runs
**Then** PRR returns classed actionable errors with sanitised diagnostics.

**Given** a checked-out PR worktree is available
**When** `prr review` invokes the agent CLI
**Then** the CLI process runs with working directory set to that worktree
**And** relative repository paths resolve against the reviewed PR state.

**Given** `--verbose` is enabled
**When** `prr review` invokes external commands
**Then** PRR logs each external command to stderr before execution.

**Given** `--what-if` is enabled
**When** `prr review` runs
**Then** PRR prints the command and input envelope details it would use
**And** performs no external command execution.

**Given** review safety options (`--max-patch-bytes`, `--max-files`) and workspace retention (`--keep`)
**When** `prr review <PR_URL>` runs
**Then** option handling remains compatible with the README command contract.

**Given** `--model <model_name>` is provided to `prr review`
**When** PRR invokes Copilot
**Then** model selection is passed through as Copilot `--model <model_name>`.

**Given** the agent review stage is executed
**When** PRR invokes the configured CLI command
**Then** invocation parameters are explicit and deterministic (binary, ordered args, optional `--model`, `cwd = workDir`, stdin/input mode, timeout)
**And** the selected `copilot` command/flags variant is pinned and covered by regression tests
**And** GitHub CLI (`gh`) is not used as the agent execution command in this story.

### Story 1.10: Replace Unix-Only Mirror Locking with Cross-Platform Lock Strategy

**FRs:** FR7

As Richard,
I want mirror-update locking to work on macOS, Linux, and Windows,
So that concurrent review safety is preserved regardless of host OS.

**Acceptance Criteria:**

**Given** PRR runs on Windows
**When** mirror locking is compiled and executed
**Then** PRR uses a supported lock implementation (no Unix-only `syscall.Flock` dependency)
**And** lock acquire/release semantics remain equivalent to macOS/Linux behaviour.

**Given** lock contention on any supported OS
**When** a run waits for lock acquisition
**Then** timeout and `--force` bypass behaviour remain consistent
**And** lock timeout failures return a stable runtime error class.

**Given** cross-platform unit tests
**When** lock tests run on macOS, Linux, and Windows
**Then** tests verify lock contention, timeout, and force-bypass behaviour
**And** do not rely on OS-specific syscall APIs in shared test files.

### Story 1.11: Normalise Cross-Platform Path and Test Contracts

**FRs:** FR2, FR28

As Richard,
I want path handling and command contracts to be OS-agnostic,
So that PRR behaves consistently with Windows path separators and shell environments.

**Acceptance Criteria:**

**Given** cache, mirror, and worktree paths
**When** PRR resolves and emits filesystem locations
**Then** paths are created using platform-safe path APIs
**And** JSON payload fields remain deterministic and valid on macOS, Linux, and Windows.

**Given** unit tests for mirror/worktree path behaviour
**When** tests run on Windows
**Then** assertions avoid hard-coded `/tmp` and `/` separator assumptions
**And** use `filepath`-safe expectations to validate deterministic path segments.

**Given** user-facing examples and diagnostics
**When** commands/logs are reviewed across platforms
**Then** examples avoid Unix-only filesystem assumptions
**And** command diagnostics remain script-compatible for automation.

### Story 1.12: Add Cross-OS Build and Smoke Verification for CLI Baseline

**FRs:** FR1, FR28

As Richard,
I want automated build/smoke verification across macOS, Linux, and Windows,
So that CLI entry and composable command contracts are validated before release.

**Acceptance Criteria:**

**Given** CI build verification
**When** pull requests are validated
**Then** `go build ./...` and `go test ./...` run in a matrix for macOS, Linux, and Windows
**And** failures are reported per OS target.

**Given** the produced CLI binary per OS
**When** smoke checks run
**Then** `prr --help`, `prr version`, and one `--what-if` composable command path succeed
**And** stdout/stderr contracts remain stable for scripts.

**Given** local developer workflows
**When** contributors run documented build/test commands
**Then** instructions use cross-platform Go commands as the source of truth
**And** Unix-only helper commands are marked optional.

