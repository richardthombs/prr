# Release Process Contract

This document defines the release artifact and version contract used by PRR release automation.
Stories 2.2 and 2.3 must consume this contract directly and must not redefine naming/version rules.

## 1. Release Target Matrix

Initial release targets are fixed and ordered for deterministic artifact generation:

1. `darwin/arm64`
2. `linux/amd64`
3. `linux/arm64`
4. `windows/amd64`

Order is part of the contract. Scripts that consume artifact lists can rely on this exact order.

## 2. Artifact Naming Contract

Each platform target produces one static CLI binary with this filename template:

`prr_<version>_<os>_<arch><extension>`

Rules:

- `<version>` is the exact validated Git tag (for example `v1.4.0` or `v1.4.0-rc.1`).
- `<os>` and `<arch>` use Go target identifiers (`darwin`, `linux`, `windows`; `amd64`, `arm64`).
- `<extension>` is empty for non-Windows targets and `.exe` for Windows targets.
- No ad hoc alternates are permitted.

Examples for `v1.4.0`:

- `prr_v1.4.0_darwin_arm64`
- `prr_v1.4.0_linux_amd64`
- `prr_v1.4.0_linux_arm64`
- `prr_v1.4.0_windows_amd64.exe`

## 3. Version Source of Truth

Release versions are sourced from Git tags only.

Canonical release tag formats:

- Stable: `vMAJOR.MINOR.PATCH`
- Pre-release candidate: `vMAJOR.MINOR.PATCH-rc.N`

Rejected tag examples:

- `1.2.3` (missing leading `v`)
- `v1.2` (not full SemVer)
- `vx.y.z` (non-numeric)
- `v1.2.3+meta` (build metadata not allowed in release tags for this repository)

## 4. Dev Build Version Contract

When a build is not sourced from a valid release tag, PRR reports a deterministic development version:

`v0.0.0-dev+<shortsha>`

Where `<shortsha>` is the first seven characters of the commit SHA, or `unknown` if unavailable.

## 5. Build Metadata Fields (`prr version` + packaging)

Build metadata is injected through link-time variables and treated as a shared contract:

- `version`: release tag (for release builds) or non-release marker (for dev builds)
- `commit`: full commit SHA used to derive `<shortsha>`
- `buildDate`: build timestamp (RFC3339 preferred)

`prr version` output expectations:

- Release build: prints exact SemVer tag (`vMAJOR.MINOR.PATCH` or `vMAJOR.MINOR.PATCH-rc.N`)
- Dev build: prints `v0.0.0-dev+<shortsha>`

## 6. Release Entry Validation (Fail Fast)

The release workflow validates the release tag before any build/publish stage runs.
If validation fails, the workflow exits with actionable diagnostics and no artifact publication job is allowed to execute.

## 7. SemVer Bump Decision Matrix

Use this matrix when selecting the next release version.
Record the selected bump and rationale in the release PR description or tag annotation during release preparation.

| Change Classification | Bump | Trigger Examples |
| --- | --- | --- |
| Breaking API/contract changes | MAJOR | CLI output contract changes, removed flags, incompatible JSON schema |
| Backwards-compatible features | MINOR | New commands/flags, additive JSON fields, new optional workflows |
| Backwards-compatible fixes only | PATCH | Bug fixes, performance fixes, documentation-only behaviour clarifications |

Release rationale template:

- `Version candidate:` `vX.Y.Z`
- `Selected bump:` `major|minor|patch`
- `Reason:` concise list of relevant merged changes
- `Risk notes:` migration or compatibility concerns

## 8. Story Boundaries and Downstream Consumers

Story 2.1 provides contract definitions only. It does not complete full packaging/publishing automation.

Story 2.2 must consume:

- Target matrix and deterministic ordering
- Artifact filename template and extension rules
- Build metadata injection fields (`version`, `commit`, `buildDate`)

Story 2.3 must consume:

- Release tag validation gate behaviour
- Artifact naming contract during upload/publication
- No-publish-on-invalid-tag rule
