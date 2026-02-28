// this file contains --full output collectors.
package system

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func osWithVersionCodename(_ bool) string {
	if runtime.GOOS != "linux" {
		return osNameVersion()
	}
	rel := linuxOSRelease()
	parts := []string{}
	if rel.PrettyName != "" {
		parts = append(parts, rel.PrettyName)
	}
	if b, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "VERSION_CODENAME=") {
				c := strings.Trim(strings.TrimPrefix(line, "VERSION_CODENAME="), "\"")
				if c != "" {
					parts = append(parts, "codename "+c)
				}
				break
			}
		}
	}
	if len(parts) == 0 {
		return osNameVersion()
	}
	return strings.Join(parts, " | ")
}

func kernelFull(_ bool) string {
	if runtime.GOOS == "windows" {
		return kernelVersion()
	}
	if out, _ := runCmd(180*time.Millisecond, "uname", "-srvmo"); out != "" {
		return out
	}
	return kernelVersion()
}

func architectureInfo(_ bool) string {
	if runtime.GOOS == "windows" {
		if wf := windowsFacts(); wf.Valid && strings.TrimSpace(wf.Architecture) != "" {
			return wf.Architecture
		}
		if !windowsSlowProbesEnabled() {
			if v := strings.TrimSpace(os.Getenv("PROCESSOR_ARCHITECTURE")); v != "" {
				return v
			}
			return runtime.GOARCH
		}
		if out, _ := runCmd(300*time.Millisecond, "powershell", "-NoProfile", "-Command", "$env:PROCESSOR_ARCHITECTURE"); out != "" {
			return out
		}
	}
	if out, _ := runCmd(120*time.Millisecond, "uname", "-m"); out != "" {
		return out
	}
	return runtime.GOARCH
}

func sessionType(_ bool) string {
	if runtime.GOOS == "linux" {
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			return "Wayland"
		}
		if os.Getenv("DISPLAY") != "" {
			return "X11"
		}
		return "TTY"
	}
	return "n/a"
}

func visualSettingsFull(fast bool) string {
	if runtime.GOOS == "linux" && !fast && commandExists("gsettings") {
		base := visualSettings(false)
		cursor, _ := runCmd(200*time.Millisecond, "gsettings", "get", "org.gnome.desktop.interface", "cursor-theme")
		if strings.TrimSpace(cursor) != "" {
			return base + " | cursor " + trimQuotes(cursor)
		}
		return base
	}
	return visualSettings(fast)
}

func terminalFull(_ bool) string {
	t := terminalInfo()
	for _, k := range []string{"TERMINAL_FONT", "KITTY_FONT", "ALACRITTY_FONT"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return t + " | font " + v
		}
	}
	return t
}

func cpuDetailed(fast bool) string {
	base := cpuInfo(fast)
	extras := []string{}
	if runtime.GOOS == "linux" {
		if v := readFirstLine("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor"); v != "" {
			extras = append(extras, "gov "+v)
		}
		if b, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cache/index3/size"); err == nil {
			c := strings.TrimSpace(string(b))
			if c != "" {
				extras = append(extras, "cache "+c)
			}
		}
	}
	if len(extras) == 0 {
		return base
	}
	return base + " | " + strings.Join(extras, " | ")
}

func gpuDetailed(fast bool) string {
	if runtime.GOOS == "linux" {
		if v := fastLinuxGPU(); v != "" {
			if !fast && commandExists("lspci") {
				if out, _ := runCmd(400*time.Millisecond, "lspci", "-nnk"); out != "" {
					driver := ""
					for _, line := range strings.Split(out, "\n") {
						t := strings.TrimSpace(line)
						if strings.HasPrefix(t, "Kernel driver in use:") {
							driver = strings.TrimSpace(strings.TrimPrefix(t, "Kernel driver in use:"))
							break
						}
					}
					if driver != "" {
						return v + " | driver " + driver
					}
				}
			}
			return v
		}
	}
	return gpuInfo(fast)
}

func memoryWithSwap(_ bool) string {
	mem := memoryInfo()
	if runtime.GOOS != "linux" {
		return mem
	}
	var totalKB, freeKB int64
	if b, err := os.ReadFile("/proc/meminfo"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "SwapTotal:") {
				totalKB = parseMeminfoKB(line)
			}
			if strings.HasPrefix(line, "SwapFree:") {
				freeKB = parseMeminfoKB(line)
			}
		}
	}
	if totalKB <= 0 {
		return mem
	}
	used := totalKB - freeKB
	return mem + " | swap " + formatBytes(used*1024) + " / " + formatBytes(totalKB*1024)
}

func diskRootUsage(_ bool) string {
	total, free, err := diskUsage("/")
	if err != nil {
		return "unknown"
	}
	used := total - free
	return fmt.Sprintf("%s / %s", formatBytes(used), formatBytes(total))
}

