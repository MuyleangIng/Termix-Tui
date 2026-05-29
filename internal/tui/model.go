package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/config"
	fontpkg "github.com/muyleanging/termix/internal/font"
	"github.com/muyleanging/termix/internal/installer"
	previewpkg "github.com/muyleanging/termix/internal/preview"
	"github.com/muyleanging/termix/internal/profile"
	"github.com/muyleanging/termix/internal/shell"
	themepkg "github.com/muyleanging/termix/internal/theme"
	"github.com/muyleanging/termix/internal/toolpath"
	"github.com/muyleanging/termix/internal/uninstaller"
)

const (
	appVersion = "v0.1.11"
	appAuthor  = "Ing Muyleang"
	appEmail   = "muyleanging@gmail.com"
)

type screen int

const (
	screenDashboard screen = iota
	screenThemes
	screenFonts
	screenDoctor
	screenProfiles
	screenSettings
	screenUpdate
	screenUninstall
	screenSetupShell
	screenSetupFont
	screenSetupTheme
	screenSetupPreview
	screenSetupApply
)

type focusArea int

const (
	focusSidebar focusArea = iota
	focusContent
	focusPreview
	focusLogs
)

type tickMsg time.Time
type previewMsg struct {
	name string
	text string
	err  error
}
type actionMsg struct {
	label         string
	err           error
	refreshThemes bool
}
type themesMsg struct {
	items []themepkg.Theme
	err   error
}

type navItem struct {
	icon  string
	label string
	view  screen
}

type Model struct {
	rt            *app.Runtime
	width         int
	height        int
	screen        screen
	focus         focusArea
	history       []screen
	setup         bool
	startup       bool
	bootStep      int
	palette       bool
	search        bool
	feedback      bool
	fontInput     bool
	fontInputMode string
	fontInputText string
	feedbackKind  string
	feedbackField int
	feedbackEmail string
	feedbackText  string
	confirm       bool
	confirmIndex  int
	lightMode     bool
	activityRows  int
	activityFold  bool
	resizingLogs  bool
	pending       pendingAction
	now           time.Time
	spinner       spinner.Model
	progress      progress.Model
	busy          string
	preview       string
	themePreviews map[string]string
	previewErrors map[string]string
	themeItems    []themepkg.Theme
	fontItems     []fontpkg.Font
	activeShell   string
	setupShell    string
	setupFont     string
	setupTheme    string
	setupButton   int
	setupNotice   string

	navIndex     int
	contentIndex int
	themeIndex   int
	previewIndex int
	logIndex     int
	scroll       int
	logs         []string
}

var nav = []navItem{
	{"⌘", "Dashboard", screenDashboard},
	{"", "Themes", screenThemes},
	{"", "Fonts", screenFonts},
	{"󰒡", "Doctor", screenDoctor},
	{"", "Profiles", screenProfiles},
	{"", "Settings", screenSettings},
	{"󰚰", "Updates", screenUpdate},
	{"", "Remove", screenUninstall},
}

var themes = []string{"catppuccin_mocha", "paradox", "atomic", "dracula", "tokyo", "night-owl", "multiverse-neon", "spaceship", "jandedobbeleer", "powerlevel10k_modern"}
var fonts = fontpkg.Choices()

type profileTarget struct {
	Name      string
	Installed bool
	Supported bool
	Detail    string
}

type pendingAction struct {
	kind    string
	label   string
	profile string
	font    string
	theme   themepkg.Theme
}

func New(rt *app.Runtime) Model {
	return base(rt, !rt.Config.SetupComplete)
}

func NewSetup(rt *app.Runtime) Model {
	m := base(rt, true)
	m.screen = screenSetupShell
	return m
}

func base(rt *app.Runtime, setup bool) Model {
	applyColorMode(false)
	applyBorderMode(rt.Config.BorderStyle)
	fonts = buildFontChoices(rt.Config)
	spin := spinner.New()
	spin.Spinner = spinner.MiniDot
	spin.Style = lipgloss.NewStyle().Foreground(cyan)
	bar := progress.New(progress.WithDefaultGradient())
	fontItems := fontpkg.Detect(userHome())
	initialLogs := []string{
		"SUCCESS terminal scan complete",
		"SUCCESS theme cache ready",
		"INFO profile watcher idle",
	}
	resolvedFont := fontpkg.ResolveAvailableFamily(userHome(), rt.Config.DefaultFont)
	if !strings.EqualFold(resolvedFont, fontpkg.ResolveFamily(rt.Config.DefaultFont)) {
		initialLogs = append([]string{"WARN Preferred font " + rt.Config.DefaultFont + " was not found. Termix is using fallback font " + resolvedFont + ". Some icons may look different."}, initialLogs...)
	}
	setupShell := rt.Config.DefaultShell
	if profileTargetByName(rt, setupShell).Name == "" {
		targets := profileTargets(rt)
		if len(targets) > 0 {
			setupShell = targets[0].Name
		}
	}
	return Model{
		rt:            rt,
		setup:         setup,
		startup:       true,
		now:           time.Now(),
		spinner:       spin,
		progress:      bar,
		preview:       previewText(),
		themePreviews: map[string]string{},
		previewErrors: map[string]string{},
		fontItems:     fontItems,
		activeShell:   setupShell,
		setupShell:    setupShell,
		setupFont:     rt.Config.DefaultFont,
		setupTheme:    firstOrDefault(rt.Config.FavoriteThemes, "catppuccin_mocha"),
		setupButton:   -1,
		activityRows:  clamp(4, 20, firstPositive(rt.Config.ActivityHeight, 7)),
		logs:          initialLogs,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tick(), loadThemesCmd(m.rt))
}

func tick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.progress.Width = max(12, min(38, msg.Width/4))
		if m.setup {
			layout := m.setupLayout()
			m.logs = cappedLogs(append([]string{fmt.Sprintf("INFO setup layout terminal=%dx%d container=%dx%d pos=%d,%d focus=%s profile=%s theme=%s", layout.terminalW, layout.terminalH, layout.containerW, layout.containerH, layout.x, layout.y, m.setupFocusedElement(), m.currentSetupProfileTarget().Name, m.setupTheme)}, m.logs...))
		}
	case tickMsg:
		m.now = time.Time(msg)
		if m.startup {
			m.bootStep++
			if m.bootStep >= 10 {
				m.startup = false
			}
		}
		cmds = append(cmds, tick())
	case tea.MouseMsg:
		m = m.handleMouse(msg)
	case tea.KeyMsg:
		oldScreen, oldIndex := m.screen, m.contentIndex
		next, quit, cmd := m.handleKey(msg)
		m = next
		if quit {
			return m, tea.Quit
		}
		cmds = append(cmds, cmd)
		if m.needsPreview() && (oldScreen != m.screen || oldIndex != m.contentIndex) {
			m.syncThemeIndex()
			cmds = append(cmds, m.renderPreviewCmd())
		}
	case previewMsg:
		m.busy = ""
		if msg.err != nil {
			m.previewErrors[msg.name] = msg.err.Error()
			delete(m.themePreviews, msg.name)
		} else if msg.text != "" {
			m.preview = msg.text
			m.themePreviews[msg.name] = msg.text
			delete(m.previewErrors, msg.name)
		}
	case actionMsg:
		m.busy = ""
		if m.setup && strings.HasPrefix(msg.label, "setup") {
			if msg.err != nil {
				m.setupNotice = "ERROR " + msg.err.Error()
				m.logs = cappedLogs(append([]string{"ERROR setup apply result: " + msg.err.Error()}, m.logs...))
			} else {
				m.setupNotice = "SUCCESS setup applied"
				m.logs = cappedLogs(append([]string{"SUCCESS setup apply result: " + m.setupShell + " / " + m.setupTheme}, m.logs...))
				m.setup = false
				m.screen = screenDashboard
			}
			if msg.refreshThemes {
				cmds = append(cmds, loadThemesCmd(m.rt))
			}
			break
		}
		if msg.err != nil {
			m.logs = cappedLogs(append([]string{"ERROR " + msg.label + ": " + msg.err.Error()}, m.logs...))
		} else {
			m.logs = cappedLogs(append([]string{"SUCCESS " + msg.label}, m.logs...))
		}
		if msg.refreshThemes {
			cmds = append(cmds, loadThemesCmd(m.rt))
		}
	case themesMsg:
		if msg.err != nil {
			m.logs = cappedLogs(append([]string{"WARN theme scan: " + msg.err.Error()}, m.logs...))
		} else if len(msg.items) > 0 {
			m.themeItems = msg.items
			names := make([]string, 0, len(msg.items))
			for _, item := range msg.items {
				names = append(names, item.Name)
			}
			themes = names
			m.themeIndex = clamp(0, max(0, len(names)-1), m.themeIndex)
			m.logs = cappedLogs(append([]string{fmt.Sprintf("SUCCESS loaded %d themes", len(msg.items))}, m.logs...))
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, bool, tea.Cmd) {
	if m.fontInput {
		return m.handleFontInputKey(msg)
	}
	if m.feedback {
		return m.handleFeedbackKey(msg)
	}
	if m.confirm {
		return m.handleConfirmKey(msg)
	}
	if m.setup {
		return m.handleSetupKey(msg)
	}
	switch msg.String() {
	case "q", "ctrl+c":
		return m, true, nil
	case "f1":
		m.logs = cappedLogs(append([]string{"INFO F1 ignored by Termix; use ? or h for help"}, m.logs...))
		return m, false, nil
	case "?", "h":
		m.palette = !m.palette
		m.search = false
	case "ctrl+p":
		m.palette = !m.palette
		m.search = false
	case "/":
		m.search = true
		m.palette = false
	case "m":
		m.lightMode = !m.lightMode
		applyColorMode(m.lightMode)
		m.spinner.Style = lipgloss.NewStyle().Foreground(cyan)
		m.logs = cappedLogs(append([]string{"SUCCESS color mode: " + m.colorModeName()}, m.logs...))
	case "esc":
		if m.palette || m.search {
			m.palette = false
			m.search = false
			return m, false, nil
		}
		m = m.goBack()
	case "backspace":
		if m.setup && m.screen > screenSetupShell {
			m.screen--
			return m, false, nil
		}
		m = m.goBack()
	case "tab":
		m.focus = (m.focus + 1) % 4
	case "shift+tab":
		m.focus = (m.focus + 3) % 4
	case "left":
		m.focus = maxFocus(focusSidebar, m.focus-1)
	case "right":
		m.focus = minFocus(focusLogs, m.focus+1)
	case "ctrl+up":
		m = m.resizeActivity(1, true)
	case "ctrl+down":
		m = m.resizeActivity(-1, true)
	case "ctrl+l":
		m.activityFold = !m.activityFold
		if !m.activityFold && m.activityRows < 4 {
			m.activityRows = 7
		}
		m.logs = cappedLogs(append([]string{"INFO activity panel toggled"}, m.logs...))
	case "a":
		if m.screen == screenFonts {
			return m.openFontInput("add"), false, nil
		}
	case "c":
		if m.screen == screenFonts {
			m.logs = cappedLogs(append([]string{"INFO custom fonts are managed with A/E/D"}, m.logs...))
		}
	case "r":
		if m.screen == screenFonts {
			m.fontItems = fontpkg.Detect(userHome())
			m.logs = cappedLogs(append([]string{"SUCCESS font scan refreshed"}, m.logs...))
		}
	case "i":
		if m.screen == screenFonts {
			return m.requestFontInstall(), false, nil
		}
	case "w":
		if m.screen == screenFonts {
			fontName := selectedText(fonts, m.contentIndex)
			m.busy = "applying font to Windows Terminal"
			return m, false, applyWindowsTerminalFontCmd(m.rt, fontName)
		}
	case "d":
		if m.screen == screenFonts {
			return m.deleteCustomFont(), false, nil
		}
	case "e":
		if m.screen == screenFonts {
			return m.openFontInput("edit"), false, nil
		}
	case "t":
		if m.screen == screenFonts {
			m.logs = cappedLogs(append([]string{"INFO glyph test: powerline/icons/borders shown in preview"}, m.logs...))
		}
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(1)
	case "home":
		m.jumpSelection(false)
	case "end":
		m.jumpSelection(true)
	case "pgup":
		m.scroll = max(0, m.scroll-5)
		m.moveSelection(-5)
	case "pgdown":
		m.scroll += 5
		m.moveSelection(5)
	case "enter":
		var cmd tea.Cmd
		m, cmd = m.activate()
		return m, false, cmd
	case "f":
		m.open(screenFonts)
	case "u":
		m.open(screenUpdate)
	case "x":
		m.open(screenUninstall)
	}
	return m, false, nil
}

