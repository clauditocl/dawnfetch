// this file has linux fast-path probes used by default fast mode.
package system

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func fastLinuxPackageCount() string {
	counts := map[string]int{}

	// read one file and count installed package status lines.
	// this is usually faster than scanning thousands of files under /var/lib/dpkg/info.
	if b, err := os.ReadFile("/var/lib/dpkg/status"); err == nil {
		count := bytes.Count(b, []byte("\nStatus: install ok installed"))
		if bytes.HasPrefix(b, []byte("Status: install ok installed")) {
			count++
		}
		if count > 0 {
			counts["dpkg"] = count
		}
	}

	if entries, err := os.ReadDir("/var/lib/pacman/local"); err == nil {
		count := 0
		for _, e := range entries {
			if e.IsDir() {
				count++
			}
		}
		if count > 0 {
			counts["pacman"] = count
		}
	}

	if len(counts) == 0 {
		return ""
	}

	keys := make([]string, 0, len(counts))
	total := 0
	for k, v := range counts {
		keys = append(keys, k)
		total += v
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", k, counts[k]))
	}
	return fmt.Sprintf("%d (%s)", total, strings.Join(parts, ", "))
}

func fastLinuxResolution() string {
	if entries, err := os.ReadDir("/sys/class/drm"); err == nil {
		for _, e := range entries {
			name := e.Name()
			// drm connector entries are commonly symlinks, not real directories.
			if !strings.Contains(name, "-") {
				continue
			}
			mode := readFirstLine(filepath.Join("/sys/class/drm", name, "modes"))
			if mode != "" {
				return mode
			}
		}
	}
	return ""
}

func fastLinuxGPU() string {
	seen := map[string]struct{}{}
	out := []string{}
	if entries, err := os.ReadDir("/sys/class/drm"); err == nil {
		for _, e := range entries {
			name := e.Name()
			// cardN entries are commonly symlinks, so don't require IsDir.
			if !strings.HasPrefix(name, "card") || strings.Contains(name, "-") {
				continue
			}
			p := filepath.Join("/sys/class/drm", name, "device", "uevent")
			b, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			driver := ""
			pciID := ""
			for _, line := range strings.Split(string(b), "\n") {
				switch {
				case strings.HasPrefix(line, "DRIVER="):
					driver = strings.TrimSpace(strings.TrimPrefix(line, "DRIVER="))
				case strings.HasPrefix(line, "PCI_ID="):
					pciID = strings.TrimSpace(strings.TrimPrefix(line, "PCI_ID="))
				}
			}
			label := strings.TrimSpace(strings.Join(nonEmpty([]string{driver, pciID}), " "))
			if label == "" {
				continue
			}
			if _, ok := seen[label]; ok {
				continue
			}
			seen[label] = struct{}{}
			out = append(out, label)
		}
	}
	return strings.Join(out, " | ")
}
