# Install prr on Windows.
# Usage: irm https://raw.githubusercontent.com/richardthombs/prr/main/scripts/install.ps1 | iex
$ErrorActionPreference = "Stop"

$Repo = "richardthombs/prr"
$Binary = "prr"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\prr\bin" }

# Detect architecture
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default {
        Write-Error "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)"
        exit 1
    }
}

# Resolve latest version
$Version = $env:VERSION
if (-not $Version) {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $Release.tag_name
}

if (-not $Version) {
    Write-Error "Failed to resolve latest release version"
    exit 1
}

$VersionNum = $Version -replace '^v', ''
$Archive = "${Binary}_${VersionNum}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$Archive"

Write-Host "Installing prr $Version (windows/$Arch) to $InstallDir..."

$Tmp = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_.FullName }
try {
    $ArchivePath = Join-Path $Tmp $Archive
    Invoke-WebRequest -Uri $Url -OutFile $ArchivePath -UseBasicParsing
    Expand-Archive -Path $ArchivePath -DestinationPath $Tmp -Force

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path (Join-Path $Tmp "$Binary.exe") -Destination (Join-Path $InstallDir "$Binary.exe") -Force
} finally {
    Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}

Write-Host "Installed: $InstallDir\$Binary.exe"

$PathDirs = $env:PATH -split ";"
if ($InstallDir -notin $PathDirs) {
    Write-Host ""
    Write-Host "NOTE: $InstallDir is not in your PATH."
    Write-Host "Add it permanently with:"
    Write-Host "  [Environment]::SetEnvironmentVariable('PATH', `$env:PATH + ';$InstallDir', 'User')"
}
