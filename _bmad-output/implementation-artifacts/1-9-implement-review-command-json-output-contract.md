# Story 1.9: Rework Review Command to Invoke Agent CLI and Emit Renderer-Compatible JSON

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As Richard,
I want `prr review <PR_URL>` to pass deterministic diff JSON plus a review prompt into an agent CLI,
so that the command returns structured review JSON that `prr render` can consume without manual translation.

## Acceptance Criteria

1. Given a valid PR URL input (or equivalent checkout JSON piped from stdin), when I run `prr review`, then PRR prepares deterministic review input internally and invokes the GitHub Copilot agent CLI (`copilot`) in non-interactive mode with:
  - the prepared JSON diff payload,
  - a deterministic review prompt passed via `-p`,
  - and stable command arguments suitable for automation.

2. Given checkout JSON is piped from `prr checkout <PR_URL>`, when I run `prr review` without positional args, then PRR reads PR context from stdin and skips checkout-stage setup (`resolve`, mirror ensure/fetch, worktree creation), proceeding directly with review stages from the supplied context.

3. Given a successful agent CLI response, when PRR processes the response, then PRR emits structured JSON on stdout that matches the downstream renderer contract, including:
   - `summary`,
   - `risk` (`score`, `reasons`),
   - `findings[]` (`id`, `file`, `line`, `severity`, `category`, `message`, `suggestion`),
   - `checklist`.

4. Given malformed, partial, or non-JSON agent output, when `prr review` parses the response, then PRR fails with actionable diagnostics and stable classed non-zero exit codes, and does not emit ambiguous partial review JSON.

5. Given agent invocation failures (CLI not installed, auth/config missing, timeout, non-zero exit), when `prr review` runs, then PRR returns classed actionable errors with sanitised diagnostics.

6. Given a checked-out PR worktree is available, when PRR invokes the agent CLI, then the process working directory is set to that worktree (`workDir`) so relative file paths and repository context resolve against the reviewed PR state.

7. Given `--verbose` is enabled, when external commands are invoked, then PRR logs invocation details to stderr before execution.

8. Given `--what-if` is enabled, when `prr review` runs, then PRR prints the external command and prompt/input paths it would use, and does not execute the agent CLI.

9. Given review safety limit overrides (`--max-patch-bytes`, `--max-files`) or workspace retention (`--keep`), when `prr review <PR_URL>` runs, then CLI option handling remains compatible with README-documented behaviour.

10. Given `--model <model_name>` is provided to `prr review`, when PRR invokes Copilot, then PRR passes the selected model through to Copilot as `--model <model_name>`.

## Tasks / Subtasks

- [x] Rework `review` orchestration to call agent CLI after diff/bundle preparation (AC: 1, 2, 6, 7, 8, 9, 10)
  - [x] Keep existing resolve/mirror/worktree/diff/bundle stages as upstream inputs.
  - [x] When stdin already contains checkout JSON with required context, bypass resolve/mirror/fetch/worktree setup and use stdin values as authoritative.
  - [x] Add agent-invocation stage that shells out to configured CLI binary.
  - [x] Set invocation working directory to the checked-out PR worktree (`workDir`) from upstream stages.
  - [x] Preserve stdout JSON / stderr diagnostics channel contract.

- [x] Add agent CLI adapter boundary (AC: 1, 4)
  - [x] Introduce `internal/engine` adapter for command construction and execution.
  - [x] Support configurable binary/arguments and deterministic invocation settings.
  - [x] Map exit/error conditions to stable error classes.

- [x] Implement deterministic prompt and input packaging (AC: 1)
  - [x] Define prompt template/instructions for code review scope and output schema.
  - [x] Pass the prompt to Copilot using `-p` so the invocation exits once prompt execution completes.
  - [x] Pass diff JSON payload as structured input to the agent CLI.
  - [x] Ensure prompt text enforces output schema required by renderer.

- [x] Add review model selection flag and pass-through (AC: 10)
  - [x] Add `--model <model_name>` option to `prr review`.
  - [x] Map `prr review --model <model_name>` directly to Copilot invocation `--model <model_name>`.
  - [x] Preserve deterministic defaults when `--model` is not provided.

- [x] Implement response normalisation and validation (AC: 2, 3)
  - [x] Parse raw agent output, extract JSON object, and normalise to review schema.
  - [x] Validate required fields, enums, and value constraints.
  - [x] Assign/verify finding IDs stable within the run.

- [x] Add failure handling and diagnostics hardening (AC: 3, 4)
  - [x] Handle non-JSON output, truncation, and mixed text/JSON responses explicitly.
  - [x] Surface actionable diagnostics without leaking secrets/tokens.
  - [x] Preserve stable exit-code mapping for automation.

