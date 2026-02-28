// this file prints command help and usage text.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
)

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