func (m Model) handleSetupKey(msg tea.KeyMsg) (Model, bool, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, true, nil
	case "f1":
		m.logs = cappedLogs(append([]string{"INFO F1 ignored by Termix setup; use the footer shortcuts"}, m.logs...))
	case "?", "h":
		m.setupNotice = "Use Up/Down to select a profile, Tab to choose Back or Apply, Enter to run the focused action."
	case "esc", "backspace":
		return m.exitSetup(), false, nil
	case "tab", "right":
		m.setupButton = nextSetupFocus(m.setupButton, 1)
		m.logs = cappedLogs(append([]string{"INFO setup focus: " + m.setupFocusedElement()}, m.logs...))
	case "shift+tab", "left":
		m.setupButton = nextSetupFocus(m.setupButton, -1)
		m.logs = cappedLogs(append([]string{"INFO setup focus: " + m.setupFocusedElement()}, m.logs...))
	case "up", "k":
		m = m.moveSetupProfileSelection(-1)
	case "down", "j":
		m = m.moveSetupProfileSelection(1)
	case "home":
		m.contentIndex = 0
		m.setupButton = -1
		m.captureSetupSelection()
	case "end":
		m.contentIndex = max(0, len(setupProfileRows(m.rt))-1)
		m.setupButton = -1
		m.captureSetupSelection()
	case "enter":
		if m.setupButton == 0 {
			return m.exitSetup(), false, nil
		}
		if m.setupButton < 0 {
			m.setupButton = 1
			m.setupNotice = "Profile selected: " + m.setupShell + ". Press Enter again to apply."
			return m, false, nil
		}
		var cmd tea.Cmd
		m, cmd = m.activate()
		return m, false, cmd
	}
	return m, false, nil
}

func (m Model) handleFeedbackKey(msg tea.KeyMsg) (Model, bool, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.feedback = false
		m.feedbackEmail = ""
		m.feedbackText = ""
		m.logs = cappedLogs(append([]string{"INFO feedback cancelled"}, m.logs...))
		return m, false, nil
	case "ctrl+s", "enter":
		if msg.String() == "enter" && m.feedbackField == 0 {
			m.feedbackField = 1
			return m, false, nil
		}
		kind := m.feedbackKind
		email := strings.TrimSpace(m.feedbackEmail)
		text := strings.TrimSpace(m.feedbackText)
		m.feedback = false
		m.feedbackEmail = ""
		m.feedbackText = ""
		if text == "" {
			m.logs = cappedLogs(append([]string{"WARN feedback message was empty"}, m.logs...))
			return m, false, nil
		}
		m.busy = "saving feedback"
		return m, false, saveFeedbackCmd(m.rt, kind, email, text)
	case "backspace":
		if m.feedbackField == 0 && len(m.feedbackEmail) > 0 {
			m.feedbackEmail = m.feedbackEmail[:len(m.feedbackEmail)-1]
		} else if m.feedbackField == 1 && len(m.feedbackText) > 0 {
			m.feedbackText = m.feedbackText[:len(m.feedbackText)-1]
		}
	case "space":
		if m.feedbackField == 0 {
			m.feedbackEmail += " "
		} else {
			m.feedbackText += " "
		}
	case "tab", "shift+tab":
		m.feedbackField = 1 - m.feedbackField
	default:
		if len(msg.Runes) > 0 {
			if m.feedbackField == 0 {
				m.feedbackEmail += string(msg.Runes)
			} else {
				m.feedbackText += string(msg.Runes)
			}
		}
	}
	return m, false, nil
}

func (m Model) handleFontInputKey(msg tea.KeyMsg) (Model, bool, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.fontInput = false
		m.fontInputText = ""
		m.fontInputMode = ""
		return m, false, nil
	case "enter":
		value := strings.TrimSpace(m.fontInputText)
		m.fontInput = false
		m.fontInputText = ""
		mode := m.fontInputMode
		m.fontInputMode = ""
		if value == "" {
			m.logs = cappedLogs(append([]string{"WARN custom font name was empty"}, m.logs...))
			return m, false, nil
		}
		custom := append([]string{}, m.rt.Config.CustomFonts...)
		if mode == "edit" && m.isSelectedCustomFont() {
			custom = replaceStringFold(custom, selectedText(fonts, m.contentIndex), value)
		} else {
			custom = append(custom, value)
		}
		m.rt.Config.CustomFonts = uniqueLocal(custom)
		fonts = buildFontChoices(m.rt.Config)
		m.contentIndex = clamp(0, max(0, len(fonts)-1), m.contentIndex)
		m.logs = cappedLogs(append([]string{"SUCCESS custom font saved: " + value}, m.logs...))
		return m, false, saveCustomFontsCmd(m.rt, m.rt.Config.CustomFonts)
	case "backspace":
		if len(m.fontInputText) > 0 {
			m.fontInputText = m.fontInputText[:len(m.fontInputText)-1]
		}
	case "space":
		m.fontInputText += " "
	default:
		if len(msg.Runes) > 0 {
			m.fontInputText += string(msg.Runes)
		}
	}
	return m, false, nil
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (Model, bool, tea.Cmd) {
	switch strings.ToLower(msg.String()) {
	case "up", "k":
		m.moveConfirmSelection(-1)
		return m, false, nil
	case "down", "j":
		m.moveConfirmSelection(1)
		return m, false, nil
	case "y", "enter":
		pending := m.pending
		if pending.kind == "apply-theme" {
			target := m.currentConfirmProfileTarget()
			pending.profile = target.Name
		}
		m.confirm = false
		m.pending = pendingAction{}
		m.busy = "running " + pending.label
		return m, false, m.pendingActionCmd(pending)
	case "n", "esc", "backspace":
		m.confirm = false
		m.pending = pendingAction{}
		m.logs = cappedLogs(append([]string{"INFO cancelled"}, m.logs...))
		return m, false, nil
	default:
		return m, false, nil
	}
}

func (m *Model) moveConfirmSelection(delta int) {
	targets := profileTargets(m.rt)
	if len(targets) == 0 {
		m.confirmIndex = 0
		return
	}
	m.confirmIndex = clamp(0, len(targets)-1, m.confirmIndex+delta)
	m.pending.profile = targets[m.confirmIndex].Name
}

func (m Model) handleMouse(msg tea.MouseMsg) Model {
	if m.resizingLogs {
		switch msg.Action {
		case tea.MouseActionMotion:
			m.activityRows = m.activityHeightFromDivider(msg.Y)
			return m
		case tea.MouseActionRelease:
			m.resizingLogs = false
			m.activityRows = m.activityHeightFromDivider(msg.Y)
			_ = config.SaveActivityHeight(m.rt.Config, m.activityRows)
			m.logs = cappedLogs(append([]string{"INFO activity height saved"}, m.logs...))
			return m
		}
	}
	if msg.Action == tea.MouseActionPress {
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Y == m.activityDividerY() {
				m.resizingLogs = true
				m.focus = focusLogs
				return m
			}
			sideW := clamp(18, 24, max(82, m.width)/5)
			if msg.X <= sideW {
				m.focus = focusSidebar
				idx := msg.Y - 4
				if idx >= 0 && idx < len(nav) {
					m.navIndex = idx
					m.open(nav[idx].view)
				}
			} else if msg.Y > m.activityDividerY() {
				m.focus = focusLogs
			} else if msg.X > sideW+(max(82, m.width)-sideW)/2 {
				m.focus = focusPreview
			} else {
				m.focus = focusContent
			}
		case tea.MouseButtonWheelUp:
			if m.focus == focusLogs || msg.Y >= m.activityDividerY() {
				m.focus = focusLogs
				m.moveSelection(-1)
			} else {
				m.moveSelection(-1)
			}
		case tea.MouseButtonWheelDown:
			if m.focus == focusLogs || msg.Y >= m.activityDividerY() {
				m.focus = focusLogs
				m.moveSelection(1)
			} else {
				m.moveSelection(1)
			}
		}
	}
	return m
}

