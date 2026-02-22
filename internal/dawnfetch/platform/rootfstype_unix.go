//go:build !windows

// this file resolves filesystem type for unix-like platforms.
package platform

import (
	"os"
	"os/exec"
	"strings"
)

func RootFSTypeOS(_ string) string {
	// linux fast path
	if b, err := os.ReadFile("/proc/mounts"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			f := strings.Fields(line)
			if len(f) >= 3 && f[1] == "/" {
				return f[2]
			}
		}
	}

	// macos/bsd fallback: parse `mount` output
	if out, err := exec.Command("mount").Output(); err == nil {
		if fs := parseMountOutputRootFS(string(out)); fs != "" {
			return fs
		}
	}
	return ""
}

func parseMountOutputRootFS(out string) string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, " on / ") {
			continue
		}

		// macos/bsd style: /dev/disk3s1 on / (apfs, ...)
		if i := strings.Index(line, " on / ("); i >= 0 {
			inside := line[i+len(" on / ("):]
			if j := strings.Index(inside, ")"); j >= 0 {
				inside = inside[:j]
			}
			first := strings.TrimSpace(strings.SplitN(inside, ",", 2)[0])
			if first != "" {
				if fields := strings.Fields(first); len(fields) > 0 {
					return fields[0]
				}
				return first
			}
		}

		// linux mount command style: /dev/sda1 on / type ext4 (...)
		if i := strings.Index(line, " on / type "); i >= 0 {
			t := strings.TrimSpace(line[i+len(" on / type "):])
			if t == "" {
				continue
			}
			end := len(t)
			for idx, r := range t {
				if r == ' ' || r == '(' || r == ',' {
					end = idx
					break
				}
			}
			if end > 0 {
				return t[:end]
			}
		}
	}
	return ""
}
