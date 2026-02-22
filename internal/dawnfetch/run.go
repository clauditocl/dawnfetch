// this file parses cli flags and coordinates one dawnfetch execution.
package dawnfetch

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
)

var version = "0.1.1"

func Run() int {
	args := os.Args
	if len(args) > 1 && strings.EqualFold(strings.TrimSpace(args[1]), "dawnfetch") {
		args = append([]string{args[0]}, args[2:]...)
	}
	args = normalizePreviewArgs(args)

	if len(args) > 1 {
		first := strings.ToLower(strings.TrimSpace(args[1]))
		if first == "--help" || first == "-h" || first == "-help" {
			printMainHelp(os.Stdout)
			return 0
		}
		cmd := strings.ToLower(strings.TrimSpace(args[1]))
		if !strings.HasPrefix(cmd, "-") {
			switch cmd {
			case "help":
				return runHelpCommand(args[2:])
			case "list-themes", "themes":
				return runListThemesCommand(args[2:])
			case "doctor":
				return runDoctorCommand(args[2:])
			case "set-default-theme":
				return runSetDefaultThemeCommand(args[2:])
			case "version", "--version", "-v":
				fmt.Println("dawnfetch", version)
				return 0
			default:
				printCLIError(fmt.Sprintf("unknown command %q", strings.TrimSpace(args[1])), "run `dawnfetch --help` to list commands")
				return 2
			}
		}
	}

	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	flag.CommandLine.Usage = func() {}

	themeName := flag.String("theme", "", "theme palette")
	themesPath := flag.String("themes", "themes.json", "path to theme palettes json")
	noColor := flag.Bool("no-color", false, "disable ANSI colors")
	noLogo := flag.Bool("no-logo", false, "disable logo rendering")
	full := flag.Bool("full", false, "run slower probes for more complete output")
	imagePath := flag.String("image", "", "path to an image logo (png/jpg/webp/gif/bmp/tiff; no svg)")

	helpFlag := flag.Bool("help", false, "show help")
	listThemesFlag := flag.Bool("list-themes", false, "list available themes and exit")
	doctorFlag := flag.Bool("doctor", false, "run environment diagnostics and exit")
	versionFlag := flag.Bool("version", false, "show version and exit")
	previewThemeFlag := flag.Bool("preview-theme", false, "open interactive preview")
	setDefaultThemeFlag := flag.String("set-default-theme", "", "set persistent default theme and exit")

	if err := flag.CommandLine.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			printMainHelp(os.Stdout)
			return 0
		}
		return handleFlagParseError(err, "run `dawnfetch --help` for usage")
	}
	if *helpFlag {
		printMainHelp(os.Stdout)
		return 0
	}
	if *versionFlag {
		fmt.Println("dawnfetch", version)
		return 0
	}
	if *listThemesFlag {
		return runListThemes(*themesPath)
	}
	if *doctorFlag {
		return runDoctor(*themesPath)
	}
	if *previewThemeFlag {
		name := strings.TrimSpace(*themeName)
		if name == "" && flag.NArg() > 0 {
			name = strings.TrimSpace(flag.Arg(0))
		}
		return runPreviewTheme(name, *themesPath, *noColor)
	}
	if strings.TrimSpace(*setDefaultThemeFlag) != "" {
		return runSetDefaultTheme(strings.TrimSpace(*setDefaultThemeFlag), *themesPath)
	}

	runFast := !*full
	isTTY := stdoutIsTerminal()
	if code := maybeRunFirstRunThemeSetup(*themesPath, *noColor, isTTY); code != 0 {
		return code
	}
	// Non-interactive fast path for benchmarks/pipes.
	nonInteractiveFast := runFast && !isTTY
	if nonInteractiveFast {
		*noLogo = true
		*noColor = true
	}

	config := defaultBrandConfig()
	skipPaletteLoad := nonInteractiveFast &&
		strings.TrimSpace(*themeName) == "" &&
		strings.TrimSpace(*themesPath) == "themes.json"
	if !skipPaletteLoad {
		if palettes, err := loadThemePalettes(*themesPath); err == nil {
			config.Palettes = palettes
		} else {
			printCLIError(err.Error(), "")
			return 1
		}
	}

	styleCfg := defaultStyleConfig()
	if *noLogo {
		styleCfg.Layout.ShowLogo = false
	}
	if !*noColor && isTTY && !enableANSIIfSupported() {
		*noColor = true
	}

	effectiveTheme := strings.TrimSpace(*themeName)
	if effectiveTheme == "" {
		if skipPaletteLoad {
			effectiveTheme = defaultPalette
		} else {
			cfg, err := loadUserConfig()
			if err != nil {
				printCLIError(fmt.Sprintf("failed to load user config: %v", err), "")
				return 1
			}
			if t := strings.TrimSpace(cfg.DefaultTheme); t != "" {
				effectiveTheme = t
			} else {
				effectiveTheme = defaultPalette
			}
		}
	} else {
		normalized, ok := resolvePaletteNameStrict(effectiveTheme, config.Palettes)
		if !ok {
			printCLIError(fmt.Sprintf("unknown theme %q", effectiveTheme), "run `dawnfetch --list-themes` to see valid themes")
			return 2
		}
		effectiveTheme = normalized
	}

	collected := collect(runFast, *full)
	printInfo(collected, config, styleCfg, effectiveTheme, *imagePath, *noColor, *noLogo)
	return 0
}

