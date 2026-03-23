# Build and Install from Source

PRR is distributed source-first. Clone the repository and build/install for your own platform.

## 1. Common Steps (all platforms)

1. Clone the repository.
2. Build and test:
   - `go build ./...`
   - `go test ./...`
3. Install:
   - `make install` (equivalent to `go install ./cmd/prr`)
4. Verify:
   - `prr version`

If `prr` is not found after install, add your Go bin directory to PATH:

- `go env GOBIN` if set
- otherwise `$(go env GOPATH)/bin`

## 2. macOS

### Prerequisites

- Git
- Go 1.25+
- Make (optional, only for `make install` convenience)

Example with Homebrew:

```bash
brew install go git make
```

### Install

```bash
git clone https://github.com/richardthombs/prr.git
cd prr
go build ./...
go test ./...
make install
```

## 3. Ubuntu Linux

### Prerequisites

- Git
- Go 1.25+
- Build tools and Make

Example:

```bash
sudo apt update
sudo apt install -y git build-essential make golang-go
```

If the packaged Go version is older than required, install Go from the official tarball and ensure `go version` reports 1.25+.

### Install

```bash
git clone https://github.com/richardthombs/prr.git
cd prr
go build ./...
go test ./...
make install
```

## 4. Windows

> **WSL users:** If you prefer to run PRR inside Windows Subsystem for Linux, see [wsl-setup.md](wsl-setup.md) for a complete WSL + Git Credential Manager + Copilot CLI setup guide instead.

### Prerequisites

- Git for Windows
- Go 1.25+
- One of:
  - GNU Make (`mingw32-make`/`make`) for Makefile commands
  - or use pure Go commands directly

Example with winget:

```powershell
winget install --id Git.Git -e
winget install --id GoLang.Go -e
```

### Install with Go commands (works without Make)

```powershell
git clone https://github.com/richardthombs/prr.git
cd prr
go build ./...
go test ./...
go install ./cmd/prr
```

### Install with Make (optional)

From Git Bash or a Make-capable shell:

```bash
git clone https://github.com/richardthombs/prr.git
cd prr
go build ./...
go test ./...
make install
```

## 5. Uninstall

Remove the `prr` binary from your Go bin directory:

- macOS/Linux: `rm "$(go env GOPATH)/bin/prr"` (or from `$(go env GOBIN)`)
- Windows PowerShell: `Remove-Item "$env:USERPROFILE\go\bin\prr.exe"`
