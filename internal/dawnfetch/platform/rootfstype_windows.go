//go:build windows

// this file resolves filesystem type for windows root drives.
package platform

import (
	"strings"
)

func RootFSTypeOS(path string) string {
	target := strings.TrimSpace(path)
	if target == "" || target == "/" {
		target = "C:\\"
	}
	return strings.TrimSpace(WindowsVolumeFSTypeAPI(target))
}
