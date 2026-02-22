//go:build windows

// this file decides whether windows should pause before exit.
package platform

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	th32csSnapProcess = 0x00000002
	maxPathW          = 260
)

var (
	modKernel32Pause             = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = modKernel32Pause.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = modKernel32Pause.NewProc("Process32FirstW")
	procProcess32NextW           = modKernel32Pause.NewProc("Process32NextW")
	procCloseHandlePause         = modKernel32Pause.NewProc("CloseHandle")
)

type processEntry32 struct {
	Size              uint32
	Usage             uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	Threads           uint32
	ParentProcessID   uint32
	PriClassBase      int32
	Flags             uint32
	ExeFile           [maxPathW]uint16
}

func stdoutIsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func shouldPauseOnExit() bool {
	if os.Getenv("DAWNFETCH_NO_PAUSE") == "1" {
		return false
	}
	if os.Getenv("DAWNFETCH_FORCE_PAUSE") == "1" {
		return true
	}
	if !stdoutIsTerminal() {
		return false
	}
	parent, grandparent, ok := windowsParentAndGrandparentProcessNames()
	if !ok {
		return false
	}
	p := strings.ToLower(strings.TrimSpace(parent))
	gp := strings.ToLower(strings.TrimSpace(grandparent))
	if isInteractiveShellProcess(p) || isInteractiveShellProcess(gp) {
		return false
	}
	return p == "explorer.exe" || gp == "explorer.exe"
}

func isInteractiveShellProcess(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "cmd.exe", "powershell.exe", "pwsh.exe", "bash.exe", "wsl.exe", "windowsterminal.exe":
		return true
	default:
		return false
	}
}

func windowsParentAndGrandparentProcessNames() (string, string, bool) {
	snap, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snap == ^uintptr(0) || snap == 0 {
		return "", "", false
	}
	defer procCloseHandlePause.Call(snap)

	curPID := uint32(os.Getpid())
	names := make(map[uint32]string, 256)
	parents := make(map[uint32]uint32, 256)
	var parentPID uint32

	var pe processEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	r1, _, _ := procProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&pe)))
	if r1 == 0 {
		return "", "", false
	}

	for {
		name := syscall.UTF16ToString(pe.ExeFile[:])
		names[pe.ProcessID] = name
		parents[pe.ProcessID] = pe.ParentProcessID
		if pe.ProcessID == curPID {
			parentPID = pe.ParentProcessID
		}

		pe.Size = uint32(unsafe.Sizeof(pe))
		r1, _, _ = procProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&pe)))
		if r1 == 0 {
			break
		}
	}

	if parentPID == 0 {
		return "", "", false
	}
	parentName := strings.TrimSpace(names[parentPID])
	grandparentName := ""
	if grandparentPID, ok := parents[parentPID]; ok && grandparentPID != 0 {
		grandparentName = strings.TrimSpace(names[grandparentPID])
	}
	if parentName == "" && grandparentName == "" {
		return "", "", false
	}
	return parentName, grandparentName, true
}