func normalizePreviewArgs(args []string) []string {
	if len(args) < 2 {
		return args
	}
	if !strings.EqualFold(strings.TrimSpace(args[1]), "preview-theme") {
		return args
	}
	out := make([]string, 0, len(args)+2)
	out = append(out, args[0], "--preview-theme")
	rest := args[2:]
	if len(rest) > 0 && !strings.HasPrefix(strings.TrimSpace(rest[0]), "-") {
		out = append(out, "--theme", strings.TrimSpace(rest[0]))
		rest = rest[1:]
	}
	out = append(out, rest...)
	return out
}

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

func runListThemes(themesPath string) int {
	palettes, err := loadThemePalettes(themesPath)
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

func initUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		printMainHelp(os.Stdout)
	}
}

func printMainHelp(out io.Writer) {
	fmt.Fprintln(out, "dawnfetch is a fast, themed, cross-platform system info CLI.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  $ dawnfetch [flags]")
	fmt.Fprintln(out, "  $ dawnfetch [command]")
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Commands")
	fmt.Fprintln(out)
	printHelpRows(out, []helpRow{
		{Name: "help [topic]", Desc: "Show help for dawnfetch or a specific topic"},
		{Name: "list-themes", Desc: "List all available themes"},
		{Name: "themes", Desc: "Alias for list-themes"},
		{Name: "preview-theme [name]", Desc: "Interactive theme preview with search"},
		{Name: "set-default-theme <name>", Desc: "Persist default theme in user config"},
		{Name: "doctor", Desc: "Run diagnostics for platform, themes, and optional tools"},
		{Name: "version", Desc: "Show dawnfetch version"},
	})
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Flags")
	fmt.Fprintln(out)
	printHelpRows(out, []helpRow{
		{Name: "--theme <name>", Desc: "Theme palette for this run"},
		{Name: "--themes <path>", Desc: "Path to theme palettes JSON"},
		{Name: "--image <path>", Desc: "Use image logo (png/jpg/webp/gif/bmp/tiff)"},
		{Name: "--full", Desc: "Collect more detailed/slower fields"},
		{Name: "--no-color", Desc: "Disable ANSI colors"},
		{Name: "--no-logo", Desc: "Disable logo rendering"},
		{Name: "--help", Desc: "Show this help output"},
		{Name: "--list-themes", Desc: "Run list-themes command"},
		{Name: "--preview-theme <name>", Desc: "Run preview-theme with initial selection"},
		{Name: "--set-default-theme <name>", Desc: "Run set-default-theme command"},
		{Name: "--doctor", Desc: "Run doctor command"},
		{Name: "--version", Desc: "Show dawnfetch version"},
	})
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Examples")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Show default output")
	fmt.Fprintln(out, "  $ dawnfetch")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Run with a custom theme")
	fmt.Fprintln(out, "  $ dawnfetch --theme transgender")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Open interactive theme preview")
	fmt.Fprintln(out, "  $ dawnfetch preview-theme")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Open interactive preview with initial selection")
	fmt.Fprintln(out, "  $ dawnfetch preview-theme transmasc")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Save your default theme")
	fmt.Fprintln(out, "  $ dawnfetch set-default-theme nonbinary")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  Run diagnostics")
	fmt.Fprintln(out, "  $ dawnfetch doctor")
}

type helpRow struct {
	Name string
	Desc string
}

func printHelpRows(out io.Writer, rows []helpRow) {
	width := 0
	for _, row := range rows {
		if n := len(row.Name); n > width {
			width = n
		}
	}
	for _, row := range rows {
		fmt.Fprintf(out, "  %-*s  %s\n", width, row.Name, row.Desc)
	}
}

func printCLIError(msg string, hint string) {
	if cliErrorColorEnabled() {
		fmt.Fprintf(os.Stderr, "\x1b[1;31merror:\x1b[0m %s\n", strings.TrimSpace(msg))
		if strings.TrimSpace(hint) != "" {
			fmt.Fprintf(os.Stderr, "\x1b[36mhint:\x1b[0m %s\n", strings.TrimSpace(hint))
		}
		return
	}
	fmt.Fprintf(os.Stderr, "error: %s\n", strings.TrimSpace(msg))
	if strings.TrimSpace(hint) != "" {
		fmt.Fprintf(os.Stderr, "hint: %s\n", strings.TrimSpace(hint))
	}
}

func cliErrorColorEnabled() bool {
	if !stderrIsTerminal() {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("NO_COLOR")), "1") {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb") {
		return false
	}
	// avoid raw escape sequences on older windows shells before ansi is initialized
	if runtime.GOOS == "windows" {
		return false
	}
	return true
}

func stderrIsTerminal() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func handleFlagParseError(err error, hint string) int {
	if err == nil {
		return 0
	}
	msg := strings.TrimSpace(err.Error())
	if strings.HasPrefix(msg, "flag provided but not defined:") {
		unknown := strings.TrimSpace(strings.TrimPrefix(msg, "flag provided but not defined:"))
		if unknown == "" {
			unknown = "(unknown)"
		}
		printCLIError(fmt.Sprintf("unknown option %s", unknown), hint)
		return 2
	}
	printCLIError(msg, hint)
	return 2
}
