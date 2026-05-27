Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
  Write-Host "Go is required. Install from https://go.dev/dl/ or winget install GoLang.Go" -ForegroundColor Yellow
  exit 1
}

Write-Host "[TERMIX] Resolving Go modules"
go mod tidy

Write-Host "[TERMIX] Building termix"
New-Item -ItemType Directory -Force -Path "$Root\bin" | Out-Null
go build -o "$Root\bin\termix.exe" .

Write-Host "[TERMIX] Built $Root\bin\termix.exe" -ForegroundColor Green
