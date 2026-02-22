// this file provides an interactive theme preview tui.
package dawnfetch

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type previewThemeModel struct {
	palettes    map[string][]string
	themes      []string
	filtered    []string
	selected    int
	confirmOpen bool
	confirmIdx  int
	confirmName string
	width       int
	height      int
	noColor     bool
	search      textinput.Model
	fields      []Field
	chosenTheme string
}

func runPreviewThemeInteractive(themesPath string, noColor bool, initial string) int {
	chosen, code := runThemeSelectionInteractive(themesPath, noColor, initial)
	if code != 0 {
		return code
	}
	if strings.TrimSpace(chosen) != "" {
		return runSetDefaultTheme(chosen, themesPath)
	}
	return 0
}

func runThemeSelectionInteractive(themesPath string, noColor bool, initial string) (string, int) {
	palettes, err := loadThemePalettes(themesPath)
	if err != nil {
		printCLIError(err.Error(), "")
		return "", 1
	}
	if len(palettes) == 0 {
		printCLIError("no themes available", "")
		return "", 1
	}

	model := newPreviewThemeModel(palettes, noColor, initial)
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		printCLIError(fmt.Sprintf("failed to run preview tui: %v", err), "")
		return "", 1
	}
	m, ok := finalModel.(previewThemeModel)
	if !ok {
		return "", 0
	}
	return strings.TrimSpace(m.chosenTheme), 0
}

func newPreviewThemeModel(palettes map[string][]string, noColor bool, initial string) previewThemeModel {
	names := make([]string, 0, len(palettes))
	for name := range palettes {
		names = append(names, name)
	}
	sort.Strings(names)

	s := textinput.New()
	s.Prompt = "search> "
	s.Placeholder = "type to filter themes"
	s.CharLimit = 120
	s.Width = 28
	s.Focus()

	m := previewThemeModel{
		palettes: palettes,
		themes:   names,
		noColor:  noColor,
		search:   s,
		fields:   collect(true, false),
	}
	if len(m.fields) == 0 {
		m.fields = []Field{
			{Label: "Operating System", Value: "Demo OS"},
			{Label: "Kernel", Value: "demo-kernel"},
			{Label: "Shell", Value: "demo-shell"},
			{Label: "Memory", Value: "1.2GiB / 8.0GiB"},
		}
	}

	// keep full theme list visible; use initial only as selected row.
	m.refreshFiltered(initial)
	return m
}

func (m previewThemeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *previewThemeModel) refreshFiltered(prefer string) {
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	prefer = strings.ToLower(strings.TrimSpace(prefer))

	m.filtered = m.filtered[:0]
	for _, t := range m.themes {
		if query == "" || strings.Contains(strings.ToLower(t), query) {
			m.filtered = append(m.filtered, t)
		}
	}
	if len(m.filtered) == 0 {
		m.selected = 0
		return
	}
	if prefer != "" {
		for i, t := range m.filtered {
			if strings.EqualFold(t, prefer) {
				m.selected = i
				return
			}
		}
	}
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
}

func (m previewThemeModel) selectedTheme() string {
	if len(m.filtered) == 0 {
		return ""
	}
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return ""
	}
	return m.filtered[m.selected]
}

func (m previewThemeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width < 40 {
			m.search.Width = 16
		} else if m.width < 90 {
			m.search.Width = 24
		} else {
			m.search.Width = 32
		}
		return m, nil
	case tea.KeyMsg:
		if m.confirmOpen {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "left", "h", "up", "k", "tab", "shift+tab", "right", "l", "down", "j":
				if m.confirmIdx == 0 {
					m.confirmIdx = 1
				} else {
					m.confirmIdx = 0
				}
				return m, nil
			case "y":
				m.chosenTheme = strings.TrimSpace(m.confirmName)
				m.confirmOpen = false
				return m, tea.Quit
			case "n", "esc", "q":
				m.confirmOpen = false
				m.confirmName = ""
				return m, nil
			case "enter":
				if m.confirmIdx == 0 {
					m.chosenTheme = strings.TrimSpace(m.confirmName)
					m.confirmOpen = false
					return m, tea.Quit
				}
				m.confirmOpen = false
				m.confirmName = ""
				return m, nil
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			name := strings.TrimSpace(m.selectedTheme())
			if name != "" {
				m.confirmOpen = true
				m.confirmIdx = 0
				m.confirmName = name
			}
			return m, nil
		case "up", "k", "ctrl+p", "left":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "j", "ctrl+n", "right":
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
			return m, nil
		case "pgup":
			m.selected -= 10
			if m.selected < 0 {
				m.selected = 0
			}
			return m, nil
		case "pgdown":
			m.selected += 10
			if m.selected >= len(m.filtered) {
				m.selected = len(m.filtered) - 1
			}
			if m.selected < 0 {
				m.selected = 0
			}
			return m, nil
		case "home":
			m.selected = 0
			return m, nil
		case "end":
			if len(m.filtered) > 0 {
				m.selected = len(m.filtered) - 1
			}
			return m, nil
		case "ctrl+l":
			m.search.SetValue("")
			m.refreshFiltered("")
			return m, nil
		}
	}

	before := m.search.Value()
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	if m.search.Value() != before {
		m.refreshFiltered("")
	}
	return m, cmd
}

