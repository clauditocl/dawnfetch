// this file implements utility subcommands like doctor and theme preview.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"dawnfetch/internal/dawnfetch/config"
	"dawnfetch/internal/dawnfetch/core"
	"dawnfetch/internal/dawnfetch/render"
	"dawnfetch/internal/dawnfetch/system"
	"dawnfetch/internal/dawnfetch/tui/preview"
)

func runListThemes(themesPath string) int {
	palettes, err := config.LoadThemePalettes(themesPath)
	if err != nil {
		printCLIError(err.Error(), "")
		return 1
	}
	names := make([]string, 0, len(palettes))
	for name := range palettes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
	return 0
}

func runDoctor(themesPath string) int {
	fmt.Println("dawnfetch doctor")
	fmt.Println()

	fmt.Printf("Version            : %s\n", version)
	fmt.Printf("Platform           : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Terminal Width     : %d\n", render.TerminalWidth())
	fmt.Printf("Themes File        : %s\n", themesPath)

	palettes, err := config.LoadThemePalettes(themesPath)
	if err != nil {
		fmt.Printf("Themes Status      : error (%v)\n", err)
		return 1
	} else {
		fmt.Printf("Themes Status      : ok (%d themes)\n", len(palettes))
	}

	cfgPath, cfgPathErr := config.UserConfigPath()
	if cfgPathErr != nil {
		fmt.Printf("User Config        : error (%v)\n", cfgPathErr)
	} else {
		if _, err := os.Stat(cfgPath); err == nil {
			fmt.Printf("User Config        : %s\n", cfgPath)
		} else if os.IsNotExist(err) {
			fmt.Printf("User Config        : not set (%s)\n", cfgPath)
		} else {
			fmt.Printf("User Config        : error (%v)\n", err)
		}
	}

	defTheme := config.LoadPersistedDefaultTheme(core.DefaultPalette)
	fmt.Printf("Default Theme      : %s\n", core.ResolvePaletteName(defTheme, palettes))
	fmt.Printf("ASCII Logo Dir     : %s\n", asciiDirStatus())

	fmt.Printf("Image Formats      : png jpg jpeg webp gif bmp tiff (svg unsupported)\n")
	fmt.Printf("Detected OS        : %s\n", system.OSNameVersion())
	fmt.Printf("Detected Shell     : %s\n", system.ShellInfo(true))

	if runtime.GOOS == "linux" {
		linuxDeps := []string{"xfconf-query", "gsettings", "xrandr", "lspci", "ip"}
		fmt.Printf("Optional Tools     : %s\n", commandStatuses(linuxDeps))
	} else if runtime.GOOS == "darwin" {
		macDeps := []string{"sw_vers", "sysctl", "system_profiler"}
		fmt.Printf("Optional Tools     : %s\n", commandStatuses(macDeps))
	} else if runtime.GOOS == "windows" {
		winDeps := []string{"powershell", "wmic"}
		fmt.Printf("Optional Tools     : %s\n", commandStatuses(winDeps))
	}

	return 0
}

func runPreviewTheme(themeName string, themesPath string, noColor bool) int {
	if stdoutIsTerminal() {
		chosen, err := preview.RunThemeSelectionInteractive(themesPath, noColor, themeName)
		if err != nil {
			printCLIError(err.Error(), "")
			return 1
		}
		if strings.TrimSpace(chosen) != "" {
			return runSetDefaultTheme(chosen, themesPath)
		}
		return 0
	}
	if strings.TrimSpace(themeName) == "" {
		printCLIError("preview-theme needs a theme name in non-interactive mode", "run `dawnfetch preview-theme <name>`")
		return 2
	}
	return runPreviewThemeStatic(themeName, themesPath, noColor)
}

func runPreviewThemeStatic(themeName string, themesPath string, noColor bool) int {
	palettes, err := config.LoadThemePalettes(themesPath)
	if err != nil {
		printCLIError(err.Error(), "")
		return 1
	}
	name, ok := core.ResolvePaletteNameStrict(themeName, palettes)
	if !ok {
		printCLIError(fmt.Sprintf("unknown theme %q", strings.TrimSpace(themeName)), "run `dawnfetch --list-themes` to see valid themes")
		return 2
	}
	p := palettes[name]
	fmt.Printf("Theme Preview: %s\n\n", name)
	for i, c := range p {
		label := fmt.Sprintf("Color %d", i+1)
		fmt.Println(render.ColorLine(c, noColor, label))
	}

	demo := []core.Field{
		{Label: "Operating System", Value: "Demo OS 1.0"},
		{Label: "Kernel", Value: "demo-kernel-0.1"},
		{Label: "Shell", Value: "demo-shell"},
		{Label: "Memory", Value: "1.2GiB / 8.0GiB"},
	}
	style := core.DefaultStyleConfig()
	style.Fields.Colorize = true
	style.Text.ShowUserHost = false
	fmt.Println()
	for _, line := range render.RenderInfoLines(demo, style, render.LabelWidth(demo), 44, p, noColor) {
		fmt.Println(line.Styled)
	}
	for _, sw := range render.PaletteSwatchLines(noColor, 44, style) {
		fmt.Println(sw)
	}
	return 0
}

func runSetDefaultTheme(themeName string, themesPath string) int {
	palettes, err := config.LoadThemePalettes(themesPath)
	if err != nil {
		printCLIError(err.Error(), "")
		return 1
	}
	name, ok := core.ResolvePaletteNameStrict(themeName, palettes)
	if !ok {
		printCLIError(fmt.Sprintf("unknown theme %q", strings.TrimSpace(themeName)), "run `dawnfetch --list-themes` to see valid themes")
		return 2
	}
	cfg, err := config.LoadUserConfig()
	if err != nil {
		printCLIError(fmt.Sprintf("failed to load user config: %v", err), "")
		return 1
	}
	cfg.DefaultTheme = name
	cfg.Initialized = true
	if err := config.SaveUserConfig(cfg); err != nil {
		printCLIError(fmt.Sprintf("failed to save user config: %v", err), "")
		return 1
	}
	path, _ := config.UserConfigPath()
	fmt.Printf("default theme set to %q (%s)\n", name, path)
	return 0
}

func asciiDirStatus() string {
	for _, path := range config.LogoTextDirCandidates() {
		ents, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		total := 0
		for _, e := range ents {
			if e.IsDir() {
				continue
			}
			if strings.EqualFold(filepath.Ext(e.Name()), ".txt") {
				total++
			}
		}
		return fmt.Sprintf("%s (%d txt logos)", path, total)
	}
	return "unavailable (no ascii logo directory found)"
}

func commandStatuses(names []string) string {
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, n := range names {
		status := "missing"
		if system.CommandExists(n) {
			status = "ok"
		}
		parts = append(parts, n+"="+status)
	}
	return strings.Join(parts, ", ")
}
