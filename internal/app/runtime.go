package app

import (
	"context"

	"github.com/muyleanging/termix/internal/config"
	"github.com/muyleanging/termix/internal/detector"
	"github.com/muyleanging/termix/internal/theme"
)

type Runtime struct {
	Config      config.Config
	Env         detector.Environment
	BootEntries []string
}

func NewRuntime(ctx context.Context, cfg config.Config) (*Runtime, error) {
	_ = theme.EnsureStarterThemes(cfg)
	env := detector.Detect(ctx)
	return &Runtime{
		Config: cfg,
		Env:    env,
		BootEntries: []string{
			"[BOOT] Detecting terminal...",
			"[BOOT] Detecting PowerShell...",
			"[BOOT] Checking Windows Terminal...",
			"[BOOT] Checking Nerd Fonts...",
			"[BOOT] Loading terminal profiles...",
			"[BOOT] Loading Oh My Posh engine...",
			"[BOOT] Building preview cache...",
			"[ OK ] Environment Ready",
		},
	}, nil
}
