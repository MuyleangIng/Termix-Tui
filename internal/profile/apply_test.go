package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepairPromptReplacesManagedBlockWithoutDuplicates(t *testing.T) {
	home := t.TempDir()
	profilePath := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	initial := "Write-Host before\n" + markerStart + "\nold snippet\n" + markerEnd + "\nWrite-Host after\n"
	if err := os.WriteFile(profilePath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RepairPrompt(home, "PowerShell 7", "C:\\Termix\\themes\\catppuccin_mocha.omp.json"); err != nil {
		t.Fatal(err)
	}
	if err := RepairPrompt(home, "PowerShell 7", "C:\\Termix\\themes\\dracula.omp.json"); err != nil {
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
	if strings.Contains(text, "catppuccin_mocha") || !strings.Contains(text, "dracula.omp.json") {
		t.Fatalf("expected old theme path replaced with new theme path, got:\n%s", text)
	}
	if !strings.Contains(text, "Write-Host before") || !strings.Contains(text, "Write-Host after") {
		t.Fatalf("expected unmanaged profile content to remain, got:\n%s", text)
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
