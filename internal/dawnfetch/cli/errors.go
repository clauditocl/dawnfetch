// this file centralizes cli error formatting and parse error handling.
package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

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
