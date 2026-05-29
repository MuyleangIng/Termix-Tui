package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/config"
	"github.com/muyleanging/termix/internal/detector"
)

func TestSetupLayoutCentersWithinTerminal(t *testing.T) {
	m := Model{width: 100, height: 30}
	layout := m.setupLayout()

	if layout.small {
		t.Fatal("expected regular setup layout")
	}
	if layout.containerW <= 0 || layout.containerH <= 0 {
		t.Fatalf("invalid container size: %dx%d", layout.containerW, layout.containerH)
	}
	if layout.x < 0 || layout.y < 0 {
		t.Fatalf("invalid position: %d,%d", layout.x, layout.y)
	}
	if layout.x+layout.containerW > layout.terminalW {
		t.Fatalf("container overflows horizontally: x=%d w=%d terminal=%d", layout.x, layout.containerW, layout.terminalW)
	}
	if layout.y+layout.containerH > layout.terminalH-layout.footerH {
		t.Fatalf("container overflows vertically: y=%d h=%d terminal=%d footer=%d", layout.y, layout.containerH, layout.terminalH, layout.footerH)
	}
}

func TestSetupLayoutUsesSmallFallback(t *testing.T) {
	m := Model{width: 50, height: 14}
	layout := m.setupLayout()

	if !layout.small {
		t.Fatal("expected small setup layout")
	}
	if layout.containerW > layout.terminalW {
		t.Fatalf("small container overflows horizontally: %d > %d", layout.containerW, layout.terminalW)
	}
	if layout.containerH > layout.bodyH {
		t.Fatalf("small container overflows body: %d > %d", layout.containerH, layout.bodyH)
	}
}

func TestSetupViewFitsTerminal(t *testing.T) {
	m := testSetupModel(100, 30)
	view := m.setupView()
	assertViewFits(t, view, 100, 30)
	if !strings.Contains(view, "TERMIX FIRST SETUP") {
		t.Fatal("setup view missing title")
	}
	if !strings.Contains(view, "CHOOSE PROFILE") {
		t.Fatal("setup view missing profile picker")
	}
	if !strings.Contains(view, m.setupShell) {
		t.Fatal("setup view missing selected profile")
	}
}

func TestSetupViewSmallTerminalMessageFits(t *testing.T) {
	m := testSetupModel(50, 14)
	view := m.setupView()
	assertViewFits(t, view, 50, 14)
	if !strings.Contains(view, "Terminal too small.") {
		t.Fatal("small setup view missing friendly message")
	}
}

func TestSetupProfileSelectionUsesArrowKeys(t *testing.T) {
	m := testSetupModel(100, 30)
	targets := profileTargets(m.rt)
	if len(targets) < 2 {
		t.Skip("profile selection needs at least two available profiles")
	}
	m.setupShell = targets[0].Name
	if m.setupShell != targets[0].Name {
		t.Fatalf("initial setup shell = %q, want %q", m.setupShell, targets[0].Name)
	}

	next, quit, cmd := m.handleSetupKey(tea.KeyMsg{Type: tea.KeyDown})
	if quit {
		t.Fatal("down should not quit")
	}
	if cmd != nil {
		t.Fatal("down should not start a command")
	}
	if next.contentIndex != 1 {
		t.Fatalf("contentIndex after down = %d, want 1", next.contentIndex)
	}
	if next.setupShell != targets[1].Name {
		t.Fatalf("setupShell after down = %q, want %q", next.setupShell, targets[1].Name)
	}
	if next.setupFocusedElement() != "Profile" {
		t.Fatalf("focus after down = %q, want Profile", next.setupFocusedElement())
	}
}

func TestSetupTabCyclesProfileBackApply(t *testing.T) {
	m := testSetupModel(100, 30)
	if got := m.setupFocusedElement(); got != "Profile" {
		t.Fatalf("initial setup focus = %q, want Profile", got)
	}
	m, _, _ = m.handleSetupKey(tea.KeyMsg{Type: tea.KeyTab})
	if got := m.setupFocusedElement(); got != "Back" {
		t.Fatalf("focus after first tab = %q, want Back", got)
	}
	m, _, _ = m.handleSetupKey(tea.KeyMsg{Type: tea.KeyTab})
	if got := m.setupFocusedElement(); got != "Apply" {
		t.Fatalf("focus after second tab = %q, want Apply", got)
	}
}

func testSetupModel(width, height int) Model {
	rt := &app.Runtime{
		Config: config.Config{
			HomeDir:        ".",
			DefaultShell:   "PowerShell 7",
			DefaultFont:    "CaskaydiaCove Nerd Font",
			FavoriteThemes: []string{"catppuccin_mocha"},
			BorderStyle:    "unicode",
		},
		Env: detector.Environment{
			PowerShell: detector.ToolState{Installed: true},
			ANSI:       true,
			Unicode:    true,
		},
	}
	m := NewSetup(rt)
	m.startup = false
	m.width = width
	m.height = height
	return m
}

func assertViewFits(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("view height = %d, want %d", len(lines), height)
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got > width {
			t.Fatalf("line %d width = %d, want <= %d\n%s", i+1, got, width, line)
		}
	}
}
