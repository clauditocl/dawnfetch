// this file parses subcommand-specific flags and dispatches command handlers.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func runHelpCommand(args []string) int {
	initUsage(flag.CommandLine)
	if len(args) == 0 {
		printMainHelp(os.Stdout)
		return 0
	}
	topic := strings.ToLower(strings.TrimSpace(args[0]))
	switch topic {
	case "themes", "list-themes":
		fmt.Println("Themes:")
		fmt.Println("  dawnfetch list-themes [--themes path]")
		fmt.Println("  Reads themes from themes.json (or custom path) and prints available names.")
		return 0
	case "doctor":
		fmt.Println("Doctor:")
		fmt.Println("  dawnfetch doctor [--themes path]")
		fmt.Println("  Runs quick diagnostics for config, themes, logos, and optional dependencies.")
		return 0
	case "preview-theme":
		fmt.Println("Preview Theme:")
		fmt.Println("  dawnfetch preview-theme [name] [--themes path] [--no-color]")
		fmt.Println("  Opens interactive preview (search + arrows + live centered output).")
		fmt.Println("  In non-interactive mode, pass a theme name for plain text preview.")
		return 0
	case "set-default-theme":
		fmt.Println("Set Default Theme:")
		fmt.Println("  dawnfetch set-default-theme <name> [--themes path]")
		fmt.Println("  Persists the default theme in your user config directory.")
		return 0
	case "version":
		fmt.Println("Version:")
		fmt.Println("  dawnfetch version")
		return 0
	default:
		printCLIError(fmt.Sprintf("unknown help topic %q", topic), "run `dawnfetch help` to see available topics")
		return 2
	}
}

func runListThemesCommand(args []string) int {
	fs := flag.NewFlagSet("list-themes", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	themesPath := fs.String("themes", "themes.json", "path to theme palettes json")
	if err := fs.Parse(args); err != nil {
		return handleFlagParseError(err, "run `dawnfetch help list-themes` for command usage")
	}
	return runListThemes(*themesPath)
}

func runDoctorCommand(args []string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	themesPath := fs.String("themes", "themes.json", "path to theme palettes json")
	if err := fs.Parse(args); err != nil {
		return handleFlagParseError(err, "run `dawnfetch help doctor` for command usage")
	}
	return runDoctor(*themesPath)
}

func runPreviewThemeCommand(args []string) int {
	fs := flag.NewFlagSet("preview-theme", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	themesPath := fs.String("themes", "themes.json", "path to theme palettes json")
	noColor := fs.Bool("no-color", false, "disable ANSI colors")
	themeName := fs.String("theme", "", "theme palette")
	if err := fs.Parse(args); err != nil {
		return handleFlagParseError(err, "run `dawnfetch help preview-theme` for command usage")
	}
	name := strings.TrimSpace(*themeName)
	if name == "" && fs.NArg() > 0 {
		name = strings.TrimSpace(fs.Arg(0))
	}
	return runPreviewTheme(name, *themesPath, *noColor)
}

func runSetDefaultThemeCommand(args []string) int {
	fs := flag.NewFlagSet("set-default-theme", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	themesPath := fs.String("themes", "themes.json", "path to theme palettes json")
	themeName := fs.String("theme", "", "theme palette")
	if err := fs.Parse(args); err != nil {
		return handleFlagParseError(err, "run `dawnfetch help set-default-theme` for command usage")
	}
	name := strings.TrimSpace(*themeName)
	if name == "" && fs.NArg() > 0 {
		name = strings.TrimSpace(fs.Arg(0))
	}
	if name == "" {
		printCLIError("set-default-theme requires a theme name", "run `dawnfetch help set-default-theme`")
		return 2
	}
	return runSetDefaultTheme(name, *themesPath)
}
