// this file detects os metadata for logo/theme selection.
package system

import (
	"os"
	"runtime"
	"strings"

	"dawnfetch/internal/dawnfetch/core"
)

func linuxOSRelease() core.OSReleaseInfo {
	return parseOSRelease()
}

func parseOSRelease() core.OSReleaseInfo {
	paths := []string{"/etc/os-release", "/usr/lib/os-release"}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		out := core.OSReleaseInfo{}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k := parts[0]
			v := strings.Trim(parts[1], "\"")
			switch k {
			case "ID":
				out.ID = strings.ToLower(v)
			case "ID_LIKE":
				out.IDLike = strings.ToLower(v)
			case "PRETTY_NAME":
				out.PrettyName = v
			}
		}
		return out
	}
	return core.OSReleaseInfo{}
}

func osLogoKey(cfg core.BrandConfig) string {
	switch runtime.GOOS {
	case "linux":
		rel := linuxOSRelease()
		if rel.ID != "" {
			if _, ok := cfg.Logos[rel.ID]; ok {
				return rel.ID
			}
		}
		for _, like := range strings.Fields(rel.IDLike) {
			if _, ok := cfg.Logos[like]; ok {
				return like
			}
		}
		if _, ok := cfg.Logos["linux"]; ok {
			return "linux"
		}
		return "generic"
	case "darwin":
		if _, ok := cfg.Logos["darwin"]; ok {
			return "darwin"
		}
	case "windows":
		if _, ok := cfg.Logos["windows"]; ok {
			return "windows"
		}
	}
	return "generic"
}

func LinuxOSRelease() core.OSReleaseInfo {
	return linuxOSRelease()
}

func OSLogoKey(cfg core.BrandConfig) string {
	return osLogoKey(cfg)
}
