package uninstaller

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/profile"
	"github.com/muyleanging/termix/internal/theme"
)

type Manager struct {
	rt *app.Runtime
}

var scheduleSelfRemoval = scheduleExecutableRemoval
var uninstallExternalDeps = uninstallExternalDependencies

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
	case "profile", "integration", "prompt":
		home, _ := os.UserHomeDir()
		return profile.RemoveAllPrompts(home)
	case "dependency", "dependencies", "deps", "tools", "oh-my-posh":
		return uninstallExternalDeps(ctx)
	case "config":
		return backupAndRemove(filepath.Join(m.rt.Config.HomeDir, "config.yaml"))
	case "app", "binary", "exe", "executable", "self":
		return scheduleSelfRemoval()
	case "all":
		home, _ := os.UserHomeDir()
		if err := profile.RemoveAllPrompts(home); err != nil {
			return err
		}
		if err := os.RemoveAll(m.rt.Config.HomeDir); err != nil {
			return err
		}
		if err := uninstallExternalDeps(ctx); err != nil {
			return err
		}
		return scheduleSelfRemoval()
	default:
		return fmt.Errorf("unknown uninstall component %q", component)
	}
}

func uninstallExternalDependencies(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	if _, err := exec.LookPath("winget"); err != nil {
		return nil
	}
	targets := []string{
		"JanDeDobbeleer.OhMyPosh",
		"DEVCOM.CascadiaCodeNerdFont",
	}
	var failures []string
	for _, id := range targets {
		cmd := exec.CommandContext(ctx, "winget", "uninstall", "--id", id, "--silent", "--accept-source-agreements")
		output, err := cmd.CombinedOutput()
		text := strings.ToLower(string(output))
		if err == nil || strings.Contains(text, "no installed package found") || strings.Contains(text, "no package found") {
			continue
		}
		failures = append(failures, fmt.Sprintf("%s: %v\n%s", id, err, string(output)))
	}
	if len(failures) > 0 {
		return fmt.Errorf("uninstall external dependencies: %s", strings.Join(failures, "\n"))
	}
	return nil
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

func scheduleExecutableRemoval() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err := os.Remove(exe); err != nil && !os.IsNotExist(err) {
			return err
		}
		alias := filepath.Join(filepath.Dir(exe), "termix-tui")
		if err := os.Remove(alias); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	dir := filepath.Dir(exe)
	removeDir := strings.EqualFold(filepath.Base(dir), "Termix")
	script, err := os.CreateTemp("", "termix-uninstall-*.ps1")
	if err != nil {
		return err
	}
	scriptPath := script.Name()
	body := windowsSelfRemoveScript(exe, dir, scriptPath, removeDir)
	if _, err := script.WriteString(body); err != nil {
		_ = script.Close()
		return err
	}
	if err := script.Close(); err != nil {
		return err
	}
	return exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden", "-File", scriptPath).Start()
}

func windowsSelfRemoveScript(exe, dir, scriptPath string, removeDir bool) string {
	removeDirLiteral := "$false"
	if removeDir {
		removeDirLiteral = "$true"
	}
	return fmt.Sprintf(`Start-Sleep -Seconds 2
$exe = %s
$dir = %s
$scriptPath = %s
$removeDir = %s
Remove-Item -LiteralPath $exe -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath (Join-Path $dir 'termix-tui.exe') -Force -ErrorAction SilentlyContinue
Get-ChildItem -LiteralPath $dir -Filter 'termix.exe.bak-*' -ErrorAction SilentlyContinue | Remove-Item -Force -ErrorAction SilentlyContinue
if ($removeDir) {
    Remove-Item -LiteralPath $dir -Recurse -Force -ErrorAction SilentlyContinue
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($userPath) {
        $parts = $userPath -split ';' | Where-Object { $_ -and ($_ -ne $dir) }
        [Environment]::SetEnvironmentVariable('Path', ($parts -join ';'), 'User')
    }
}
Remove-Item -LiteralPath $scriptPath -Force -ErrorAction SilentlyContinue
`, psSingleQuoted(exe), psSingleQuoted(dir), psSingleQuoted(scriptPath), removeDirLiteral)
}

func psSingleQuoted(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
