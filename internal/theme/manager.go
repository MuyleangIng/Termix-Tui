package theme

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/muyleanging/termix/internal/config"
)

type Theme struct {
	Name          string
	Path          string
	Category      string
	Segments      int
	Shells        []string
	Favorite      bool
	Compatibility string
	Source        string
	LastModified  time.Time
	Valid         bool
}

type Manager struct {
	cfg config.Config
}

func NewManager(cfg config.Config) Manager {
	return Manager{cfg: cfg}
}

func (m Manager) Scan(ctx context.Context) ([]Theme, error) {
	if err := EnsureAvailable(ctx, m.cfg); err != nil {
		return nil, err
	}
	return m.scan(ctx)
}

func (m Manager) ScanInstalled(ctx context.Context) ([]Theme, error) {
	return m.scan(ctx)
}

func (m Manager) scan(ctx context.Context) ([]Theme, error) {
	favorites := map[string]bool{}
	for _, f := range m.cfg.FavoriteThemes {
		favorites[f] = true
	}

	var themes []Theme
	for _, dir := range m.cfg.ThemeDirs {
		err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".omp.json") {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			item := inspectTheme(path)
			item.Favorite = favorites[item.Name]
			themes = append(themes, item)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	sort.Slice(themes, func(i, j int) bool {
		if themes[i].Favorite != themes[j].Favorite {
			return themes[i].Favorite
		}
		return themes[i].Name < themes[j].Name
	})
	return themes, WriteCache(m.cfg, themes)
}

func inspectTheme(path string) Theme {
	name := strings.TrimSuffix(filepath.Base(path), ".omp.json")
	item := Theme{Name: name, Path: path, Category: categorize(name), Shells: []string{"pwsh", "bash", "zsh", "fish"}, Compatibility: "universal", Source: "official-oh-my-posh", Valid: true}
	if info, err := os.Stat(path); err == nil {
		item.LastModified = info.ModTime()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return item
	}
	var raw struct {
		Blocks []struct {
			Segments []json.RawMessage `json:"segments"`
		} `json:"blocks"`
	}
	if json.Unmarshal(data, &raw) == nil {
		for _, block := range raw.Blocks {
			item.Segments += len(block.Segments)
		}
	}
	return item
}

func EnsureAvailable(ctx context.Context, cfg config.Config) error {
	if hasThemeFiles(cfg) {
		return nil
	}
	_, err := InstallOfficialThemes(ctx, cfg)
	return err
}

func RebuildCache(ctx context.Context, cfg config.Config) ([]Theme, error) {
	if err := EnsureAvailable(ctx, cfg); err != nil {
		return nil, err
	}
	return Manager{cfg: cfg}.scan(ctx)
}

func ReadCache(cfg config.Config) ([]Theme, error) {
	data, err := os.ReadFile(filepath.Join(cfg.HomeDir, "cache", "themes.json"))
	if err != nil {
		return nil, err
	}
	var themes []Theme
	if err := json.Unmarshal(data, &themes); err != nil {
		return nil, err
	}
	return themes, nil
}

func ClearCache(cfg config.Config) error {
	return os.RemoveAll(filepath.Join(cfg.HomeDir, "cache"))
}

func WriteCache(cfg config.Config, themes []Theme) error {
	cacheDir := filepath.Join(cfg.HomeDir, "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(themes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cacheDir, "themes.json"), data, 0o644)
}

func hasThemeFiles(cfg config.Config) bool {
	for _, dir := range cfg.ThemeDirs {
		matches, _ := filepath.Glob(filepath.Join(dir, "*.omp.json"))
		if len(matches) > 0 {
			return true
		}
	}
	return false
}

func categorize(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "catppuccin"), strings.Contains(n, "dracula"), strings.Contains(n, "tokyo"):
		return "premium"
	case strings.Contains(n, "powerline"), strings.Contains(n, "atomic"), strings.Contains(n, "paradox"):
		return "powerline"
	default:
		return "community"
	}
}
