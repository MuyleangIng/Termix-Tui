
# Termix

**Termix** is an open-source CLI/TUI terminal theme manager for Oh My Posh, Nerd Fonts, Windows Terminal, PowerShell, Git Bash, WSL, macOS, and Linux.

It helps developers install, preview, repair, and apply terminal themes and fonts without manually editing shell profile files.

## Topics

`terminal` `tui` `cli` `oh-my-posh` `nerd-fonts` `windows-terminal` `powershell` `git-bash` `wsl` `macos` `linux` `golang` `go` `themes` `terminal-themes` `shell` `developer-tools` `open-source` `profile-manager` `font-manager`

## Homepage

https://muyleanging.github.io/Termix-Tui/

---

## Theme Gallery Showcase

Browse a growing gallery of 120+ terminal themes inside an interactive, layout-aware TUI dashboard. 

### Main Dashboard & Overview

<p align="center">
  <img src="https://muyleanging.github.io/Termix-Tui/assets/themes.png" alt="Termix Themes Grid" width="100%" />
</p>

<p align="center">
  <img src="https://muyleanging.github.io/Termix-Tui/assets/termix-tui.png" alt="Termix TUI Dashboard Interface" width="49%" />
  <img src="https://muyleanging.github.io/Termix-Tui/assets/termix-terminal-demo.svg" alt="Termix Terminal Dynamic Preview" width="49%" />
</p>

### Interactive Shell Previews

Below are live environment render demos showing how the theme configs adapt fluidly to modern text styling, specific color palettes, and complex Nerd Font prompt segments:

#### Terminal Vector Simulations
<p align="center">
  <img src="https://muyleanging.github.io/Termix-Tui/assets/termix-terminal-demo-dark.svg" alt="Dark Mode Terminal Variant" width="49%" />
  <img src="https://muyleanging.github.io/Termix-Tui/assets/termix-terminal-demo-light.svg" alt="Light Mode Terminal Variant" width="49%" />
</p>

#### Snapshot Catalog
Here is a look at selected configurations from the automated scanner:

| **Theme Variant** | **Live Render Pipeline Snapshot** |
| :--- | :--- |
| `theme-1_shell` | <img src="https://muyleanging.github.io/Termix-Tui/assets/theme-1_shell.png" alt="theme-1_shell" width="400" height="220" /> |
| `theme-amro` | <img src="https://muyleanging.github.io/Termix-Tui/assets/theme-amro.png" alt="theme-amro" width="400" height="220" /> |
| `theme-atomicBit` | <img src="https://muyleanging.github.io/Termix-Tui/assets/theme-atomicBit.png" alt="theme-atomicBit" width="400" height="220" /> |
| `theme-cert` | <img src="https://muyleanging.github.io/Termix-Tui/assets/theme-cert.png" alt="theme-cert" width="400" height="220" /> |
| `theme-clean-detailed` | <img src="https://muyleanging.github.io/Termix-Tui/assets/theme-clean-detailed.png" alt="theme-clean-detailed" width="400" height="220" /> |
| `theme-gmay` | <img src="https://muyleanging.github.io/Termix-Tui/assets/theme-gmay.png" alt="theme-gmay" width="400" height="220" /> |
| `theme-m365princess` | <img src="https://muyleanging.github.io/Termix-Tui/assets/theme-m365princess.png" alt="theme-m365princess" width="400" height="220" /> |

---

## Quick Install

Normal users do not need Go, `git clone`, or `go install`. Use the GitHub Pages installer. It downloads the latest binary from GitHub Releases. On Windows, the installer also bootstraps the default Termix tools, CascadiaCode Nerd Font, and official Oh My Posh themes. After installing, open a new terminal and run `termix setup` so the interactive picker receives arrow keys correctly.

### Windows

```powershell
irm [https://muyleanging.github.io/Termix-Tui/install.ps1](https://muyleanging.github.io/Termix-Tui/install.ps1) | iex
termix setup
termix-tui

```

### macOS / Linux

```bash
curl -fsSL [https://muyleanging.github.io/Termix-Tui/install.sh](https://muyleanging.github.io/Termix-Tui/install.sh) | bash
termix setup
termix-tui

```

## Manual Download

Download the latest binary from GitHub Releases:

https://github.com/MuyleangIng/Termix-Tui/releases/latest

## Developer Build

Only use this if you want to contribute or build from source:

```bash
git clone [https://github.com/MuyleangIng/Termix-Tui.git](https://github.com/MuyleangIng/Termix-Tui.git)
cd Termix-Tui
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

---

## Features

- Cobra CLI command: `termix`
- Bubble Tea and Lip Gloss TUI with keyboard, mouse, resize-aware layouts, panels, status indicators, and Nerd Font glyphs
- Quick first-time setup that picks the target shell profile and applies the default font/theme
- Environment detection for Windows Terminal, PowerShell 7, Git Bash, WSL, Oh My Posh, ANSI, Unicode, and profile hints
- Real Oh My Posh preview engine using `oh-my-posh print primary --config "<theme>.omp.json"`
- Theme scanner for `.omp.json` files, metadata extraction, favorites, categories, cache rebuild, and official theme import
- Font manager for Nerd Fonts, installed fallbacks, custom font names, and terminal font integration on Windows Terminal, Apple Terminal, and VS Code
- Safe profile writer that backs up profile files and replaces one managed Termix block
- Repair, reinstall, cache rebuild, doctor, installer, updater, and uninstaller commands

## Commands

```text
termix                                      Show CLI help and commands
termix tui                                  Launch the main dashboard
termix-tui                                  Launch the main dashboard from installer alias
termix setup                                Pick a profile and apply the default setup
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
termix fonts apply <font>                    Save/apply a font to supported terminal settings
termix uninstall                            Confirm, then fully remove Termix profiles, data, themes, tools, and executable
termix uninstall profile                    Remove only Termix shell profile blocks

```

---

## Release

Create a public release by pushing a tag:

```bash
git checkout main
git pull
go test ./...
git tag -a v0.1.0 -m "Termix v0.1.0 first public release"
git push origin v0.1.0

```

The release workflow automatically compiles and packages:

* `termix_Windows_x86_64.zip`
* `termix_Windows_arm64.zip`
* `termix_Linux_x86_64.tar.gz`
* `termix_Linux_arm64.tar.gz`
* `termix_Darwin_x86_64.tar.gz`
* `termix_Darwin_arm64.tar.gz`
* `checksums.txt`

## GitHub Pages

Public install URL:

```text
[https://muyleanging.github.io/Termix-Tui/](https://muyleanging.github.io/Termix-Tui/)

```

Installer and Uninstaller scripts:

```text
[https://muyleanging.github.io/Termix-Tui/install.ps1](https://muyleanging.github.io/Termix-Tui/install.ps1)
[https://muyleanging.github.io/Termix-Tui/install.sh](https://muyleanging.github.io/Termix-Tui/install.sh)
[https://muyleanging.github.io/Termix-Tui/uninstall.ps1](https://muyleanging.github.io/Termix-Tui/uninstall.ps1)
[https://muyleanging.github.io/Termix-Tui/uninstall.sh](https://muyleanging.github.io/Termix-Tui/uninstall.sh)

```

## Configuration

Termix reads `~/.termix/config.yaml` when present. Defaults include official Oh My Posh theme paths, favorite themes, PowerShell 7 as the default shell, and a safe fallback font stack.
