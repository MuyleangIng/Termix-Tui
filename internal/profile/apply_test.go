package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepairPromptReplacesManagedBlockWithoutDuplicates(t *testing.T) {
	home := t.TempDir()
	firstTheme := writeTheme(t, home, "catppuccin_mocha")
	secondTheme := writeTheme(t, home, "dracula")
	profilePath := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	initial := "Write-Host before\n" + markerStart + "\nold snippet\n" + markerEnd + "\nWrite-Host after\n"
	if err := os.WriteFile(profilePath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RepairPrompt(home, "PowerShell 7", firstTheme); err != nil {
		t.Fatal(err)
	}
	if err := RepairPrompt(home, "PowerShell 7", secondTheme); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Count(text, markerStart) != 1 || strings.Count(text, markerEnd) != 1 {
		t.Fatalf("expected one managed block, got:\n%s", text)
	}
	if strings.Contains(text, "catppuccin_mocha") || !strings.Contains(text, secondTheme) {
		t.Fatalf("expected old theme path replaced with new theme path, got:\n%s", text)
	}
	if !strings.Contains(text, "Write-Host before") || !strings.Contains(text, "Write-Host after") {
		t.Fatalf("expected unmanaged profile content to remain, got:\n%s", text)
	}
}

func TestApplyPromptCreatesWindowsPowerShellProfileWithVerifiedBlock(t *testing.T) {
	home := t.TempDir()
	themePath := writeTheme(t, home, "amro")
	profilePath := filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")

	if err := ApplyPrompt(home, "Windows PowerShell", themePath); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, markerStart) || !strings.Contains(text, markerEnd) {
		t.Fatalf("expected managed block, got:\n%s", text)
	}
	for _, want := range []string{"init pwsh", `--config "` + themePath + `"`, "Invoke-Expression"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected Windows PowerShell block to contain %q, got:\n%s", want, text)
		}
	}
}

func TestApplyPromptRejectsMissingThemeFile(t *testing.T) {
	home := t.TempDir()
	err := ApplyPrompt(home, "PowerShell 7", filepath.Join(home, "missing.omp.json"))
	if err == nil {
		t.Fatal("expected missing theme file to fail")
	}
	if ProfilePath(home, "PowerShell 7") != "" {
		if _, statErr := os.Stat(ProfilePath(home, "PowerShell 7")); !os.IsNotExist(statErr) {
			t.Fatalf("expected profile not to be written after missing theme failure, stat err: %v", statErr)
		}
	}
}

func TestApplyPromptRejectsUnknownShell(t *testing.T) {
	err := ApplyPrompt(t.TempDir(), "PowerShell Typo", filepath.Join(t.TempDir(), "theme.omp.json"))
	if err == nil || !strings.Contains(err.Error(), `unsupported shell "PowerShell Typo"`) {
		t.Fatalf("expected unsupported shell error, got %v", err)
	}
}

func TestRemovePromptRemovesOnlyManagedBlock(t *testing.T) {
	home := t.TempDir()
	profilePath := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	initial := "Write-Host before\n" + markerStart + "\nold snippet\n" + markerEnd + "\nWrite-Host after\n"
	if err := os.WriteFile(profilePath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemovePrompt(home, "PowerShell 7"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, markerStart) || strings.Contains(text, markerEnd) || strings.Contains(text, "old snippet") {
		t.Fatalf("expected managed block removed, got:\n%s", text)
	}
	if !strings.Contains(text, "Write-Host before") || !strings.Contains(text, "Write-Host after") {
		t.Fatalf("expected unmanaged profile content to remain, got:\n%s", text)
	}
}

func TestRemoveAllPromptsRemovesEverySupportedProfileBlock(t *testing.T) {
	home := t.TempDir()
	themePath := writeTheme(t, home, "amro")

	for _, shellName := range []string{"PowerShell 7", "Windows PowerShell", "Git Bash", "Fish", "Nushell"} {
		if err := ApplyPrompt(home, shellName, themePath); err != nil {
			t.Fatalf("apply %s: %v", shellName, err)
		}
	}

	if err := RemoveAllPrompts(home); err != nil {
		t.Fatal(err)
	}

	for _, shellName := range []string{"PowerShell 7", "Windows PowerShell", "Git Bash", "Fish", "Nushell"} {
		if HasPromptBlock(home, shellName) {
			t.Fatalf("expected %s prompt block to be removed", shellName)
		}
	}
}

func writeTheme(t *testing.T, home, name string) string {
	t.Helper()
	path := filepath.Join(home, "themes", name+".omp.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":3}`), 0o644); err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
