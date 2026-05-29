package font

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Font struct {
	Name      string
	Family    string
	Installed bool
}

var Supported = []Font{
	{Name: "MesloLGM Nerd Font", Family: "MesloLGM Nerd Font"},
}

var FallbackStack = []string{
	"MesloLGM Nerd Font",
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
	if runtime.GOOS == "windows" {
		return "MesloLGM Nerd Font"
	}
	return "monospace"
}

func Detect(home string) []Font {
	return detect(home, true)
}

func DetectQuick(home string) []Font {
	return detect(home, false)
}

func detect(home string, includeSystemCommands bool) []Font {
	fontDirs := []string{
		filepath.Join(os.Getenv("WINDIR"), "Fonts"),
		filepath.Join(home, "AppData", "Local", "Microsoft", "Windows", "Fonts"),
		filepath.Join(home, "Library", "Fonts"),
		"/Library/Fonts",
		"/System/Library/Fonts",
		filepath.Join(home, ".local", "share", "fonts"),
		"/usr/local/share/fonts",
		"/usr/share/fonts",
	}
	items := make([]Font, len(Supported))
	copy(items, Supported)
	systemFonts := ""
	if includeSystemCommands {
		systemFonts = installedFontNames()
	}
	for i := range items {
		if matchesInstalledName(systemFonts, items[i].Name, items[i].Family) {
			items[i].Installed = true
			continue
		}
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

func installedFontNames() string {
	var parts []string
	if out, err := exec.Command("fc-list").CombinedOutput(); err == nil {
		parts = append(parts, string(out))
	}
	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("system_profiler", "SPFontsDataType").CombinedOutput(); err == nil {
			parts = append(parts, string(out))
		}
		if out, err := exec.Command("mdfind", "kMDItemKind == 'Font'").CombinedOutput(); err == nil {
			parts = append(parts, string(out))
		}
	}
	return strings.ToLower(strings.Join(parts, "\n"))
}

func matchesInstalledName(haystack string, names ...string) bool {
	if haystack == "" {
		return false
	}
	for _, name := range names {
		for _, alias := range fontAliases(name) {
			if strings.Contains(haystack, strings.ToLower(alias)) {
				return true
			}
		}
	}
	return false
}

func fontAliases(name string) []string {
	compact := strings.NewReplacer(" ", "", "-", "", "_", "").Replace(name)
	return []string{
		name,
		compact,
		strings.ReplaceAll(name, "MesloLGM Nerd Font", "MesloLGM Nerd Font Mono"),
		strings.ReplaceAll(name, "MesloLGM Nerd Font", "MesloLGM NF"),
	}
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
		"/Library/Fonts",
		"/System/Library/Fonts",
		filepath.Join(home, ".local", "share", "fonts"),
		"/usr/local/share/fonts",
		"/usr/share/fonts",
	}
}

func isNerdFamily(family string) bool {
	name := strings.ToLower(family)
	return strings.Contains(name, "nerd") || strings.Contains(name, "nf") || strings.Contains(name, "caskaydia")
}

func isGenericFamily(family string) bool {
	return strings.EqualFold(family, "monospace")
}
