package terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type FontApplyResult struct {
	Target  string
	Changed bool
	Detail  string
}

func ApplyFont(home, family string) ([]FontApplyResult, error) {
	face := terminalFace(family)
	var results []FontApplyResult
	var errs []string

	switch runtime.GOOS {
	case "windows":
		if err := SetFont(home, face); err != nil {
			errs = append(errs, "Windows Terminal: "+err.Error())
			results = append(results, FontApplyResult{Target: "Windows Terminal", Detail: "settings.json not updated"})
		} else {
			results = append(results, FontApplyResult{Target: "Windows Terminal", Changed: true, Detail: "set defaults font.face to " + face})
		}
	case "darwin":
		result, err := applyAppleTerminalFont(face)
		results = append(results, result)
		if err != nil {
			errs = append(errs, err.Error())
		}
	default:
		results = append(results, FontApplyResult{Target: "Linux terminal", Detail: "set the font in your terminal app preferences to " + face})
	}

	if result, err := applyVSCodeFont(home, face); err != nil {
		errs = append(errs, err.Error())
		results = append(results, result)
	} else if result.Target != "" {
		results = append(results, result)
	}

	if len(errs) > 0 {
		return results, errors.New(strings.Join(errs, "\n"))
	}
	return results, nil
}

func terminalFace(family string) string {
	switch strings.ToLower(strings.TrimSpace(family)) {
	case "meslolgm nerd font", "meslolgm nf":
		return "MesloLGM Nerd Font Mono"
	default:
		return family
	}
}

func applyAppleTerminalFont(face string) (FontApplyResult, error) {
	script := fmt.Sprintf(`tell application "Terminal"
	set font of settings set "Basic" to "%s"
	set font of default settings to "%s"
	set font of startup settings to "%s"
end tell`, escapeAppleScript(face), escapeAppleScript(face), escapeAppleScript(face))
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	result := FontApplyResult{Target: "Apple Terminal", Detail: "Terminal > Settings > Profiles > Text > Font > " + face + "; then restart Terminal"}
	if err != nil {
		return result, fmt.Errorf("Apple Terminal font was not updated automatically: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	result.Changed = true
	result.Detail = "set Basic, default, and startup profile font to " + face + "; restart Terminal"
	return result, nil
}

func escapeAppleScript(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}

func applyVSCodeFont(home, face string) (FontApplyResult, error) {
	path := vscodeSettingsPath(home)
	result := FontApplyResult{Target: "VS Code", Detail: "settings.json not found; set terminal.integrated.fontFamily to " + face}
	if path == "" {
		return result, nil
	}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return result, fmt.Errorf("VS Code settings could not be read: %w", err)
	}
	settings := map[string]any{}
	if len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			result.Detail = "settings.json has comments or invalid JSON; set terminal.integrated.fontFamily to " + face
			return result, fmt.Errorf("VS Code settings were not updated automatically: %w", err)
		}
	}
	settings["terminal.integrated.fontFamily"] = face
	next, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return result, err
	}
	if len(data) > 0 {
		if err := os.WriteFile(path+".termix."+time.Now().Format("20060102150405")+".bak", data, 0o644); err != nil {
			return result, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return result, err
	}
	if err := os.WriteFile(path, append(next, '\n'), 0o644); err != nil {
		return result, err
	}
	result.Changed = true
	result.Detail = "set terminal.integrated.fontFamily to " + face
	return result, nil
}

func vscodeSettingsPath(home string) string {
	var paths []string
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			paths = append(paths, filepath.Join(appData, "Code", "User", "settings.json"))
		}
		paths = append(paths, filepath.Join(home, "AppData", "Roaming", "Code", "User", "settings.json"))
	case "darwin":
		paths = append(paths, filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json"))
	default:
		paths = append(paths, filepath.Join(home, ".config", "Code", "User", "settings.json"))
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	for _, path := range paths {
		if info, err := os.Stat(filepath.Dir(path)); err == nil && info.IsDir() {
			return path
		}
	}
	return ""
}