func (m Model) resizeActivity(delta int, persist bool) Model {
	m.activityFold = false
	m.activityRows = clampActivityHeight(m.height, m.activityRows+delta)
	if persist {
		_ = config.SaveActivityHeight(m.rt.Config, m.activityRows)
	}
	m.logs = cappedLogs(append([]string{fmt.Sprintf("INFO activity height: %d rows", m.activityRows)}, m.logs...))
	return m
}

func (m Model) activityDividerY() int {
	headerH := 1
	footerH := footerHeight(m.width)
	activityH := m.effectiveActivityHeight()
	return max(headerH, max(8, m.height)-footerH-activityH-1)
}

func (m Model) activityHeightFromDivider(y int) int {
	h := max(8, m.height)
	footerH := footerHeight(m.width)
	return clampActivityHeight(h, h-footerH-y-1)
}

func (m Model) effectiveActivityHeight() int {
	if m.activityFold {
		return 1
	}
	return clampActivityHeight(max(8, m.height), firstPositive(m.activityRows, 7))
}

func (m *Model) moveSelection(delta int) {
	switch m.focus {
	case focusSidebar:
		m.navIndex = clamp(0, len(nav)-1, m.navIndex+delta)
		m.open(nav[m.navIndex].view)
	case focusLogs:
		m.logIndex = clamp(0, max(0, len(m.logs)-1), m.logIndex+delta)
	case focusPreview:
		m.previewIndex = clamp(0, 4, m.previewIndex+delta)
	default:
		m.contentIndex = clamp(0, max(0, len(m.currentItems())-1), m.contentIndex+delta)
		m.syncThemeIndex()
	}
}

func (m *Model) jumpSelection(end bool) {
	target := 0
	if end {
		switch m.focus {
		case focusSidebar:
			target = len(nav) - 1
		case focusLogs:
			target = len(m.logs) - 1
		case focusPreview:
			target = 4
		default:
			target = len(m.currentItems()) - 1
		}
	}
	switch m.focus {
	case focusSidebar:
		m.navIndex = max(0, target)
		m.open(nav[m.navIndex].view)
	case focusLogs:
		m.logIndex = max(0, target)
	case focusPreview:
		m.previewIndex = max(0, target)
	default:
		m.contentIndex = max(0, target)
		m.syncThemeIndex()
	}
}

func (m *Model) syncThemeIndex() {
	if m.screen == screenThemes {
		m.themeIndex = clamp(0, max(0, len(themes)-1), m.contentIndex)
	}
}

func (m Model) activate() (Model, tea.Cmd) {
	if m.setup {
		m.captureSetupSelection()
		m.busy = "applying setup"
		m.setupNotice = "Applying setup..."
		layout := m.setupLayout()
		m.logs = cappedLogs(append([]string{
			fmt.Sprintf("INFO setup apply terminal=%dx%d container=%dx%d pos=%d,%d focus=%s profile=%s theme=%s", layout.terminalW, layout.terminalH, layout.containerW, layout.containerH, layout.x, layout.y, m.setupFocusedElement(), m.setupShell, m.setupTheme),
		}, m.logs...))
		return m, setupApplyCmd(m.rt, m.setupShell, m.setupFont, m.setupTheme)
	}
	if m.focus == focusSidebar {
		m.open(nav[m.navIndex].view)
		return m, nil
	}
	if m.focus == focusPreview || (m.needsPreview() && m.screen != screenThemes) {
		m.busy = "rendering preview"
		return m, m.renderPreviewCmd()
	}
	if m.screen == screenProfiles {
		target := m.currentProfileTarget()
		if err := validateProfileTarget(target); err != nil {
			m.logs = cappedLogs(append([]string{"ERROR switch profile: " + err.Error()}, m.logs...))
			return m, nil
		}
		m.activeShell = target.Name
		m.logs = cappedLogs(append([]string{"SUCCESS active profile: " + target.Name}, m.logs...))
		return m, nil
	}
	if m.screen == screenThemes {
		theme := m.currentTheme()
		m.confirm = true
		m.confirmIndex = profileIndexByName(m.rt, m.activeShell)
		m.pending = pendingAction{
			kind:    "apply-theme",
			label:   "apply " + theme.Name + " to " + m.activeShell,
			profile: m.activeShell,
			theme:   theme,
		}
		return m, nil
	}
	if m.screen == screenSettings && selectedText(m.currentItems(), m.contentIndex) == "Toggle Light/Dark Mode" {
		m.lightMode = !m.lightMode
		applyColorMode(m.lightMode)
		m.spinner.Style = lipgloss.NewStyle().Foreground(cyan)
		m.logs = cappedLogs(append([]string{"SUCCESS color mode: " + m.colorModeName()}, m.logs...))
		return m, nil
	}
	if m.screen == screenSettings && selectedText(m.currentItems(), m.contentIndex) == "Toggle Border Style" {
		next := nextBorderStyle(m.rt.Config.BorderStyle)
		m.rt.Config.BorderStyle = next
		applyBorderMode(next)
		_ = config.SaveBorderStyle(m.rt.Config, next)
		m.logs = cappedLogs(append([]string{"SUCCESS border style: " + next}, m.logs...))
		return m, nil
	}
	if m.screen == screenSettings {
		switch selectedText(m.currentItems(), m.contentIndex) {
		case "Send Feedback":
			m.openFeedbackComposer("Feedback")
			return m, nil
		case "Report Issue":
			m.openFeedbackComposer("Issue")
			return m, nil
		case "Request Update":
			m.openFeedbackComposer("Update Request")
			return m, nil
		}
	}
	item := selectedText(m.currentItems(), m.contentIndex)
	if item != "" {
		cmd := m.actionCmd(item)
		if cmd != nil {
			m.busy = "running " + item
			return m, cmd
		}
		m.logs = cappedLogs(append([]string{"✓ selected " + item}, m.logs...))
	}
	return m, nil
}

func (m Model) requestFontInstall() Model {
	fontName := selectedText(fonts, m.contentIndex)
	if !isRecommendedNerdFont(fontName) {
		m.logs = cappedLogs(append([]string{"WARN selected font has no automated installer: " + fontName}, m.logs...))
		return m
	}
	if m.isFontInstalled(fontName) {
		m.logs = cappedLogs(append([]string{"INFO font already installed: " + fontName}, m.logs...))
		return m
	}
	m.confirm = true
	m.pending = pendingAction{kind: "install-font", label: "install " + fontName, font: fontName}
	return m
}

func (m Model) openFontInput(mode string) Model {
	m.fontInput = true
	m.fontInputMode = mode
	m.fontInputText = ""
	if mode == "edit" {
		if !m.isSelectedCustomFont() {
			m.fontInput = false
			m.logs = cappedLogs(append([]string{"WARN only custom fonts can be edited"}, m.logs...))
			return m
		}
		m.fontInputText = selectedText(fonts, m.contentIndex)
	}
	return m
}

func (m Model) deleteCustomFont() Model {
	name := selectedText(fonts, m.contentIndex)
	if !m.isSelectedCustomFont() {
		m.logs = cappedLogs(append([]string{"WARN only custom fonts can be deleted"}, m.logs...))
		return m
	}
	m.rt.Config.CustomFonts = removeStringFold(m.rt.Config.CustomFonts, name)
	fonts = buildFontChoices(m.rt.Config)
	m.contentIndex = clamp(0, max(0, len(fonts)-1), m.contentIndex)
	_ = config.SaveCustomFonts(m.rt.Config, m.rt.Config.CustomFonts)
	m.logs = cappedLogs(append([]string{"SUCCESS custom font removed: " + name}, m.logs...))
	return m
}

func (m Model) exitSetup() Model {
	m.setup = false
	m.screen = screenDashboard
	m.navIndex = 0
	m.busy = ""
	m.setupNotice = ""
	m.logs = cappedLogs(append([]string{"INFO setup exited safely"}, m.logs...))
	return m
}

func (m Model) setupFocusedElement() string {
	if m.setupButton < 0 {
		return "Profile"
	}
	if m.setupButton == 0 {
		return "Back"
	}
	return "Apply"
}

func (m Model) moveSetupProfileSelection(delta int) Model {
	rows := setupProfileRows(m.rt)
	m.contentIndex = clamp(0, max(0, len(rows)-1), m.contentIndex+delta)
	m.setupButton = -1
	m.captureSetupSelection()
	target := m.currentSetupProfileTarget()
	if target.Name != "" {
		m.setupNotice = "Selected profile: " + target.Name + ". Tab to Apply."
	}
	return m
}

func nextSetupFocus(current, delta int) int {
	order := []int{-1, 0, 1}
	index := 0
	for i, value := range order {
		if value == current {
			index = i
			break
		}
	}
	index = (index + delta + len(order)) % len(order)
	return order[index]
}

func (m Model) isSelectedCustomFont() bool {
	name := selectedText(fonts, m.contentIndex)
	for _, custom := range m.rt.Config.CustomFonts {
		if strings.EqualFold(custom, name) {
			return true
		}
	}
	return false
}

func (m *Model) open(s screen) {
	if m.screen != s && !isSetupScreen(m.screen) {
		m.history = append(m.history, m.screen)
	}
	m.screen = s
	m.contentIndex = 0
	m.previewIndex = 0
	for i, item := range nav {
		if item.view == s {
			m.navIndex = i
			break
		}
	}
}

func (m Model) goBack() Model {
	if len(m.history) == 0 {
		m.screen = screenDashboard
		m.navIndex = 0
		return m
	}
	last := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	m.screen = last
	for i, item := range nav {
		if item.view == last {
			m.navIndex = i
			break
		}
	}
	return m
}

