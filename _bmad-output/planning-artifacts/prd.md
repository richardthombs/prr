---
stepsCompleted:
  - step-01-init
  - step-02-discovery
  - step-02b-vision
  - step-02c-executive-summary
  - step-03-success
  - step-04-journeys
  - step-05-domain
  - step-06-innovation
  - step-07-project-type
  - step-08-scoping
  - step-09-functional
  - step-10-nonfunctional
  - step-11-polish
  - step-12-complete
inputDocuments:
  - /Users/richardthombs/dev/prr/docs/initial_specification.md
documentCounts:
  briefCount: 0
  researchCount: 0
  brainstormingCount: 0
  projectDocsCount: 1
classification:
  projectType: cli_tool
  domain: general
  complexity: low
  projectContext: greenfield
date: '2026-03-04'
workflowType: 'prd'
---

# Product Requirements Document - prr

**Author:** Richard
**Date:** 2026-03-04

## Executive Summary

PRR is a greenfield CLI tool for Richard that automates on-demand pull request review from a single command. The tool fetches a PR merge snapshot into an isolated Git worktree, computes a deterministic merge-based diff, packages review inputs as structured JSON, and sends them to a review engine for findings generation and Markdown output. The primary problem solved is reducing review effort and inconsistency while preserving trust in outputs through reproducible diff construction and strict non-interference with any existing local working copy.

### What Makes This Special

This project is intentionally personal and non-commercial: success is measured by practical utility in Richard’s own workflow, not market differentiation. Its core value is reliable time savings from repeatable, low-friction PR reviews with minimal manual setup. The core technical insight is that review quality and confidence improve when input generation is deterministic, isolated, and composable.

## Project Classification

- Project type: CLI tool (with developer-tool characteristics).
- Domain: General software tooling.
- Complexity: Low domain complexity (with moderate implementation complexity).
- Project context: Greenfield.

## Success Criteria

### User Success

- Richard can run one command to review a PR end-to-end without touching an existing working copy.
- Review output is clear enough to act on immediately (summary, risk, findings, checklist).
- Typical review flow feels faster and less mentally taxing than manual setup.

### Business Success

- Personal productivity improves: PR review setup time is reduced by at least 70% versus current manual process.
- Tool is used as the default path for most eligible PR reviews within 4 weeks of first usable version.
- Maintenance burden remains low enough to keep the tool actively used (no frequent manual babysitting).

### Technical Success

- Deterministic diff generation using merge-based snapshot and isolated worktree.
- Zero interference with local repo state across runs.
- Clear failure modes for missing merge refs, limit exceedance, and provider errors.

### Measurable Outcomes

- Median time from command start to rendered review ≤ 90 seconds for typical PRs.
- 100% of runs leave the user’s existing working tree untouched.
- ≥ 95% successful runs on supported PR/provider scenarios in v1 test set.
- 100% of oversized patch/file-limit cases fail with explicit, actionable errors.

## Product Scope

### MVP - Minimum Viable Product

- `prr review PR_ID` flow with provider resolution, mirror/worktree management, deterministic diff, v1 bundle, review engine invocation, Markdown rendering, optional publish, cleanup/`--keep`.
- Configurable patch/file safety limits.
- Support providers where merge ref is available (with clear error otherwise).

### Growth Features (Post-MVP)

- Incremental review since last run.
- Smarter diff chunking/grouping for large PRs.
- Better inline comment line mapping fidelity.
- Broader provider behaviours and richer publish integrations.

### Vision (Future)

- Highly trusted personal review assistant embedded in Richard’s daily development loop.
- Optional policy packs/templates for different repo types.
- Potential local+remote hybrid review strategies for speed/cost control.

## User Journeys

### 1) Primary User (Richard) — Success Path

Richard gets a message asking for a PR review and runs `prr review PR12345`. PRR resolves repo/provider context, fetches the PR merge ref into a cached bare mirror, creates an isolated detached worktree, computes `HEAD^1..HEAD`, builds the v1 bundle, sends it to the review engine, and renders Markdown findings. Richard reads a concise summary, risk view, and actionable findings, then decides whether to publish back to the PR. He finishes with confidence that nothing in his active working copy was touched.

### 2) Primary User (Richard) — Edge Case / Recovery

Richard runs the same command, but the provider has no usable merge ref (or limits are exceeded). PRR fails fast with a clear diagnostic describing the exact constraint or missing ref and next action (e.g., provider limitation, reduce patch size, adjust limits). Richard does not need to inspect Git internals; he gets an explicit, actionable error and can retry safely. Trust is preserved because failure is explicit and non-destructive.

### 3) Operations User (Richard-as-Admin) — Configuration & Maintenance

Richard configures defaults for provider, output behaviour, and safety limits, then periodically verifies cache/worktree hygiene. He can preserve a worktree with `--keep` for investigation and remove stale artifacts safely afterward. He expects predictable storage locations, lock-safe concurrency on mirrors, and transparent operational behaviour so the tool remains low-maintenance.