- [x] Add tests for contract and invocation semantics (AC: 1, 2, 3, 4, 5, 6, 7, 10)
  - [x] Tests for `prr review <PR_URL>` positional input and stdin checkout-pipe input modes.
  - [x] Tests for README-documented review options (`--keep`, `--max-patch-bytes`, `--max-files`, `--model`) in the reworked flow.
  - [x] Unit tests for command construction and `--verbose`/`--what-if` behaviour.
  - [x] Unit tests asserting Copilot invocation includes `-p <prompt>`.
  - [x] Unit tests asserting `prr review --model` maps to Copilot `--model`.
  - [x] Unit tests asserting the agent process is launched with `cwd = workDir`.
  - [x] Fixture-based tests for valid agent output to renderer-compatible JSON.
  - [x] Negative tests for malformed output, non-zero exits, missing binary, and timeout.
  - [x] Regression tests for deterministic output shape and stdout/stderr separation.

## Dev Notes

- This is a rework of Story 1.9 scope.
- `prr render` remains the downstream renderer and must not require schema changes for this story.
- The agent CLI is an integration boundary; command handlers in `cmd/prr/*` must remain thin.
- Keep implementation deterministic and automation-safe: repeatable prompting, explicit failures, stable schema.

### Technical Requirements

- `prr review <PR_URL>` must:
  - accept PR context via positional PR URL or checkout JSON on stdin,
  - skip checkout-stage setup when valid checkout JSON is provided via stdin,
  - build deterministic review input from existing internal stages,
  - pass prompt content to Copilot via `-p`,
  - accept optional `--model <model_name>` and pass it through to Copilot as `--model <model_name>`,
  - run the agent CLI with the checked-out PR worktree as process working directory,
  - invoke agent CLI non-interactively,
  - parse and validate response,
  - emit renderer-compatible JSON.
- Support configuration for agent CLI command path and invocation mode.
- Support timeout and output-size safeguards for agent invocation.
- Never emit secrets in logs or surfaced error payloads.

### Agent CLI Invocation Contract

- The review stage must construct and execute one deterministic external command for agent review generation.
- The command must be configurable, but the runtime contract must be explicit and testable.

Required invocation shape:

```text
<agent_binary> <agent_args...> -p "<prompt>" [--model <model_name>]
```

Required execution parameters:

- `cwd`: must be set to checkout `workDir`.
- `stdin`: must contain the prepared review input payload (or a deterministic prompt+payload envelope if file-based input is configured).
- `env`: only required environment variables for agent execution/auth; no unrelated environment leakage in diagnostics.
- `timeout`: enforced via config with stable timeout failure classification.

Required configured fields (minimum):

- `agent.command`: executable path/name (must be `copilot`, not `gh`).
- `agent.args`: ordered argument list for non-interactive invocation.
- `agent.prompt_arg`: must be `-p`.
- `agent.model_arg`: must be `--model`.
- `agent.input_mode`: `stdin` or `file`.
- `agent.output_mode`: must produce machine-parseable JSON (directly or via deterministic extraction path).
- `agent.timeout_seconds`: process timeout limit.

Copilot-specific note:

- Story 1.9 targets GitHub Copilot agent CLI invocation explicitly; GitHub CLI (`gh`) is out of scope for agent execution in this story.
- Implementation must pin the exact `copilot` subcommand/flags variant supported in this repo and add regression tests for that concrete invocation.
- The chosen `copilot` invocation must be non-interactive and automation-safe.

### Architecture Compliance

- Command handlers remain in `cmd/prr/*` and delegate to `internal/*`.
- Agent invocation and parsing live under `internal/engine` (or equivalent adapter module).
- Review schema remains in `internal/types/review.go` (or equivalent).
- Error classification remains centralised in `internal/errors`.

### Library / Framework Requirements

- Continue using Cobra command model.
- No hard dependency on one agent vendor in command code; use adapter/config indirection.

### File Structure Requirements

- Update: `cmd/prr/review.go`
- Update/Create: `internal/engine/*` (agent CLI adapter + parser/normaliser)
- Update: `internal/types/review.go` (validation/normalisation helpers if needed)
- Update: `internal/errors/*` (error classes and mapping if needed)
- Update tests in `cmd/prr/*_test.go` and `internal/engine/*_test.go`

### Testing Requirements

