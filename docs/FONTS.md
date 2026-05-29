# Fonts

Termix recommends Nerd Fonts for Oh My Posh icons, but missing fonts do not block startup.

Recommended font:

- MesloLGM Nerd Font

Useful commands:

```bash
termix fonts list
termix fonts install "MesloLGM Nerd Font" --yes
termix fonts apply "MesloLGM Nerd Font"
```

Termix installs Meslo through Oh My Posh when possible:

```bash
oh-my-posh font install meslo
```

`termix fonts apply` saves the Termix font and applies it where Termix can safely edit settings:

- Windows Terminal: updates `settings.json` defaults with a backup.
- macOS Terminal.app: sets Basic, default, and startup profiles through AppleScript when available.
- VS Code: sets `terminal.integrated.fontFamily` when the user settings file is valid JSON.
- Linux terminals: prints the exact font to select in the terminal preferences.

For Apple Terminal manually open `Terminal > Settings > Profiles > Text > Font`, choose `MesloLGM Nerd Font Mono`, then restart Terminal.

Fallback stack:

```text
MesloLGM Nerd Font
monospace
```
