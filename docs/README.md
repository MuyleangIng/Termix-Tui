# Termix GitHub Pages

This folder contains the static GitHub Pages site for Termix.

## Deploy

1. Push the repository to GitHub.
2. Open repository settings.
3. Go to **Pages**.
4. Set source to **Deploy from a branch**.
5. Choose the main branch and `/docs` folder.
6. Save.

GitHub will publish the static site from `docs/index.html`.

## Local Preview

Open `docs/index.html` in a browser. No build step is required.

## Public Install URLs

GitHub Pages hosts only the website and installer scripts. The scripts download real binaries from GitHub Releases. On Windows, the installer also bootstraps the default tools, CascadiaCode Nerd Font, and official Oh My Posh themes.
After installing, open a new terminal before running `termix setup`; setup is not launched from inside the installer so arrow-key selection works correctly.

```powershell
irm https://muyleanging.github.io/Termix-Tui/install.ps1 | iex
```

```bash
curl -fsSL https://muyleanging.github.io/Termix-Tui/install.sh | bash
```

## Sections

- Home
- Install
- Setup and recovery
- CLI usage
- Docs
- Fonts install and custom font guide
- Keyboard notes
- Contribute
- Feedback and contact

## Fonts

Termix recommends `MesloLGM Nerd Font` through Oh My Posh but does not require one to run. Use `termix fonts install "MesloLGM Nerd Font" --yes`, then `termix fonts apply "MesloLGM Nerd Font"` to update supported terminal settings.

Useful Windows installs:

```powershell
winget install DEVCOM.CascadiaCodeNerdFont
winget install DEVCOM.JetBrainsMonoNerdFont
winget install DEVCOM.FiraCodeNerdFont
winget install DEVCOM.HackNerdFont
```

In the Termix Fonts page:

- `Enter` saves the selected font to Termix config.
- `W` applies it to Windows Terminal.
- `I` installs a missing recommended Nerd Font after confirmation.
- `A` adds a custom font family name.
- `E` edits a selected custom font.
- `D` removes a selected custom font.
- `R` rescans installed fonts.

## Setup And Recovery

Use these commands for production setup and repair:

```powershell
termix setup
termix tui
termix repair --dry-run
termix repair
termix reinstall
termix uninstall profile
termix cache rebuild
termix themes update
termix themes apply catppuccin_mocha --profile "PowerShell 7"
termix fonts apply "JetBrainsMono Nerd Font" --windows-terminal
termix uninstall
```

Important behavior:

- `termix cache clear` removes cache metadata only.
- `termix` shows CLI help; use `termix tui` or the installer alias `termix-tui` for the full-screen dashboard.
- `termix uninstall` asks for confirmation, then performs a full uninstall: profile blocks, config, cache, downloaded themes, external tools, and executable.
- Downloaded themes are removed only with `termix uninstall downloaded-themes` when you are doing a partial uninstall.
- Profile repair replaces the managed Termix block instead of appending duplicates.
- Missing Nerd Fonts are warnings; Termix resolves an installed fallback font.
- Official themes come from <https://github.com/JanDeDobbeleer/oh-my-posh/tree/main/themes>.

## Windows Terminal F1

Termix uses `?` or `h` for in-app help and does not use `F1`.

If Windows Terminal still opens its Help/About popup, unbind `F1` in Windows Terminal `settings.json`:

```json
{
  "command": "unbound",
  "keys": "f1"
}
```
