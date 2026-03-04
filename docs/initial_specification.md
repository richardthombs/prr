# PRR – Pull Request Review Tool (Specification)

## Purpose

PRR is a CLI tool that performs an **on-demand** review of a pull request without disrupting a developer’s existing working copy.

Typical usage:

- A person messages: “Please review PR12345.”
- The reviewer runs: `prr review PR12345`
- PRR fetches the PR snapshot, computes a diff, sends it to a review engine (e.g., an LLM), and outputs a review (optionally publishing it back to the PR).

This document specifies **Version 1**.

---

## Goals

- **On-demand** PR review triggered by a human running a command.
- **No interference** with any existing local repository state.
- **Deterministic Git-based diff** based on the PR’s merge result.
- **Simple v1 bundle**: one unified diff + minimal metadata.
- **Composable design**: small commands that can be piped (JSON in/out).
- **Extensible**: clear interfaces for alternate PR providers and review engines.

---

## Non-Goals (Version 1)

- Incremental “review only changes since last review”.
- Diff chunking/grouping heuristics.
- Inline comment line-mapping correctness guarantees.
- Auto-fix commits.
- Running tests/lint in the workflow (can be added later).

---

## Platform Compatibility

### Git transport assumptions

PRR relies on Git’s ability to fetch a PR snapshot via remote refs.

- **Azure DevOps** supports:
  - `refs/pull/<PR_ID>/head`
  - `refs/pull/<PR_ID>/merge`
- **GitHub** supports:
  - `refs/pull/<PR_ID>/head`
  - `refs/pull/<PR_ID>/merge` (may be absent when PR cannot be merged cleanly)

Version 1 assumes a **merge ref exists**. If a provider cannot supply a merge ref, the provider must fail with a clear error.

---

## High-Level Flow

1. Resolve PR identity (PR id + repo URL and remote).
2. Ensure a cached **bare mirror** exists locally.
3. Fetch the PR **merge ref** into the bare mirror.
4. Create an **isolated worktree** checked out at the merge ref.
5. Compute the PR’s contribution diff: `HEAD^1..HEAD`.
6. Build a v1 review bundle (metadata + file list + diff stat + unified patch).
7. Run a review engine to produce structured findings.
8. Render results to Markdown (default).
9. Optionally publish results back to the PR.
10. Cleanup the worktree unless the user asks to keep it.

---

## Repository Management

### Bare mirror cache

PRR maintains a cached bare mirror per repository URL.

Default location:

```
~/.cache/prr/repos/<repoHash>.git
```

Creation:

```
git clone --mirror <repoUrl> <bareDir>
```

Update:

```
git -C <bareDir> remote update --prune
```

### Concurrency

Implementations must prevent concurrent corruption of the same bare mirror.

- Use a per-repo lock (e.g., file lock adjacent to `<bareDir>`).

---

## PR Snapshot Fetching

PRR fetches the PR merge ref into the bare mirror.

Remote ref:

- `pull/<PR_ID>/merge` (Git allows omitting the `refs/` prefix)

Local ref target (within the mirror):

- `refs/prr/pull/<PR_ID>/merge`

Fetch command:

```
git -C <bareDir> fetch <remote> pull/<PR_ID>/merge:refs/prr/pull/<PR_ID>/merge
```

Optionally (not required for v1) the tool may also fetch the PR head ref.

---

## Workspace (Worktree) Management

PRR must not use the developer’s existing working directory.

### Worktree location

Default location:

```
~/.cache/prr/work/<repoHash>/pr-<PR_ID>/<runId>/
```

Where `<runId>` is a unique identifier (e.g., timestamp).

### Worktree creation

Create a detached worktree at the merge ref:

```
git -C <bareDir> worktree add --detach <workDir> refs/prr/pull/<PR_ID>/merge
```

### Cleanup

By default, PRR removes the worktree after completing the review:

```
git -C <bareDir> worktree remove --force <workDir>
git -C <bareDir> worktree prune
```

A `--keep` option must exist to preserve the worktree for inspection.

---

## Diff Generation

Within the worktree:

- `HEAD` is the PR merge commit.
- `HEAD^1` is the target branch parent.

### Required diff outputs

1. Diff stat:

```
git -C <workDir> diff --stat HEAD^1..HEAD
```

2. Changed file list:

```
git -C <workDir> diff --name-only HEAD^1..HEAD
```

3. Unified patch:

```
git -C <workDir> diff -M --patch HEAD^1..HEAD
```

Version 1 does **not** require any additional context capture beyond these outputs.

---

## Review Bundle (Version 1)

The v1 bundle is a single JSON document.

### Required fields

- `meta.prId` (int)
- `meta.repoUrl` (string)
- `meta.provider` (string; e.g., `azure-devops`, `github`)
- `stat` (string)
- `files` (array of file paths)
- `patch` (string; unified diff)

### Example

```json
{
  "meta": {
    "prId": 12345,
    "repoUrl": "https://dev.azure.com/org/project/_git/repo",
    "provider": "azure-devops"
  },
  "stat": "4 files changed, 40 insertions(+), 5 deletions(-)",
  "files": [
    "internal/git/worktree.go",
    "internal/git/fetch.go"
  ],
  "patch": "diff --git a/internal/git/worktree.go b/internal/git/worktree.go\n..."
}
```

