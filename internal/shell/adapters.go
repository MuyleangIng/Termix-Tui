package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Adapter interface {
	Name() string
	ProfilePath(home string) string
	InitSnippet(themePath string) string
}

type PowerShell struct{}
type WindowsPowerShell struct{}
type Bash struct{ Label string }
type Fish struct{}
type Nushell struct{}

func (PowerShell) Name() string { return "PowerShell 7" }
func (PowerShell) ProfilePath(home string) string {
	return filepath.Join(documentsDir(home), "PowerShell", "Microsoft.PowerShell_profile.ps1")
}
func (PowerShell) InitSnippet(themePath string) string {
	return fmt.Sprintf("%s init pwsh --config %s | Invoke-Expression", powerShellCommand(ohMyPoshCommand()), quotePowerShellPath(themePath))
}

func (WindowsPowerShell) Name() string { return "Windows PowerShell" }
func (WindowsPowerShell) ProfilePath(home string) string {
	return filepath.Join(documentsDir(home), "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
}
func (WindowsPowerShell) InitSnippet(themePath string) string {
	return fmt.Sprintf("%s init pwsh --config %s | Invoke-Expression", powerShellCommand(ohMyPoshCommand()), quotePowerShellPath(themePath))
}

func (b Bash) Name() string {
	if b.Label != "" {
		return b.Label
	}
	return "Git Bash"
}
func (Bash) ProfilePath(home string) string { return filepath.Join(home, ".bashrc") }
func (Bash) InitSnippet(themePath string) string {
	return fmt.Sprintf("eval \"$(%s init bash --config %q)\"", shellCommand(ohMyPoshCommand()), themePath)
}

func (Fish) Name() string { return "Fish" }
func (Fish) ProfilePath(home string) string {
	return filepath.Join(home, ".config", "fish", "config.fish")
}
func (Fish) InitSnippet(themePath string) string {
	return fmt.Sprintf("%s init fish --config %q | source", shellCommand(ohMyPoshCommand()), themePath)
}

func (Nushell) Name() string { return "Nushell" }
func (Nushell) ProfilePath(home string) string {
	return filepath.Join(home, "AppData", "Roaming", "nushell", "config.nu")
}
func (Nushell) InitSnippet(themePath string) string {
	return fmt.Sprintf("%s init nu --config %q", shellCommand(ohMyPoshCommand()), themePath)
}

func Supported() []Adapter {
	return []Adapter{PowerShell{}, WindowsPowerShell{}, Bash{Label: "Git Bash"}, Bash{Label: "WSL Ubuntu"}, Fish{}, Nushell{}}
}

func quotePowerShellPath(path string) string {
	path = strings.ReplaceAll(path, "`", "``")
	path = strings.ReplaceAll(path, `"`, "`\"")
	return `"` + path + `"`
}

func ohMyPoshCommand() string {
	if path, err := exec.LookPath("oh-my-posh"); err == nil {
		return path
	}
	if runtime.GOOS != "windows" {
		return "oh-my-posh"
	}
	exe := "oh-my-posh.exe"
	candidates := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WindowsApps", exe),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "oh-my-posh", exe),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "oh-my-posh", "bin", exe),
		filepath.Join(os.Getenv("ProgramFiles"), "oh-my-posh", exe),
		filepath.Join(os.Getenv("ProgramFiles"), "oh-my-posh", "bin", exe),
	}
	for _, candidate := range candidates {
		if candidate == "" || !filepath.IsAbs(candidate) {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return "oh-my-posh"
}

func powerShellCommand(path string) string {
	if strings.ContainsAny(path, ` \/`) {
		return "& " + quotePowerShellPath(path)
	}
	return path
}

func shellCommand(path string) string {
	if strings.ContainsAny(path, ` \`) {
		return fmt.Sprintf("%q", path)
	}
	return path
}
