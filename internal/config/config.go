package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HomeDir        string   `mapstructure:"home_dir"`
	ThemeDirs      []string `mapstructure:"theme_dirs"`
	FavoriteThemes []string `mapstructure:"favorite_themes"`
	DefaultShell   string   `mapstructure:"default_shell"`
	DefaultFont    string   `mapstructure:"default_font"`
	FontStack      []string `mapstructure:"font_stack"`
	CustomFonts    []string `mapstructure:"custom_fonts"`
	ActivityHeight int      `mapstructure:"activity_height"`
	BorderStyle    string   `mapstructure:"border_style"`
	SetupComplete  bool     `mapstructure:"setup_complete"`
}

func Load(path string) (Config, error) {
	home, _ := os.UserHomeDir()
	cfgDir := filepath.Join(home, ".termix")

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(cfgDir)
	if path != "" {
		v.SetConfigFile(path)
	}

	v.SetDefault("home_dir", cfgDir)
	v.SetDefault("theme_dirs", []string{
		filepath.Join(cfgDir, "themes"),
		filepath.Join(home, "AppData", "Local", "Programs", "oh-my-posh", "themes"),
	})
	v.SetDefault("favorite_themes", []string{"catppuccin_mocha", "paradox", "atomic", "dracula", "tokyo"})
	v.SetDefault("default_shell", "PowerShell 7")
	v.SetDefault("default_font", "CaskaydiaCove Nerd Font")
	v.SetDefault("font_stack", []string{"CaskaydiaCove Nerd Font", "Cascadia Code", "JetBrains Mono", "Fira Code", "Consolas", "Courier New", "monospace"})
	v.SetDefault("custom_fonts", []string{})
	v.SetDefault("activity_height", 7)
	v.SetDefault("border_style", "unicode")

	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	if _, err := os.Stat(setupMarker(cfg.HomeDir)); err == nil {
		cfg.SetupComplete = true
	}
	return cfg, nil
}

func MarkSetupComplete(homeDir string) error {
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(setupMarker(homeDir), []byte("complete\n"), 0o644)
}

func SaveSetupChoices(cfg Config, shellName, fontName, themeName string) error {
	if err := os.MkdirAll(cfg.HomeDir, 0o755); err != nil {
		return err
	}
	if shellName == "" {
		shellName = cfg.DefaultShell
	}
	if fontName == "" {
		fontName = cfg.DefaultFont
	}
	if themeName == "" || strings.EqualFold(themeName, "No prompt style") {
		themeName = firstOrDefault(cfg.FavoriteThemes, "catppuccin_mocha")
	}
	data := fmt.Sprintf(`home_dir: %q
theme_dirs:
%s
favorite_themes:
  - %q
default_shell: %q
default_font: %q
font_stack:
%s
custom_fonts:
%s
activity_height: %d
border_style: %q
`, cfg.HomeDir, yamlList(cfg.ThemeDirs), themeName, shellName, fontName, yamlList(firstNonEmptyList(cfg.FontStack, []string{fontName, "Cascadia Code", "JetBrains Mono", "Fira Code", "Consolas", "Courier New", "monospace"})), yamlList(cfg.CustomFonts), normalizeActivityHeight(cfg.ActivityHeight), normalizeBorderStyle(cfg.BorderStyle))
	return os.WriteFile(filepath.Join(cfg.HomeDir, "config.yaml"), []byte(data), 0o644)
}

func SaveFontChoice(cfg Config, fontName string) error {
	if strings.TrimSpace(fontName) == "" {
		fontName = cfg.DefaultFont
	}
	cfg.DefaultFont = fontName
	cfg.FontStack = prependUnique(fontName, firstNonEmptyList(cfg.FontStack, []string{"Cascadia Code", "JetBrains Mono", "Fira Code", "Consolas", "Courier New", "monospace"}))
	return SaveSetupChoices(cfg, cfg.DefaultShell, cfg.DefaultFont, firstOrDefault(cfg.FavoriteThemes, "catppuccin_mocha"))
}

func SaveActivityHeight(cfg Config, height int) error {
	cfg.ActivityHeight = normalizeActivityHeight(height)
	return SaveSetupChoices(cfg, cfg.DefaultShell, cfg.DefaultFont, firstOrDefault(cfg.FavoriteThemes, "catppuccin_mocha"))
}

func SaveBorderStyle(cfg Config, style string) error {
	cfg.BorderStyle = normalizeBorderStyle(style)
	return SaveSetupChoices(cfg, cfg.DefaultShell, cfg.DefaultFont, firstOrDefault(cfg.FavoriteThemes, "catppuccin_mocha"))
}

func SaveCustomFonts(cfg Config, fonts []string) error {
	cfg.CustomFonts = uniqueStrings(fonts)
	return SaveSetupChoices(cfg, cfg.DefaultShell, cfg.DefaultFont, firstOrDefault(cfg.FavoriteThemes, "catppuccin_mocha"))
}

func setupMarker(homeDir string) string {
	return filepath.Join(homeDir, "setup.done")
}

func yamlList(values []string) string {
	var b strings.Builder
	for _, value := range values {
		fmt.Fprintf(&b, "  - %q\n", value)
	}
	return strings.TrimRight(b.String(), "\n")
}

func firstOrDefault(values []string, fallback string) string {
	if len(values) == 0 || values[0] == "" {
		return fallback
	}
	return values[0]
}

func firstNonEmptyList(values, fallback []string) []string {
	if len(values) == 0 {
		return fallback
	}
	return values
}

func prependUnique(first string, values []string) []string {
	out := []string{first}
	seen := map[string]bool{strings.ToLower(first): true}
	for _, value := range values {
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func normalizeActivityHeight(height int) int {
	if height <= 0 {
		return 7
	}
	if height < 4 {
		return 4
	}
	if height > 20 {
		return 20
	}
	return height
}

func normalizeBorderStyle(style string) string {
	switch strings.ToLower(strings.TrimSpace(style)) {
	case "ascii", "none":
		return strings.ToLower(strings.TrimSpace(style))
	default:
		return "unicode"
	}
}