func (m Model) currentItems() []string {
	switch m.screen {
	case screenSetupShell:
		return setupProfileRows(m.rt)
	case screenSetupFont:
		return fonts
	case screenSetupTheme:
		return setupThemeRows()
	case screenSetupPreview, screenSetupApply:
		return []string{"Apply setup"}
	case screenThemes:
		return themes
	case screenFonts:
		return fonts
	case screenProfiles:
		return profileRows(m.rt)
	case screenSettings:
		return []string{"Toggle Light/Dark Mode", "Toggle Border Style", "Send Feedback", "Report Issue", "Request Update", "Shell profiles", "Theme paths", "Font settings", "Backups", "Restore points", "Reset workspace"}
	case screenUpdate:
		return []string{"Oh My Posh engine", "Nerd Fonts", "PowerShell profile", "Theme cache", "Terminal profiles"}
	case screenUninstall:
		return []string{"Oh My Posh integration", "Theme cache", "Downloaded themes", "Font config", "Shell snippets", "TERMIX cache"}
	case screenDoctor:
		return []string{"ANSI support", "Unicode support", "Nerd Font", "PowerShell profile", "Windows Terminal", "Oh My Posh"}
	default:
		return []string{"Theme Manager", "Fonts", "Profile Manager", "Doctor", "Settings"}
	}
}

func (m Model) needsPreview() bool {
	return m.screen == screenThemes || m.screen == screenSetupTheme || m.screen == screenSetupPreview
}

func (m Model) currentTheme() themepkg.Theme {
	if len(m.themeItems) > 0 {
		return m.themeItems[clamp(0, len(m.themeItems)-1, m.themeIndex)]
	}
	name := selectedText(themes, m.themeIndex)
	return themepkg.Theme{Name: name}
}

func (m Model) themeByName(name string) themepkg.Theme {
	for _, item := range m.themeItems {
		if item.Name == name {
			return item
		}
	}
	return themepkg.Theme{Name: name, Path: findThemePath(m.rt, name), Category: "community", Shells: []string{"pwsh", "bash", "zsh"}, Compatibility: "universal"}
}

func (m Model) renderPreviewCmd() tea.Cmd {
	item := m.currentTheme()
	return func() tea.Msg {
		text, err := previewpkg.New().Render(context.Background(), item)
		return previewMsg{name: item.Name, text: text, err: err}
	}
}

func (m Model) actionCmd(item string) tea.Cmd {
	switch m.screen {
	case screenUpdate:
		return func() tea.Msg {
			err := runUpdateAction(m.rt, item)
			return actionMsg{label: "updated " + item, err: err, refreshThemes: item == "Theme cache"}
		}
	case screenUninstall:
		return func() tea.Msg {
			err := runRemoveAction(m.rt, item)
			return actionMsg{label: "removed " + item, err: err}
		}
	case screenSettings:
		return func() tea.Msg {
			err := runSettingsAction(m.rt, item)
			return actionMsg{label: "checked " + item, err: err, refreshThemes: item == "Theme paths"}
		}
	case screenFonts:
		return func() tea.Msg {
			err := config.SaveFontChoice(m.rt.Config, item)
			if err == nil {
				m.rt.Config.DefaultFont = item
			}
			resolved := fontpkg.ResolveAvailableFamily(userHome(), item)
			label := "selected font " + item
			if !strings.EqualFold(resolved, fontpkg.ResolveFamily(item)) {
				label = "selected font " + item + "; using fallback " + resolved
			}
			return actionMsg{label: label, err: err}
		}
	case screenDoctor:
		return func() tea.Msg {
			err := runDoctorAction(m.rt, item)
			return actionMsg{label: "fixed " + item, err: err}
		}
	}
	return nil
}

func (m Model) pendingActionCmd(pending pendingAction) tea.Cmd {
	return func() tea.Msg {
		switch pending.kind {
		case "apply-theme":
			if err := validateProfileTarget(profileTargetByName(m.rt, pending.profile)); err != nil {
				return actionMsg{label: "apply theme " + pending.theme.Name + " to " + pending.profile, err: err}
			}
			err := applySelectedPromptToShell(m.rt, pending.profile, pending.theme)
			return actionMsg{label: "applied theme " + pending.theme.Name + " to " + pending.profile + " profile; restart shell to see it", err: err}
		case "install-font":
			err := installer.New(m.rt).Install(context.Background(), "font:"+pending.font)
			if err == nil {
				m.rt.Config.DefaultFont = pending.font
			}
			return actionMsg{label: "installed font " + pending.font, err: err}
		default:
			return actionMsg{label: pending.label, err: fmt.Errorf("unknown action %q", pending.kind)}
		}
	}
}

func saveCustomFontsCmd(rt *app.Runtime, custom []string) tea.Cmd {
	return func() tea.Msg {
		err := config.SaveCustomFonts(rt.Config, custom)
		return actionMsg{label: "saved custom fonts", err: err}
	}
}

func applyWindowsTerminalFontCmd(rt *app.Runtime, fontName string) tea.Cmd {
	return func() tea.Msg {
		err := profile.ApplyWindowsTerminalFont(userHome(), fontName)
		return actionMsg{label: "Windows Terminal font " + fontpkg.ResolveAvailableFamily(userHome(), fontName), err: err}
	}
}

func runUpdateAction(rt *app.Runtime, item string) error {
	switch item {
	case "Oh My Posh engine":
		return installer.New(rt).Install(context.Background(), "oh-my-posh")
	case "PowerShell profile":
		return installer.New(rt).Install(context.Background(), "powershell")
	case "Nerd Fonts":
		return installer.New(rt).Install(context.Background(), "fonts")
	case "Terminal profiles":
		return installer.New(rt).Install(context.Background(), "terminal")
	case "Theme cache":
		_, err := themepkg.InstallOfficialThemes(context.Background(), rt.Config)
		return err
	default:
		return nil
	}
}

func runDoctorAction(rt *app.Runtime, item string) error {
	switch item {
	case "Nerd Font":
		return installer.New(rt).Install(context.Background(), "fonts")
	case "PowerShell profile":
		return applySelectedPrompt(rt, themepkg.Theme{Name: "catppuccin_mocha", Path: findThemePath(rt, "catppuccin_mocha")})
	case "Windows Terminal":
		return installer.New(rt).Install(context.Background(), "terminal")
	case "Oh My Posh":
		return installer.New(rt).Install(context.Background(), "oh-my-posh")
	default:
		return nil
	}
}

func runRemoveAction(rt *app.Runtime, item string) error {
	switch item {
	case "Theme cache":
		return uninstaller.New(rt).Uninstall(context.Background(), "cache")
	case "Downloaded themes":
		return uninstaller.New(rt).Uninstall(context.Background(), "downloaded-themes")
	case "TERMIX cache":
		return uninstaller.New(rt).Uninstall(context.Background(), "cache")
	case "Oh My Posh integration", "Font config", "Shell snippets":
		return uninstaller.New(rt).Uninstall(context.Background(), "profile")
	default:
		return nil
	}
}

func runSettingsAction(rt *app.Runtime, item string) error {
	switch item {
	case "Toggle Light/Dark Mode":
		return nil
	case "Toggle Border Style":
		return nil
	case "Send Feedback":
		return nil
	case "Report Issue":
		return nil
	case "Request Update":
		return nil
	case "Shell profiles":
		return applySelectedPrompt(rt, themepkg.Theme{Name: "catppuccin_mocha", Path: findThemePath(rt, "catppuccin_mocha")})
	case "Theme paths":
		return themepkg.EnsureAvailable(context.Background(), rt.Config)
	case "Font settings":
		return profile.ApplyWindowsTerminalFont(userHome(), rt.Config.DefaultFont)
	case "Reset workspace":
		return uninstaller.New(rt).Uninstall(context.Background(), "cache")
	default:
		return config.MarkSetupComplete(rt.Config.HomeDir)
	}
}

func (m *Model) openFeedbackComposer(kind string) {
	m.feedback = true
	m.feedbackKind = kind
	m.feedbackField = 0
	m.feedbackEmail = ""
	m.feedbackText = ""
}

func saveFeedbackCmd(rt *app.Runtime, kind, email, text string) tea.Cmd {
	return func() tea.Msg {
		err := saveFeedback(rt.Config.HomeDir, kind, email, text)
		return actionMsg{label: "saved " + strings.ToLower(kind), err: err}
	}
}

func saveFeedback(homeDir, kind, email, text string) error {
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		return err
	}
	entry := fmt.Sprintf("time: %s\nkind: %s\nto: %s\nfrom: %s\nmessage:\n%s\n---\n", time.Now().Format(time.RFC3339), kind, appEmail, email, text)
	return appendFile(filepath.Join(homeDir, "feedback.log"), entry)
}

func appendFile(path, data string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(data)
	return err
}

func setupApplyCmd(rt *app.Runtime, shellName, fontName, themeName string) tea.Cmd {
	return func() tea.Msg {
		if err := installer.New(rt).Install(context.Background(), "oh-my-posh"); err != nil {
			return actionMsg{label: "setup", err: err}
		}
		if runtime.GOOS == "windows" {
			if err := installer.New(rt).Install(context.Background(), "fonts"); err != nil {
				return actionMsg{label: "setup", err: err}
			}
		}
		if runtime.GOOS != "linux" && isRecommendedNerdFont(fontName) {
			_ = installer.New(rt).Install(context.Background(), "font:"+fontName)
		}
		if err := themepkg.EnsureAvailable(context.Background(), rt.Config); err != nil {
			return actionMsg{label: "setup", err: err}
		}
		if err := profile.ApplyWindowsTerminalFont(userHome(), fontName); err != nil {
			return actionMsg{label: "setup", err: err}
		}
		if !strings.EqualFold(themeName, "No prompt style") {
			theme := themepkg.Theme{Name: themeName, Path: findThemePath(rt, themeName)}
			if err := applyProfileTarget(rt, profileTargetByName(rt, shellName), theme); err != nil {
				return actionMsg{label: "setup", err: err}
			}
		}
		if err := config.SaveSetupChoices(rt.Config, shellName, fontName, themeName); err != nil {
			return actionMsg{label: "setup", err: err}
		}
		if err := config.MarkSetupComplete(rt.Config.HomeDir); err != nil {
			return actionMsg{label: "setup", err: err}
		}
		return actionMsg{label: "setup applied", err: nil}
	}
}