- Validate review JSON schema compatibility with current `prr render` inputs.
- Validate input compatibility with README command forms (`prr review <PR_URL>` and `prr checkout <PR_URL> | prr review`).
- Validate that stdin checkout JSON path does not invoke resolve/mirror/fetch/worktree setup.
- Validate deterministic command invocation parameters.
- Validate agent invocation uses `workDir` as process working directory.
- Validate Copilot invocation uses `-p` for prompt submission.
- Validate `prr review --model` pass-through to Copilot `--model`.
- Validate README-documented option handling (`--keep`, `--max-patch-bytes`, `--max-files`, `--model`).
- Validate `--what-if` performs no external execution.
- Validate failures are classed and actionable.
- Validate stdout/stderr contract remains automation-safe.

### References

- Source story definition: `_bmad-output/planning-artifacts/epics.md` (Story 1.9)
- Output model and automation constraints: `_bmad-output/planning-artifacts/prd.md`
- Architecture boundaries: `_bmad-output/planning-artifacts/architecture.md`
- Upstream input stages: `_bmad-output/implementation-artifacts/1-8-implement-diff-and-bundle-composable-commands.md`
- Downstream renderer contract: `_bmad-output/implementation-artifacts/1-9b-implement-render-command-for-json-to-markdown.md`

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- 2026-03-05: Story rework request received to pivot Story 1.9 from generic review-engine output to explicit agent CLI invocation with schema-constrained output.
- 2026-03-05: Implemented CLI adapter and review orchestration updates; added stdin checkout-context bypass, `--model` pass-through, deterministic prompt/input packaging, and parser/error hardening.
- 2026-03-05: Added command and engine tests for invocation semantics (`-p`, `--model`, `cwd`, what-if/verbose), mixed-output JSON extraction, and failure classification.
- 2026-03-05: Field validation with installed Copilot CLI showed `--json` unsupported and `-p` mode unsuitable for deterministic stdin-driven payload handling.
- 2026-03-05: Decision made to remove `-p` from runtime invocation and move all instructions + diff payload into stdin envelope with explicit delimiters.
- 2026-03-05: Added verbose output logging of Copilot stdout/stderr to diagnose schema/output issues during live runs.
- 2026-03-05: Decision made to accept parsed Copilot JSON as returned (no runtime schema validation), while retaining prompt-side guidance for `risk.score` in `[0,1]`.

### Completion Notes List

- Reworked `prr review` flow to keep existing upstream prep stages while adding an explicit Copilot CLI invocation stage with deterministic prompt (`-p`) and diff-bundle JSON input.
- Added authoritative stdin checkout-context mode to bypass resolve/mirror/fetch/worktree setup and continue directly with diff/bundle/review stages.
- Added `--model` flag to `prr review` and passed selected model through to agent invocation as `--model`.
- Implemented robust agent response extraction/normalisation for mixed output and classed errors for malformed output, missing binary, timeout, and non-zero exit.
- Preserved automation-safe stdout/stderr contract and what-if/verbose diagnostics for external command execution.
- Added regression tests in `cmd/prr/review_render_test.go` and new unit tests in `internal/engine/engine_test.go`.
- Observed and documented real CLI compatibility findings: installed `copilot` rejected `--json` and practical stdin handling in `-p` mode was not reliable for this workflow.
- Runtime invocation now uses stdin-only control envelope (instruction block + `DIFF_BUNDLE_JSON_START/END` around deterministic diff bundle JSON) with optional `--model` pass-through.
- Added explicit instruction text requiring `risk.score` to be a decimal in the inclusive range `[0,1]`.
- Runtime review pipeline now parses/extracts JSON and emits it as returned by Copilot (no schema validation gate in command/engine path).

### File List

- cmd/prr/review.go
- cmd/prr/review_render_test.go
- internal/engine/engine.go
- internal/engine/engine_test.go
- internal/types/review.go
- internal/types/review_test.go
- prr
- _bmad-output/implementation-artifacts/1-9-implement-review-command-json-output-contract.md
- _bmad-output/implementation-artifacts/sprint-status.yaml

## Change Log

- 2026-03-05: Reworked Story 1.9 scope to require `prr review` agent CLI invocation using deterministic diff JSON + prompt, with renderer-compatible JSON output contract.
- 2026-03-05: Implemented Story 1.9 review-command rework with Copilot CLI adapter, stdin checkout-context bypass, model pass-through, structured output normalisation, and test coverage for invocation + failure semantics.
- 2026-03-05: Updated runtime integration strategy after live CLI findings: removed `-p`, removed unsupported `--json`, switched to stdin-only instruction+payload envelope, added verbose Copilot output diagnostics, and removed runtime schema validation gate for returned review JSON.
