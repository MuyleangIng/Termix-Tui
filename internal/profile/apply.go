package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/muyleanging/termix/internal/font"
	"github.com/muyleanging/termix/internal/shell"
	"github.com/muyleanging/termix/internal/terminal"
)

const markerStart = "# BEGIN TERMIX THEME"
const markerEnd = "# END TERMIX THEME"

var legacyMarkers = [][2]string{
	{"# >>> Termix Oh My Posh >>>", "# <<< Termix Oh My Posh <<<"},
	{"# >>> TERMIX managed prompt >>>", "# <<< TERMIX managed prompt <<<"},
}

func ApplyPrompt(home, shellName, themePath string) error {
	adapter := adapterFor(shellName)
	if adapter == nil {
		return fmt.Errorf("unsupported shell %q", shellName)
	}
	themePath, err := absoluteThemePath(themePath)
	if err != nil {
		return err
	}
	path := adapter.ProfilePath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create profile directory %q: %w", filepath.Dir(path), err)
	}
	existing, _ := os.ReadFile(path)
	if len(existing) > 0 {
		backup := fmt.Sprintf("%s.termix.%s.bak", path, time.Now().Format("20060102150405"))
		if err := os.WriteFile(backup, existing, 0o644); err != nil {
			return fmt.Errorf("backup profile %q: %w", path, err)
		}
	}
	snippet := adapter.InitSnippet(themePath)
	next := replaceManagedBlock(string(existing), snippet)
	if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
		return fmt.Errorf("write profile %q: %w", path, err)
	}
	if err := verifyPromptBlock(path, snippet); err != nil {
		return err
	}
	return nil
}

func RepairPrompt(home, shellName, themePath string) error {
	return ApplyPrompt(home, shellName, themePath)
}

func RemovePrompt(home, shellName string) error {
	adapter := adapterFor(shellName)
	if adapter == nil {
		return fmt.Errorf("unsupported shell %q", shellName)
	}
	path := adapter.ProfilePath(home)
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	backup := fmt.Sprintf("%s.termix.%s.bak", path, time.Now().Format("20060102150405"))
	if err := os.WriteFile(backup, existing, 0o644); err != nil {
		return err
	}
	next := removeManagedBlocks(string(existing))
	return os.WriteFile(path, []byte(next), 0o644)
}

func HasPromptBlock(home, shellName string) bool {
	adapter := adapterFor(shellName)
	if adapter == nil {
		return false
	}
	data, err := os.ReadFile(adapter.ProfilePath(home))
	if err != nil {
		return false
	}
	existing := string(data)
	if strings.Contains(existing, markerStart) && strings.Contains(existing, markerEnd) {
		return true
	}
	for _, markers := range legacyMarkers {
		if strings.Contains(existing, markers[0]) && strings.Contains(existing, markers[1]) {
			return true
		}
	}
	return false
}

func ProfilePath(home, shellName string) string {
	adapter := adapterFor(shellName)
	if adapter == nil {
		return ""
	}
	return adapter.ProfilePath(home)
}

func ApplyWindowsTerminalFont(home, family string) error {
	return terminal.SetFont(home, font.ResolveAvailableFamily(home, family))
}

func absoluteThemePath(themePath string) (string, error) {
	themePath = strings.TrimSpace(themePath)
	if themePath == "" {
		return "", fmt.Errorf("theme path is empty")
	}
	abs, err := filepath.Abs(themePath)
	if err != nil {
		return "", fmt.Errorf("resolve absolute theme path %q: %w", themePath, err)
	}
	if info, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("theme file %q is not accessible: %w", abs, err)
	} else if info.IsDir() {
		return "", fmt.Errorf("theme path %q is a directory, expected an .omp.json file", abs)
	}
	return abs, nil
}

func verifyPromptBlock(path, snippet string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("verify profile %q: %w", path, err)
	}
	text := string(data)
	if !strings.Contains(text, markerStart) || !strings.Contains(text, markerEnd) {
		return fmt.Errorf("profile %q was written but the Termix theme block is missing", path)
	}
	if strings.Count(text, markerStart) != 1 || strings.Count(text, markerEnd) != 1 {
		return fmt.Errorf("profile %q contains duplicate Termix theme blocks", path)
	}
	if !strings.Contains(text, snippet) {
		return fmt.Errorf("profile %q was written but does not contain the expected Oh My Posh init command", path)
	}
	return nil
}

func replaceManagedBlock(existing, snippet string) string {
	block := markerStart + "\n" + snippet + "\n" + markerEnd
	cleaned := removeManagedBlocks(existing)
	if strings.TrimSpace(cleaned) == "" {
		return block + "\n"
	}
	return strings.TrimRight(cleaned, "\r\n") + "\n\n" + block + "\n"
}

func removeManagedBlocks(existing string) string {
	existing = removeBlockPair(existing, markerStart, markerEnd)
	for _, markers := range legacyMarkers {
		existing = removeBlockPair(existing, markers[0], markers[1])
	}
	return strings.TrimSpace(existing) + "\n"
}

func removeBlockPair(existing, startMarker, endMarker string) string {
	for {
		start := strings.Index(existing, startMarker)
		end := strings.Index(existing, endMarker)
		if start < 0 || end < start {
			break
		}
		end += len(endMarker)
		existing = strings.TrimRight(existing[:start], "\r\n") + "\n" + strings.TrimLeft(existing[end:], "\r\n")
	}
	return existing
}

func adapterFor(name string) shell.Adapter {
	for _, adapter := range shell.Supported() {
		if strings.EqualFold(adapter.Name(), name) {
			return adapter
		}
	}
	return nil
}