func (m previewThemeModel) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "loading preview..."
	}

	appW := tuiMin(180, m.width-2)
	if appW < 48 {
		appW = tuiMin(m.width, 48)
	}
	appH := tuiMin(m.height-2, 44)
	if appH < 16 {
		appH = tuiMin(m.height, 16)
	}

	listW := tuiMax(22, tuiMin(30, appW/5))
	if appW-listW < 24 {
		listW = appW / 3
	}
	previewW := appW - listW - 3
	if previewW < 20 {
		previewW = 20
	}
	bodyH := appH - 8
	if bodyH < 8 {
		bodyH = 8
	}

	header := lipgloss.NewStyle().Bold(true).Render("dawnfetch theme preview")
	search := m.search.View()
	if strings.TrimSpace(search) == "" {
		search = "search> "
	}

	left := m.renderThemeList(listW-2, bodyH)
	right, stacked := m.renderPreview(previewW-2, bodyH)

	panelStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	leftPanel := panelStyle.Width(listW).Height(bodyH + 2).Render(left)
	rightPanel := panelStyle.Width(previewW).Height(bodyH + 2).Render(right)
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)

	parts := []string{header, search}
	if stacked {
		parts = append(parts, m.stackWarningBanner(appW))
	}
	if m.confirmOpen {
		parts = append(parts, m.renderConfirmBox(tuiMin(appW, 84)))
	}
	parts = append(parts, panes)
	footer := "keys: ↑/↓/←/→ or j/k move • type to search • PgUp/PgDn jump • Enter select • q/esc quit"
	if m.confirmOpen {
		footer = "confirm: ←/→ choose • Enter apply • Esc/q cancel"
	}
	parts = append(parts, footer)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m previewThemeModel) renderConfirmBox(width int) string {
	name := strings.TrimSpace(m.confirmName)
	if name == "" {
		name = "(none)"
	}
	msg := fmt.Sprintf("set %q as the default theme?", name)
	yes := "[ yes ]"
	cancel := "[ cancel ]"
	if m.confirmIdx == 0 {
		yes = lipgloss.NewStyle().Bold(true).Render(yes)
	} else {
		cancel = lipgloss.NewStyle().Bold(true).Render(cancel)
	}
	body := lipgloss.JoinVertical(
		lipgloss.Center,
		msg,
		"",
		lipgloss.JoinHorizontal(lipgloss.Center, yes, "  ", cancel),
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(width).
		Padding(0, 1).
		Align(lipgloss.Center).
		Render(body)
}

func (m previewThemeModel) stackWarningBanner(width int) string {
	msg := "screen size is small, preview is stacked"
	if m.noColor {
		return "[!] " + msg
	}
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("160")).
		Padding(0, 1)
	line := style.Render(msg)
	return lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, line)
}