### Size limits

Implementations must include configurable safety limits (with sensible defaults), for example:

- maximum patch bytes
- maximum number of files

If limits are exceeded, PRR should fail with a clear error describing what exceeded the limit.

---

## Review Output

The review engine produces a structured JSON review.

### Required fields

- `summary` (string)
- `risk.score` (0–1 float)
- `risk.reasons` (array of strings)
- `findings[]` (array)
  - `id` (string)
  - `file` (string)
  - `line` (int; may be 0 if unknown)
  - `severity` (`blocker|important|suggestion|nit`)
  - `category` (`correctness|security|performance|readability|api|tests|other`)
  - `message` (string)
  - `suggestion` (string)
- `checklist` (array of strings)

### Finding identity

The `findings[].id` is a per-review reference identifier. It does not need to be stable across reruns of the same PR.

---

## Rendering

### Default renderer

PRR must render the review as Markdown to stdout.

Minimum required sections:

- Summary
- Risk (score + reasons)
- Findings (grouped by severity)
- Checklist / Next actions

---

## Publishing (Optional in Version 1)

Publishing is optional. If implemented, it must be a separate step/command.

Supported publish targets (at most):

- One top-level PR comment containing the Markdown review.

Inline comments and status checks are out of scope for v1.

---

## CLI Command Model

PRR is designed as a set of composable commands that read JSON from stdin and emit JSON to stdout.

### Required commands

- `prr review <PR_ID>`
  - High-level orchestration command.
  - Must not require the user to be inside a repo.
  - Must accept `--repo <repoUrl>` if the PR provider cannot resolve a repo URL from PR id alone.

### Recommended composable commands

These are recommended to support piping and testing, but may be implemented internally initially:

- `prr checkout <PR_URL>` → emits workspace payload including `PRRef` + `bareDir` + `mergeRef` + `workDir`
- `prr diff` → emits diff outputs
- `prr bundle` → emits `Bundle`
- `prr review-engine` → emits `Review`
- `prr render` → prints Markdown
- `prr publish` → posts results (optional)

All commands must also support equivalent flags (not only stdin).

For `checkout`, implementation should auto-detect provider/context from supported PR URL formats:

- Azure DevOps: `https://dev.azure.com/<org>/<project>/_git/<repo>/pullrequest/<id>`
- GitHub: `https://github.com/<owner>/<repo>/pull/<id>`

---

## Data Structures (JSON + Go)

### PRRef

```json
{ "prId": 12345, "repoUrl": "...", "remote": "origin", "provider": "azure-devops" }
```

```go
type PRRef struct {
  PRID     int    `json:"prId"`
  RepoURL  string `json:"repoUrl"`
  Remote   string `json:"remote"`
  Provider string `json:"provider"`
}
```

### Workspace

```json
{ "bareDir": "...", "workDir": "...", "mergeRef": "..." }
```

```go
type Workspace struct {
  BareDir   string `json:"bareDir"`
  WorkDir   string `json:"workDir"`
  MergeRef  string `json:"mergeRef"`
}
```

### Bundle

```go
type Bundle struct {
  Meta  map[string]any `json:"meta"`
  Stat  string         `json:"stat"`
  Files []string       `json:"files"`
  Patch string         `json:"patch"`
}
```

### Review

```go
type Risk struct {
  Score   float64  `json:"score"`
  Reasons []string `json:"reasons"`
}

type Finding struct {
  ID         string `json:"id"`
  File       string `json:"file"`
  Line       int    `json:"line"`
  Severity   string `json:"severity"`
  Category   string `json:"category"`
  Message    string `json:"message"`
  Suggestion string `json:"suggestion"`
}

type Review struct {
  Summary   string    `json:"summary"`
  Risk      Risk      `json:"risk"`
  Findings  []Finding `json:"findings"`
  Checklist []string  `json:"checklist"`
}
```

---

## Interfaces

Implementations must keep provider and engine swappable.

### Git operations

```go
type Git interface {
  EnsureMirror(ctx context.Context, repoURL string) (bareDir string, err error)
  UpdateMirror(ctx context.Context, bareDir string) error
  FetchPRMergeRef(ctx context.Context, bareDir, remote string, prID int) (mergeRef string, err error)
  CreateWorktree(ctx context.Context, bareDir, mergeRef, workDir string) error
  RemoveWorktree(ctx context.Context, bareDir, workDir string) error
  DiffStat(ctx context.Context, workDir string) (string, error)
  DiffNames(ctx context.Context, workDir string) ([]string, error)
  DiffPatch(ctx context.Context, workDir string) (string, error)
}
```

### PR provider

```go
type PRProvider interface {
  // Resolve may require repoUrl depending on provider capabilities.
  Resolve(ctx context.Context, prID int, opts map[string]string) (PRRef, error)

  // Optional publish support.
  PublishComment(ctx context.Context, pr PRRef, markdown string) error
}
```

### Bundler

```go
type Bundler interface {
  Build(ctx context.Context, pr PRRef, stat string, files []string, patch string) (Bundle, error)
}
```

### Review engine

```go
type ReviewEngine interface {
  Review(ctx context.Context, bundle Bundle) (Review, error)
}
```

### Renderer

```go
y itype Renderer interface {
  RenderMarkdown(ctx context
```
