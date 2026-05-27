package ansi

import "github.com/muesli/termenv"

type Renderer struct{}

func (Renderer) Render(input string) string {
	profile := termenv.ColorProfile()
	return termenv.String(input).Foreground(profile.Color("#7dd3fc")).String()
}
