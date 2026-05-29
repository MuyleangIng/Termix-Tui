package app

import (
	"context"
	"os"

	"github.com/muyleanging/termix/internal/config"
	"github.com/muyleanging/termix/internal/detector"
	"github.com/muyleanging/termix/internal/glyph"
	"github.com/muyleanging/termix/internal/theme"
)

type Runtime struct {
	Config      config.Config
	Env         detector.Environment
	Glyph       glyph.Capability
	BootEntries []string
}

func NewRuntime(ctx context.Context, cfg config.Config) (*Runtime, error) {
	_ = theme.EnsureStarterThemes(cfg)
	env := detector.Detect(ctx)
	home, _ := os.UserHomeDir()
	glyphs := glyph.Detect(ctx, cfg, home)
	return &Runtime{
		Config: cfg,
		Env:    env,
		Glyph:  glyphs,
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
