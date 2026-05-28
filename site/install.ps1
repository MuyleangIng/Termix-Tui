# Termix Windows Installer
# Public URL:
# irm https://muyleanging.github.io/Termix-Tui/install.ps1 | iex

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\Termix",
    [switch]$NoPath,
    [switch]$NoBuildFallback,
    [switch]$NoBootstrap
)

$ErrorActionPreference = "Stop"

$RepoOwner = "MuyleangIng"
$RepoName = "Termix-Tui"
$ExeName = "termix.exe"
$TuiExeName = "termix-tui.exe"

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

function Clear-DownloadMark {
    param([string]$Path)

    if (-not (Test-Path $Path)) {
        return
    }

    Unblock-File -Path $Path -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath "${Path}:Zone.Identifier" -Force -ErrorAction SilentlyContinue
}

function Get-GoCommandPath {
    $cmd = Get-Command go -ErrorAction SilentlyContinue
    if ($cmd) {
        return $cmd.Source
    }

    $defaultGo = "C:\Program Files\Go\bin\go.exe"
    if (Test-Path $defaultGo) {
        return $defaultGo
    }

    return $null
}

function Test-TermixBinary {
    param([string]$Path)

    try {
        $output = & $Path --version 2>&1
        if ($LASTEXITCODE -eq 0) {
            return $true
        }
        $script:TermixVerifyError = ($output | Out-String).Trim()
        return $false
    }
    catch {
        $script:TermixVerifyError = $_.Exception.Message
        return $false
    }
}

function Bootstrap-TermixDefaults {
    param([string]$Path)

    if ($NoBootstrap) {
        Write-TermixInfo "Skipped dependency/theme bootstrap."
        return
    }

    Write-TermixInfo "Installing default tools, Nerd Font, and official themes..."
    try {
        & $Path install
        if ($LASTEXITCODE -eq 0) {
            Write-TermixSuccess "Default tools, font, and themes are ready."
            return
        }
        Write-TermixWarn "Bootstrap exited with code $LASTEXITCODE. You can retry later with: termix install"
    }
    catch {
        Write-TermixWarn "Bootstrap did not complete. You can retry later with: termix install"
        Write-TermixWarn $_
    }
}

function Build-TermixFromSource {
    param(
        [string]$Tag,
        [string]$TargetExe
    )

    $goPath = Get-GoCommandPath
    if (-not $goPath) {
        throw "Go is not installed, so the source-build fallback cannot run."
    }

    $sourceUrl = "https://github.com/$RepoOwner/$RepoName/archive/refs/tags/$Tag.zip"
    $sourceRoot = Join-Path $env:TEMP "termix-source-build"
    $sourceZip = Join-Path $sourceRoot "source.zip"
    $sourceExtract = Join-Path $sourceRoot "extract"

    if (Test-Path $sourceRoot) {
        Remove-Item $sourceRoot -Recurse -Force
    }

    New-Item -ItemType Directory -Path $sourceRoot | Out-Null
    New-Item -ItemType Directory -Path $sourceExtract | Out-Null

    Write-TermixWarn "Windows blocked the unsigned release binary."
    Write-TermixInfo "Go was found, so Termix will build from source locally as a fallback."
    Write-TermixInfo "Downloading source for $Tag..."
    Invoke-WebRequest -Uri $sourceUrl -OutFile $sourceZip -UseBasicParsing
    Clear-DownloadMark -Path $sourceZip

    Write-TermixInfo "Extracting source..."
    Expand-Archive -Path $sourceZip -DestinationPath $sourceExtract -Force

    $sourceDir = Get-ChildItem -Path $sourceExtract -Recurse -Filter "go.mod" |
        Where-Object { (Get-Content $_.FullName -Raw) -match "github.com/muyleanging/termix" } |
        Select-Object -First 1 |
        ForEach-Object { $_.Directory.FullName }

    if (-not $sourceDir) {
        throw "Could not find Termix source after downloading $Tag."
    }

    Write-TermixInfo "Building Termix locally..."
    $buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    $ldflags = "-X main.version=$Tag -X main.commit=source -X main.date=$buildDate"
    Push-Location $sourceDir
    try {
        & $goPath build -ldflags $ldflags -o $TargetExe .
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed with exit code $LASTEXITCODE."
        }
    }
    finally {
        Pop-Location
    }

    Clear-DownloadMark -Path $TargetExe
    Write-TermixSuccess "Built local Termix binary at $TargetExe"
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
    Clear-DownloadMark -Path $tempZip

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
    $targetTuiExe = Join-Path $InstallDir $TuiExeName

    if (Test-Path $targetExe) {
        $backupName = "termix.exe.bak-" + (Get-Date -Format "yyyyMMdd-HHmmss")
        $backupPath = Join-Path $InstallDir $backupName
        Copy-Item $targetExe $backupPath -Force
        Write-TermixInfo "Backed up existing termix.exe to $backupPath"
    }

    Copy-Item $binary.FullName $targetExe -Force
    Clear-DownloadMark -Path $binary.FullName
    Clear-DownloadMark -Path $targetExe
    Write-TermixSuccess "Installed Termix to $targetExe"

    if (-not $NoPath) {
        Add-ToUserPath -PathToAdd $InstallDir
    }

    Write-TermixInfo "Verifying install..."
    if (-not (Test-TermixBinary -Path $targetExe)) {
        if (-not $NoBuildFallback) {
            try {
                Build-TermixFromSource -Tag $tag -TargetExe $targetExe
            }
            catch {
                Write-TermixWarn $_
            }
        }

        if (-not (Test-TermixBinary -Path $targetExe)) {
            Write-TermixError "Termix installed, but Windows blocked it from running."
            Write-Host ""
            Write-Host "Installed file:"
            Write-Host "  $targetExe"
            Write-Host ""
            Write-Host "Reason:"
            Write-Host "  $script:TermixVerifyError"
            Write-Host ""
            Write-Host "Fix options:"
            Write-Host "  1. Ask Windows Security or your organization admin to allow this unsigned open-source binary."
            Write-Host "  2. Install Go, then rerun this installer so it can build Termix locally."
            Write-Host "  3. Build manually from source:"
            Write-Host "     git clone https://github.com/$RepoOwner/$RepoName.git"
            Write-Host "     cd $RepoName"
            Write-Host "     go build -o `"$targetExe`" ."
            throw "Windows Application Control blocked Termix."
        }
    }

    Copy-Item $targetExe $targetTuiExe -Force
    Clear-DownloadMark -Path $targetTuiExe
    Write-TermixSuccess "Installed TUI launcher to $targetTuiExe"
    Bootstrap-TermixDefaults -Path $targetExe
    Write-TermixSuccess "Termix installed successfully."

    Write-Host ""
    Write-TermixSuccess "Done."
    Write-Host "Next steps:"
    Write-Host "  1. Open a new terminal."
    Write-Host "  2. Run: termix setup"
    Write-Host "  3. Open the dashboard with: termix-tui"
}

try {
    Install-Termix
}
catch {
    Write-TermixError $_
    exit 1
}
