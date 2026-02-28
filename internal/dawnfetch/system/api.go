// this file exposes stable system probing entrypoints to other packages.
package system

import (
	"time"

	"dawnfetch/internal/dawnfetch/core"
)

type WindowsFactsData = windowsFactsData

func OSNameVersion() string { return osNameVersion() }
func KernelVersion() string { return kernelVersion() }
func ShellInfo(fast bool) string { return shellInfo(fast) }
func TerminalInfo() string { return terminalInfo() }
func LocalIPSummary() string { return localIPSummary() }
func VisualSettings(fast bool) string { return visualSettings(fast) }
func DesktopEnvironment() string { return desktopEnvironment() }
func WindowManager(fast bool) string { return windowManager(fast) }
func ResolutionInfo(fast bool) string { return resolutionInfo(fast) }
func CPUInfo(fast bool) string { return cpuInfo(fast) }
func GPUInfo(fast bool) string { return gpuInfo(fast) }
func MemoryInfo() string { return memoryInfo() }
func SwapUsageSummary() string { return swapUsageSummary() }
func DiskRootUsageDetailed() string { return diskRootUsageDetailed() }
func UptimeString() string { return uptimeString() }
func PackagesCount(fast bool) string { return packagesCount(fast) }
func HostModel(fast bool) string { return hostModel(fast) }
func BatteryInfo(fast bool) string { return batteryInfo(fast) }
func WindowsFacts() windowsFactsData { return windowsFacts() }
func CommandExists(name string) bool { return commandExists(name) }
func RunCmd(timeoutMs int, name string, args ...string) (string, error) {
	return runCmd(time.Duration(timeoutMs)*time.Millisecond, name, args...)
}

func CollectDefaultFast() []core.Field { return Collect(true, false) }
