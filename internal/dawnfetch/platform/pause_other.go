//go:build !windows

// this file disables pause-on-exit behavior on non-windows platforms.
package platform

func shouldPauseOnExit() bool { return false }
