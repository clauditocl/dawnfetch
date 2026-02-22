// this file contains reusable parsing and formatting helpers.
package dawnfetch

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func stdoutIsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func parseXrandr(out string) string {
	lines := strings.Split(out, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, " connected") {
			for j := i + 1; j < len(lines); j++ {
				next := lines[j]
				if strings.TrimSpace(next) == "" || !strings.HasPrefix(next, "   ") {
					break
				}
				if strings.Contains(next, "*") {
					fields := strings.Fields(next)
					if len(fields) > 0 {
						res := fields[0]
						rr := ""
						for _, f := range fields[1:] {
							if strings.Contains(f, "*") {
								rr = strings.TrimSuffix(strings.TrimSuffix(f, "*"), "+")
								break
							}
						}
						if rr != "" {
							return res + " @ " + rr + "Hz"
						}
						return res
					}
				}
			}
		}
	}
	return ""
}

func parseWlrRandr(out string) string {
	resRe := regexp.MustCompile(`(\d+)x(\d+)`)
	rrRe := regexp.MustCompile(`@\s*([\d.]+)Hz`)
	res := resRe.FindString(out)
	rr := ""
	if m := rrRe.FindStringSubmatch(out); len(m) == 2 {
		rr = m[1]
	}
	if res != "" && rr != "" {
		return res + " @ " + rr + "Hz"
	}
	return res
}

func parseMacDisplay(out string) (string, string) {
	res := ""
	rr := ""
	for _, line := range strings.Split(out, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "Resolution:") {
			res = strings.TrimSpace(strings.TrimPrefix(t, "Resolution:"))
		}
		if strings.HasPrefix(t, "Refresh Rate:") {
			rr = strings.TrimSpace(strings.TrimPrefix(t, "Refresh Rate:"))
		}
	}
	return res, rr
}

func parseMeminfoKB(line string) int64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	v, _ := strconv.ParseInt(fields[1], 10, 64)
	return v
}

func formatBytes(b int64) string {
	if b < 0 {
		b = 0
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func parseVMStatUsed(vm string) int64 {
	if vm == "" {
		return 0
	}
	pageSize := int64(4096)
	if m := regexp.MustCompile(`page size of (\d+) bytes`).FindStringSubmatch(vm); len(m) == 2 {
		if v, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			pageSize = v
		}
	}
	var active, wired, compressed int64
	for _, line := range strings.Split(vm, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Pages active:"):
			active = parseVMStatPages(line)
		case strings.HasPrefix(line, "Pages wired down:"):
			wired = parseVMStatPages(line)
		case strings.HasPrefix(line, "Pages occupied by compressor:"):
			compressed = parseVMStatPages(line)
		}
	}
	return (active + wired + compressed) * pageSize
}

func parseVMStatPages(line string) int64 {
	re := regexp.MustCompile(`(\d+)`)
	m := re.FindStringSubmatch(line)
	if len(m) != 2 {
		return 0
	}
	v, _ := strconv.ParseInt(m[1], 10, 64)
	return v
}

func hzToGHz(raw string) string {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || v <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2fGHz", v/1_000_000_000)
}

func trimQuotes(s string) string {
	return strings.Trim(strings.TrimSpace(s), "'\"")
}

func readFirstLine(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, l := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(l)
		if t != "" {
			return t
		}
	}
	return ""
}

func firstNonEmptyLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			return l
		}
	}
	return ""
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = compactSpaces(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func compactSpaces(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func countLines(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

func nonEmpty(items []string) []string {
	out := make([]string, 0, len(items))
	for _, i := range items {
		i = strings.TrimSpace(i)
		if i != "" {
			out = append(out, i)
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
