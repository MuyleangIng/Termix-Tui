package terminal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type WindowsTerminalSettings struct {
	Profiles struct {
		Defaults map[string]any   `json:"defaults,omitempty"`
		List     []map[string]any `json:"list,omitempty"`
	} `json:"profiles"`
}

func SettingsPath(home string) string {
	return filepath.Join(home, "AppData", "Local", "Packages", "Microsoft.WindowsTerminal_8wekyb3d8bbwe", "LocalState", "settings.json")
}

func SetFont(home, family string) error {
	path, err := findSettingsPath(home)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var settings WindowsTerminalSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}
	if settings.Profiles.Defaults == nil {
		settings.Profiles.Defaults = map[string]any{}
	}
	settings.Profiles.Defaults["font"] = map[string]any{"face": family}
	next, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(backupPath(path), data, 0o644); err != nil {
		return err
	}
	return os.WriteFile(path, next, 0o644)
}

func findSettingsPath(home string) (string, error) {
	paths := []string{
		SettingsPath(home),
		filepath.Join(home, "AppData", "Local", "Microsoft", "Windows Terminal", "settings.json"),
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return paths[0], os.ErrNotExist
}

func backupPath(path string) string {
	return path + ".termix." + time.Now().Format("20060102150405") + ".bak"
}
