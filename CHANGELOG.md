# Changelog

## v0.2.8

- Made TUI startup faster by avoiding network theme installation and full macOS font probing during launch.
- Added local starter themes when no theme cache exists, so `termix-tui` opens immediately after install.
- Added uninstall confirmation in CLI and TUI remove actions.
- Made full uninstall remove Homebrew Oh My Posh on macOS and local Oh My Posh binaries on Linux.
- Updated uninstall scripts to pass `--yes` after their own confirmation prompt.

## v0.2.7

- Restored the v0.2.5 TUI rendering behavior while keeping the font workflow focused.
- Switched the default Nerd Font install path to `oh-my-posh font install meslo`.
- Added cross-platform terminal font apply support for Windows Terminal, Apple Terminal, and VS Code.
- Improved macOS font detection through `fc-list`, `system_profiler`, `mdfind`, and standard font folders.
- Updated docs for Apple Terminal manual font selection and restart steps.

## v0.2.5

- Added Oh My Posh font install fallback when macOS Homebrew cask install fails.
- Prevented first-time setup from failing only because a Nerd Font install failed.
- Added Linux Nerd Font install support through Oh My Posh where available.

## v0.2.4

- Made macOS Nerd Font installs idempotent by checking `brew list --cask` before installing.
- Treated Homebrew “already installed” font output as success.
- Improved macOS font install errors with the exact manual `brew install --cask` command.

## v0.2.3

- Added macOS Homebrew cask mappings for recommended Nerd Fonts.
- Made setup install the selected recommended Nerd Font on macOS and Windows.
- Improved font detection for Homebrew, system, and Linux user font directories.
- Updated the public site font instructions for macOS terminal apps.

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
