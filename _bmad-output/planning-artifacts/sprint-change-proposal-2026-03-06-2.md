# Sprint Change Proposal – 2026-03-06 (2)

## 1. Issue Summary

- **Change description:** Remove the `render` command and make `review` emit Markdown directly by default. Add a `--json` flag to `review` for structured JSON output.
- **Trigger type:** UX simplification — deliberate course correction during active sprint.
- **Context:** Stories 1.9 and 1.9b were implemented as a two-command pipeline (`prr review | prr render`). In use, the render step is always required for human-readable output, making it mandatory plumbing rather than a composable option. Merging the two steps removes friction and makes the primary use case a single command.
- **Evidence:**
  - Current flow requires `prr review <PR_URL> | prr render` for Markdown output — an unnecessary burden for the common case.
  - FR25 ("Richard can receive a Markdown review report as the default output") and FR26 ("Richard can request structured JSON output for automation workflows") are both better served by a single command with a `--json` flag.
  - Story 1.9b (`render` command) is entirely superseded by this change.

---

## 2. Impact Analysis

### Epic Impact

- **Epic 1:** Remains in-progress. No epic-level goals change — FR25 and FR26 remain satisfied.
- **Story impact:**
  - **Story 1.9** (`review` command): Acceptance criteria require update — default output becomes Markdown, `--json` flag emits structured JSON.
  - **Story 1.9b** (`render` command): Superseded and removed from the active plan. Sprint-status entry removed.
- **No other stories affected.** Epic 2 (release pipeline) is unaffected.

### Artifact Conflict Analysis

- **PRD:** No structural conflict. FR25 and FR26 remain fully satisfied by the updated `review` command contract.
- **Epics document:** Story 1.9 AC needs rewriting. Story 1.9b entry should be removed or annotated as superseded.
- **Architecture:** No architecture changes required. The review command's output contract simplifies; no new components introduced.
- **Tests:** `review_render_test.go` requires significant updates (tests checking JSON-only `review` output need `--json` added or rewritten; all `render`-specific tests removed). `root_test.go` needs `render` removed from expected command list.
- **Sprint tracking:** `sprint-status.yaml` entry for `1-9b` removed.

### Technical Impact

- **Code changes:** Modify `review.go` (add `--json` flag, branch output based on flag). Delete `render.go`. Update `review_render_test.go` and `root_test.go`.
- **No infrastructure or CI changes required.**
- **Risk level:** Low. `renderMarkdown()` already exists in `render.go` and will be moved/kept in the codebase. Test coverage is comprehensive and will be updated to cover new contract.

---

## 3. Recommended Approach

- **Selected path:** Option 1 — Direct Adjustment.
- **Rationale:**
  - Both stories (1.9 and 1.9b) are already complete — no rollback needed, just a forward amendment.
  - All the rendering logic already exists; the change is purely about when and how it is invoked.
  - The revised contract still satisfies all FRs and simplifies the user-facing API.
- **Effort estimate:** Low (< 2 hours implementation).
- **Timeline impact:** Same sprint, same day.
- **Risk:** Low — all existing tests will be updated to cover both default Markdown and `--json` JSON modes.

---

## 4. Detailed Change Proposals

### A) `cmd/prr/review.go` — Add `--json` flag, emit Markdown by default

**Section:** `init()` flag registration

OLD:
```go
reviewCmd.Flags().String("model", "", "Copilot model to use for review generation")
```

NEW:
```go
reviewCmd.Flags().String("model", "", "Copilot model to use for review generation")
reviewCmd.Flags().Bool("json", false, "Emit structured JSON output instead of Markdown")
```

**Section:** Final output block in `RunE`

OLD:
```go
encoded, err := json.Marshal(reviewOutput)
if err != nil {
    return apperrors.WrapRuntime("failed to encode review JSON", err)
}
if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(encoded)); err != nil {
    return apperrors.WrapRuntime("failed to write output", err)
}
```

