package detector

import (
	"context"
	"os"
	"os/exec"
	"runtime"
)

type Environment struct {
	OS              string
	Terminal        string
	PowerShell      ToolState
	WindowsTerminal ToolState
	OhMyPosh        ToolState
	GitBash         ToolState
	WSL             ToolState
	Unicode         bool
	ANSI            bool
	GPU             bool
}

type ToolState struct {
	Name      string
	Path      string
	Installed bool
	Version   string
}

func Detect(ctx context.Context) Environment {
	return Environment{
		OS:              runtime.GOOS,
		Terminal:        firstNonEmpty(os.Getenv("WT_SESSION"), os.Getenv("TERM_PROGRAM"), os.Getenv("TERM"), "unknown"),
		PowerShell:      detectTool(ctx, "PowerShell 7", "pwsh", "--version"),
		WindowsTerminal: detectTool(ctx, "Windows Terminal", "wt", "--version"),
		OhMyPosh:        detectTool(ctx, "Oh My Posh", "oh-my-posh", "version"),
		GitBash:         detectTool(ctx, "Git Bash", "bash", "--version"),
		WSL:             detectTool(ctx, "WSL", "wsl", "--status"),
		Unicode:         true,
		ANSI:            true,
		GPU:             os.Getenv("WT_SESSION") != "",
	}
}

func detectTool(ctx context.Context, name, bin, versionArg string) ToolState {
	path, err := exec.LookPath(bin)
	if err != nil {
		return ToolState{Name: name}
	}
	out, _ := exec.CommandContext(ctx, bin, versionArg).CombinedOutput()
	return ToolState{Name: name, Path: path, Installed: true, Version: compact(string(out))}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func compact(s string) string {
	for i, r := range s {
		if r == '\n' || r == '\r' {
			return s[:i]
		}
	}
	return s
}
