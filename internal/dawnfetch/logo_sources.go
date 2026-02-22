// this file resolves ascii/image logo sources and platform logo routing.
package dawnfetch

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/BigJk/imeji"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var (
	distroKeyRE = regexp.MustCompile(`[^a-z0-9._-]+`)
)

const (
	macOSSmallLogoWidth  = 120
	macOSSmallLogoHeight = 36
)

func resolveLogoSet(imagePath string, cfg BrandConfig) LogoSet {
	if p := strings.TrimSpace(imagePath); p != "" {
		if ls, err := imageLogoSet(p); err == nil {
			return ls
		}
	}

	if runtime.GOOS == "windows" {
		for _, key := range windowsLogoCandidates() {
			if ls, ok := loadTextLogoSetForKey(key); ok {
				return ls
			}
		}
		// Hard fallback requested: always use windows.txt on Windows when detection is unclear.
		if ls, ok := loadTextLogoSetForKey("windows"); ok {
			return ls
		}
	}

	if runtime.GOOS == "darwin" {
		if ls, ok := macOSLogoSet(); ok {
			return ls
		}
	}

	rel := linuxOSRelease()
	distroID := rel.ID
	if distroID == "" {
		distroID = osLogoKey(cfg)
	}

	for _, key := range distroCandidates(distroID, rel.IDLike) {
		if ls, ok := loadTextLogoSetForKey(key); ok {
			return ls
		}
	}

	if p := findImageLogoPath(distroCandidates(distroID, rel.IDLike)); p != "" {
		if ls, err := imageLogoSet(p); err == nil {
			return ls
		}
	}

	k := osLogoKey(cfg)
	if ls, ok := cfg.Logos[k]; ok {
		return ls
	}
	if ls, ok := cfg.Logos["generic"]; ok {
		return ls
	}
	return LogoSet{Tiny: []LogoLine{{Text: "dawnfetch", ColorIndex: 0}}}
}

func macOSLogoSet() (LogoSet, bool) {
	w := getTerminalWidth()
	h := getTerminalHeight()
	if (w > 0 && w < macOSSmallLogoWidth) || (h > 0 && h < macOSSmallLogoHeight) {
		if ls, ok := loadTextLogoSetForKey("macos_small"); ok {
			return ls, true
		}
	}
	if ls, ok := loadTextLogoSetForKey("macos"); ok {
		return ls, true
	}
	if ls, ok := loadTextLogoSetForKey("darwin"); ok {
		return ls, true
	}
	return LogoSet{}, false
}

func textLogoSet(normal []string, small []string) LogoSet {
	if len(normal) == 0 {
		normal = []string{"dawnfetch"}
	}
	if len(small) == 0 {
		small = normal
	}
	mk := func(src []string) []LogoLine {
		out := make([]LogoLine, 0, len(src))
		for i, s := range src {
			out = append(out, LogoLine{Text: s, ColorIndex: i})
		}
		return out
	}
	return LogoSet{
		Normal:  mk(normal),
		Compact: mk(small),
		Tiny:    mk(small),
	}
}

func loadTextLogoSetForKey(key string) (LogoSet, bool) {
	var normal []string
	var small []string
	dirs := logoTextDirCandidates()
	seen := map[string]struct{}{}
	for _, dir := range dirs {
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		if normal == nil {
			normal = readLogoTextFile(filepath.Join(dir, key+".txt"))
		}
		if small == nil {
			small = readLogoTextFile(filepath.Join(dir, key+"_small.txt"))
		}
		if normal != nil && small != nil {
			break
		}
	}
	if len(normal) == 0 && len(small) == 0 {
		return LogoSet{}, false
	}
	return textLogoSet(normal, small), true
}

func readLogoTextFile(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := splitLinesPreserveSpacing(string(b))
	if len(lines) == 0 {
		return nil
	}
	return lines
}

func imageLogoSet(path string) (LogoSet, error) {
	// Keep startup fast: render once here and let layout-specific fitting adjust if needed.
	normal, err := renderImageANSI(path, 26)
	if err != nil {
		return LogoSet{}, err
	}

	mk := func(src []string) []LogoLine {
		out := make([]LogoLine, 0, len(src))
		for _, s := range src {
			out = append(out, LogoLine{Text: s, RawANSI: true})
		}
		return out
	}
	one := mk(normal)
	return LogoSet{Normal: one, Compact: one, Tiny: one}, nil
}

