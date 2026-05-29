package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	switch runtime.GOOS {
	case "windows":
		return e.windowsPlan(component)
	case "darwin":
		return e.unixPlan(component, macPackageManager{})
	case "linux":
		return e.unixPlan(component, linuxPackageManager{})
	default:
		return []func(context.Context) error{func(context.Context) error {
			return fmt.Errorf("automated installer is not implemented for %s", runtime.GOOS)
		}}
	}
}

func (e Engine) windowsPlan(component string) []func(context.Context) error {
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
			if _, err := resolveToolPath(name); err == nil {
				return nil
			}
			return install(ctx)
		}
	}

	installCascadiaCode := func(ctx context.Context) error {
		omp, err := resolveToolPath("oh-my-posh")
		if err != nil {
			return fmt.Errorf("oh-my-posh was installed but is not available in PATH yet; open a new terminal or run termix install again: %w", err)
		}
		cmd := exec.CommandContext(ctx, omp, "font", "install", "CascadiaCode", "--headless", "--plain")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("oh-my-posh font install CascadiaCode failed: %w\n%s", err, string(output))
		}
		return nil
	}

	switch component {
	case "oh-my-posh":
		return []func(context.Context) error{requireTool("oh-my-posh", winget("JanDeDobbeleer.OhMyPosh"))}
	case "font", "fonts", "nerd-font", "nerd-fonts":
		return []func(context.Context) error{
			requireTool("oh-my-posh", winget("JanDeDobbeleer.OhMyPosh")),
			installCascadiaCode,
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
			installCascadiaCode,
			func(ctx context.Context) error {
				_, err := theme.InstallOfficialThemes(ctx, e.rt.Config)
				return err
			},
		}
	}
}

type packageManager interface {
	Install(ctx context.Context, pkg string) error
	OhMyPosh(ctx context.Context) error
}

type macPackageManager struct{}

func (macPackageManager) Install(ctx context.Context, pkg string) error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew is required to install %s automatically on macOS; install Homebrew first or install %s manually", pkg, pkg)
	}
	return run(ctx, "brew", "install", pkg)
}

func (m macPackageManager) OhMyPosh(ctx context.Context) error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew is required to install Oh My Posh automatically on macOS; install Homebrew first")
	}
	cmd := exec.CommandContext(ctx, "brew", "install", "jandedobbeleer/oh-my-posh/oh-my-posh")
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	fallback := exec.CommandContext(ctx, "brew", "install", "oh-my-posh")
	fallbackOutput, fallbackErr := fallback.CombinedOutput()
	if fallbackErr != nil {
		return fmt.Errorf("brew install oh-my-posh failed: %w\n%s\nfallback failed: %w\n%s", err, string(output), fallbackErr, string(fallbackOutput))
	}
	return nil
}

type linuxPackageManager struct{}

func (linuxPackageManager) Install(ctx context.Context, pkg string) error {
	switch {
	case commandExists("apt-get"):
		return run(ctx, "sudo", "apt-get", "install", "-y", pkg)
	case commandExists("dnf"):
		return run(ctx, "sudo", "dnf", "install", "-y", pkg)
	case commandExists("yum"):
		return run(ctx, "sudo", "yum", "install", "-y", pkg)
	case commandExists("pacman"):
		return run(ctx, "sudo", "pacman", "-S", "--noconfirm", pkg)
	case commandExists("zypper"):
		return run(ctx, "sudo", "zypper", "install", "-y", pkg)
	default:
		return fmt.Errorf("no supported Linux package manager found for %s; supported: apt-get, dnf, yum, pacman, zypper", pkg)
	}
}

func (linuxPackageManager) OhMyPosh(ctx context.Context) error {
	if commandExists("curl") {
		return runShell(ctx, "curl -s https://ohmyposh.dev/install.sh | bash -s")
	}
	if commandExists("wget") {
		return runShell(ctx, "wget -qO- https://ohmyposh.dev/install.sh | bash -s")
	}
	return fmt.Errorf("curl or wget is required to install Oh My Posh automatically on Linux")
}

func (e Engine) unixPlan(component string, pm packageManager) []func(context.Context) error {
	requireTool := func(name string, install func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error {
			if _, err := resolveToolPath(name); err == nil {
				return nil
			}
			return install(ctx)
		}
	}
	installCascadiaCode := func(ctx context.Context) error {
		omp, err := resolveToolPath("oh-my-posh")
		if err != nil {
			return fmt.Errorf("oh-my-posh was installed but is not available in PATH yet; open a new terminal or run termix install again: %w", err)
		}
		return run(ctx, omp, "font", "install", "CascadiaCode", "--headless", "--plain")
	}
	installThemes := func(ctx context.Context) error {
		_, err := theme.InstallOfficialThemes(ctx, e.rt.Config)
		return err
	}

	switch component {
	case "oh-my-posh":
		return []func(context.Context) error{requireTool("oh-my-posh", pm.OhMyPosh)}
	case "zsh":
		return []func(context.Context) error{requireTool("zsh", func(ctx context.Context) error { return pm.Install(ctx, "zsh") })}
	case "bash":
		return []func(context.Context) error{requireTool("bash", func(ctx context.Context) error { return pm.Install(ctx, "bash") })}
	case "fish":
		return []func(context.Context) error{requireTool("fish", func(ctx context.Context) error { return pm.Install(ctx, "fish") })}
	case "nushell":
		return []func(context.Context) error{requireTool("nu", func(ctx context.Context) error { return pm.Install(ctx, "nushell") })}
	case "powershell":
		return []func(context.Context) error{requireTool("pwsh", func(ctx context.Context) error { return pm.Install(ctx, "powershell") })}
	case "font", "fonts", "nerd-font", "nerd-fonts":
		return []func(context.Context) error{requireTool("oh-my-posh", pm.OhMyPosh), installCascadiaCode}
	case "theme", "themes", "all-themes":
		return []func(context.Context) error{installThemes}
	default:
		return []func(context.Context) error{
			requireTool("oh-my-posh", pm.OhMyPosh),
			requireTool("zsh", func(ctx context.Context) error { return pm.Install(ctx, "zsh") }),
			installCascadiaCode,
			installThemes,
		}
	}
}

func resolveToolPath(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("%s not found", name)
	}
	exe := name
	if filepath.Ext(exe) == "" {
		exe += ".exe"
	}
	candidates := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WindowsApps", exe),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", name, exe),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", name, "bin", exe),
		filepath.Join(os.Getenv("ProgramFiles"), name, exe),
		filepath.Join(os.Getenv("ProgramFiles"), name, "bin", exe),
	}
	for _, candidate := range candidates {
		if candidate == "" || !filepath.IsAbs(candidate) {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%s not found", name)
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v failed: %w\n%s", name, args, err, string(output))
	}
	return nil
}

func runShell(ctx context.Context, script string) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("shell install command failed: %w\n%s", err, string(output))
	}
	return nil
}