### 4) Support/Troubleshooting User (Richard-debug mode)

A review result looks unexpected, so Richard inspects logs/metadata and (if needed) kept worktree state to trace where inputs came from. He verifies diff stat, changed files, and patch consistency against the merge snapshot, then reruns. The tool supports diagnosis by making each stage observable (fetch, worktree, diff, bundle, review engine), reducing ambiguity in failure analysis.

### 5) Integration User (Automation / Script Consumer)

Richard scripts PRR in a shell pipeline using JSON-friendly inputs/outputs. A wrapper command triggers reviews for selected PRs and aggregates results without bespoke parsing hacks. Predictable structure allows simple automation, while non-zero exits and explicit error payloads make CI/local automation robust.

### Journey Requirements Summary

- Isolated execution model (bare mirror + detached worktree) with strict non-interference guarantees.
- Deterministic diff construction from merge snapshot (`HEAD^1..HEAD`) and reproducible bundle generation.
- Clear operational and error surfaces for missing refs, limit exceedance, and provider failures.
- Configurable limits and behaviour suitable for both interactive and scripted usage.
- Observability/debuggability across each pipeline stage.
- Safe cleanup defaults with optional retained state for investigation.

## CLI Tool Specific Requirements

### Project-Type Overview

PRR is a script-friendly CLI for deterministic review-input generation and on-demand PR review execution. The command surface prioritises single-command interactive use while remaining composable for automation workflows. The tool behaves predictably across repeated runs for input preparation and provides explicit, machine-parsable failure behaviour for integration scenarios.

### Technical Architecture Considerations

The CLI should be structured around deterministic input-generation stages: PR resolution, mirror/worktree preparation, diff generation, and review bundle construction, followed by review engine invocation, rendering, optional publication, and cleanup. Review-engine output content is expected to vary between invocations. Each stage should expose clear error boundaries and stable contracts to support troubleshooting and scripting. Command execution must preserve local repository isolation and enforce configurable safety limits before review engine invocation.

### Command Structure

- Primary command: `prr review <PR_ID>`.
- MVP composable commands (must be implemented):
  - `prr resolve <PR_ID>`
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

### Output Formats

- Default human-readable output: Markdown review report.
- Structured output mode for automation (JSON) for summary/risk/findings/checklist payloads.
- Error output should be explicit and actionable, with machine-consumable fields in structured mode.
- Keep stdout/stderr behaviour consistent to avoid breaking shell pipelines.

### Configuration Schema

- Configurable defaults for provider, remote naming assumptions, output mode, publish behaviour, and safety limits.
- Safety limits must include at least max patch bytes and max changed files.
- Configuration sources should support a clear precedence model (e.g., CLI flags override config file defaults).
- Validation should fail fast on invalid config with precise diagnostics.

### Scripting Support

- JSON in/out compatibility for chaining with shell tools and CI scripts.
- Deterministic review-input generation across identical source refs.
- Non-zero exit codes and explicit failure metadata for robust automation.
- Idempotent operational side effects where feasible (safe retries after transient failures).

### Implementation Considerations

- Concurrency-safe mirror updates via per-repo locking.
- Explicit lifecycle management for cached mirrors and transient worktrees.
- Provider abstraction should isolate transport/ref differences from core review flow.
- Keep v1 intentionally narrow: one unified diff bundle with no incremental review logic.

## Project Scoping & Phased Development

### MVP Strategy & Philosophy

**MVP Approach:** Problem-solving MVP focused on a reliable primary review workflow plus explicit composable commands for deterministic stage-by-stage execution and automation.
**Resource Requirements:** Solo builder (Richard), with skills in CLI engineering, Git plumbing, and basic prompt/review-engine integration.

### MVP Feature Set (Phase 1)

**Core User Journeys Supported:**

- Primary success path (`prr review <PR_ID>` end-to-end).
- Primary edge-case recovery (clear failure for missing merge refs/limit breaches).
- Operations/maintenance path (config + cache/worktree lifecycle).
- Automation path (JSON-friendly scripting support).

**Must-Have Capabilities:**

- PR context resolution and provider abstraction for merge-ref fetch.
- Bare mirror cache + lock-safe update + isolated detached worktree.
- Deterministic diff generation (`HEAD^1..HEAD`) with stat/files/patch outputs.
- v1 review bundle schema and review engine invocation.
- Markdown default rendering + optional publish integration.
- Configurable safety limits (patch bytes, changed files) and clear errors.
- Cleanup by default, `--keep` override, stable exit-code semantics.
- Explicit MVP command surface includes `review`, `resolve`, `mirror ensure`, `prref fetch`, `worktree add`, `diff`, `bundle`, `review-engine`, `render`, and `publish`.

### Post-MVP Features

**Phase 2 (Post-MVP):**

- Incremental review since last run.
- Improved diff chunking/grouping for large PRs.
- Stronger inline comment mapping precision.
- Better diagnostics and richer provider-specific ergonomics.

**Phase 3 (Expansion):**

