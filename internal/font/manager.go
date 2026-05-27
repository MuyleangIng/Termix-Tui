package font

import (
	"os"
	"path/filepath"
	"strings"
)

type Font struct {
	Name      string
	Family    string
	Installed bool
}

var Supported = []Font{
	{Name: "Cascadia Code Nerd Font", Family: "CaskaydiaCove Nerd Font"},
	{Name: "CaskaydiaCove Nerd Font", Family: "CaskaydiaCove Nerd Font"},
	{Name: "CascadiaCode Nerd Font", Family: "CaskaydiaCove Nerd Font"},
	{Name: "JetBrainsMono Nerd Font", Family: "JetBrainsMono Nerd Font"},
	{Name: "FiraCode Nerd Font", Family: "FiraCode Nerd Font"},
	{Name: "Hack Nerd Font", Family: "Hack Nerd Font"},
	{Name: "MesloLGS Nerd Font", Family: "MesloLGS NF"},
	{Name: "MesloLGM Nerd Font", Family: "MesloLGM Nerd Font"},
	{Name: "UbuntuMono Nerd Font", Family: "UbuntuMono Nerd Font"},
	{Name: "Cascadia Code", Family: "Cascadia Code"},
	{Name: "JetBrains Mono", Family: "JetBrains Mono"},
	{Name: "Fira Code", Family: "Fira Code"},
	{Name: "Consolas", Family: "Consolas"},
	{Name: "Courier New", Family: "Courier New"},
}

var FallbackStack = []string{
	"CaskaydiaCove Nerd Font",
	"Cascadia Code",
	"JetBrains Mono",
	"Fira Code",
	"Consolas",
	"Courier New",
	"monospace",
}

func Choices() []string {
	items := make([]string, 0, len(Supported))
	seen := map[string]bool{}
	for _, item := range Supported {
		if seen[item.Name] {
			continue
		}
		seen[item.Name] = true
		items = append(items, item.Name)
	}
	return items
}

func ResolveFamily(name string) string {
	for _, item := range Supported {
		if strings.EqualFold(item.Name, name) || strings.EqualFold(item.Family, name) {
			return item.Family
		}
	}
	return name
}

func ResolveAvailableFamily(home, name string) string {
	requested := ResolveFamily(name)
	if requested == "" {
		requested = FallbackStack[0]
	}
	if installed(home, requested) {
		return requested
	}
	for _, candidate := range FallbackStack {
		family := ResolveFamily(candidate)
		if isGenericFamily(family) {
			continue
		}
		if installed(home, family) {
			return family
		}
	}
	return "Consolas"
}

func Detect(home string) []Font {
	fontDirs := []string{
		filepath.Join(os.Getenv("WINDIR"), "Fonts"),
		filepath.Join(home, "AppData", "Local", "Microsoft", "Windows", "Fonts"),
		filepath.Join(home, "Library", "Fonts"),
	}
	items := make([]Font, len(Supported))
	copy(items, Supported)
	for i := range items {
		for _, fontDir := range fontDirs {
			if fontDir == "Fonts" {
				continue
			}
			if installedInDir(fontDir, items[i].Family) {
				items[i].Installed = true
				break
			}
		}
	}
	return items
}

func installed(home, family string) bool {
	for _, fontDir := range fontDirs(home) {
		if installedInDir(fontDir, family) {
			return true
		}
	}
	return false
}

func installedInDir(fontDir, family string) bool {
	if fontDir == "" {
		return false
	}
	compact := strings.NewReplacer(" ", "", "-", "", "_", "").Replace(family)
	patterns := []string{
		filepath.Join(fontDir, "*"+family+"*"),
		filepath.Join(fontDir, "*"+compact+"*"),
	}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			return true
		}
	}
	return false
}

func fontDirs(home string) []string {
	return []string{
		filepath.Join(os.Getenv("WINDIR"), "Fonts"),
		filepath.Join(home, "AppData", "Local", "Microsoft", "Windows", "Fonts"),
		filepath.Join(home, "Library", "Fonts"),
	}
}

func isNerdFamily(family string) bool {
	name := strings.ToLower(family)
	return strings.Contains(name, "nerd") || strings.Contains(name, "nf") || strings.Contains(name, "caskaydia")
}

func isGenericFamily(family string) bool {
	return strings.EqualFold(family, "monospace")
}