func renderImageANSI(path string, width int) ([]string, error) {
	if width < 8 {
		width = 8
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".svg" {
		return nil, fmt.Errorf("svg is not supported")
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	ascii, err := imeji.FileString(
		path,
		imeji.WithMaxWidth(width),
		imeji.WithTrueColor(),
	)
	if err != nil {
		return nil, err
	}
	ascii = strings.TrimRight(ascii, "\n")
	if ascii == "" {
		return nil, fmt.Errorf("empty ascii conversion")
	}
	return splitLinesPreserveSpacing(ascii), nil
}

func fitImageLogoForHeight(path string, preferredWidth int, maxAllowedWidth int, maxHeight int, noColor bool) ([]RenderedLine, int, bool) {
	if maxHeight <= 0 {
		return nil, 0, false
	}

	const minWidth = 10
	const cellAspectComp = 1.9

	if maxAllowedWidth < minWidth {
		maxAllowedWidth = minWidth
	}
	estimate := preferredWidth
	if estimate < minWidth {
		estimate = minWidth
	}
	if estimate > maxAllowedWidth {
		estimate = maxAllowedWidth
	}
	if iw, ih, ok := imageDimensions(path); ok && ih > 0 {
		est := int(float64(maxHeight)*(float64(iw)/float64(ih))*cellAspectComp + 0.5)
		if est > 0 {
			estimate = est
		}
	}
	if estimate < minWidth {
		estimate = minWidth
	}
	if estimate > maxAllowedWidth {
		estimate = maxAllowedWidth
	}

	renderAt := func(w int) []string {
		if w < minWidth {
			w = minWidth
		}
		if w > maxAllowedWidth {
			w = maxAllowedWidth
		}
		lines, err := renderImageANSI(path, w)
		if err != nil || len(lines) == 0 {
			return nil
		}
		return lines
	}

	// Find widest width that does not exceed maxHeight.
	lo := minWidth
	hi := maxAllowedWidth
	bestFitW := 0
	bestFitLines := []string(nil)
	for lo <= hi {
		mid := (lo + hi) / 2
		lines := renderAt(mid)
		if len(lines) == 0 {
			hi = mid - 1
			continue
		}
		if len(lines) <= maxHeight {
			bestFitW = mid
			bestFitLines = lines
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	// If it still underfills, prefer the nearest overflow width and crop to exact height.
	if bestFitW > 0 && len(bestFitLines) < maxHeight && bestFitW < maxAllowedWidth {
		lo = bestFitW + 1
		hi = maxAllowedWidth
		overflowW := 0
		overflowLines := []string(nil)
		for lo <= hi {
			mid := (lo + hi) / 2
			lines := renderAt(mid)
			if len(lines) == 0 {
				hi = mid - 1
				continue
			}
			if len(lines) >= maxHeight {
				overflowW = mid
				overflowLines = lines
				hi = mid - 1
			} else {
				lo = mid + 1
			}
		}
		if overflowW > 0 && len(overflowLines) > 0 {
			if len(overflowLines) > maxHeight {
				overflowLines = overflowLines[:maxHeight]
			}
			return ansiLinesToRendered(overflowLines, noColor), visibleWidth(overflowLines), true
		}
	}

	if bestFitW > 0 && len(bestFitLines) > 0 {
		return ansiLinesToRendered(bestFitLines, noColor), visibleWidth(bestFitLines), true
	}

	// Last resort: use estimate and crop if required.
	lines := renderAt(estimate)
	if len(lines) == 0 {
		return nil, 0, false
	}
	if len(lines) > maxHeight {
		lines = lines[:maxHeight]
	}
	return ansiLinesToRendered(lines, noColor), visibleWidth(lines), true
}

func ansiLinesToRendered(lines []string, noColor bool) []RenderedLine {
	out := make([]RenderedLine, 0, len(lines))
	for _, line := range lines {
		clean := stripANSI(line)
		styled := line
		if noColor {
			styled = clean
		}
		out = append(out, RenderedLine{Raw: clean, Styled: styled})
	}
	return out
}

func visibleWidth(lines []string) int {
	maxW := 0
	for _, line := range lines {
		w := displayWidth(stripANSI(line))
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

func imageDimensions(path string) (int, int, bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

func findImageLogoPath(candidates []string) string {
	dirs := logoImageDirCandidates()
	for _, key := range candidates {
		for _, dir := range dirs {
			for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp", ".tiff"} {
				p := filepath.Join(dir, key+ext)
				if _, err := os.Stat(p); err == nil {
					return p
				}
			}
		}
	}
	return ""
}

func windowsLogoCandidates() []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 5)
	add := func(v string) {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	w := getTerminalWidth()
	h := getTerminalHeight()
	if (w > 0 && w < 145) || (h > 0 && h < 45) {
		// always use windows_small for TIGHT TIGHT TIGHT Windows terminals.
		add("windows_small")
	}

	add(windowsVersionLogoKey())

	add("windows")
	return out
}

func windowsVersionLogoKey() string {
	wf := windowsFacts()
	caption := strings.ToLower(strings.TrimSpace(wf.Caption))
	if strings.Contains(caption, "windows 11") {
		return "windows_11"
	}
	if build, ok := firstPositiveInt(wf.Build); ok && build >= 22000 {
		return "windows_11"
	}
	return "windows_8_10"
}

func firstPositiveInt(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func distroCandidates(id, idLike string) []string {
	out := []string{}
	seen := map[string]struct{}{}
	add := func(v string) {
		v = strings.ToLower(strings.TrimSpace(v))
		v = distroKeyRE.ReplaceAllString(v, "")
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	add(id)
	for _, v := range strings.Fields(idLike) {
		add(v)
	}
	add("linux")
	add("generic")
	return out
}

func splitLinesPreserveSpacing(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	parts := strings.Split(s, "\n")
	for len(parts) > 0 && isVisuallyEmpty(parts[len(parts)-1]) {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func isVisuallyEmpty(s string) bool {
	return strings.TrimSpace(stripANSI(s)) == ""
}
