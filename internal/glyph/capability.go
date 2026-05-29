package glyph

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/muyleanging/termix/internal/config"
	"github.com/muyleanging/termix/internal/font"
)

type Capability struct {
	Mode          string
	Terminal      string
	UTF8          bool
	NerdInstalled bool
	NerdActive    bool
	ASCII         bool
	Warning       string
	Fix           string
}

func Detect(ctx context.Context, cfg config.Config, home string) Capability {
	term := terminalName()
	utf8 := localeUTF8()
	installed := hasNerdFont(home)
	active := terminalUsesNerdFont(ctx, term)
	mode := strings.ToLower(strings.TrimSpace(cfg.Icons))
	if envTruthy("TERMIX_ASCII_ICONS") {
		mode = "ascii"
	}
	if mode == "" {
		mode = "nerd"
	}
	capability := Capability{
		Mode:          mode,
		Terminal:      term,
		UTF8:          utf8,
		NerdInstalled: installed,
		NerdActive:    active,
	}
	if mode == "ascii" || !utf8 || !installed || !active {
		capability.ASCII = true
		capability.Mode = "ascii"
	}
	switch {
	case !utf8:
		capability.Warning = "Terminal locale is not UTF-8; using ASCII icons."
		capability.Fix = "Set LANG or LC_ALL to a UTF-8 locale, then restart the terminal."
	case !installed:
		capability.Warning = "No supported Nerd Font was detected; using ASCII icons."
		capability.Fix = "Install a Nerd Font with termix fonts install \"FiraCode Nerd Font\" --yes."
	case !active:
		capability.Warning = "Nerd Font installed but current terminal profile may not be using it; using ASCII icons."
		capability.Fix = macFontFix(term)
	case capability.ASCII:
		capability.Warning = "ASCII icon fallback is active."
		capability.Fix = "Unset TERMIX_ASCII_ICONS or set icons: \"nerd\" after configuring a Nerd Font."
	}
	return capability
}

func hasNerdFont(home string) bool {
	for _, item := range font.Detect(home) {
		name := strings.ToLower(item.Name + " " + item.Family)
		if item.Installed && (strings.Contains(name, "nerd") || strings.Contains(name, "caskaydia")) {
			return true
		}
	}
	return false
}

func terminalUsesNerdFont(ctx context.Context, term string) bool {
	if envTruthy("TERMIX_NERD_FONT_ACTIVE") {
		return true
	}
	if envTruthy("TERMIX_ASCII_ICONS") {
		return false
	}
	if runtime.GOOS != "darwin" {
		return true
	}
	if strings.Contains(strings.ToLower(term), "apple_terminal") || strings.EqualFold(term, "Apple Terminal") {
		return appleTerminalUsesNerdFont(ctx)
	}
	return false
}

func appleTerminalUsesNerdFont(ctx context.Context) bool {
	profile := commandOutput(ctx, "defaults", "read", "com.apple.Terminal", "Default Window Settings")
	if strings.TrimSpace(profile) == "" {
		profile = commandOutput(ctx, "defaults", "read", "com.apple.Terminal", "Startup Window Settings")
	}
	settings := commandOutput(ctx, "defaults", "read", "com.apple.Terminal", "Window Settings")
	text := strings.ToLower(profile + "\n" + settings)
	for _, name := range []string{"caskaydiacove", "caskaydia cove", "cascadia code nerd", "jetbrainsmono nerd", "jetbrains mono nerd", "firacode nerd", "fira code nerd", "hack nerd", "meslo"} {
		if strings.Contains(text, name) {
			return true
		}
	}
	return false
}

func commandOutput(ctx context.Context, name string, args ...string) string {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	return string(out)
}

func terminalName() string {
	for _, key := range []string{"TERM_PROGRAM", "WT_SESSION", "TERM"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return "unknown"
}

func localeUTF8() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	joined := strings.ToLower(os.Getenv("LC_ALL") + " " + os.Getenv("LC_CTYPE") + " " + os.Getenv("LANG"))
	return strings.Contains(joined, "utf-8") || strings.Contains(joined, "utf8")
}

func envTruthy(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on", "ascii":
		return true
	default:
		return false
	}
}

func macFontFix(term string) string {
	if runtime.GOOS == "darwin" {
		return "Terminal: Settings > Profiles > Text > Font, choose CaskaydiaCove Nerd Font or FiraCode Nerd Font, then restart Terminal."
	}
	return "Configure your terminal profile to use an installed Nerd Font, then restart the terminal."
}

func ReplaceUnsafe(text string) string {
	replacer := strings.NewReplacer(
		"", ">", "", ">", "", "<", "", "<", "", ">", "", ">",
		"", "[", "", "]", "", "git:", "", "git:",
		"", "dir:", "", "dir:", "", "git:",
		"", "[ok]", "✔", "[ok]", "✓", "[ok]",
		"", "[!]", "⚠", "[!]",
		"", "[x]", "✗", "[x]",
		"⚙", "cfg", "󰒡", "Doctor", "󰚰", "Update", "", "Themes", "", "Fonts",
		"❯", ">", "➜", ">", "╰─", ">",
		"🚀", "ship", "☾", "moon", "λ", "lambda",
		"┌", "+", "┐", "+", "└", "+", "┘", "+", "─", "-", "│", "|",
	)
	return replacer.Replace(text)
}