func diskMountsSummary(_ bool) string {
	if runtime.GOOS != "linux" {
		return diskRootUsage(false)
	}
	mountType := map[string]string{}
	if b, err := os.ReadFile("/proc/mounts"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			f := strings.Fields(line)
			if len(f) >= 3 {
				if _, ok := mountType[f[1]]; !ok {
					mountType[f[1]] = f[2]
				}
			}
		}
	}
	mounts := []string{"/", "/home", "/boot"}
	parts := []string{}
	for _, m := range mounts {
		total, free, err := diskUsage(m)
		if err != nil {
			continue
		}
		used := total - free
		fs := mountType[m]
		if fs == "" {
			fs = "fs"
		}
		parts = append(parts, fmt.Sprintf("%s (%s) %s/%s", m, fs, formatBytes(used), formatBytes(total)))
	}
	if len(parts) == 0 {
		return diskRootUsage(false)
	}
	return strings.Join(parts, " | ")
}

func batteryInfo(fast bool) string {
	switch runtime.GOOS {
	case "linux":
		batDir, ok := linuxBatteryDir()
		if !ok {
			return "n/a"
		}
		capacity := strings.TrimSpace(readFirstLine(filepath.Join(batDir, "capacity")))
		status := strings.ToLower(strings.TrimSpace(readFirstLine(filepath.Join(batDir, "status"))))
		if capacity == "" {
			return "n/a"
		}
		switch status {
		case "charging", "full", "not charging":
			return capacity + "% (plugged in)"
		case "discharging":
			return capacity + "% (not plugged in)"
		default:
			return capacity + "%"
		}
	case "darwin":
		out, _ := runCmd(350*time.Millisecond, "pmset", "-g", "batt")
		if strings.TrimSpace(out) == "" {
			return "n/a"
		}
		pctRe := regexp.MustCompile(`(\d+)%`)
		m := pctRe.FindStringSubmatch(out)
		if len(m) != 2 {
			return "n/a"
		}
		pct := m[1]
		lower := strings.ToLower(out)
		if strings.Contains(lower, "ac power") || strings.Contains(lower, "charging") || strings.Contains(lower, "charged") {
			return pct + "% (plugged in)"
		}
		if strings.Contains(lower, "battery power") || strings.Contains(lower, "discharging") {
			return pct + "% (not plugged in)"
		}
		return pct + "%"
	case "windows":
		pct, pluggedState, hasBattery, ok := windowsBatteryAPI()
		if !ok || !hasBattery {
			return "n/a"
		}
		if pct < 0 {
			if pluggedState == 1 {
				return "unknown% (plugged in)"
			}
			if pluggedState == 0 {
				return "unknown% (not plugged in)"
			}
			return "unknown"
		}
		if pluggedState == 1 {
			return fmt.Sprintf("%d%% (plugged in)", pct)
		}
		if pluggedState == 0 {
			return fmt.Sprintf("%d%% (not plugged in)", pct)
		}
		if fast {
			return fmt.Sprintf("%d%%", pct)
		}
		return fmt.Sprintf("%d%% (power state unknown)", pct)
	default:
		return "n/a"
	}
}

func linuxBatteryDir() (string, bool) {
	entries, err := os.ReadDir("/sys/class/power_supply")
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "BAT") {
			return filepath.Join("/sys/class/power_supply", name), true
		}
	}
	return "", false
}

func networkSummary(_ bool) string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil || ipNet.IP.IsLoopback() {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue
			}
			return iface.Name + " " + ip4.String()
		}
	}
	return "disconnected"
}

func localeInfo(_ bool) string {
	if runtime.GOOS == "windows" {
		if wf := windowsFacts(); wf.Valid && strings.TrimSpace(wf.Locale) != "" {
			return wf.Locale
		}
		if !windowsSlowProbesEnabled() {
			if v := strings.TrimSpace(os.Getenv("LANG")); v != "" {
				return v
			}
		} else {
			if out, _ := runCmd(300*time.Millisecond, "powershell", "-NoProfile", "-Command", "(Get-WinUserLanguageList | Select-Object -First 1 -ExpandProperty LanguageTag)"); out != "" {
				return out
			}
		}
	}
	for _, k := range []string{"LC_ALL", "LANG", "LC_MESSAGES"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return "unknown"
}

func loadSummary(_ bool) string {
	if runtime.GOOS == "linux" {
		if b, err := os.ReadFile("/proc/loadavg"); err == nil {
			parts := strings.Fields(string(b))
			if len(parts) >= 3 {
				return fmt.Sprintf(
					"1m %s | 5m %s | 15m %s (%d cores)",
					parts[0],
					parts[1],
					parts[2],
					runtime.NumCPU(),
				)
			}
		}
	}
	return "n/a"
}

func temperaturesSummary(_ bool) string {
	if runtime.GOOS != "linux" {
		return "n/a"
	}
	zones, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	if len(zones) == 0 {
		return "n/a"
	}
	temps := []string{}
	for i, p := range zones {
		if i >= 2 {
			break
		}
		raw := readFirstLine(p)
		if raw == "" {
			continue
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}
		if v > 1000 {
			v = v / 1000.0
		}
		temps = append(temps, fmt.Sprintf("%.1fC", v))
	}
	if len(temps) == 0 {
		return "n/a"
	}
	return strings.Join(temps, " | ")
}
