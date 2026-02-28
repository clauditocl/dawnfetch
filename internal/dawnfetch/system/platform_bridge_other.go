//go:build !windows

// this file provides non-windows stubs for windows-only api helpers.
package system

func windowsUptimeSecondsAPI() int64 {
	return 0
}

func windowsVersionAPI() (major int, minor int, build int, ok bool) {
	return 0, 0, 0, false
}

func windowsResolutionAPI() (int, int, bool) {
	return 0, 0, false
}

func windowsMemoryAPI() (total int64, avail int64, ok bool) {
	return 0, 0, false
}

func windowsBatteryAPI() (percent int, pluggedState int, hasBattery bool, ok bool) {
	return -1, -1, false, false
}