func applySelectedPrompt(rt *app.Runtime, item themepkg.Theme) error {
	return applySelectedPromptToShell(rt, rt.Config.DefaultShell, item)
}

func applySelectedPromptToShell(rt *app.Runtime, shellName string, item themepkg.Theme) error {
	path := item.Path
	if path == "" {
		path = findThemePath(rt, item.Name)
	}
	if path == "" {
		return fmt.Errorf("theme %q was not found; install Oh My Posh themes or add theme_dirs in config", item.Name)
	}
	return profile.ApplyPrompt(userHome(), shellName, path)
}

func applyProfileTarget(rt *app.Runtime, target profileTarget, item themepkg.Theme) error {
	if err := validateProfileTarget(target); err != nil {
		return err
	}
	return applySelectedPromptToShell(rt, target.Name, item)
}

func validateProfileTarget(target profileTarget) error {
	if target.Name == "" {
		return fmt.Errorf("no profile selected")
	}
	if !target.Supported {
		return fmt.Errorf("%s is detected but Termix does not configure that profile yet", target.Name)
	}
	if !target.Installed {
		return fmt.Errorf("%s is not installed; install it from Settings first", target.Name)
	}
	return nil
}

func (m Model) profileTheme() themepkg.Theme {
	if len(m.themeItems) > 0 {
		return m.themeItems[0]
	}
	name := selectedText(themes, 0)
	return themepkg.Theme{Name: name, Path: findThemePath(m.rt, name)}
}

func (m Model) currentProfileTarget() profileTarget {
	items := profileTargets(m.rt)
	if len(items) == 0 {
		return profileTarget{}
	}
	return items[clamp(0, len(items)-1, m.contentIndex)]
}

func profileRows(rt *app.Runtime) []string {
	targets := profileTargets(rt)
	rows := make([]string, 0, len(targets))
	for _, target := range targets {
		state := "missing"
		if target.Installed {
			state = "ready"
		}
		if !target.Supported {
			state = "view only"
		}
		rows = append(rows, fmt.Sprintf("%s  %s", target.Name, state))
	}
	return rows
}

func setupProfileRows(rt *app.Runtime) []string {
	targets := profileTargets(rt)
	rows := make([]string, 0, len(targets))
	for _, target := range targets {
		state := "missing"
		if target.Installed {
			state = "ready"
		}
		if !target.Supported {
			state = "view only"
		}
		rows = append(rows, fmt.Sprintf("%s  %s", target.Name, state))
	}
	return rows
}

func profileTargets(rt *app.Runtime) []profileTarget {
	home := userHome()
	adapters := shell.Available(home)
	targets := make([]profileTarget, 0, len(adapters)+1)
	for _, adapter := range adapters {
		targets = append(targets, profileTarget{
			Name:      adapter.Name(),
			Installed: profileInstalled(rt, adapter.Name()),
			Supported: true,
			Detail:    adapter.ProfilePath(home),
		})
	}
	if runtime.GOOS == "windows" {
		targets = append(targets, profileTarget{Name: "CMD", Installed: commandExists("cmd"), Supported: false, Detail: "CMD prompt setup needs a separate Clink-style integration"})
	}
	return targets
}

func profileTargetByName(rt *app.Runtime, name string) profileTarget {
	for _, target := range profileTargets(rt) {
		if strings.EqualFold(target.Name, name) {
			return target
		}
	}
	return profileTarget{}
}

func profileInstalled(rt *app.Runtime, name string) bool {
	switch name {
	case "PowerShell 7":
		return rt.Env.PowerShell.Installed
	case "Windows PowerShell":
		return commandExists("powershell")
	case "Git Bash", "Bash":
		return rt.Env.GitBash.Installed
	case "WSL Bash":
		return rt.Env.WSL.Installed
	case "Zsh":
		return rt.Env.Zsh.Installed
	case "Fish":
		return rt.Env.Fish.Installed
	case "Nushell":
		return rt.Env.Nushell.Installed
	default:
		return false
	}
}

func profileIndexByName(rt *app.Runtime, name string) int {
	for i, target := range profileTargets(rt) {
		if strings.EqualFold(target.Name, name) {
			return i
		}
	}
	return 0
}

func commandExists(name string) bool {
	return toolpath.Exists(name)
}

func setupThemeRows() []string {
	items := make([]string, 0, min(12, len(themes)+1))
	items = append(items, "No prompt style")
	for _, item := range themes {
		items = append(items, item)
		if len(items) >= 12 {
			break
		}
	}
	return items
}

func (m *Model) captureSetupSelection() {
	switch m.screen {
	case screenSetupShell:
		target := m.currentSetupProfileTarget()
		if target.Name != "" {
			m.setupShell = target.Name
		}
	case screenSetupFont:
		m.setupFont = selectedText(fonts, m.contentIndex)
	case screenSetupTheme:
		m.setupTheme = selectedText(setupThemeRows(), m.contentIndex)
	}
}

func (m Model) currentSetupProfileTarget() profileTarget {
	items := profileTargets(m.rt)
	if len(items) == 0 {
		return profileTarget{}
	}
	return items[clamp(0, len(items)-1, m.contentIndex)]
}

