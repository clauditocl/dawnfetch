//go:build windows

// this file provides terminal size and ansi support checks on windows.
package platform

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func GetTerminalWidth() int {
	envW := 0
	if c := os.Getenv("COLUMNS"); c != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(c)); err == nil && n > 0 {
			envW = n
		}
	}
	apiW, _ := getConsoleViewportSize()
	if envW > 0 && apiW > 0 {
		// choose the safer bound to avoid side-by-side overestimation.
		if envW < apiW {
			return envW
		}
		return apiW
	}
	if envW > 0 {
		return envW
	}
	if apiW > 0 {
		return apiW
	}
	return 0
}

func GetTerminalHeight() int {
	envH := 0
	if l := os.Getenv("LINES"); l != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(l)); err == nil && n > 0 {
			envH = n
		}
	}
	_, apiH := getConsoleViewportSize()
	if envH > 0 && apiH > 0 {
		if envH < apiH {
			return envH
		}
		return apiH
	}
	if envH > 0 {
		return envH
	}
	if apiH > 0 {
		return apiH
	}
	return 0
}

func getConsoleViewportSize() (int, int) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleScreenBufferInfo := kernel32.NewProc("GetConsoleScreenBufferInfo")

	const stdOutputHandle = uint32(0xFFFFFFF5) // STD_OUTPUT_HANDLE (-11 as DWORD)

	type coord struct {
		X int16
		Y int16
	}
	type smallRect struct {
		Left   int16
		Top    int16
		Right  int16
		Bottom int16
	}
	type consoleScreenBufferInfo struct {
		DwSize              coord
		DwCursorPosition    coord
		WAttributes         uint16
		SrWindow            smallRect
		DwMaximumWindowSize coord
	}

	h, _, _ := getStdHandle.Call(uintptr(stdOutputHandle))
	if h == 0 || h == uintptr(^uintptr(0)) {
		return 0, 0
	}
	var info consoleScreenBufferInfo
	r1, _, _ := getConsoleScreenBufferInfo.Call(h, uintptr(unsafe.Pointer(&info)))
	if r1 == 0 {
		return 0, 0
	}
	w := int(info.SrWindow.Right-info.SrWindow.Left) + 1
	hh := int(info.SrWindow.Bottom-info.SrWindow.Top) + 1
	if w < 0 {
		w = 0
	}
	if hh < 0 {
		hh = 0
	}
	return w, hh
}

func EnableANSIIfSupported() bool {
	// modern Windows terminals usually already support ANSI
	if os.Getenv("WT_SESSION") != "" || os.Getenv("ANSICON") != "" || strings.EqualFold(os.Getenv("ConEmuANSI"), "ON") {
		return true
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	const (
		stdOutputHandle                 = uint32(0xFFFFFFF5) // STD_OUTPUT_HANDLE (-11 as DWORD)
		enableVirtualTerminalProcessing = 0x0004
	)

	h, _, _ := getStdHandle.Call(uintptr(stdOutputHandle))
	if h == 0 || h == uintptr(^uintptr(0)) {
		return false
	}

	var mode uint32
	r1, _, _ := getConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		return false
	}

	if mode&enableVirtualTerminalProcessing != 0 {
		return true
	}

	r2, _, _ := setConsoleMode.Call(h, uintptr(mode|enableVirtualTerminalProcessing))
	return r2 != 0
}
