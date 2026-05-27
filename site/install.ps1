# Termix Windows Installer
# Public URL:
# irm https://muyleanging.github.io/termix/install.ps1 | iex

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\Termix",
    [switch]$NoSetup,
    [switch]$NoPath
)

$ErrorActionPreference = "Stop"

$RepoOwner = "muyleanging"
$RepoName = "termix"
$ExeName = "termix.exe"

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

function Write-TermixError {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Get-WindowsArch {
    $arch = $env:PROCESSOR_ARCHITECTURE

    if ($arch -eq "AMD64") {
        return "x86_64"
    }

    if ($arch -eq "ARM64") {
        return "arm64"
    }

    throw "Unsupported Windows architecture: $arch"
}

function Get-Release {
    if ($Version -eq "latest") {
        $url = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
        Write-TermixInfo "Checking latest release..."
    } else {
        $url = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/tags/$Version"
        Write-TermixInfo "Checking release $Version..."
    }

    try {
        return Invoke-RestMethod -Uri $url -Headers @{ "User-Agent" = "TermixInstaller" }
    }
    catch {
        throw "Unable to fetch release from GitHub. $_"
    }
}

function Add-ToUserPath {
    param([string]$PathToAdd)

    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ($currentPath -split ";" | Where-Object { $_ -eq $PathToAdd }) {
        Write-TermixInfo "Install directory already exists in User PATH."
        return
    }

    $newPath = if ([string]::IsNullOrWhiteSpace($currentPath)) {
        $PathToAdd
    } else {
        "$currentPath;$PathToAdd"
    }

    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $env:Path = "$env:Path;$PathToAdd"

    Write-TermixSuccess "Added Termix to User PATH."
}

function Install-Termix {
    $arch = Get-WindowsArch
    $assetName = "termix_Windows_$arch.zip"
    $release = Get-Release
    $tag = $release.tag_name

    Write-TermixInfo "Selected release: $tag"

    $asset = $release.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1

    if (-not $asset) {
        Write-TermixError "Could not find release asset: $assetName"
        Write-Host "Available assets:"
        $release.assets | ForEach-Object { Write-Host " - $($_.name)" }
        throw "Missing release asset for this system."
    }

    $tempRoot = Join-Path $env:TEMP "termix-install"
    $tempZip = Join-Path $tempRoot $assetName
    $extractDir = Join-Path $tempRoot "extract"

    if (Test-Path $tempRoot) {
        Remove-Item $tempRoot -Recurse -Force
    }

    New-Item -ItemType Directory -Path $tempRoot | Out-Null
    New-Item -ItemType Directory -Path $extractDir | Out-Null

    Write-TermixInfo "Downloading $assetName..."
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tempZip -UseBasicParsing

    Write-TermixInfo "Extracting..."
    Expand-Archive -Path $tempZip -DestinationPath $extractDir -Force

    $binary = Get-ChildItem -Path $extractDir -Recurse -Filter $ExeName | Select-Object -First 1

    if (-not $binary) {
        throw "termix.exe was not found inside the release archive."
    }

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir | Out-Null
    }

    $targetExe = Join-Path $InstallDir $ExeName

    if (Test-Path $targetExe) {
        $backupName = "termix.exe.bak-" + (Get-Date -Format "yyyyMMdd-HHmmss")
        $backupPath = Join-Path $InstallDir $backupName
        Copy-Item $targetExe $backupPath -Force
        Write-TermixInfo "Backed up existing termix.exe to $backupPath"
    }

    Copy-Item $binary.FullName $targetExe -Force
    Write-TermixSuccess "Installed Termix to $targetExe"

    if (-not $NoPath) {
        Add-ToUserPath -PathToAdd $InstallDir
    }

    Write-TermixInfo "Verifying install..."
    & $targetExe --version

    if ($LASTEXITCODE -ne 0) {
        throw "Termix installed, but version check failed."
    }

    Write-TermixSuccess "Termix installed successfully."

    if (-not $NoSetup) {
        $answer = Read-Host "Run termix setup now? [Y/n]"
        if ($answer -eq "" -or $answer.ToLower() -eq "y" -or $answer.ToLower() -eq "yes") {
            & $targetExe setup
        } else {
            Write-TermixInfo "You can run setup later with: termix setup"
        }
    }

    Write-Host ""
    Write-TermixSuccess "Done."
    Write-Host "Open a new terminal if 'termix' is not found immediately."
}

try {
    Install-Termix
}
catch {
    Write-TermixError $_
    exit 1
}
