// this file handles the one-time first-run onboarding flow.
package dawnfetch

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const firstRunSuggestedTheme = "transgender"
const firstRunAuthor = "almightynan"
const firstRunGitHubURL = "https://github.com/almightynan"

var firstRunTransColors = []string{
	"38;2;91;206;250",  // blue
	"38;2;245;169;184", // pink
	"97",               // white
	"38;2;245;169;184", // pink
	"38;2;91;206;250",  // blue
}

var firstRunWelcomeASCII = []string{
	"	      __                     ____     __       __  ",
	"  ____/ /___ __      ______  / __/__  / /______/ /_ ",
	" / __  / __ `/ | /| / / __ \\/ /_/ _ \\/ __/ ___/ __ \\",
	"/ /_/ / /_/ /| |/ |/ / / / / __/  __/ /_/ /__/ / / /",
	"\\__,_/\\__,_/ |__/|__/_/ /_/_/  \\___/\\__/\\___/_/ /_/ ",
	"                                                   ",
	"                                                   ",
}

type onboardingLine struct {
	Plain  string
	Styled string
}

func maybeRunFirstRunThemeSetup(themesPath string, noColor bool, isTTY bool) int {
	if !isTTY {
		return 0
	}

	cfg, err := loadUserConfig()
	if err != nil {
		printCLIError(fmt.Sprintf("failed to load user config: %v", err), "")
		return 1
	}
	if cfg.Initialized {
		return 0
	}
	// if a default theme already exists from older config format, assume setup as done.
	if strings.TrimSpace(cfg.DefaultTheme) != "" {
		cfg.Initialized = true
		if err := saveUserConfig(cfg); err != nil {
			printCLIError(fmt.Sprintf("failed to save user config: %v", err), "")
			return 1
		}
		return 0
	}

	showFirstRunWelcome(noColor)

	chosen, code := runThemeSelectionInteractive(themesPath, noColor, firstRunSuggestedTheme)
	if code != 0 {
		return code
	}

	cfg.Initialized = true
	if name := strings.TrimSpace(chosen); name != "" {
		cfg.DefaultTheme = name
	}
	if err := saveUserConfig(cfg); err != nil {
		printCLIError(fmt.Sprintf("failed to save user config: %v", err), "")
		return 1
	}
	clearOnboardingScreen(noColor)
	return 0
}

