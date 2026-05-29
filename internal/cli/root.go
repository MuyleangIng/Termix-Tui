package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muyleanging/termix/internal/app"
	"github.com/muyleanging/termix/internal/config"
	"github.com/muyleanging/termix/internal/doctor"
	"github.com/muyleanging/termix/internal/font"
	"github.com/muyleanging/termix/internal/installer"
	"github.com/muyleanging/termix/internal/profile"
	"github.com/muyleanging/termix/internal/shell"
	"github.com/muyleanging/termix/internal/theme"
	"github.com/muyleanging/termix/internal/tui"
	"github.com/muyleanging/termix/internal/uninstaller"
	"github.com/muyleanging/termix/internal/updater"
	"github.com/spf13/cobra"
)

var cfgPath string

func Execute() error {
	return ExecuteWithVersion("dev", "none", "unknown")
}

func ExecuteWithVersion(version, commit, date string) error {
	root := &cobra.Command{
		Use:     "termix",
		Short:   "Termix - Modern Terminal Experience Manager",
		Long:    "Termix is a premium CLI and TUI platform for shells, Oh My Posh, fonts, Windows Terminal, WSL, and terminal profiles.",
		Version: fmt.Sprintf("%s\ncommit: %s\nbuilt: %s", version, commit, date),
		RunE: func(cmd *cobra.Command, args []string) error {
			if isTUIExecutable(os.Args[0]) {
				return runTUI(cmd.Context())
			}
			return cmd.Help()
		},
	}

	root.PersistentFlags().StringVar(&cfgPath, "config", "", "config file path")
	root.AddCommand(tuiCommand(), setupCommand(), doctorCommand(), installCommand(), updateCommand(), uninstallCommand(), cleanCommand(), resetCommand(), reinstallCommand(), repairCommand(), cacheCommand(), themesCommand(), fontsCommand(), profileCommand(), applyCommand())
	return root.Execute()
}

