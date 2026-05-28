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

Write-TermixWarn "This removes Termix, Termix profile blocks, Termix data, and Termix-installed dependencies."
Write-Host ""

$answer = Read-Host "Continue uninstalling Termix executable? [y/N]"
if ($answer.ToLower() -ne "y" -and $answer.ToLower() -ne "yes") {
    Write-TermixInfo "Cancelled."
    exit 0
}

$termixCmd = Get-Command termix -ErrorAction SilentlyContinue
if ($termixCmd) {
    Write-TermixInfo "Running Termix full uninstall first..."
    & $termixCmd.Source uninstall
    if ($LASTEXITCODE -eq 0) {
        Write-TermixSuccess "Termix full uninstall completed."
    } else {
        Write-TermixWarn "Termix full uninstall exited with code $LASTEXITCODE. Continuing fallback cleanup."
    }
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
$termixHome = Join-Path $env:USERPROFILE ".termix"

if ($CleanData) {
    if (Test-Path $dataDir) {
        $dataAnswer = Read-Host "Remove Termix data folder $dataDir ? This removes config/cache/themes. [y/N]"
        if ($dataAnswer.ToLower() -eq "y" -or $dataAnswer.ToLower() -eq "yes") {
            Remove-Item $dataDir -Recurse -Force
            Write-TermixSuccess "Removed Termix data folder."
        }
    }
    if (Test-Path $termixHome) {
        Remove-Item $termixHome -Recurse -Force
        Write-TermixSuccess "Removed Termix home folder: $termixHome"
    }
} else {
    Write-TermixInfo "Termix data folder was kept: $dataDir"
    Write-TermixInfo "Termix home folder was kept: $termixHome"
    Write-TermixInfo "Run with -CleanData if you want to remove config/cache/themes."
}

Write-TermixSuccess "Uninstall complete."
