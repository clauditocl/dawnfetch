//go:build !windows

// this file provides terminal size and ansi support checks on unix.
package platform

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func GetTerminalWidth() int {
	if c := os.Getenv("COLUMNS"); c != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(c)); err == nil && n > 0 {
			return n
		}
	}

	ws := &struct {
		row    uint16
		col    uint16
		xpixel uint16
		ypixel uint16
	}{}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if errno == 0 && ws.col > 0 {
		return int(ws.col)
	}

	return 0
}

func GetTerminalHeight() int {
	if l := os.Getenv("LINES"); l != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(l)); err == nil && n > 0 {
			return n
		}
	}

	ws := &struct {
		row    uint16
		col    uint16
		xpixel uint16
		ypixel uint16
	}{}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if errno == 0 && ws.row > 0 {
		return int(ws.row)
	}

	return 0
}

func EnableANSIIfSupported() bool {
	return true
}