func (m previewThemeModel) renderThemeList(width, height int) string {
	if width < 8 {
		width = 8
	}
	if height < 3 {
		height = 3
	}
	if len(m.filtered) == 0 {
		return "themes\n\n(no match)"
	}

	head := "themes"
	itemsH := height - 2
	if itemsH < 1 {
		itemsH = 1
	}

	start := m.selected - (itemsH / 2)
	if start < 0 {
		start = 0
	}
	if start+itemsH > len(m.filtered) {
		start = len(m.filtered) - itemsH
		if start < 0 {
			start = 0
		}
	}
	end := start + itemsH
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	lines := []string{head, strings.Repeat("-", tuiMin(width, 20))}
	for i := start; i < end; i++ {
		name := m.filtered[i]
		rawName := name
		if displayWidth(rawName) > width-2 {
			rawName = truncateRunes(rawName, width-2)
		}
		prefix := "  "
		line := rawName
		if i == m.selected {
			prefix = "▶ "
			line = lipgloss.NewStyle().Bold(true).Render(rawName)
		}
		line = prefix + line
		lines = append(lines, line)
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (m previewThemeModel) renderPreview(width, height int) (string, bool) {
	theme := m.selectedTheme()
	if strings.TrimSpace(theme) == "" {
		return "preview\n\nno theme selected", false
	}
	palette := m.palettes[theme]
	previewLines, stacked := buildThemePreviewLines(theme, palette, m.fields, m.noColor, width, height)

	for i := range previewLines {
		plain := stripANSI(previewLines[i])
		if displayWidth(plain) > width {
			previewLines[i] = truncateRunes(plain, width)
		}
	}
	if len(previewLines) > height {
		previewLines = previewLines[:height]
	}
	for len(previewLines) < height {
		previewLines = append(previewLines, "")
	}
	content := strings.Join(previewLines, "\n")
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, content), stacked
}

func buildThemePreviewLines(theme string, palette []string, fields []Field, noColor bool, maxWidth, maxHeight int) ([]string, bool) {
	if maxWidth < 20 {
		maxWidth = 20
	}
	if maxHeight < 8 {
		maxHeight = 8
	}

	style := defaultStyleConfig()
	style.Layout.LogoLeftPadding = 1
	style.Layout.SideBySideGap = 3
	style.Layout.OuterTopSpacing = 0
	style.Layout.OuterBottomSpacing = 0
	style.Swatch.Center = true

	if len(fields) > 12 {
		fields = fields[:12]
	}
	labelW := labelWidth(fields)
	logoSet := resolveLogoSet("", defaultBrandConfig())
	sideBySide, logoSize, logoWidth, valueWidth := chooseLayout(logoSet, style, maxWidth, labelW)
	logoLines, _ := renderLogoLines(logoSet, logoSize, palette, noColor)
	infoLines := renderInfoLines(fields, style, labelW, valueWidth, palette, noColor)
	pre := renderUserHostLines(style, palette, noColor)
	infoWidth := renderedBlockWidth(infoLines, pre)
	if sideBySide {
		needed := style.Layout.LogoLeftPadding + logoWidth + style.Layout.SideBySideGap + infoWidth + sideBySideSafetyMargin()
		if needed > maxWidth {
			sideBySide = false
		}
	}
	if !sideBySide {
		infoLines = renderInfoLines(fields, style, labelW, 0, palette, noColor)
		infoWidth = renderedBlockWidth(infoLines, pre)
	}
	swatchLines := paletteSwatchLines(noColor, infoWidth, style)

	lines := make([]string, 0, maxHeight)
	header := "theme: " + theme
	if !noColor && len(palette) > 0 {
		header = colorLine(palette[0], false, header)
	}
	lines = append(lines, header, "")

	if sideBySide {
		right := make([]RenderedLine, 0, len(pre)+len(infoLines))
		right = append(right, pre...)
		right = append(right, infoLines...)
		left, right := alignSideBlocks(logoLines, right, "top")
		total := maxInt(len(left), len(right))
		leftPad := strings.Repeat(" ", style.Layout.LogoLeftPadding)
		for i := 0; i < total; i++ {
			lRaw := ""
			lStyled := ""
			rStyled := ""
			if i < len(left) {
				lRaw = left[i].Raw
				lStyled = left[i].Styled
			}
			if i < len(right) {
				rStyled = right[i].Styled
			}
			lines = append(lines, leftPad+padRightStyled(lStyled, lRaw, logoWidth)+strings.Repeat(" ", style.Layout.SideBySideGap)+rStyled)
		}
		for _, sw := range swatchLines {
			lines = append(lines, leftPad+strings.Repeat(" ", logoWidth+style.Layout.SideBySideGap)+sw)
		}
	} else {
		leftPad := strings.Repeat(" ", style.Layout.LogoLeftPadding)
		for _, l := range logoLines {
			lines = append(lines, leftPad+l.Styled)
		}
		lines = append(lines, "")
		for _, l := range pre {
			lines = append(lines, l.Styled)
		}
		for _, l := range infoLines {
			lines = append(lines, l.Styled)
		}
		for _, sw := range swatchLines {
			lines = append(lines, sw)
		}
	}

	if len(lines) > maxHeight {
		lines = lines[:maxHeight]
	}
	return lines, !sideBySide
}

func tuiMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func tuiMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