func isTUIExecutable(path string) bool {
	if index := strings.LastIndexAny(path, `/\`); index >= 0 {
		path = path[index+1:]
	}
	name := strings.TrimSuffix(strings.ToLower(path), ".exe")
	return name == "termix-tui"
}

func tuiCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "tui",
		Aliases: []string{"ui", "app"},
		Short:   "Open the Termix full-screen TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI(cmd.Context())
		},
	}
}

func applyCommand() *cobra.Command {
	var themeName string
	var shellName string
	var fontName string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the selected theme, shell profile, and terminal font",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if shellName == "" {
				shellName = rt.Config.DefaultShell
			}
			if fontName == "" {
				fontName = rt.Config.DefaultFont
			}
			themes, err := theme.NewManager(rt.Config).Scan(cmd.Context())
			if err != nil {
				return err
			}
			themePath := ""
			for _, item := range themes {
				if item.Name == themeName {
					themePath = item.Path
					break
				}
			}
			if themePath == "" {
				return fmt.Errorf("theme %q not found; run termix themes", themeName)
			}
			home, _ := os.UserHomeDir()
			if err := profile.ApplyPrompt(home, shellName, themePath); err != nil {
				return err
			}
			_ = profile.ApplyWindowsTerminalFont(home, fontName)
			fmt.Fprintf(os.Stdout, "Applied %s to %s with %s\n", themeName, shellName, fontName)
			return nil
		},
	}
	cmd.Flags().StringVar(&themeName, "theme", "catppuccin_mocha", "Oh My Posh theme name")
	cmd.Flags().StringVar(&shellName, "shell", "", "shell profile to configure")
	cmd.Flags().StringVar(&fontName, "font", "", "Windows Terminal font family")
	return cmd
}

func runtime(ctx context.Context) (*app.Runtime, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	return app.NewRuntime(ctx, cfg)
}

func runTUI(ctx context.Context) error {
	rt, err := runtime(ctx)
	if err != nil {
		return err
	}
	program := tea.NewProgram(tui.New(rt), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = program.Run()
	return err
}

func setupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Pick a shell profile and apply the default Termix setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			program := tea.NewProgram(tui.NewSetup(rt), tea.WithAltScreen(), tea.WithMouseCellMotion())
			_, err = program.Run()
			return err
		},
	}
}

func cleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove broken cache and temporary Termix state without deleting downloaded themes",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if err := theme.ClearCache(rt.Config); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Cleaned Termix cache metadata. Downloaded themes were kept.")
			return nil
		},
	}
}

func resetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Backup and reset Termix config while keeping themes and profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if err := uninstaller.New(rt).Uninstall(cmd.Context(), "config"); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Reset Termix config. Existing config was backed up when present.")
			return nil
		},
	}
}

func reinstallCommand() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "reinstall",
		Short: "Clean reinstall Termix-managed cache, themes, profile block, and config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepairWorkflow(cmd.Context(), dryRun, true)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show planned changes without writing files")
	return cmd
}

func repairCommand() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "repair",
		Short: "Repair CONFIG NOT FOUND, theme cache, missing theme path, and profile integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepairWorkflow(cmd.Context(), dryRun, false)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show planned changes without writing files")
	return cmd
}

func doctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run terminal, shell, WSL, Unicode, and Oh My Posh diagnostics",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			report := doctor.New(rt).Run(cmd.Context())
			fmt.Fprint(os.Stdout, report.RenderText())
			return nil
		},
	}
}

func installCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install [component]",
		Short: "Install Oh My Posh, shells, terminal tools, fonts, or themes for the current OS",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			component := "all"
			if len(args) == 1 {
				component = args[0]
			}
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			return installer.New(rt).Install(cmd.Context(), component)
		},
	}
}

func updateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Detect and update Termix-managed tools, themes, fonts, and shell configs",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			return updater.New(rt).Run(cmd.Context())
		},
	}
}

func uninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall [component]",
		Short: "Remove Termix profiles, dependencies, data, themes, and executable",
		Long: `Remove Termix-managed profiles, config, cache, downloaded themes, external dependencies, and the executable.

With no component, Termix performs a full uninstall:
  termix uninstall

To remove only one part:
  termix uninstall profile
  termix uninstall cache
  termix uninstall downloaded-themes
  termix uninstall config
  termix uninstall dependencies
  termix uninstall executable`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			component := "all"
			if len(args) == 0 {
				component = "all"
			} else {
				component = args[0]
			}
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if err := uninstaller.New(rt).Uninstall(cmd.Context(), component); err != nil {
				return err
			}
			if strings.EqualFold(component, "all") {
				fmt.Fprintln(os.Stdout, "Uninstalled Termix profiles, dependencies, cache, config, downloaded themes, and scheduled executable removal.")
			} else {
				fmt.Fprintf(os.Stdout, "Uninstalled Termix component: %s\n", component)
			}
			return nil
		},
	}
}

func themesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "themes",
		Short: "List, update, rebuild, and apply Oh My Posh themes",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			themes, err := theme.NewManager(rt.Config).Scan(cmd.Context())
			if err != nil {
				return err
			}
			for _, item := range themes {
				fmt.Fprintf(os.Stdout, "%-28s %s\n", item.Name, item.Path)
			}
			return nil
		},
	}
	cmd.AddCommand(themesUpdateCommand(), themesApplyCommand())
	return cmd
}

func themesUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Download or refresh official Oh My Posh themes and rebuild cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if _, err := theme.InstallOfficialThemes(cmd.Context(), rt.Config); err != nil {
				return err
			}
			themes, err := theme.RebuildCache(cmd.Context(), rt.Config)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Updated official themes and rebuilt cache: %d themes\n", len(themes))
			return nil
		},
	}
}

func themesApplyCommand() *cobra.Command {
	var shellName string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "apply <theme>",
		Short: "Apply a theme to a selected shell profile without duplicating Termix blocks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if shellName == "" {
				shellName = rt.Config.DefaultShell
			}
			item, err := findTheme(cmd.Context(), rt.Config, args[0])
			if err != nil {
				return err
			}
			home, _ := os.UserHomeDir()
			if dryRun {
				fmt.Fprintf(os.Stdout, "DRY RUN\nProfile: %s\nProfile file: %s\nTheme: %s\nTheme file: %s\n", shellName, profile.ProfilePath(home, shellName), item.Name, item.Path)
				return nil
			}
			if err := profile.ApplyPrompt(home, shellName, item.Path); err != nil {
				return err
			}
			if err := config.SaveSetupChoices(rt.Config, shellName, rt.Config.DefaultFont, item.Name); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Applied %s to %s\nRestart your terminal or reload the profile.\n", item.Name, shellName)
			return nil
		},
	}
	cmd.Flags().StringVar(&shellName, "profile", "", "target shell profile")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show planned profile change without writing files")
	return cmd
}

func cacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage rebuildable Termix cache metadata",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild theme cache from real theme files, downloading official themes if needed",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			themes, err := theme.RebuildCache(cmd.Context(), rt.Config)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Rebuilt theme cache: %d themes\n", len(themes))
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "clear",
		Short: "Clear cache metadata only; downloaded theme files are kept",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if err := theme.ClearCache(rt.Config); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Cleared cache metadata. Theme files were kept.")
			return nil
		},
	})
	return cmd
}

func fontsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fonts",
		Short: "List, install, and apply terminal fonts with fallback handling",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List recommended fonts and installed status",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			for _, item := range font.Detect(home) {
				status := "MISSING"
				if item.Installed {
					status = "INSTALLED"
				}
				fmt.Fprintf(os.Stdout, "%-30s %s\n", item.Name, status)
			}
			return nil
		},
	})
	cmd.AddCommand(fontsInstallCommand())
	cmd.AddCommand(fontsApplyCommand())
	return cmd
}

func fontsInstallCommand() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "install <font>",
		Short: "Install a supported Nerd Font after explicit confirmation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("install requires confirmation; rerun with --yes to install %q", args[0])
			}
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			return installer.New(rt).Install(cmd.Context(), "font:"+args[0])
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm font installation")
	return cmd
}

func fontsApplyCommand() *cobra.Command {
	var windowsTerminal bool
	cmd := &cobra.Command{
		Use:   "apply <font>",
		Short: "Save selected font in Termix config and optionally Windows Terminal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if err := config.SaveFontChoice(rt.Config, args[0]); err != nil {
				return err
			}
			home, _ := os.UserHomeDir()
			resolved := font.ResolveAvailableFamily(home, args[0])
			fmt.Fprintf(os.Stdout, "Saved Termix font: %s\nResolved available face: %s\n", args[0], resolved)
			if windowsTerminal {
				if err := profile.ApplyWindowsTerminalFont(home, args[0]); err != nil {
					return err
				}
				fmt.Fprintln(os.Stdout, "Updated Windows Terminal font with backup.")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&windowsTerminal, "windows-terminal", false, "also apply font to Windows Terminal defaults")
	return cmd
}

func profileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Inspect and repair shell profile integration",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List supported shell profiles and paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			for _, adapter := range shell.Available(home) {
				shellName := adapter.Name()
				path := adapter.ProfilePath(home)
				status := "missing"
				if profile.HasPromptBlock(home, shellName) {
					status = "integrated"
				}
				fmt.Fprintf(os.Stdout, "%-18s %-10s %s\n", shellName, status, path)
			}
			return nil
		},
	})
	cmd.AddCommand(profileRepairCommand())
	return cmd
}

func profileRepairCommand() *cobra.Command {
	var themeName string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "repair <profile>",
		Short: "Deduplicate and rewrite Termix Oh My Posh block in a shell profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtime(cmd.Context())
			if err != nil {
				return err
			}
			if themeName == "" {
				themeName = firstThemeName(rt.Config)
			}
			item, err := findTheme(cmd.Context(), rt.Config, themeName)
			if err != nil {
				return err
			}
			home, _ := os.UserHomeDir()
			if dryRun {
				fmt.Fprintf(os.Stdout, "DRY RUN\nProfile: %s\nProfile file: %s\nTheme file: %s\n", args[0], profile.ProfilePath(home, args[0]), item.Path)
				return nil
			}
			return profile.RepairPrompt(home, args[0], item.Path)
		},
	}
	cmd.Flags().StringVar(&themeName, "theme", "", "theme to write into profile")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show planned profile repair without writing files")
	return cmd
}

func runRepairWorkflow(ctx context.Context, dryRun, reinstall bool) error {
	rt, err := runtime(ctx)
	if err != nil {
		return err
	}
	home, _ := os.UserHomeDir()
	shellName := rt.Config.DefaultShell
	themeName := firstThemeName(rt.Config)
	resolvedFont := font.ResolveAvailableFamily(home, rt.Config.DefaultFont)
	if dryRun {
		fmt.Fprintf(os.Stdout, "DRY RUN\nMode: %s\nWould clear cache metadata: %t\nWould rebuild official theme cache\nWould repair profile: %s\nProfile file: %s\nTheme: %s\nResolved font: %s\n", workflowName(reinstall), reinstall, shellName, profile.ProfilePath(home, shellName), themeName, resolvedFont)
		return nil
	}
	if reinstall {
		if err := theme.ClearCache(rt.Config); err != nil {
			return err
		}
	}
	themes, err := theme.RebuildCache(ctx, rt.Config)
	if err != nil {
		return err
	}
	item, err := findThemeInList(themes, themeName)
	if err != nil {
		return err
	}
	if err := profile.RepairPrompt(home, shellName, item.Path); err != nil {
		return err
	}
	if err := config.SaveSetupChoices(rt.Config, shellName, resolvedFont, item.Name); err != nil {
		return err
	}
	if err := config.MarkSetupComplete(rt.Config.HomeDir); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "SUCCESS %s complete\nProfile: %s\nTheme: %s\nTheme file: %s\nFont: %s\nTheme cache: %d themes\nNext step: restart your terminal or reload the profile.\n", strings.Title(workflowName(reinstall)), shellName, item.Name, item.Path, resolvedFont, len(themes))
	return nil
}

func findTheme(ctx context.Context, cfg config.Config, name string) (theme.Theme, error) {
	themes, err := theme.NewManager(cfg).Scan(ctx)
	if err != nil {
		return theme.Theme{}, err
	}
	return findThemeInList(themes, name)
}

func findThemeInList(themes []theme.Theme, name string) (theme.Theme, error) {
	for _, item := range themes {
		if strings.EqualFold(item.Name, name) {
			return item, nil
		}
	}
	return theme.Theme{}, fmt.Errorf("theme %q not found; run termix themes update or termix cache rebuild", name)
}

func firstThemeName(cfg config.Config) string {
	if len(cfg.FavoriteThemes) > 0 && cfg.FavoriteThemes[0] != "" {
		return cfg.FavoriteThemes[0]
	}
	return "catppuccin_mocha"
}

func workflowName(reinstall bool) string {
	if reinstall {
		return "reinstall"
	}
	return "repair"
}
