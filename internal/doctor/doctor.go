package doctor

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/detector"
	"github.com/muyleanging/termix/internal/glyph"
)

type Check struct {
	Name   string
	OK     bool
	Detail string
	Repair string
}

type Report struct {
	Checks []Check
}

type Doctor struct {
	rt *app.Runtime
}

func New(rt *app.Runtime) Doctor { return Doctor{rt: rt} }

func (d Doctor) Run(ctx context.Context) Report {
	env := d.rt.Env
	checks := []Check{
		tool("Oh My Posh", env.OhMyPosh, "Install with: termix install oh-my-posh"),
		{Name: "Current terminal", OK: d.rt.Glyph.Terminal != "", Detail: d.rt.Glyph.Terminal, Repair: "Run inside Terminal, iTerm2, Warp, VS Code, or another modern terminal"},
		{Name: "ANSI support", OK: env.ANSI, Detail: boolDetail(env.ANSI), Repair: "Use Windows Terminal or a modern ANSI terminal"},
		{Name: "Unicode support", OK: env.Unicode && d.rt.Glyph.UTF8, Detail: boolDetail(env.Unicode && d.rt.Glyph.UTF8), Repair: "Enable UTF-8 locale and install a Nerd Font"},
		{Name: "Nerd Font installed", OK: d.rt.Glyph.NerdInstalled, Detail: boolDetail(d.rt.Glyph.NerdInstalled), Repair: "Install with: termix fonts install \"FiraCode Nerd Font\" --yes"},
		{Name: "Nerd Font active", OK: d.rt.Glyph.NerdActive, Detail: glyphDetail(d.rt.Glyph), Repair: d.rt.Glyph.Fix},
		{Name: "ASCII fallback", OK: true, Detail: asciiDetail(d.rt.Glyph), Repair: "Set TERMIX_ASCII_ICONS=1 to force ASCII icons"},
	}
	if runtime.GOOS == "windows" {
		checks = append(checks,
			tool("PowerShell 7", env.PowerShell, "Install with: termix install powershell"),
			tool("Windows Terminal", env.WindowsTerminal, "Install with: termix install terminal"),
			tool("WSL", env.WSL, "Install with: termix install wsl"),
			Check{Name: "Profile manager", OK: env.PowerShell.Installed || env.GitBash.Installed || env.WSL.Installed, Detail: "shell profiles available", Repair: "Install PowerShell 7, Git Bash, or WSL"},
		)
	} else {
		checks = append(checks,
			tool("Zsh", env.Zsh, "Install with: termix install zsh"),
			tool("Bash", env.GitBash, "Install with: termix install bash"),
			Check{Name: "Profile manager", OK: env.Zsh.Installed || env.GitBash.Installed || env.PowerShell.Installed || env.Fish.Installed || env.Nushell.Installed, Detail: "shell profiles available", Repair: "Install zsh, bash, PowerShell 7, fish, or nushell"},
		)
	}
	return Report{Checks: checks}
}

func glyphDetail(capability glyph.Capability) string {
	if capability.NerdActive {
		return "terminal profile appears to use a Nerd Font"
	}
	if capability.Warning != "" {
		return capability.Warning
	}
	return "not detected"
}

func asciiDetail(capability glyph.Capability) string {
	if capability.ASCII {
		return "active"
	}
	return "inactive"
}

func tool(name string, state detector.ToolState, repair string) Check {
	detail := "missing"
	if state.Installed {
		detail = strings.TrimSpace(state.Version)
		if detail == "" {
			detail = state.Path
		}
	}
	return Check{Name: name, OK: state.Installed, Detail: detail, Repair: repair}
}

func boolDetail(ok bool) string {
	if ok {
		return "available"
	}
	return "not detected"
}

func (r Report) RenderText() string {
	var b strings.Builder
	b.WriteString("TERMIX DOCTOR\n\n")
	for _, check := range r.Checks {
		mark := "✗"
		if check.OK {
			mark = "✓"
		}
		fmt.Fprintf(&b, "%s %-20s %s\n", mark, check.Name, check.Detail)
		if !check.OK {
			fmt.Fprintf(&b, "  repair: %s\n", check.Repair)
		}
	}
	return b.String()
}