func showFirstRunWelcome(noColor bool) {
	ansiOK := false
	if !noColor {
		// enable ansi early so the welcome screen can render nicely on supported terminals.
		ansiOK = enableANSIIfSupported()
	}
	useColor := !noColor && ansiOK

	width := terminalWidth()
	if width <= 0 {
		width = 100
	}
	height := getTerminalHeight()
	if height <= 0 {
		height = 30
	}

	intro := []string{
		"",
		"welcome to dawnfetch! <3",
		"",
		"dawnfetch is a fast, themed, cross-platform system info tool.",
		"before your first run, you'll choose a default theme.",
		"",
		"press enter to continue, ^C to exit",
	}

	lines := make([]string, 0, len(firstRunWelcomeASCII)+len(intro))
	lines = append(lines, firstRunWelcomeASCII...)
	lines = append(lines, intro...)
	footer := buildOnboardingFooter(width, useColor, ansiOK)

	maxLine := 0
	for _, l := range lines {
		w := displayWidth(stripANSI(l))
		if w > maxLine {
			maxLine = w
		}
	}

	// keep the whole welcome block centered with simple, deterministic spacing.
	blockHeight := len(lines)
	topPad := (height - blockHeight) / 2
	if topPad < 0 {
		topPad = 0
	}

	if ansiOK {
		fmt.Print("\x1b[2J\x1b[H")
	}
	printedLines := 0
	for i := 0; i < topPad; i++ {
		fmt.Println()
		printedLines++
	}

	for i, raw := range lines {
		line := raw
		switch {
		case i < len(firstRunWelcomeASCII):
			if useColor {
				code := firstRunTransColors[i%len(firstRunTransColors)]
				line = colorLine(code, false, raw)
			}
		case strings.EqualFold(strings.TrimSpace(raw), "welcome to dawnfetch! <3"):
			if useColor {
				line = colorLine("1;97", false, raw)
			}
		case strings.HasPrefix(strings.ToLower(strings.TrimSpace(raw)), "dawnfetch is"):
			if useColor {
				line = colorLine("38;2;91;206;250", false, raw)
			}
		case strings.HasPrefix(strings.ToLower(strings.TrimSpace(raw)), "before your first run"):
			if useColor {
				line = colorLine("38;2;91;206;250", false, raw)
			}
		case strings.HasPrefix(strings.ToLower(strings.TrimSpace(raw)), "press enter"):
			if useColor {
				line = colorLine("90", false, raw)
			}
		default:
			if useColor && strings.TrimSpace(raw) != "" {
				line = colorLine("38;2;156;163;175", false, raw)
			}
		}

		pad := (width - maxLine) / 2
		if pad < 0 {
			pad = 0
		}
		fmt.Printf("%s%s\n", strings.Repeat(" ", pad), line)
		printedLines++
	}

	// keep footer pinned near the bottom while preserving centered onboarding content.
	footerStart := height - len(footer)
	if footerStart < printedLines+1 {
		footerStart = printedLines + 1
	}
	for printedLines < footerStart {
		fmt.Println()
		printedLines++
	}

	for _, l := range footer {
		pad := (width - displayWidth(l.Plain)) / 2
		if pad < 0 {
			pad = 0
		}
		fmt.Printf("%s%s\n", strings.Repeat(" ", pad), l.Styled)
		printedLines++
	}

	// wait for explicit confirmation so users can read the first-run message.
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func clearOnboardingScreen(noColor bool) {
	if !noColor && enableANSIIfSupported() {
		fmt.Print("\x1b[2J\x1b[H")
		return
	}
	// basic fallback for terminals without ansi clear support.
	fmt.Print(strings.Repeat("\n", 40))
}

func buildOnboardingFooter(width int, useColor bool, ansiOK bool) []onboardingLine {
	sepPlain := " | "
	sepStyled := sepPlain
	if useColor {
		sepStyled = colorLine("90", false, sepPlain)
	}

	versionPlain := fmt.Sprintf("version %s", version)
	versionStyled := versionPlain
	if useColor {
		versionStyled = colorLine("38;2;91;206;250", false, versionPlain)
	}

	authorNamePlain := firstRunAuthor
	authorNameStyled := authorNamePlain
	if useColor {
		authorNameStyled = colorLine("38;2;245;169;184", false, authorNamePlain)
	}
	if ansiOK {
		authorNameStyled = osc8Link(firstRunGitHubURL, authorNameStyled)
	}

	authorPrefix := "dawnfetch by "
	authorPlain := authorPrefix + authorNamePlain
	authorStyled := authorPrefix + authorNameStyled

	githubLabelPlain := "github:"
	githubLabelStyled := githubLabelPlain

	urlPlain := firstRunGitHubURL
	urlStyled := urlPlain
	if ansiOK {
		urlStyled = osc8Link(firstRunGitHubURL, urlStyled)
	}

	fullPrimaryPlain := versionPlain + sepPlain + authorPlain + sepPlain + githubLabelPlain + " " + urlPlain
	fullPrimaryStyled := versionStyled + sepStyled + authorStyled + sepStyled + githubLabelStyled + " " + urlStyled

	primaryPlain := versionPlain + sepPlain + authorPlain + sepPlain + githubLabelPlain
	primaryStyled := versionStyled + sepStyled + authorStyled + sepStyled + githubLabelStyled

	compactVersionPlain := fmt.Sprintf("v%s", version)
	compactVersionStyled := compactVersionPlain
	if useColor {
		compactVersionStyled = colorLine("38;2;91;206;250", false, compactVersionPlain)
	}

	compactPrimaryPlain := compactVersionPlain + sepPlain + "by " + firstRunAuthor
	compactPrimaryStyled := compactVersionStyled + sepStyled
	compactPrimaryStyled += "by " + authorNameStyled

	switch {
	case width >= 2 && displayWidth(fullPrimaryPlain) <= width-2:
		return []onboardingLine{
			{Plain: fullPrimaryPlain, Styled: fullPrimaryStyled},
		}
	case width >= 2 && displayWidth(primaryPlain) <= width-2:
		return []onboardingLine{
			{Plain: primaryPlain, Styled: primaryStyled},
			{Plain: githubLabelPlain + " " + urlPlain, Styled: githubLabelStyled + " " + urlStyled},
		}
	case width >= 2 && displayWidth(compactPrimaryPlain) <= width-2:
		return []onboardingLine{
			{Plain: compactPrimaryPlain, Styled: compactPrimaryStyled},
			{Plain: githubLabelPlain + " " + urlPlain, Styled: githubLabelStyled + " " + urlStyled},
		}
	default:
		linkLabel := "github: " + firstRunAuthor
		linkStyled := linkLabel
		return []onboardingLine{
			{Plain: versionPlain, Styled: versionStyled},
			{Plain: linkLabel, Styled: linkStyled},
		}
	}
}

func osc8Link(url, label string) string {
	if strings.TrimSpace(url) == "" || strings.TrimSpace(label) == "" {
		return label
	}
	return "\x1b]8;;" + url + "\x1b\\" + label + "\x1b]8;;\x1b\\"
}
