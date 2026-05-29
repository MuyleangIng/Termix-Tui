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
type Zsh struct{}
type Fish struct{}
type Nushell struct{}

func (PowerShell) Name() string { return "PowerShell 7" }
func (PowerShell) ProfilePath(home string) string {
	if runtime.GOOS != "windows" {
		return filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	}
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

func (Zsh) Name() string                   { return "Zsh" }
func (Zsh) ProfilePath(home string) string { return filepath.Join(home, ".zshrc") }
func (Zsh) InitSnippet(themePath string) string {
	return fmt.Sprintf("eval \"$(%s init zsh --config %q)\"", shellCommand(ohMyPoshCommand()), themePath)
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
	if runtime.GOOS != "windows" {
		return filepath.Join(home, ".config", "nushell", "config.nu")
	}
	return filepath.Join(home, "AppData", "Roaming", "nushell", "config.nu")
}
func (Nushell) InitSnippet(themePath string) string {
	return fmt.Sprintf("%s init nu --config %q", shellCommand(ohMyPoshCommand()), themePath)
}

func Supported() []Adapter {
	return SupportedForOS(runtime.GOOS)
}

func SupportedForOS(goos string) []Adapter {
	if goos == "windows" {
		return []Adapter{PowerShell{}, WindowsPowerShell{}, Bash{Label: "Git Bash"}, Bash{Label: "WSL Bash"}, Fish{}, Nushell{}}
	}
	return []Adapter{Zsh{}, Bash{Label: "Bash"}, PowerShell{}, Fish{}, Nushell{}}
}

func Available(home string) []Adapter {
	all := Supported()
	out := make([]Adapter, 0, len(all))
	for _, adapter := range all {
		if profileExists(adapter, home) || commandAvailable(adapter) || currentShellMatches(adapter) {
			out = append(out, adapter)
		}
	}
	if len(out) == 0 {
		for _, adapter := range all {
			switch adapter.(type) {
			case Zsh, Bash:
				out = append(out, adapter)
			}
		}
	}
	return out
}

func profileExists(adapter Adapter, home string) bool {
	if _, err := os.Stat(adapter.ProfilePath(home)); err == nil {
		return true
	}
	return false
}

func commandAvailable(adapter Adapter) bool {
	switch adapter.Name() {
	case "PowerShell 7":
		return commandExists("pwsh")
	case "Windows PowerShell":
		return commandExists("powershell")
	case "Git Bash", "WSL Bash", "Bash":
		return commandExists("bash")
	case "Zsh":
		return commandExists("zsh")
	case "Fish":
		return commandExists("fish")
	case "Nushell":
		return commandExists("nu")
	default:
		return false
	}
}

func currentShellMatches(adapter Adapter) bool {
	shellPath := strings.ToLower(os.Getenv("SHELL"))
	if shellPath == "" {
		return false
	}
	switch adapter.Name() {
	case "Bash", "Git Bash", "WSL Bash":
		return strings.HasSuffix(shellPath, "bash")
	case "Zsh":
		return strings.HasSuffix(shellPath, "zsh")
	case "Fish":
		return strings.HasSuffix(shellPath, "fish")
	default:
		return false
	}
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
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
