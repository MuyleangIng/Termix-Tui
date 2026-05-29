package preview

import (
	"bytes"
	"context"
	"errors"
	"os/exec"

	"github.com/muyleanging/termix/internal/ansi"
	"github.com/muyleanging/termix/internal/theme"
	"github.com/muyleanging/termix/internal/toolpath"
)

type Engine struct {
	Renderer ansi.Renderer
}

func New() Engine {
	return Engine{Renderer: ansi.Renderer{}}
}

func (e Engine) Render(ctx context.Context, item theme.Theme) (string, error) {
	if item.Path == "" {
		return "", errors.New("theme path is empty")
	}
	omp, err := toolpath.Resolve("oh-my-posh")
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, omp, "print", "primary", "--config", item.Path)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return e.Renderer.Render(stdout.String()), nil
}
