package updater

import (
	"context"
	"os/exec"

	"github.com/muyleanging/termix/internal/app"
)

type Manager struct {
	rt *app.Runtime
}

func New(rt *app.Runtime) Manager { return Manager{rt: rt} }

func (m Manager) Run(ctx context.Context) error {
	commands := [][]string{
		{"winget", "upgrade", "--id", "JanDeDobbeleer.OhMyPosh", "--silent"},
		{"oh-my-posh", "cache", "clear"},
	}
	for _, args := range commands {
		if err := exec.CommandContext(ctx, args[0], args[1:]...).Run(); err != nil {
			return err
		}
	}
	return nil
}
