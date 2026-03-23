# WSL Setup Guide

This guide covers everything you need to run PRR inside Windows Subsystem for Linux (WSL) on Windows 11.

## 1. Install WSL and Ubuntu

Open **PowerShell** or **Command Prompt** as Administrator and run:

```powershell
wsl --install
```

This installs the WSL 2 kernel and the default Ubuntu distribution in one step. Restart your machine when prompted.

After restarting, Ubuntu will launch automatically and ask you to create a UNIX username and password. Choose any username and a memorable password — these are separate from your Windows credentials.

Verify the installation:

```powershell
wsl --list --verbose
```

You should see your Ubuntu distribution listed with `VERSION 2`.

> **Note:** If WSL was already installed but your distribution is running version 1, upgrade it with:
> `wsl --set-version Ubuntu 2`

## 2. Update Ubuntu

Once inside the Ubuntu shell, bring it fully up to date:

```bash
sudo apt update && sudo apt upgrade -y
```

## 3. Install Prerequisites

Install the tools that PRR requires:

```bash
sudo apt install -y curl
```

## 4. Install and Configure the Git Credential Manager

The **Git Credential Manager (GCM)** lets the git inside WSL reuse your Windows credential store, so you do not need to re-enter passwords or re-authenticate for HTTPS remotes.

### 4.1 Install GCM inside WSL

GCM ships as a self-contained binary. Install the latest release:

```bash
GCM_VERSION=2.6.1
curl -fsSL "https://github.com/git-ecosystem/git-credential-manager/releases/download/v${GCM_VERSION}/gcm-linux_amd64.${GCM_VERSION}.tar.gz" \
  | sudo tar -C /usr/local/bin -xz
```

> Check https://github.com/git-ecosystem/git-credential-manager/releases for the latest version and update `GCM_VERSION` accordingly.

### 4.2 Configure GCM as your git credential helper

```bash
git config --global credential.helper /usr/local/bin/git-credential-manager
```

### 4.3 Configure GCM to use the Windows credential store

Tell GCM to delegate storage to the Windows Credential Manager so that tokens are shared with the Windows side of your machine:

```bash
git config --global credential.credentialStore wincredman
```

Optionally, allow GCM to open Windows GUI prompts from inside WSL:

```bash
git config --global credential.guiPrompt true
```

### 4.4 Verify

```bash
git config --global --get credential.helper
# Expected: /usr/local/bin/git-credential-manager
```

The first time you `git clone` or `git push` a private HTTPS repository, a Windows authentication prompt will appear and the resulting token will be stored for subsequent operations.

## 5. Install and Configure the GitHub Copilot CLI

PRR uses the **GitHub Copilot CLI** (`copilot`) — **not** the `gh` CLI — as its review engine. Follow these steps to install and authenticate it inside WSL.

### 5.1 Install Node.js (required for `copilot`)

```bash
# Install nvm (Node Version Manager)
curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
source ~/.bashrc

# Install and use the latest LTS release of Node.js
nvm install --lts
nvm use --lts
```

Verify:

```bash
node --version   # should be v20+ or the current LTS
npm --version
```

### 5.2 Install the Copilot CLI

```bash
npm install -g @githubnext/github-copilot-cli
```

Verify the binary is available:

```bash
copilot --version
```

### 5.3 Authenticate Copilot

Run the one-time authentication flow:

```bash
copilot auth
```

A browser window (on the Windows side) will open. Sign in with the GitHub account that holds an active Copilot subscription and grant the requested permissions. The token is saved locally and reused for all subsequent `copilot` invocations.

### 5.4 Verify end-to-end

```bash
echo "What time is it?" | copilot suggest
```

If Copilot returns a suggestion, authentication is working correctly.

## 6. Install PRR

With all prerequisites in place, install PRR using the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/richardthombs/prr/main/scripts/install.sh | bash
```

The script downloads the latest pre-built binary and installs it to `~/.local/bin`. If `prr` is not found after install, add that directory to your PATH:

```bash
echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.bashrc
source ~/.bashrc
```

Verify:

```bash
prr version
```

## 7. Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `git credential-manager: command not found` | PATH not updated after install | Run `source ~/.bashrc` or open a new terminal |
| GCM prompts for credentials every time | `credentialStore` not set to `wincredman` | See [§ 4.3](#43-configure-gcm-to-use-the-windows-credential-store) |
| `copilot: command not found` | npm global bin directory not in PATH | Run `npm bin -g` to find the path and add it to `~/.bashrc` |
| `copilot` reports auth error | Token expired or not stored | Re-run `copilot auth` |
| `prr: command not found` | `~/.local/bin` not in PATH | Add `$HOME/.local/bin` to PATH (see [§ 6](#6-install-prr)) |
