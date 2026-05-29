# Install Termix

## Windows

```powershell
irm https://muyleanging.github.io/Termix-Tui/install.ps1 | iex
termix setup
```

## macOS / Linux

```bash
curl -fsSL https://muyleanging.github.io/Termix-Tui/install.sh | bash
termix setup
```

The GitHub Pages scripts download the latest binary from GitHub Releases.
If Termix is already installed, the installer asks before replacing it. Use `-Yes` on Windows or `--yes` on macOS/Linux for unattended updates.

The installer also bootstraps Oh My Posh, MesloLGM Nerd Font, official Oh My Posh themes, and terminal/VS Code font settings.

Close and reopen your terminal after install, then run `termix setup`. The setup screen asks for the target profile and applies the default Termix font/theme automatically.

Pinned release examples:

```powershell
& ([scriptblock]::Create((irm https://muyleanging.github.io/Termix-Tui/install.ps1))) -Version v0.2.13 -Yes
```

```bash
curl -fsSL https://muyleanging.github.io/Termix-Tui/install.sh | bash -s -- --version v0.2.13 --yes
```

Manual downloads are available at:

https://github.com/MuyleangIng/Termix-Tui/releases/latest