func findThemePath(rt *app.Runtime, name string) string {
	for _, dir := range rt.Config.ThemeDirs {
		candidate := filepath.Join(expandHome(dir), name+".omp.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func userHome() string {
	home, _ := os.UserHomeDir()
	return home
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		return filepath.Join(userHome(), path[2:])
	}
	return path
}

func loadThemesCmd(rt *app.Runtime) tea.Cmd {
	return func() tea.Msg {
		items, err := themepkg.NewManager(rt.Config).Scan(context.Background())
		return themesMsg{items: items, err: err}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return startupShell()
	}
	if m.startup {
		return m.startupView()
	}
	if m.fontInput {
		return appShell.Render(fitScreen(lipgloss.Place(max(40, m.width), max(8, m.height), lipgloss.Center, lipgloss.Center, m.fontInputBox(max(40, m.width))), max(40, m.width), max(8, m.height)))
	}
	if m.feedback {
		return appShell.Render(fitScreen(lipgloss.Place(max(40, m.width), max(8, m.height), lipgloss.Center, lipgloss.Center, m.feedbackBox(max(40, m.width))), max(40, m.width), max(8, m.height)))
	}
	if m.setup {
		return m.setupView()
	}

	w := max(44, m.width)
	h := max(8, m.height)
	headerH := 1
	footerH := footerHeight(w)
	dividerH := 1
	activityH := m.effectiveActivityHeight()
	bodyH := max(4, h-headerH-dividerH-activityH-footerH)
	sideW := clamp(18, 24, w/5)
	contentW := max(20, w-sideW)

	body := fitScreen(lipgloss.JoinHorizontal(lipgloss.Top,
		m.sidebar(sideW, bodyH),
		m.content(contentW, bodyH),
	), w, bodyH)
	if m.palette {
		body = fitScreen(lipgloss.Place(w, bodyH, lipgloss.Center, lipgloss.Center, m.commandPalette(w)), w, bodyH)
	}
	if m.search {
		body = fitScreen(lipgloss.Place(w, bodyH, lipgloss.Center, lipgloss.Center, m.searchBox(w)), w, bodyH)
	}
	out := lipgloss.JoinVertical(lipgloss.Left,
		fitScreen(m.header(w), w, headerH),
		body,
		m.activityDivider(w),
		fitScreen(m.activityPanel(w, activityH), w, activityH),
		fitScreen(m.footer(w), w, footerH),
	)
	if m.confirm {
		return appShell.Render(fitScreen(lipgloss.Place(w, max(8, m.height), lipgloss.Center, lipgloss.Center, m.confirmBox(w)), w, max(8, m.height)))
	}
	return appShell.Render(fitScreen(out, w, h))
}

func (m Model) content(w, h int) string {
	contentH := max(3, h)
	mainW := clamp(22, 48, w/2)
	previewW := max(20, w-mainW)
	main := m.mainPanel(mainW, contentH)
	preview := m.previewPanel(previewW, contentH)
	return fitScreen(lipgloss.JoinHorizontal(lipgloss.Top, main, preview), w, h)
}

func (m Model) header(w int) string {
	shellName := m.activeShell
	if shellName == "" {
		shellName = m.rt.Config.DefaultShell
	}
	themeName := selectedText(themes, clamp(0, len(themes)-1, m.themeIndex))
	if m.screen == screenProfiles {
		target := m.currentProfileTarget()
		if target.Name != "" {
			shellName = target.Name
		}
	}
	if m.setup {
		shellName = m.setupHeaderShell()
		themeName = m.setupHeaderTheme()
	}
	cells := []string{
		title.Render("TERMIX " + appVersion),
		pill.Render(shellName),
		pill.Render(themeName),
		pill.Render(m.colorModeName()),
		pill.Render("ANSI " + stateIcon(m.rt.Env.ANSI)),
		pill.Render("UTF8 " + stateIcon(m.rt.Env.Unicode)),
		pill.Render(m.now.Format("15:04:05")),
	}
	return header.Width(w).Render(fitLine(joinFit(cells, "  ", w-2), w-2))
}

func (m Model) setupHeaderShell() string {
	if m.screen == screenSetupShell {
		target := m.currentSetupProfileTarget()
		if target.Name != "" {
			return target.Name
		}
	}
	return m.setupShell
}

func (m Model) setupHeaderTheme() string {
	switch m.screen {
	case screenSetupFont:
		return selectedText(fonts, m.contentIndex)
	case screenSetupTheme:
		return selectedText(setupThemeRows(), m.contentIndex)
	case screenSetupShell, screenSetupPreview, screenSetupApply:
		if m.setupTheme != "" {
			return m.setupTheme
		}
	}
	return firstOrDefault(m.rt.Config.FavoriteThemes, "catppuccin_mocha")
}

func (m Model) footer(w int) string {
	left := "Made by MuyleangIng"
	right := "Open Source • Modern Terminal Experience Manager"
	shortcuts := "↑↓ Select  Enter Open  Tab Focus  ? Help  Q Quit"
	if m.setup {
		shortcuts = "↑↓ Select  Tab Focus  Enter Apply  Esc Back  Q Quit"
		right = "Configure your terminal experience"
	}
	line := footerLine(w-2, left, right, shortcuts)
	return footer.Width(w).Render(line)
}

func footerHeight(w int) int {
	return 1
}

func footerLine(width int, left, right, shortcuts string) string {
	parts := []string{left, shortcuts, right}
	if width < 98 {
		parts = []string{left, shortcuts}
	}
	if width < 48 {
		parts = []string{left}
	}
	return fitLine(joinFit(parts, "  │  ", width), width)
}

func (m Model) colorModeName() string {
	if m.lightMode {
		return "Light"
	}
	return "Dark"
}

func nextBorderStyle(current string) string {
	switch strings.ToLower(strings.TrimSpace(current)) {
	case "unicode":
		return "ascii"
	case "ascii":
		return "none"
	default:
		return "unicode"
	}
}

func (m Model) sidebar(w, h int) string {
	lines := []string{label.Render(" NAVIGATION"), ""}
	for i, item := range nav {
		row := fmt.Sprintf("%s  %s", item.icon, item.label)
		style := menuRow(i == m.navIndex, max(8, w-4))
		if i == m.navIndex {
			row = "❯ " + row
		} else {
			row = "  " + row
		}
		lines = append(lines, style.Render(fitLine(row, max(4, w-6))))
	}
	hints := hintStyle().Width(max(8, w-4)).Render(fitBlock("↑↓ select\nEnter open\n←→ focus", max(4, w-6), 3))
	lines = append(lines, "", hints)
	return focusedPanel(m.focus == focusSidebar).Width(w).Height(h).Render(fitBlock(strings.Join(lines, "\n"), max(8, w-4), max(4, h-2)))
}

func (m Model) mainPanel(w, h int) string {
	items := m.currentItems()
	rows := []string{sectionTitle(screenName(m.screen))}
	rows = append(rows, "")
	rows = append(rows, m.renderList(items, max(1, h-5), w-4)...)
	body := fitBlock(strings.Join(rows, "\n"), max(4, w-4), max(3, h-2))
	return m.panelStyle(focusContent).Width(w).Height(h).Render(body)
}

func (m Model) previewPanel(w, h int) string {
	var body string
	switch m.screen {
	case screenThemes:
		theme := selectedText(themes, m.themeIndex)
		item := m.currentTheme()
		body = sectionTitle("THEME DETAILS") + "\n" + theme +
			"\nCategory: " + firstOrDefault([]string{item.Category}, "community") +
			"\nSegments: " + fmt.Sprintf("%d", item.Segments) +
			"\n\nEnter opens apply confirmation.\nR refreshes real preview.\n\n" + m.realThemePreviewCard(item, max(24, w-4), max(8, h-9))
	case screenFonts:
		font := selectedText(fonts, m.contentIndex)
		resolved := fontpkg.ResolveAvailableFamily(userHome(), font)
		warning := ""
		if !strings.EqualFold(resolved, fontpkg.ResolveFamily(font)) {
			warning = "\n" + warn.Render("Preferred font not found. Termix is using fallback "+resolved+". Some icons may look different.")
		}
		body = sectionTitle("FONT MANAGER") +
			"\nCurrent font: " + m.rt.Config.DefaultFont +
			"\nHighlighted: " + font +
			"\nResolved face: " + resolved +
			"\nStatus: " + m.fontStatus(font) + m.fontTags(font) + warning +
			"\n\nFallback stack:\n" + fitLine(strings.Join(fontpkg.FallbackStack, " -> "), max(16, w-6)) +
			"\n\n" + fontSampleText() +
			"\n\n" + badgeRow("Enter config", "I install", "A add", "E edit", "D delete", "W Windows Terminal", "R rescan")
	case screenDoctor:
		body = sectionTitle("TERMINAL HEALTH") + "\n" + statusLine("ANSI", m.rt.Env.ANSI) + "\n" + statusLine("Unicode", m.rt.Env.Unicode) + "\n" + statusLine("Font fallback", true) + "\n" + statusLine("Oh My Posh", m.rt.Env.OhMyPosh.Installed) + "\n\n" + sectionTitle("FIXES") + "\n• Install Oh My Posh\n• Configure PowerShell profile\n• Select terminal font"
	case screenProfiles:
		target := m.currentProfileTarget()
		state := "missing"
		if target.Installed {
			state = "ready"
		}
		if !target.Supported {
			state = "view only"
		}
		active := ""
		if target.Name == m.activeShell {
			active = "\nActive: yes"
		}
		body = sectionTitle("PROFILE MANAGER") + "\n" + target.Name + "\n\nStatus: " + state + active + "\nTarget: " + target.Detail + "\n\nEnter switches active profile.\nThen choose a theme in Themes."
	default:
		body = sectionTitle("WORKSPACE") +
			"\nActive profile: " + m.activeShell +
			"\nTheme: " + selectedText(themes, clamp(0, len(themes)-1, m.themeIndex)) +
			"\nMode: " + m.colorModeName() +
			"\n\n" + m.preview +
			"\n\n" + badgeRow("Profiles first", "Themes apply there", "Fonts", "Renderer")
	}
	return m.panelStyle(focusPreview).Width(w).Height(h).Render(fitBlock(body, max(4, w-4), max(3, h-2)))
}

func (m Model) renderList(items []string, maxRows, width int) []string {
	if len(items) == 0 {
		return []string{label.Render("no items")}
	}
	selected := clamp(0, len(items)-1, m.contentIndex)
	start := clamp(0, max(0, len(items)-maxRows), selected-maxRows/2)
	end := min(len(items), start+maxRows)
	var rows []string
	for i := start; i < end; i++ {
		line := fitLine(m.listLabel(items[i]), max(8, width-4))
		if i == selected {
			rows = append(rows, menuRow(true, max(8, width)).Render("❯ "+line))
		} else {
			rows = append(rows, menuRow(false, max(8, width)).Render("  "+line))
		}
	}
	return rows
}

func (m Model) listLabel(item string) string {
	switch m.screen {
	case screenFonts:
		return item + "  " + m.fontStatus(item) + m.fontTags(item)
	default:
		return item
	}
}

func (m Model) panelStyle(area focusArea) lipgloss.Style {
	return focusedPanel(m.focus == area)
}

func (m Model) focusColor(area focusArea) lipgloss.Color {
	if m.focus == area {
		return cyan
	}
	return theme.border
}

func (m Model) activityPanel(w, h int) string {
	if h <= 1 || m.activityFold {
		return focusedPanel(m.focus == focusLogs).Width(w).Height(max(1, h)).Render(fitLine("ACTIVITY  "+m.latestLogSummary(), max(4, w-4)))
	}
	rows := []string{activityHeader(w, len(m.logs), m.activityRows)}
	if m.busy != "" {
		rows = append(rows, "  "+m.spinner.View()+" "+fitLine(m.busy, max(8, w-8)))
	}
	limit := max(1, h-3)
	start := clamp(0, max(0, len(m.logs)-limit), m.logIndex-limit/2)
	end := min(len(m.logs), start+limit)
	for i := start; i < end; i++ {
		entry := renderLogLine(m.logs[i], max(8, w-8))
		if i == m.logIndex && m.focus == focusLogs {
			rows = append(rows, menuRow(true, max(8, w-4)).Render("❯ "+entry))
		} else {
			rows = append(rows, menuRow(false, max(8, w-4)).Render("  "+entry))
		}
	}
	return m.panelStyle(focusLogs).Width(w).Height(h).Render(fitBlock(strings.Join(rows, "\n"), max(4, w-4), max(2, h-2)))
}

func (m Model) activityDivider(w int) string {
	label := " drag to resize  Ctrl+Up/Down resize  Ctrl+L collapse "
	lineW := max(1, w-lipgloss.Width(label)-2)
	left := strings.Repeat("─", lineW/2)
	right := strings.Repeat("─", max(0, lineW-lineW/2))
	return lipgloss.NewStyle().Foreground(cyan).Background(bg).Width(w).Render(fitLine(left+label+right, w))
}

func activityHeader(w, count, rows int) string {
	meta := fmt.Sprintf("%d logs  %d rows  Ctrl+L collapse", count, rows)
	return fitLine(sectionTitle("ACTIVITY")+"    "+label.Render(meta), max(4, w-4))
}

func (m Model) latestLogSummary() string {
	if len(m.logs) == 0 {
		return "no logs"
	}
	return renderLogLine(m.logs[0], 80)
}

func renderLogLine(entry string, width int) string {
	kind, text := splitLogEntry(entry)
	badge := badgeStyle(kind).Render(fitLine(kind, 7))
	return fitLine(fmt.Sprintf("%s  %s", badge, text), width)
}

func splitLogEntry(entry string) (string, string) {
	for _, kind := range []string{"SUCCESS", "ERROR", "WARN", "INFO"} {
		prefix := kind + " "
		if strings.HasPrefix(entry, prefix) {
			return kind, strings.TrimPrefix(entry, prefix)
		}
	}
	return "INFO", entry
}

func (m Model) realThemePreviewCard(item themepkg.Theme, width, height int) string {
	status := ok.Render("preview ready")
	preview := strings.TrimSpace(m.themePreviews[item.Name])
	if preview == "" {
		if errText := m.previewErrors[item.Name]; errText != "" {
			status = warn.Render("missing dependency")
			preview = dependencyPreviewMessage(errText)
		} else {
			status = label.Render("press R to render")
			preview = "Real preview pending.\nPress R to render with Oh My Posh."
		}
	}
	warning := ""
	if m.preferredNerdFontMissing() {
		warning = "\n" + warn.Render("Preferred Nerd Font missing. Glyphs may differ.")
	}
	body := sectionTitle(item.Name) + " " + status + warning + "\n\n" + preview
	return card.Width(width).Height(height).Render(fitBlock(body, max(8, width-4), max(4, height-2)))
}

func (m Model) previewLabBody(themeName, preview string, width, height int) string {
	if strings.TrimSpace(preview) == "" {
		if errText := m.previewErrors[themeName]; errText != "" {
			return dependencyPreviewMessage(errText)
		}
		return "Real preview pending.\nPress R to render with Oh My Posh."
	}
	return fitScreen(preview, width, height)
}

func (m Model) fontStatus(name string) string {
	family := fontpkg.ResolveFamily(name)
	for _, item := range m.fontItems {
		if strings.EqualFold(item.Name, name) || strings.EqualFold(item.Family, family) {
			if item.Installed {
				return ok.Render("Available")
			}
			break
		}
	}
	if !strings.Contains(strings.ToLower(name), "nerd") {
		return warn.Render("Fallback")
	}
	return warn.Render("Missing")
}

func (m Model) isFontInstalled(name string) bool {
	family := fontpkg.ResolveFamily(name)
	for _, item := range m.fontItems {
		if (strings.EqualFold(item.Name, name) || strings.EqualFold(item.Family, family)) && item.Installed {
			return true
		}
	}
	return false
}

func (m Model) isCustomFont(name string) bool {
	for _, custom := range m.rt.Config.CustomFonts {
		if strings.EqualFold(custom, name) {
			return true
		}
	}
	return false
}

func (m Model) fontGlyphStatus(name string) string {
	if strings.Contains(strings.ToLower(name), "nerd") && strings.Contains(m.fontStatus(name), "Available") {
		return ok.Render("Nerd glyphs likely")
	}
	return label.Render("glyph fallback possible")
}

func (m Model) fontTags(name string) string {
	var tags []string
	if strings.EqualFold(name, m.rt.Config.DefaultFont) {
		tags = append(tags, ok.Render("ACTIVE"))
	}
	if isRecommendedNerdFont(name) {
		tags = append(tags, accent.Render("RECOMMENDED"))
	}
	if m.isCustomFont(name) {
		tags = append(tags, label.Render("CUSTOM"))
	}
	if !strings.Contains(strings.ToLower(name), "nerd") {
		tags = append(tags, warn.Render("FALLBACK"))
	}
	if len(tags) == 0 {
		return ""
	}
	return "  " + strings.Join(tags, " ")
}

func fontSampleText() string {
	return "ABC abc 123\nPowerline:      \nIcons:      ⚙\nGit:  main ✔\nPrompt: Admin   ~/workspace   main ✔\nBorders: ┌ ─ ┐ │ └ ┘"
}

func dependencyPreviewMessage(errText string) string {
	return "Oh My Posh real preview is unavailable.\nRun termix setup or termix install to install required tools.\n\n" + fitLine(errText, 80)
}

func (m Model) preferredNerdFontMissing() bool {
	for _, item := range m.fontItems {
		if strings.Contains(strings.ToLower(item.Name), "cascadia") && strings.Contains(strings.ToLower(item.Name), "nerd") {
			return !item.Installed
		}
	}
	return true
}

type setupLayoutMetrics struct {
	terminalW  int
	terminalH  int
	headerH    int
	footerH    int
	bodyH      int
	containerW int
	containerH int
	x          int
	y          int
	small      bool
}

func (m Model) setupLayout() setupLayoutMetrics {
	w := max(1, m.width)
	h := max(1, m.height)
	headerH := 1
	footerH := footerHeight(w)
	bodyH := max(1, h-headerH-footerH)
	layout := setupLayoutMetrics{
		terminalW: w,
		terminalH: h,
		headerH:   headerH,
		footerH:   footerH,
		bodyH:     bodyH,
		small:     w < 68 || h < 20,
	}
	if layout.small {
		layout.containerW = min(w, min(56, max(20, w-4)))
		layout.containerH = min(bodyH, min(12, max(6, bodyH-2)))
	} else {
		layout.containerW = clamp(64, 96, w-4)
		layout.containerH = min(24, max(18, bodyH-2))
	}
	layout.x = max(0, (w-layout.containerW)/2)
	layout.y = headerH + max(0, (bodyH-layout.containerH)/2)
	return layout
}

func (m Model) setupView() string {
	layout := m.setupLayout()
	var body string
	if layout.small {
		body = m.setupSmallView(layout.containerW, layout.containerH)
	} else {
		body = m.setupOnboarding(layout.containerW, layout.containerH)
	}
	centered := fitScreen(lipgloss.Place(layout.terminalW, layout.bodyH, lipgloss.Center, lipgloss.Center, body), layout.terminalW, layout.bodyH)
	return appShell.Render(fitScreen(lipgloss.JoinVertical(lipgloss.Left,
		fitScreen(m.header(layout.terminalW), layout.terminalW, layout.headerH),
		centered,
		fitScreen(m.footer(layout.terminalW), layout.terminalW, layout.footerH),
	), layout.terminalW, layout.terminalH))
}

func (m Model) setupSmallView(w, h int) string {
	lines := []string{
		title.Render("TERMIX FIRST SETUP"),
		"",
		"Terminal too small.",
		"Resize to at least 68x20.",
		"",
		"Q quits safely.",
	}
	return cardWarn.Width(w).Height(h).Render(fitBlock(strings.Join(lines, "\n"), max(12, w-4), max(4, h-2)))
}

func (m Model) setupOnboarding(w, h int) string {
	innerW := max(32, w-4)
	panelH := 12
	colW := max(24, (innerW-3)/2)
	target := m.currentSetupProfileTarget()
	resolvedFont := fontpkg.ResolveAvailableFamily(userHome(), m.setupFont)
	notice := m.setupNotice
	if notice == "" {
		notice = "Ready to configure " + target.Name + "."
	}
	if m.busy != "" {
		notice = m.spinner.View() + " " + m.busy + "  65%"
	}

	leftLines := []string{
		sectionTitle("CHOOSE PROFILE"),
		"",
		setupList(setupProfileRows(m.rt), m.contentIndex),
		"",
		sectionTitle("APPLY AUTOMATICALLY"),
		ok.Render("✓") + " install missing tools",
		ok.Render("✓") + " import official themes",
		ok.Render("✓") + " write selected profile",
	}
	rightLines := []string{
		sectionTitle("PREVIEW"),
		"",
		"Profile: " + target.Name,
		"Font: " + resolvedFont,
		"Theme: " + m.setupTheme,
		"",
		"This updates your shell profile",
		"and Windows Terminal config.",
		"Target: " + target.Detail,
		"State: " + setupProfileState(target),
	}
	left := focusedPanel(m.setupButton < 0).Width(colW).Height(panelH).Render(fitBlock(strings.Join(leftLines, "\n"), max(12, colW-4), max(5, panelH-2)))
	right := focusedPanel(false).Width(colW).Height(panelH).Render(fitBlock(strings.Join(rightLines, "\n"), max(12, colW-4), max(5, panelH-2)))
	buttons := lipgloss.NewStyle().Align(lipgloss.Center).Width(innerW).Render(
		m.setupButtonView("Back", m.setupButton == 0) + "          " + m.setupButtonView("Apply", m.setupButton == 1),
	)
	status := lipgloss.NewStyle().Foreground(theme.muted).Width(innerW).Render(fitLine(notice, innerW))
	progressLine := ""
	if m.busy != "" {
		m.progress.Width = min(34, max(18, innerW/2))
		progressLine = lipgloss.NewStyle().Align(lipgloss.Center).Width(innerW).Render(m.progress.ViewAs(0.65))
	}
	headerBlock := lipgloss.JoinVertical(lipgloss.Left,
		title.Render("TERMIX FIRST SETUP"),
		label.Render("Configure your terminal profile, theme, and font"),
	)
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerBlock,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right),
		"",
		status,
		progressLine,
		"",
		buttons,
	)
	return cardHot.Width(w).Height(h).Render(fitBlock(content, innerW, max(8, h-2)))
}

