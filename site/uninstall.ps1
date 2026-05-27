# Termix Windows Uninstaller
# Public URL:
# irm https://muyleanging.github.io/Termix-Tui/uninstall.ps1 | iex

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\Termix",
    [switch]$CleanData,
    [switch]$KeepPath
)

$ErrorActionPreference = "Stop"

function Write-TermixInfo {
    param([string]$Message)
    Write-Host "[TERMIX] $Message" -ForegroundColor Cyan
}

function Write-TermixSuccess {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-TermixWarn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Remove-FromUserPath {
    param([string]$PathToRemove)

    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ([string]::IsNullOrWhiteSpace($currentPath)) {
        return
    }

    $parts = $currentPath -split ";" | Where-Object {
        $_ -and ($_ -ne $PathToRemove)
    }

    $newPath = ($parts -join ";")
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

    Write-TermixSuccess "Removed Termix from User PATH."
}

Write-TermixWarn "This removes the Termix executable."
Write-TermixWarn "To remove shell profile integration safely, run this before deleting Termix:"
Write-Host "  termix uninstall"
Write-Host ""

$answer = Read-Host "Continue uninstalling Termix executable? [y/N]"
if ($answer.ToLower() -ne "y" -and $answer.ToLower() -ne "yes") {
    Write-TermixInfo "Cancelled."
    exit 0
}

if (Test-Path $InstallDir) {
    Remove-Item $InstallDir -Recurse -Force
    Write-TermixSuccess "Removed install directory: $InstallDir"
} else {
    Write-TermixInfo "Install directory not found: $InstallDir"
}

if (-not $KeepPath) {
    Remove-FromUserPath -PathToRemove $InstallDir
}

$dataDir = Join-Path $env:LOCALAPPDATA "Termix"

if ($CleanData) {
    if (Test-Path $dataDir) {
        $dataAnswer = Read-Host "Remove Termix data folder $dataDir ? This removes config/cache/themes. [y/N]"
        if ($dataAnswer.ToLower() -eq "y" -or $dataAnswer.ToLower() -eq "yes") {
            Remove-Item $dataDir -Recurse -Force
            Write-TermixSuccess "Removed Termix data folder."
        }
    }
} else {
    Write-TermixInfo "Termix data folder was kept: $dataDir"
    Write-TermixInfo "Run with -CleanData if you want to remove config/cache/themes."
}

Write-TermixSuccess "Uninstall complete."
