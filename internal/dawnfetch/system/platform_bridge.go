// this file bridges common platform-specific implementations into the dawnfetch package.
package system

import (
	"runtime"
	"strings"
	"dawnfetch/internal/dawnfetch/platform"
)

func getTerminalWidth() int {
	return platform.GetTerminalWidth()
}

func getTerminalHeight() int {
	return platform.GetTerminalHeight()
}

func enableANSIIfSupported() bool {
	return platform.EnableANSIIfSupported()
}

func diskUsage(path string) (total int64, free int64, err error) {
	// keep current windows behavior and prefer to use live facts when available
	if runtime.GOOS == "windows" {
		target := strings.TrimSpace(path)
		if target == "" || target == "/" {
			if wf := windowsFacts(); wf.Valid && wf.DiskTotalB > 0 {
				return wf.DiskTotalB, wf.DiskFreeB, nil
			}
		}
	}
	return platform.DiskUsage(path)
}

func rootFSTypeOS(path string) string {
	// keep current windows behavior and prefer to use live facts when available
	if runtime.GOOS == "windows" {
		target := strings.TrimSpace(path)
		if target == "" || target == "/" {
			if wf := windowsFacts(); wf.Valid && strings.TrimSpace(wf.DiskFS) != "" {
				return strings.TrimSpace(wf.DiskFS)
			}
		}
	}
	return platform.RootFSTypeOS(path)
}

func MaybePauseOnExit(code int) {
	platform.MaybePauseOnExit(code)
}
