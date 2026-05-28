package uninstaller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/profile"
	"github.com/muyleanging/termix/internal/theme"
)

type Manager struct {
	rt *app.Runtime
}

func New(rt *app.Runtime) Manager { return Manager{rt: rt} }

func (m Manager) Uninstall(ctx context.Context, component string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	switch component {
	case "cache":
		return theme.ClearCache(m.rt.Config)
	case "themes", "downloaded-themes":
		return os.RemoveAll(filepath.Join(m.rt.Config.HomeDir, "themes"))
	case "profile", "integration", "prompt", "oh-my-posh":
		home, _ := os.UserHomeDir()
		return profile.RemoveAllPrompts(home)
	case "config":
		return backupAndRemove(filepath.Join(m.rt.Config.HomeDir, "config.yaml"))
	case "all":
		if err := theme.ClearCache(m.rt.Config); err != nil {
			return err
		}
		home, _ := os.UserHomeDir()
		if err := profile.RemoveAllPrompts(home); err != nil {
			return err
		}
		return backupAndRemove(filepath.Join(m.rt.Config.HomeDir, "config.yaml"))
	default:
		return fmt.Errorf("unknown uninstall component %q", component)
	}
}

func backupAndRemove(path string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	backup := path + ".termix.bak"
	if err := os.WriteFile(backup, data, 0o644); err != nil {
		return err
	}
	return os.Remove(path)
}
