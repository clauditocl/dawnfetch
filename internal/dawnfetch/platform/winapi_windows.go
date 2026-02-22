//go:build windows

// this file wraps windows api calls used by collectors.
package platform

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modKernel32               = syscall.NewLazyDLL("kernel32.dll")
	modNtdll                  = syscall.NewLazyDLL("ntdll.dll")
	procGetTickCount64        = modKernel32.NewProc("GetTickCount64")
	procGetTickCount          = modKernel32.NewProc("GetTickCount")
	procGlobalMemoryStatusEx  = modKernel32.NewProc("GlobalMemoryStatusEx")
	procGetSystemPowerStatus  = modKernel32.NewProc("GetSystemPowerStatus")
	procGetDiskFreeSpaceExW   = modKernel32.NewProc("GetDiskFreeSpaceExW")
	procGetSystemMetrics      = modKernel32.NewProc("GetSystemMetrics")
	procGetVolumeInformationW = modKernel32.NewProc("GetVolumeInformationW")
	procRtlGetVersion         = modNtdll.NewProc("RtlGetVersion")
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	Reserved1           byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

type rtlOSVersionInfoEx struct {
	OSVersionInfoSize uint32
	MajorVersion      uint32
	MinorVersion      uint32
	BuildNumber       uint32
	PlatformID        uint32
	CSDVersion        [128]uint16
}

func windowsRootPath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" || p == "/" {
		if d := strings.TrimSpace(os.Getenv("SystemDrive")); d != "" {
			if strings.HasSuffix(d, ":") {
				return d + `\`
			}
			return d
		}
		return `C:\`
	}
	if len(p) == 2 && p[1] == ':' {
		return p + `\`
	}
	return p
}

func WindowsVersionAPI() (major int, minor int, build int, ok bool) {
	if procRtlGetVersion.Find() != nil {
		return 0, 0, 0, false
	}
	var info rtlOSVersionInfoEx
	info.OSVersionInfoSize = uint32(unsafe.Sizeof(info))
	r1, _, _ := procRtlGetVersion.Call(uintptr(unsafe.Pointer(&info)))
	if r1 != 0 {
		return 0, 0, 0, false
	}
	return int(info.MajorVersion), int(info.MinorVersion), int(info.BuildNumber), true
}

func WindowsUptimeSecondsAPI() int64 {
	if procGetTickCount64.Find() == nil {
		r1, _, _ := procGetTickCount64.Call()
		return int64(uint64(r1) / 1000)
	}
	if procGetTickCount.Find() == nil {
		r1, _, _ := procGetTickCount.Call()
		return int64(uint32(r1) / 1000)
	}
	return 0
}

func WindowsResolutionAPI() (int, int, bool) {
	if procGetSystemMetrics.Find() != nil {
		return 0, 0, false
	}
	// SM_CXSCREEN=0, SM_CYSCREEN=1
	w, _, _ := procGetSystemMetrics.Call(uintptr(0))
	h, _, _ := procGetSystemMetrics.Call(uintptr(1))
	if int(w) <= 0 || int(h) <= 0 {
		return 0, 0, false
	}
	return int(w), int(h), true
}

func WindowsMemoryAPI() (total int64, avail int64, ok bool) {
	if procGlobalMemoryStatusEx.Find() != nil {
		return 0, 0, false
	}
	var st memoryStatusEx
	st.Length = uint32(unsafe.Sizeof(st))
	r1, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&st)))
	if r1 == 0 || st.TotalPhys == 0 {
		return 0, 0, false
	}
	return int64(st.TotalPhys), int64(st.AvailPhys), true
}

func WindowsDiskUsageAPI(path string) (total int64, free int64, ok bool) {
	if procGetDiskFreeSpaceExW.Find() != nil {
		return 0, 0, false
	}
	root := windowsRootPath(path)
	p, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return 0, 0, false
	}
	var freeToCaller uint64
	var totalBytes uint64
	var totalFree uint64
	r1, _, _ := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(&freeToCaller)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if r1 == 0 || totalBytes == 0 {
		return 0, 0, false
	}
	return int64(totalBytes), int64(totalFree), true
}

func WindowsVolumeFSTypeAPI(path string) string {
	if procGetVolumeInformationW.Find() != nil {
		return ""
	}
	root := windowsRootPath(path)
	p, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return ""
	}
	fsName := make([]uint16, 64)
	r1, _, _ := procGetVolumeInformationW.Call(
		uintptr(unsafe.Pointer(p)),
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&fsName[0])),
		uintptr(len(fsName)),
	)
	if r1 == 0 {
		return ""
	}
	return strings.TrimSpace(syscall.UTF16ToString(fsName))
}

// WindowsBatteryAPI returns:
// - percent: 0...100, or -1 when unknown
// - pluggedState: 1 plugged in, 0 not plugged in, -1 unknown
// - hasBattery: false when system reports no battery
// - ok: false when API call fails
func WindowsBatteryAPI() (percent int, pluggedState int, hasBattery bool, ok bool) {
	if procGetSystemPowerStatus.Find() != nil {
		return -1, -1, false, false
	}
	var s systemPowerStatus
	r1, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&s)))
	if r1 == 0 {
		return -1, -1, false, false
	}

	// BatteryFlag 128 means no system battery
	// how would we show this to the user? :P
	if s.BatteryFlag&0x80 != 0 {
		return -1, -1, false, true
	}

	pct := -1
	if s.BatteryLifePercent != 255 {
		pct = int(s.BatteryLifePercent)
	}

	plug := -1
	if s.ACLineStatus == 1 {
		plug = 1
	} else if s.ACLineStatus == 0 {
		plug = 0
	}
	return pct, plug, true, true
}
