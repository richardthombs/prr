# Distribution and Install Contract

This document defines the current PRR distribution strategy.

## 1. Strategy

PRR is source-first.

- Users clone the repository.
- Users build and install locally for their own OS.
- Cross-platform support is validated through build/test/smoke automation, not prebuilt binary publication.

## 2. Canonical Commands

These commands are the source of truth on macOS, Ubuntu Linux, and Windows:

- `go build ./...`
- `go test ./...`
- `go install ./cmd/prr`

`make build`, `make test`, and `make install` are convenience wrappers where Make is available.

## 3. Platform Guidance

Required coverage:

- macOS
- Ubuntu Linux
- Windows

Each platform section must include:

- prerequisites
- build/test/install commands
- install verification (`prr version`)
- troubleshooting for PATH issues
- uninstall steps

## 4. Version Semantics

`prr version` continues to expose build metadata from the existing version contract.
No binary artifact naming contract is required for source-first distribution.

## 5. CI Expectations

CI must validate cross-platform source usability:

- build (`go build ./...`)
- tests (`go test ./...`)
- CLI smoke checks (`prr --help`, `prr version`, and one what-if flow)

## 6. Scope Boundary

Out of scope for the current plan:

- GitHub Release binary publication
- checksum publication for downloadable binaries
- release artifact matrix management

If binary distribution is reintroduced later, this contract will be superseded by a dedicated release artifact contract.
