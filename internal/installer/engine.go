package installer

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/theme"
)

type Engine struct {
	rt *app.Runtime
}

func New(rt *app.Runtime) Engine {
	return Engine{rt: rt}
}

func (e Engine) Install(ctx context.Context, component string) error {
	plan := e.plan(component)
	for _, step := range plan {
		if err := step(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (e Engine) plan(component string) []func(context.Context) error {
	if runtime.GOOS != "windows" {
		return []func(context.Context) error{func(context.Context) error {
			return fmt.Errorf("automated installer is currently implemented for Windows; detected %s", runtime.GOOS)
		}}
	}

	winget := func(id string) func(context.Context) error {
		return func(ctx context.Context) error {
			cmd := exec.CommandContext(ctx, "winget", "install", "--id", id, "--silent", "--accept-package-agreements", "--accept-source-agreements")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("winget install %s failed: %w\n%s", id, err, string(output))
			}
			return nil
		}
	}

	requireTool := func(name string, install func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error {
			if _, err := exec.LookPath(name); err == nil {
				return nil
			}
			return install(ctx)
		}
	}

	switch component {
	case "oh-my-posh":
		return []func(context.Context) error{requireTool("oh-my-posh", winget("JanDeDobbeleer.OhMyPosh"))}
	case "font", "fonts", "nerd-font", "nerd-fonts":
		return []func(context.Context) error{
			requireTool("oh-my-posh", winget("JanDeDobbeleer.OhMyPosh")),
			func(ctx context.Context) error {
				cmd := exec.CommandContext(ctx, "oh-my-posh", "font", "install", "CascadiaCode", "--headless", "--plain")
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("oh-my-posh font install CascadiaCode failed: %w\n%s", err, string(output))
				}
				return nil
			},
		}
	case "font:CaskaydiaCove Nerd Font", "font:CascadiaCode Nerd Font", "font:Cascadia Code Nerd Font":
		return []func(context.Context) error{winget("DEVCOM.CascadiaCodeNerdFont")}
	case "font:JetBrainsMono Nerd Font":
		return []func(context.Context) error{winget("DEVCOM.JetBrainsMonoNerdFont")}
	case "font:FiraCode Nerd Font":
		return []func(context.Context) error{winget("DEVCOM.FiraCodeNerdFont")}
	case "font:Hack Nerd Font":
		return []func(context.Context) error{winget("DEVCOM.HackNerdFont")}
	case "theme", "themes", "all-themes":
		return []func(context.Context) error{func(ctx context.Context) error {
			_, err := theme.InstallOfficialThemes(ctx, e.rt.Config)
			return err
		}}
	case "powershell":
		return []func(context.Context) error{requireTool("pwsh", winget("Microsoft.PowerShell"))}
	case "terminal":
		return []func(context.Context) error{requireTool("wt", winget("Microsoft.WindowsTerminal"))}
	case "git":
		return []func(context.Context) error{requireTool("git", winget("Git.Git"))}
	case "wsl":
		return []func(context.Context) error{requireTool("wsl", func(ctx context.Context) error {
			cmd := exec.CommandContext(ctx, "wsl", "--install")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("wsl install failed: %w\n%s", err, string(output))
			}
			return nil
		})}
	default:
		return []func(context.Context) error{
			requireTool("oh-my-posh", winget("JanDeDobbeleer.OhMyPosh")),
			requireTool("pwsh", winget("Microsoft.PowerShell")),
			requireTool("wt", winget("Microsoft.WindowsTerminal")),
			requireTool("git", winget("Git.Git")),
			func(ctx context.Context) error {
				cmd := exec.CommandContext(ctx, "oh-my-posh", "font", "install", "CascadiaCode", "--headless", "--plain")
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("oh-my-posh font install CascadiaCode failed: %w\n%s", err, string(output))
				}
				return nil
			},
			func(ctx context.Context) error {
				_, err := theme.InstallOfficialThemes(ctx, e.rt.Config)
				return err
			},
		}
	}
}
