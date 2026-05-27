package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/muyleanging/termix/internal/config"
)

func EnsureStarterThemes(cfg config.Config) error {
	if len(cfg.ThemeDirs) == 0 {
		return nil
	}
	dir := cfg.ThemeDirs[0]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	for _, item := range starterThemes {
		path := filepath.Join(dir, item.name+".omp.json")
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte(renderStarterTheme(item)), 0o644); err != nil {
			return err
		}
	}
	return nil
}

type starterTheme struct {
	name       string
	userBG     string
	pathBG     string
	gitBG      string
	promptFG   string
	darkText   string
	lightText  string
	promptIcon string
}

var starterThemes = []starterTheme{
	{name: "catppuccin_mocha", userBG: "#67E8F9", pathBG: "#7C3AED", gitBG: "#34D399", promptFG: "#C084FC", darkText: "#0B1020", lightText: "#E5E7EB", promptIcon: "╰─❯"},
	{name: "dracula", userBG: "#BD93F9", pathBG: "#FF79C6", gitBG: "#50FA7B", promptFG: "#F8F8F2", darkText: "#282A36", lightText: "#F8F8F2", promptIcon: "❯"},
	{name: "tokyo", userBG: "#7AA2F7", pathBG: "#BB9AF7", gitBG: "#9ECE6A", promptFG: "#7DCFFF", darkText: "#1A1B26", lightText: "#C0CAF5", promptIcon: "❯"},
	{name: "atomic", userBG: "#F97316", pathBG: "#2563EB", gitBG: "#22C55E", promptFG: "#FACC15", darkText: "#111827", lightText: "#F9FAFB", promptIcon: "λ"},
	{name: "paradox", userBG: "#00D7FF", pathBG: "#AF00FF", gitBG: "#FFD700", promptFG: "#00FF87", darkText: "#080808", lightText: "#FFFFFF", promptIcon: "➜"},
	{name: "spaceship", userBG: "#38BDF8", pathBG: "#6366F1", gitBG: "#F472B6", promptFG: "#A7F3D0", darkText: "#020617", lightText: "#E0F2FE", promptIcon: "🚀"},
	{name: "night_owl", userBG: "#82AAFF", pathBG: "#C792EA", gitBG: "#C3E88D", promptFG: "#7FDBCA", darkText: "#011627", lightText: "#D6DEEB", promptIcon: "☾"},
}

func renderStarterTheme(theme starterTheme) string {
	return fmt.Sprintf(`{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "version": 3,
  "final_space": true,
  "blocks": [
    {
      "type": "prompt",
      "alignment": "left",
      "segments": [
        {
          "type": "session",
          "style": "diamond",
          "foreground": "%s",
          "background": "%s",
          "leading_diamond": "\ue0b6",
          "trailing_diamond": "\ue0b0",
          "template": " {{ .UserName }} "
        },
        {
          "type": "path",
          "style": "powerline",
          "foreground": "%s",
          "background": "%s",
          "powerline_symbol": "\ue0b0",
          "template": " \uf07c {{ .Path }} "
        },
        {
          "type": "git",
          "style": "powerline",
          "foreground": "%s",
          "background": "%s",
          "powerline_symbol": "\ue0b0",
          "template": " {{ .HEAD }}{{ if .Working.Changed }} *{{ end }} "
        }
      ]
    },
    {
      "type": "prompt",
      "alignment": "left",
      "newline": true,
      "segments": [
        {
          "type": "text",
          "style": "plain",
          "foreground": "%s",
          "template": "%s "
        }
      ]
    }
  ]
}
`, theme.darkText, theme.userBG, theme.lightText, theme.pathBG, theme.darkText, theme.gitBG, theme.promptFG, theme.promptIcon)
}