func (m Model) setupButtonView(name string, focused bool) string {
	text := "[ " + name + " ]"
	if focused {
		return lipgloss.NewStyle().Foreground(theme.inkInverse).Background(theme.cyan).Bold(true).Padding(0, 1).Render(text)
	}
	return lipgloss.NewStyle().Foreground(theme.ink).Background(theme.elevated).Padding(0, 1).Render(text)
}

func (m Model) setupPreviewText() string {
	if strings.EqualFold(m.setupTheme, "No prompt style") {
		return "No prompt style will be written.\n" + m.setupShell + " keeps its current prompt."
	}
	return "Themes page will render the real Oh My Posh prompt for " + m.setupTheme + "."
}

func (m Model) startupView() string {
	w := max(44, m.width)
	h := max(14, m.height)
	checks := []string{"Detect terminal", "Load themes", "Resolve font stack", "Scan profiles", "Prepare renderer"}
	step := clamp(0, len(checks), m.bootStep/2)
	boxW := clamp(42, 64, w-6)
	innerW := max(24, boxW-4)
	pct := float64(step) / float64(len(checks))
	if m.bootStep >= 9 {
		pct = 1
	}
	m.progress.Width = max(18, min(42, innerW-2))

	lines := []string{
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerW).Render(title.Render("TERMIX")),
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerW).Render(label.Render(appVersion + "  •  Modern Terminal Experience Manager")),
		"",
		m.progress.ViewAs(pct),
		"",
	}
	for i, check := range checks {
		state := label.Render("○")
		if i < step {
			state = ok.Render("●")
		} else if i == step && pct < 1 {
			state = accent.Render("●")
		}
		lines = append(lines, fmt.Sprintf("%s  %-20s %s", state, check, startupStatus(i, step, pct)))
	}
	lines = append(lines, "", label.Render("Font fallback: "+strings.Join(fontpkg.FallbackStack[:5], " → ")))
	body := fitBlock(strings.Join(lines, "\n"), innerW, min(max(10, h-4), 14))
	return appShell.Render(lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, cardHot.Width(boxW).Render(body)))
}

