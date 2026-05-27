package doctor

import (
	"context"
	"fmt"
	"strings"

	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/detector"
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
	return Report{Checks: []Check{
		tool("Oh My Posh", env.OhMyPosh, "Install with: termix install oh-my-posh"),
		tool("PowerShell 7", env.PowerShell, "Install with: termix install powershell"),
		tool("Windows Terminal", env.WindowsTerminal, "Install with: termix install terminal"),
		tool("WSL", env.WSL, "Install with: termix install wsl"),
		{Name: "ANSI support", OK: env.ANSI, Detail: boolDetail(env.ANSI), Repair: "Use Windows Terminal or a modern ANSI terminal"},
		{Name: "Unicode support", OK: env.Unicode, Detail: boolDetail(env.Unicode), Repair: "Enable UTF-8 locale and install a Nerd Font"},
		{Name: "Profile manager", OK: env.PowerShell.Installed || env.GitBash.Installed || env.WSL.Installed, Detail: "shell profiles available", Repair: "Install PowerShell 7, Git Bash, or WSL"},
	}}
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
