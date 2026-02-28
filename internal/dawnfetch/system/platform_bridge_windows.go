//go:build windows

// this file bridges windows-only api helpers into the dawnfetch package.
package system

import "dawnfetch/internal/dawnfetch/platform"

func windowsUptimeSecondsAPI() int64 {
	return platform.WindowsUptimeSecondsAPI()
}

func windowsVersionAPI() (major int, minor int, build int, ok bool) {
	return platform.WindowsVersionAPI()
}

func windowsResolutionAPI() (int, int, bool) {
	return platform.WindowsResolutionAPI()
}

func windowsMemoryAPI() (total int64, avail int64, ok bool) {
	return platform.WindowsMemoryAPI()
}

func windowsBatteryAPI() (percent int, pluggedState int, hasBattery bool, ok bool) {
	return platform.WindowsBatteryAPI()
}