func startupStatus(index, step int, pct float64) string {
	switch {
	case index < step:
		return ok.Render("READY")
	case index == step && pct < 1:
		return accent.Render("RUNNING")
	default:
		return label.Render("PENDING")
	}
}

func (m Model) commandPalette(w int) string {
	boxW := clamp(34, 56, w-4)
	return cardHot.Width(boxW).Render(fitBlock(sectionTitle("HELP / COMMANDS")+"\n? or h toggles this help. F1 is ignored by Termix.\n\n❯ Open Theme Manager\n  Manage Fonts\n  Run Terminal Doctor\n  Configure Profile\n  Settings", max(20, boxW-4), 9))
}

func (m Model) searchBox(w int) string {
	boxW := clamp(34, 56, w-4)
	return cardHot.Width(boxW).Render(fitBlock(sectionTitle("SEARCH")+"\nType to filter current panel\n\n/", max(20, boxW-4), 5))
}

func (m Model) feedbackBox(w int) string {
	boxW := clamp(48, 78, w-4)
	innerW := max(24, boxW-4)
	body := sectionTitle("SEND "+strings.ToUpper(m.feedbackKind)) +
		"\n" + label.Render("Internal Termix form. Saved locally to ~/.termix/feedback.log.") +
		"\n\nTo: " + appEmail +
		"\n\n" + feedbackInput("Your email", m.feedbackEmail, m.feedbackField == 0, innerW) +
		"\n\n" + feedbackInput("Message", m.feedbackText, m.feedbackField == 1, innerW) +
		"\n\n" + badgeRow("Tab switch", "Enter next/save", "Ctrl+S save", "Esc cancel")
	return cardHot.Width(boxW).Render(fitBlock(body, innerW, 18))
}

func (m Model) fontInputBox(w int) string {
	boxW := clamp(44, 72, w-4)
	titleText := "ADD CUSTOM FONT"
	if m.fontInputMode == "edit" {
		titleText = "EDIT CUSTOM FONT"
	}
	body := sectionTitle(titleText) +
		"\nEnter font family name exactly as your terminal sees it." +
		"\n\n" + feedbackInput("Font family", m.fontInputText, true, max(24, boxW-4)) +
		"\n\nEnter save  Esc cancel"
	return cardHot.Width(boxW).Render(fitBlock(body, max(24, boxW-4), 10))
}

func feedbackInput(name string, value string, active bool, width int) string {
	marker := "  "
	border := lipgloss.Color("#3F3F46")
	if active {
		marker = "❯ "
		border = cyan
	}
	display := strings.TrimSpace(value)
	if display == "" {
		display = label.Render("type here")
	}
	fieldStyle := lipgloss.NewStyle().
		Foreground(ink).
		Background(bgSoft).
		Padding(0, 1).
		Width(max(18, width-4))
	if activeBorderStyle != "none" {
		fieldStyle = fieldStyle.Border(borderForMode()).BorderForeground(border)
	}
	field := fieldStyle.Render(fitBlock(display, max(10, width-8), 3))
	return marker + accent.Render(name) + "\n" + field
}

func (m Model) confirmBox(w int) string {
	boxW := clamp(44, 72, w-4)
	if m.pending.kind == "install-font" {
		body := sectionTitle("INSTALL FONT") +
			"\nInstall: " + m.pending.font +
			"\n\nTermix will run the supported installer for this Nerd Font." +
			"\nNo install runs until you confirm." +
			"\n\nY/Enter install  Esc/N cancel"
		return cardHot.Width(boxW).Render(fitBlock(body, max(24, boxW-4), 11))
	}
	body := sectionTitle("APPLY THEME") +
		"\nTheme: " + m.pending.theme.Name +
		"\n\nChoose profile:\n" + m.confirmProfileList() +
		"\n\n↑↓ select  Enter/Y apply  Esc/N cancel"
	return cardHot.Width(boxW).Render(fitBlock(body, max(24, boxW-4), 13))
}

func (m Model) confirmProfileList() string {
	targets := profileTargets(m.rt)
	if len(targets) == 0 {
		return "  no profiles"
	}
	selected := clamp(0, len(targets)-1, m.confirmIndex)
	start := clamp(0, max(0, len(targets)-6), selected-3)
	end := min(len(targets), start+6)
	var rows []string
	for i := start; i < end; i++ {
		target := targets[i]
		prefix := "  "
		if i == selected {
			prefix = "❯ "
		}
		rows = append(rows, prefix+fitLine(fmt.Sprintf("%-18s %s", target.Name, setupProfileState(target)), 44))
	}
	return strings.Join(rows, "\n")
}

func (m Model) currentConfirmProfileTarget() profileTarget {
	targets := profileTargets(m.rt)
	if len(targets) == 0 {
		return profileTarget{}
	}
	return targets[clamp(0, len(targets)-1, m.confirmIndex)]
}

func startupShell() string {
	return `TERMIX
Modern Terminal Experience Manager
Initializing workspace...`
}

func previewText() string {
	return " Admin   ~/workspace\n  main ✔\n PowerShell UTF8 ANSI\n╰─❯"
}

func selectedText(items []string, index int) string {
	if len(items) == 0 {
		return ""
	}
	return items[clamp(0, len(items)-1, index)]
}

func buildFontChoices(cfg config.Config) []string {
	out := append([]string{}, fontpkg.Choices()...)
	out = append(out, cfg.CustomFonts...)
	return uniqueLocal(out)
}

func uniqueLocal(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func replaceStringFold(values []string, oldValue, newValue string) []string {
	out := append([]string{}, values...)
	for i, value := range out {
		if strings.EqualFold(value, oldValue) {
			out[i] = newValue
			return out
		}
	}
	return append(out, newValue)
}

func removeStringFold(values []string, target string) []string {
	var out []string
	for _, value := range values {
		if !strings.EqualFold(value, target) {
			out = append(out, value)
		}
	}
	return out
}

func isRecommendedNerdFont(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "nerd font")
}

func setupList(items []string, selected int) string {
	if len(items) == 0 {
		return "no items"
	}
	selected = clamp(0, len(items)-1, selected)
	start := clamp(0, max(0, len(items)-7), selected-3)
	end := min(len(items), start+7)
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		prefix := "  "
		if i == selected {
			prefix = "❯ "
		}
		lines = append(lines, prefix+items[i])
	}
	return strings.Join(lines, "\n")
}

func setupProfileState(target profileTarget) string {
	if !target.Supported {
		return "view only"
	}
	if target.Installed {
		return "ready"
	}
	return "missing"
}

func firstOrDefault(values []string, fallback string) string {
	if len(values) == 0 || values[0] == "" {
		return fallback
	}
	return values[0]
}

func sectionTitle(s string) string { return title.Render(s) }

func stateIcon(healthy bool) string {
	if healthy {
		return ok.Render("✓")
	}
	return warn.Render("!")
}

func statusLine(name string, healthy bool) string {
	if healthy {
		return fmt.Sprintf("%-14s %s", name, ok.Render("✓"))
	}
	return fmt.Sprintf("%-14s %s", name, bad.Render("×"))
}

func badgeRow(values ...string) string {
	var out []string
	for _, value := range values {
		out = append(out, pill.Render(value))
	}
	return strings.Join(out, " ")
}

func cappedLogs(lines []string) []string {
	if len(lines) <= 100 {
		return lines
	}
	return lines[:100]
}

func fitBlock(s string, width, height int) string {
	width = max(1, width)
	height = max(1, height)
	raw := strings.Split(s, "\n")
	out := make([]string, 0, height)
	for _, line := range raw {
		if len(out) >= height {
			break
		}
		out = append(out, fitLine(line, width))
	}
	for len(out) < height {
		out = append(out, "")
	}
	return strings.Join(out, "\n")
}

func fitScreen(s string, width, height int) string {
	width = max(1, width)
	height = max(1, height)
	raw := strings.Split(s, "\n")
	out := make([]string, 0, height)
	for _, line := range raw {
		if len(out) >= height {
			break
		}
		out = append(out, fitLineClip(line, width))
	}
	for len(out) < height {
		out = append(out, "")
	}
	return strings.Join(out, "\n")
}

func fitLine(s string, width int) string {
	width = max(1, width)
	if strings.TrimSpace(s) == "" || lipgloss.Width(s) == 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width < 4 {
		return ansi.Truncate(s, width, "")
	}
	return ansi.Truncate(s, width, "…")
}

func fitLineClip(s string, width int) string {
	width = max(1, width)
	if strings.TrimSpace(s) == "" || lipgloss.Width(s) == 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, "")
}

func joinFit(parts []string, sep string, w int) string {
	out := strings.Join(parts, sep)
	if lipgloss.Width(out) <= w {
		return out
	}
	for len(parts) > 1 && lipgloss.Width(strings.Join(parts, sep)) > w {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, sep)
}

func clampActivityHeight(totalHeight, rows int) int {
	maxRows := max(4, totalHeight*4/10)
	maxRows = min(maxRows, max(4, totalHeight-11))
	return clamp(4, maxRows, rows)
}

func firstPositive(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(lo, hi, v int) int {
	return max(lo, min(hi, v))
}

func minFocus(a, b focusArea) focusArea {
	if a < b {
		return a
	}
	return b
}

func maxFocus(a, b focusArea) focusArea {
	if a > b {
		return a
	}
	return b
}

func isSetupScreen(s screen) bool {
	return s >= screenSetupShell
}

func screenName(s screen) string {
	for _, item := range nav {
		if item.view == s {
			return item.label
		}
	}
	switch s {
	case screenSetupShell:
		return "Choose Terminal"
	case screenSetupFont:
		return "Choose Font"
	case screenSetupTheme:
		return "Choose Theme"
	case screenSetupPreview:
		return "Live Preview"
	case screenSetupApply:
		return "Apply Setup"
	default:
		return "Dashboard"
	}
}

func focusName(f focusArea) string {
	switch f {
	case focusSidebar:
		return "Sidebar"
	case focusPreview:
		return "Preview"
	case focusLogs:
		return "Logs"
	default:
		return "Content"
	}
}
