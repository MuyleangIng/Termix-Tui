package shell

import (
	"fmt"
	"path/filepath"
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
	return fmt.Sprintf("oh-my-posh init pwsh --config %s | Invoke-Expression", quotePowerShellPath(themePath))
}

func (WindowsPowerShell) Name() string { return "Windows PowerShell" }
func (WindowsPowerShell) ProfilePath(home string) string {
	return filepath.Join(documentsDir(home), "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
}
func (WindowsPowerShell) InitSnippet(themePath string) string {
	return fmt.Sprintf("oh-my-posh init pwsh --config %s | Invoke-Expression", quotePowerShellPath(themePath))
}

func (b Bash) Name() string {
	if b.Label != "" {
		return b.Label
	}
	return "Git Bash"
}
func (Bash) ProfilePath(home string) string { return filepath.Join(home, ".bashrc") }
func (Bash) InitSnippet(themePath string) string {
	return fmt.Sprintf("eval \"$(oh-my-posh init bash --config %q)\"", themePath)
}

func (Fish) Name() string { return "Fish" }
func (Fish) ProfilePath(home string) string {
	return filepath.Join(home, ".config", "fish", "config.fish")
}
func (Fish) InitSnippet(themePath string) string {
	return fmt.Sprintf("oh-my-posh init fish --config %q | source", themePath)
}

func (Nushell) Name() string { return "Nushell" }
func (Nushell) ProfilePath(home string) string {
	return filepath.Join(home, "AppData", "Roaming", "nushell", "config.nu")
}
func (Nushell) InitSnippet(themePath string) string {
	return fmt.Sprintf("oh-my-posh init nu --config %q", themePath)
}

func Supported() []Adapter {
	return []Adapter{PowerShell{}, WindowsPowerShell{}, Bash{Label: "Git Bash"}, Bash{Label: "WSL Ubuntu"}, Fish{}, Nushell{}}
}

func quotePowerShellPath(path string) string {
	path = strings.ReplaceAll(path, "`", "``")
	path = strings.ReplaceAll(path, `"`, "`\"")
	return `"` + path + `"`
}
