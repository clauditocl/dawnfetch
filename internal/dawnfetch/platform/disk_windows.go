//go:build windows

// this file reads disk usage using windows api calls.
package platform

import (
	"errors"
	"strings"
)

func DiskUsage(path string) (total int64, free int64, err error) {
	target := strings.TrimSpace(path)
	if target == "" || target == "/" {
		target = "C:\\"
	}

	if total, free, ok := WindowsDiskUsageAPI(target); ok {
		return total, free, nil
	}

	return 0, 0, errors.New("failed to query disk usage")
}
