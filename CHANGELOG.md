# Changelog

## v0.2.2

- Resolved Homebrew and Oh My Posh paths on macOS when `/opt/homebrew/bin` or `/usr/local/bin` are missing from the app PATH.
- Made real preview, doctor, setup, and shell profile snippets use the same executable resolver.
- Ensured setup installs/checks Oh My Posh on macOS and Linux before applying zsh/bash profiles.

## v0.2.1

- Fixed macOS/Linux installer argument parsing and cleanup under `set -u`.
- Skipped Windows Terminal font writes on non-Windows systems during setup/apply.

## v0.2.0

- Added OS-aware installer flows for Windows, macOS, and Linux.
- Added macOS/Linux shell profile support for zsh, bash, PowerShell 7, fish, and nushell through Oh My Posh.
- Kept Windows profile choices focused on PowerShell, Windows PowerShell, Git Bash, WSL Bash, Nushell, Fish, and CMD view-only status.
- Made CLI/TUI profile lists adapt to the current OS and available shells.
- Removed Oh My Zsh install/detection support so Termix consistently uses Oh My Posh as the prompt engine.

## v0.1.0

- First public release packaging for GitHub Releases.
- GitHub Pages install scripts for Windows, macOS, and Linux.
- Termix setup, repair, reinstall, cache, theme, font, profile, and doctor workflows.
