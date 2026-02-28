// this file centralizes path resolution for bundled assets.
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func ExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exePath)
}

func ThemeFileCandidates(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "themes.json"
	}
	if filepath.IsAbs(path) {
		return []string{path}
	}

	out := make([]string, 0, 12)
	seen := map[string]struct{}{}
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	// cwd relative
	add(path)

	// executable relative
	exeDir := ExecutableDir()
	if exeDir != "" {
		add(filepath.Join(exeDir, path))
		add(filepath.Join(exeDir, "assets", path))
		add(filepath.Join(exeDir, "share", "dawnfetch", path))
		add(filepath.Clean(filepath.Join(exeDir, "..", "share", "dawnfetch", path)))
	}

	// system-wide unix installs
	if runtime.GOOS != "windows" {
		add(filepath.Join("/usr/share/dawnfetch", path))
		add(filepath.Join("/usr/local/share/dawnfetch", path))
	}

	return out
}

func LogoTextDirCandidates() []string {
	out := make([]string, 0, 16)
	seen := map[string]struct{}{}
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	add("./ascii")
	add("./logos")

	exeDir := ExecutableDir()
	if exeDir != "" {
		add(filepath.Join(exeDir, "ascii"))
		add(filepath.Join(exeDir, "logos"))
		add(filepath.Join(exeDir, "assets", "ascii"))
		add(filepath.Join(exeDir, "assets", "logos"))
		add(filepath.Join(exeDir, "share", "dawnfetch", "ascii"))
		add(filepath.Clean(filepath.Join(exeDir, "..", "share", "dawnfetch", "ascii")))
		add(filepath.Join(exeDir, "share", "dawnfetch", "logos"))
		add(filepath.Clean(filepath.Join(exeDir, "..", "share", "dawnfetch", "logos")))
	}

	if runtime.GOOS != "windows" {
		add("/usr/share/dawnfetch/ascii")
		add("/usr/local/share/dawnfetch/ascii")
		add("/usr/share/dawnfetch/logos")
		add("/usr/local/share/dawnfetch/logos")
	}

	return out
}

func LogoImageDirCandidates() []string {
	// image lookup reuses the same candidate roots as text logos.
	return LogoTextDirCandidates()
}
