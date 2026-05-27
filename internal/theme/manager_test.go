package theme

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/muyleanging/termix/internal/config"
)

func TestScanThemesExtractsMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catppuccin_mocha.omp.json")
	data := `{"blocks":[{"segments":[{"type":"session"},{"type":"path"}]},{"segments":[{"type":"git"}]}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	manager := NewManager(config.Config{
		ThemeDirs:      []string{dir},
		FavoriteThemes: []string{"catppuccin_mocha"},
	})

	got, err := manager.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 theme, got %d", len(got))
	}
	if got[0].Name != "catppuccin_mocha" || got[0].Segments != 3 || !got[0].Favorite {
		t.Fatalf("unexpected theme metadata: %#v", got[0])
	}
}

func TestClearCacheKeepsDownloadedThemes(t *testing.T) {
	home := t.TempDir()
	themeDir := filepath.Join(home, "themes")
	cacheDir := filepath.Join(home, "cache")
	themePath := filepath.Join(themeDir, "dracula.omp.json")
	cachePath := filepath.Join(cacheDir, "themes.json")
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(themePath, []byte(`{"blocks":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachePath, []byte(`[]`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ClearCache(config.Config{HomeDir: home}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(themePath); err != nil {
		t.Fatalf("theme file should remain after cache clear: %v", err)
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("cache file should be removed, got %v", err)
	}
}
