@echo off
setlocal
set "ROOT=%~dp0"
if exist "%ROOT%bin\termix.exe" (
  "%ROOT%bin\termix.exe" %*
) else (
  go run "%ROOT%" %*
)
