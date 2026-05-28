# Termix

**Termix** is an open-source CLI/TUI terminal theme manager for Oh My Posh, Nerd Fonts, Windows Terminal, PowerShell, Git Bash, WSL, macOS, and Linux.

It helps developers install, preview, repair, and apply terminal themes and fonts without manually editing shell profile files.

## Topics

`terminal` `tui` `cli` `oh-my-posh` `nerd-fonts` `windows-terminal` `powershell` `git-bash` `wsl` `macos` `linux` `golang` `go` `themes` `terminal-themes` `shell` `developer-tools` `open-source` `profile-manager` `font-manager`

## Homepage

https://muyleanging.github.io/Termix-Tui/

## Quick Install

Normal users do not need Go, `git clone`, or `go install`. Use the GitHub Pages installer. It downloads the latest binary from GitHub Releases.

### Windows

```powershell
irm https://muyleanging.github.io/Termix-Tui/install.ps1 | iex
termix setup
termix-tui
```

### macOS / Linux

```bash
curl -fsSL https://muyleanging.github.io/Termix-Tui/install.sh | bash
termix setup
termix-tui
```

## Manual Download

Download the latest binary from GitHub Releases:

https://github.com/MuyleangIng/Termix-Tui/releases/latest

## Developer Build

Only use this if you want to contribute or build from source:

```bash
git clone https://github.com/MuyleangIng/Termix-Tui.git
cd termix
go mod tidy
go test ./...
go build -o bin/termix .
./bin/termix tui
```

On Windows:

```powershell
go build -o bin\termix.exe .
.\bin\termix.exe tui
```

## Features

- Cobra CLI command: `termix`
- Bubble Tea and Lip Gloss TUI with keyboard, mouse, resize-aware layouts, panels, status indicators, and Nerd Font glyphs
- First-time setup wizard for shell, font, and theme selection
- Environment detection for Windows Terminal, PowerShell 7, Git Bash, WSL, Oh My Posh, ANSI, Unicode, and profile hints
- Real Oh My Posh preview engine using `oh-my-posh print primary --config "<theme>.omp.json"`
- Theme scanner for `.omp.json` files, metadata extraction, favorites, categories, cache rebuild, and official theme import
- Font manager for Nerd Fonts, installed fallbacks, custom font names, and Windows Terminal font integration
- Safe profile writer that backs up profile files and replaces one managed Termix block
- Repair, reinstall, cache rebuild, doctor, installer, updater, and uninstaller commands

## Commands

```text
termix                                      Show CLI help and commands
termix tui                                  Launch the main dashboard
termix-tui                                  Launch the main dashboard from installer alias
termix setup                                Run the first-time setup wizard
termix doctor                               Run diagnostics
termix repair --dry-run                     Preview repair actions
termix repair                               Repair cache, config, theme path, and profile integration
termix reinstall                            Clean cache metadata, rebuild themes, repair profile, save config
termix cache rebuild                        Rebuild theme cache from real files
termix cache clear                          Clear cache metadata only
termix themes update                        Download official Oh My Posh themes
termix themes apply <theme> --profile <p>   Apply a theme to a profile
termix fonts list                           Show recommended font status
termix fonts install <font> --yes           Install a supported Nerd Font
termix fonts apply <font> --windows-terminal Save/apply a font
termix uninstall                            Fully remove Termix profiles, data, themes, and executable
termix uninstall profile                    Remove only Termix shell profile blocks
```

## Release

Create a public release by pushing a tag:

```bash
git checkout main
git pull
go test ./...
git tag -a v0.1.0 -m "Termix v0.1.0 first public release"
git push origin v0.1.0
```

The release workflow publishes:

- `termix_Windows_x86_64.zip`
- `termix_Windows_arm64.zip`
- `termix_Linux_x86_64.tar.gz`
- `termix_Linux_arm64.tar.gz`
- `termix_Darwin_x86_64.tar.gz`
- `termix_Darwin_arm64.tar.gz`
- `checksums.txt`

## GitHub Pages

Public install URL:

```text
https://muyleanging.github.io/Termix-Tui/
```

Installer scripts:

```text
https://muyleanging.github.io/Termix-Tui/install.ps1
https://muyleanging.github.io/Termix-Tui/install.sh
https://muyleanging.github.io/Termix-Tui/uninstall.ps1
https://muyleanging.github.io/Termix-Tui/uninstall.sh
```

Use only the GitHub Pages URL above unless a custom domain is configured later.

## Configuration

Termix reads `~/.termix/config.yaml` when present. Defaults include official Oh My Posh theme paths, favorite themes, PowerShell 7 as the default shell, and a safe fallback font stack.
