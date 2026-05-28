package uninstaller

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/config"
)

func TestUninstallAllRemovesTermixHomeAndSchedulesExecutableRemoval(t *testing.T) {
	termixHome := filepath.Join(t.TempDir(), ".termix")
	if err := os.MkdirAll(filepath.Join(termixHome, "themes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(termixHome, "config.yaml"), []byte("default_shell: PowerShell 7\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(termixHome, "themes", "amro.omp.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	called := false
	original := scheduleSelfRemoval
	scheduleSelfRemoval = func() error {
		called = true
		return nil
	}
	defer func() { scheduleSelfRemoval = original }()

	rt := &app.Runtime{Config: config.Config{HomeDir: termixHome, DefaultShell: "PowerShell 7"}}
	if err := New(rt).Uninstall(context.Background(), "all"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected executable removal to be scheduled")
	}
	if _, err := os.Stat(termixHome); !os.IsNotExist(err) {
		t.Fatalf("expected Termix home to be removed, stat err: %v", err)
	}
}
