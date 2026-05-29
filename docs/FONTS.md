# Fonts

Termix uses `MesloLGM Nerd Font` through Oh My Posh for prompt icons. Fonts are installed on the host system and selected by the terminal app, not by the shell profile. For WSL or containers, install and configure the font in the host terminal, such as Windows Terminal or VS Code.

Useful commands:

```bash
oh-my-posh font install meslo
termix fonts apply "MesloLGM Nerd Font"
```

`termix fonts apply` saves the Termix font and applies it where Termix can safely edit settings:

- Windows Terminal: updates `settings.json` defaults with a backup.
- macOS Terminal.app: sets Apple Terminal profiles to `MesloLGM Nerd Font Mono` through AppleScript when available.
- VS Code: sets `terminal.integrated.fontFamily` to `MesloLGM Nerd Font` when the user settings file is valid JSON.
- Linux terminals: prints the exact font to select in the terminal preferences.

After installing or applying a font, close and reopen the terminal so the terminal app reloads the profile font.

## Manual terminal settings

Windows Terminal:

```json
{
  "profiles": {
    "defaults": {
      "font": {
        "face": "MesloLGM Nerd Font"
      }
    }
  }
}
```

VS Code:

```json
"terminal.integrated.fontFamily": "MesloLGM Nerd Font"
```

Apple Terminal:

```bash
osascript -e 'tell application "Terminal" to set font of settings set "Basic" to "MesloLGM Nerd Font Mono"'
```

You can also open `Terminal > Settings > Profiles > Text > Font`, choose `MesloLGM Nerd Font Mono`, then restart Terminal.

## Common fixes

- Icons show boxes or question marks: run `termix fonts apply "MesloLGM Nerd Font"`, then close all terminal tabs and open a new terminal.
- Installer font step failed: run `oh-my-posh font install meslo`, then `termix fonts apply "MesloLGM Nerd Font"`.
- Themes only show a small starter set: run `termix install themes`, `termix cache rebuild`, then reopen `termix-tui`.