NEW:
```go
emitJSON, err := cmd.Flags().GetBool("json")
if err != nil {
    return apperrors.WrapRuntime("failed to parse json flag", err)
}

if emitJSON {
    encoded, err := json.Marshal(reviewOutput)
    if err != nil {
        return apperrors.WrapRuntime("failed to encode review JSON", err)
    }
    if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(encoded)); err != nil {
        return apperrors.WrapRuntime("failed to write output", err)
    }
} else {
    markdown := renderMarkdown(reviewOutput)
    if _, err := fmt.Fprintln(cmd.OutOrStdout(), markdown); err != nil {
        return apperrors.WrapRuntime("failed to write markdown output", err)
    }
}
```

Rationale: Markdown is the default because it satisfies FR25 directly. `--json` satisfies FR26 for automation consumers. `renderMarkdown()` already exists and is moved out of the deleted `render.go`.

---

### B) `cmd/prr/render.go` — DELETE

The file is removed entirely. The `renderMarkdown()` and `severityHeading()` functions are retained in the codebase (they were already in `render.go` and will remain there until the file is deleted — they will need to either remain in `render.go` temporarily and be moved, or `render.go` is converted to just hold these helper functions without registering the command).

**Implementation note:** Since `renderMarkdown()` is defined in `render.go`, the cleanest approach is:
- Remove the `renderCmd` variable, `init()` registration, and the `var renderCmd` command from `render.go`
- Rename the file to `render_helpers.go` (or move the helpers into `review.go`)

Alternatively: keep `render.go` as a helper-only file (no command registration).

Rationale: Eliminates the user-facing `render` subcommand while retaining the rendering logic used by `review`.

---

### C) `cmd/prr/review_render_test.go` — Update tests

- Tests that validate stdout as JSON (without `--json` flag) → add `--json` to args.
- `TestReviewOutputCanBePipedIntoRenderDeterministically` → rewrite to `TestReviewCommandEmitsDeterministicMarkdown` — run review twice with same stub, assert identical Markdown output with required headings.
- `TestRenderCommandProducesDeterministicMarkdown` → remove (superseded by Markdown-default review tests).
- `TestRenderCommandRejectsMalformedJSON`, `TestRenderCommandRejectsMissingRequiredFields`, `TestRenderCommandRejectsMissingFindingID`, `TestRenderCommandRejectsNonPositiveFindingLine`, `TestRenderCommandVerboseWhatIfDiagnostics` → remove (render command removed).
- `resetRenderFlagState()` helper → remove.
- `resetReviewFlagState()` → add `json` flag reset entry.

---

### D) `cmd/prr/root_test.go` — Remove `render` from expected commands

OLD:
```go
expected := map[string]bool{
    "checkout": false,
    "render":   false,
    "review":   false,
    "version":  false,
}
```

NEW:
```go
expected := map[string]bool{
    "checkout": false,
    "review":   false,
    "version":  false,
}
```

---

### E) `_bmad-output/planning-artifacts/epics.md` — Update Story 1.9 AC, annotate 1.9b

Story 1.9 acceptance criteria updated to reflect: default Markdown output, `--json` for JSON. Story 1.9b description updated to note it is superseded.

---

### F) `_bmad-output/implementation-artifacts/sprint-status.yaml` — Remove 1.9b

Remove the entry: `1-9b-implement-render-command-for-json-to-markdown: done`

---

## 5. Implementation Handoff

- **Scope classification:** Minor — direct development team implementation.
- **Handoff:** Development team (Amelia / dev agent).
- **Success criteria:**
  - `prr review <PR_URL>` emits Markdown to stdout by default.
  - `prr review <PR_URL> --json` emits structured JSON to stdout.
  - `prr render` is no longer a registered command.
  - All tests pass (`go test ./...`).
  - `root_test.go` test `TestPlaceholderCommandsRegistered` passes without `render`.
  - FR25 and FR26 verified by updated test coverage.
