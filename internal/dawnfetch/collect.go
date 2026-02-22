// this file wires field collection and controls fast/full collector modes.
package dawnfetch

import (
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

var fastCollectMode atomic.Bool
var windowsSlowProbeMode atomic.Bool

type fieldCollector struct {
	label string
	fn    func(bool) string
}

var defaultCollectors = []fieldCollector{
	{label: "Operating System", fn: collectOSNameVersion},
	{label: "Host", fn: hostModel},
	{label: "Kernel", fn: collectKernelVersion},
	{label: "Uptime", fn: collectUptimeString},
	{label: "Packages", fn: packagesCount},
	{label: "Shell", fn: shellInfo},
	{label: "Resolution", fn: resolutionInfo},
	{label: "Desktop Environment", fn: collectDesktopEnvironment},
	{label: "Window Manager", fn: windowManager},
	{label: "Theme/Icons/Fonts", fn: visualSettings},
	{label: "Terminal", fn: collectTerminalInfo},
	{label: "CPU", fn: cpuInfo},
	{label: "GPU", fn: gpuInfo},
	{label: "Memory", fn: collectMemoryInfo},
	{label: "Battery", fn: batteryInfo},
	{label: "Swap", fn: collectSwapUsageSummary},
	{label: "Disk (/)", fn: collectDiskRootUsageDetailed},
	{label: "Local IP", fn: collectLocalIPSummary},
}

var fullCollectors = []fieldCollector{
	{label: "OS", fn: osWithVersionCodename},
	{label: "Kernel", fn: kernelFull},
	{label: "Architecture", fn: architectureInfo},
	{label: "Host", fn: hostModel},
	{label: "Packages", fn: packagesCount},
	{label: "Shell", fn: shellInfo},
	{label: "Desktop Environment", fn: collectDesktopEnvironment},
	{label: "Window Manager", fn: windowManager},
	{label: "Session Type", fn: sessionType},
	{label: "Theme/Icons/Fonts/Cursor", fn: visualSettingsFull},
	{label: "Terminal", fn: terminalFull},
	{label: "CPU", fn: cpuDetailed},
	{label: "GPU", fn: gpuDetailed},
	{label: "Memory", fn: memoryWithSwap},
	{label: "Disk", fn: diskMountsSummary},
	{label: "Battery", fn: batteryInfo},
	{label: "Network", fn: networkSummary},
	{label: "Locale", fn: localeInfo},
	{label: "Load", fn: loadSummary},
	{label: "Temperatures", fn: temperaturesSummary},
}

func collectOSNameVersion(_ bool) string { return osNameVersion() }
func collectKernelVersion(_ bool) string { return kernelVersion() }
func collectUptimeString(_ bool) string { return uptimeString() }
func collectDesktopEnvironment(_ bool) string { return desktopEnvironment() }
func collectTerminalInfo(_ bool) string { return terminalInfo() }
func collectMemoryInfo(_ bool) string { return memoryInfo() }
func collectSwapUsageSummary(_ bool) string { return swapUsageSummary() }
func collectDiskRootUsageDetailed(_ bool) string { return diskRootUsageDetailed() }
func collectLocalIPSummary(_ bool) string { return localIPSummary() }

func collect(fast bool, full bool) []Field {
	fastCollectMode.Store(fast && !full)
	windowsSlowProbeMode.Store(runtime.GOOS == "windows" && full)
	defer fastCollectMode.Store(false)
	defer windowsSlowProbeMode.Store(false)

	items := defaultCollectors
	if full {
		items = fullCollectors
	}

	fields := make([]Field, len(items))
	var wg sync.WaitGroup
	wg.Add(len(items))
	for i := range items {
		i := i
		go func() {
			defer wg.Done()
			v := items[i].fn(fast)
			if strings.TrimSpace(v) == "" {
				v = "unknown"
			}
			fields[i] = Field{Label: items[i].label, Value: v}
		}()
	}
	wg.Wait()

	return fields
}

func windowsSlowProbesEnabled() bool {
	return runtime.GOOS == "windows" && windowsSlowProbeMode.Load()
}
