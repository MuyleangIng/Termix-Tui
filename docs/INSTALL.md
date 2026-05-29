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
The installer also bootstraps Oh My Posh, MesloLGM Nerd Font, and official Oh My Posh themes.

Open a new terminal after install, then run `termix setup`. The setup screen asks for the target profile and applies the default Termix font/theme automatically.

Manual downloads are available at:

https://github.com/MuyleangIng/Termix-Tui/releases/latest