- Policy/profile presets per repo type.
- Advanced multi-engine review orchestration.
- Local+remote hybrid review modes for speed/cost tuning.
- Broader automation surfaces and optional team-sharing workflows.

### Risk Mitigation Strategy

- **Technical Risks:** Git/provider edge cases and ref availability; mitigate with strict provider contracts, explicit error taxonomy, and focused integration tests on merge-ref scenarios.
- **Market Risks:** Minimal (personal project); core risk is self-adoption, mitigated by prioritising daily usability and fast feedback loops.
- **Resource Risks:** Solo-time constraints; mitigate by aggressively narrowing MVP scope and deferring non-essential polish to Phase 2+.

## Functional Requirements

### PR Identification & Context Resolution

- FR1: Richard can start a review by providing a pull request identifier.
- FR2: Richard can run a review without switching away from his current working copy.
- FR3: Richard can have PRR resolve repository and remote context for the requested PR.
- FR4: Richard can override inferred repository/provider context when defaults are incorrect.

### Repository Snapshot Management

- FR5: PRR can maintain a cached mirror per repository for repeat reviews.
- FR6: PRR can update cached repository state before review processing.
- FR7: PRR can prevent concurrent corruption when up to 5 concurrent reviews target the same repository.
- FR8: PRR can fetch a PR merge snapshot into an internal review namespace.
- FR9: PRR can fail with an explicit message when required merge snapshot refs are unavailable.

### Isolated Review Workspace

- FR10: PRR can create an isolated workspace for each review run.
- FR11: PRR can ensure review execution does not modify Richard’s active local working copy.
- FR12: Richard can keep an isolated workspace for investigation when requested.
- FR13: PRR can remove transient review workspaces automatically when retention is not requested.

### Diff & Review Bundle Generation

- FR14: PRR can compute the PR contribution diff using merge-parent comparison semantics.
- FR15: PRR can produce a changed-file list for the PR contribution.
- FR16: PRR can produce a diff stat summary for the PR contribution.
- FR17: PRR can produce a unified patch for the PR contribution.
- FR18: PRR can build a review bundle containing required metadata, stat, files, and patch fields.
- FR19: PRR can enforce configurable review-input size limits before engine invocation.
- FR20: PRR can fail with clear diagnostics when review-input limits are exceeded.

### Review Engine Orchestration

- FR21: PRR can submit the generated review bundle to a configured review engine.
- FR22: PRR can receive structured review output containing summary, risk, findings, and checklist.
- FR23: PRR can include finding identifiers for references within a single review result, with no requirement to correlate findings across reruns.
- FR24: PRR can surface review-engine failures with actionable error context.

### Output, Publication & Automation

- FR25: Richard can receive a Markdown review report as the default output.
- FR26: Richard can request structured JSON output for automation workflows.
- FR27: Richard can optionally publish review results back to the pull request.
- FR28: PRR can return stable outcome signalling suitable for shell/CI scripting.
- FR29: PRR can expose stage-level diagnostics to support troubleshooting.
- FR30: Richard can configure default behaviours and override them per run.

## Non-Functional Requirements

### Performance

- NFR1: For typical PRs within configured size limits, PRR should produce a rendered review within 90 seconds on Richard’s normal development machine and network.
- NFR2: PRR should provide visible stage progress or clear terminal feedback at each major pipeline stage to avoid perceived hangs.
- NFR3: PRR should fail fast (within 5 seconds of detection) when mandatory preconditions are missing (e.g., merge ref unavailable, invalid config).

### Security

- NFR4: PRR must not persist secrets in logs, review artifacts, or temporary files.
- NFR5: PRR must use least-privilege credentials for provider/review-engine operations and rely on externally managed auth mechanisms.
- NFR6: PRR must isolate review workspaces so no writes occur in Richard’s active repository working copy.
- NFR7: PRR must sanitise error output to avoid leaking tokens, secret URLs, or sensitive headers.

### Reliability

- NFR8: PRR must complete or fail with a deterministic terminal state; partial runs must not leave ambiguous review outcomes.
- NFR9: PRR must clean transient worktrees by default and support explicit retention only via `--keep`.
- NFR10: PRR must prevent concurrent corruption of shared mirror state via per-repository locking.
- NFR11: Re-running the same review command against unchanged source refs should produce functionally equivalent bundle content.

### Integration

- NFR12: PRR must support stable, machine-readable JSON output for automation use cases.
- NFR13: PRR must return stable non-zero exit codes by error class (configuration, provider/ref, limit, engine/runtime).
- NFR14: PRR must keep stdout/stderr behaviour consistent across versions for script compatibility.

### Maintainability & Operability

- NFR15: PRR must emit stage-level diagnostics sufficient to troubleshoot failures without manual Git forensics in most cases.
- NFR16: Configuration validation errors must identify offending fields and expected value format.
- NFR17: Internal module boundaries (provider, git workspace, bundle, engine, renderer) must remain separable to enable incremental changes without full rewrites.
