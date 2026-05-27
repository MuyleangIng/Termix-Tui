package tui

import "github.com/charmbracelet/lipgloss"

type palette struct {
	cyan       lipgloss.Color
	green      lipgloss.Color
	yellow     lipgloss.Color
	red        lipgloss.Color
	muted      lipgloss.Color
	ink        lipgloss.Color
	inkInverse lipgloss.Color
	bg         lipgloss.Color
	surface    lipgloss.Color
	elevated   lipgloss.Color
	border     lipgloss.Color
	borderSoft lipgloss.Color
}

var theme = darkPalette()
var activeBorderStyle = "unicode"
var activeLightMode bool

var (
	cyan    = theme.cyan
	blue    = theme.cyan
	purple  = theme.cyan
	magenta = theme.cyan
	green   = theme.green
	yellow  = theme.yellow
	red     = theme.red
	muted   = theme.muted
	ink     = theme.ink
	bg      = theme.bg
	bgSoft  = theme.surface

	appShell = lipgloss.NewStyle()
	header   = lipgloss.NewStyle()
	footer   = lipgloss.NewStyle()
	sidebar  = lipgloss.NewStyle()
	card     = lipgloss.NewStyle()
	cardHot  = lipgloss.NewStyle()
	cardWarn = lipgloss.NewStyle()
	cardBad  = lipgloss.NewStyle()
	logCard  = lipgloss.NewStyle()

	title  = lipgloss.NewStyle()
	accent = lipgloss.NewStyle()
	label  = lipgloss.NewStyle()
	ok     = lipgloss.NewStyle()
	warn   = lipgloss.NewStyle()
	bad    = lipgloss.NewStyle()
	pill   = lipgloss.NewStyle()
)

func init() {
	applyColorMode(false)
}

func applyColorMode(light bool) {
	activeLightMode = light
	if light {
		theme = lightPalette()
	} else {
		theme = darkPalette()
	}

	cyan = theme.cyan
	blue = theme.cyan
	purple = theme.cyan
	magenta = theme.cyan
	green = theme.green
	yellow = theme.yellow
	red = theme.red
	muted = theme.muted
	ink = theme.ink
	bg = theme.bg
	bgSoft = theme.surface

	appShell = lipgloss.NewStyle().Foreground(theme.ink).Background(theme.bg)
	header = lipgloss.NewStyle().Foreground(theme.ink).Background(theme.bg).Padding(0, 1)
	footer = lipgloss.NewStyle().Foreground(theme.muted).Background(theme.bg).Padding(0, 1)
	sidebar = basePanel().BorderForeground(theme.border)
	card = basePanel()
	cardHot = card.BorderForeground(theme.cyan)
	cardWarn = card.BorderForeground(theme.yellow)
	cardBad = card.BorderForeground(theme.red)
	logCard = card.BorderForeground(theme.border)

	title = lipgloss.NewStyle().Foreground(theme.cyan).Bold(true)
	accent = lipgloss.NewStyle().Foreground(theme.cyan).Bold(true)
	label = lipgloss.NewStyle().Foreground(theme.muted)
	ok = lipgloss.NewStyle().Foreground(theme.green).Bold(true)
	warn = lipgloss.NewStyle().Foreground(theme.yellow).Bold(true)
	bad = lipgloss.NewStyle().Foreground(theme.red).Bold(true)
	pill = lipgloss.NewStyle().
		Padding(0, 1).
		Background(theme.elevated).
		Foreground(theme.cyan)
}

func applyBorderMode(mode string) {
	switch mode {
	case "ascii", "none":
		activeBorderStyle = mode
	default:
		activeBorderStyle = "unicode"
	}
	applyColorMode(activeLightMode)
}

func darkPalette() palette {
	return palette{
		cyan:       lipgloss.Color("#06B6D4"),
		green:      lipgloss.Color("#22C55E"),
		yellow:     lipgloss.Color("#EAB308"),
		red:        lipgloss.Color("#EF4444"),
		muted:      lipgloss.Color("#94A3B8"),
		ink:        lipgloss.Color("#F8FAFC"),
		inkInverse: lipgloss.Color("#031A20"),
		bg:         lipgloss.Color("#09090B"),
		surface:    lipgloss.Color("#111113"),
		elevated:   lipgloss.Color("#18181B"),
		border:     lipgloss.Color("#27272A"),
		borderSoft: lipgloss.Color("#1F2937"),
	}
}

func lightPalette() palette {
	return palette{
		cyan:       lipgloss.Color("#0891B2"),
		green:      lipgloss.Color("#15803D"),
		yellow:     lipgloss.Color("#A16207"),
		red:        lipgloss.Color("#BE123C"),
		muted:      lipgloss.Color("#5F6368"),
		ink:        lipgloss.Color("#0A0A0A"),
		inkInverse: lipgloss.Color("#ECFEFF"),
		bg:         lipgloss.Color("#FFFFFF"),
		surface:    lipgloss.Color("#F8FAFC"),
		elevated:   lipgloss.Color("#ECFEFF"),
		border:     lipgloss.Color("#E5E7EB"),
		borderSoft: lipgloss.Color("#CBD5E1"),
	}
}

func basePanel() lipgloss.Style {
	style := lipgloss.NewStyle().
		Background(theme.surface).
		Foreground(theme.ink).
		Padding(0, 1)
	if activeBorderStyle == "none" {
		return style
	}
	return style.
		Border(borderForMode()).
		BorderForeground(theme.border)
}

func focusedPanel(focused bool) lipgloss.Style {
	if focused {
		return basePanel().BorderForeground(theme.cyan)
	}
	return basePanel()
}

func menuRow(selected bool, width int) lipgloss.Style {
	style := lipgloss.NewStyle().Padding(0, 1).Width(max(1, width))
	if selected {
		return style.Foreground(theme.inkInverse).Background(theme.cyan).Bold(true)
	}
	return style.Foreground(theme.ink).Background(theme.surface)
}

func hintStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.muted).Background(theme.surface)
}

func badgeStyle(kind string) lipgloss.Style {
	style := lipgloss.NewStyle().Padding(0, 1).Bold(true)
	switch kind {
	case "SUCCESS":
		return style.Foreground(theme.bg).Background(theme.green)
	case "WARN":
		return style.Foreground(theme.bg).Background(theme.yellow)
	case "ERROR":
		return style.Foreground(theme.inkInverse).Background(theme.red)
	default:
		return style.Foreground(theme.inkInverse).Background(theme.cyan)
	}
}

func borderForMode() lipgloss.Border {
	if activeBorderStyle == "ascii" {
		return lipgloss.Border{
			Top:         "-",
			Bottom:      "-",
			Left:        "|",
			Right:       "|",
			TopLeft:     "+",
			TopRight:    "+",
			BottomLeft:  "+",
			BottomRight: "+",
		}
	}
	return lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}
}
