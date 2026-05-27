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

const markerStart = "# >>> Termix Oh My Posh >>>"
const markerEnd = "# <<< Termix Oh My Posh <<<"

var legacyMarkers = [][2]string{
	{"# >>> TERMIX managed prompt >>>", "# <<< TERMIX managed prompt <<<"},
}

func ApplyPrompt(home, shellName, themePath string) error {
	adapter := adapterFor(shellName)
	if adapter == nil {
		return fmt.Errorf("unsupported shell %q", shellName)
	}
	path := adapter.ProfilePath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	existing, _ := os.ReadFile(path)
	if len(existing) > 0 {
		backup := fmt.Sprintf("%s.termix.%s.bak", path, time.Now().Format("20060102150405"))
		if err := os.WriteFile(backup, existing, 0o644); err != nil {
			return err
		}
	}
	next := replaceManagedBlock(string(existing), adapter.InitSnippet(themePath))
	return os.WriteFile(path, []byte(next), 0o644)
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
	return shell.PowerShell{}
}
