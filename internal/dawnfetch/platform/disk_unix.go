//go:build !windows

// this file reads disk usage using unix statfs.
package platform

import "syscall"

func DiskUsage(path string) (total int64, free int64, err error) {
	var st syscall.Statfs_t
	if err = syscall.Statfs(path, &st); err != nil {
		return 0, 0, err
	}
	total = int64(st.Blocks) * int64(st.Bsize)
	free = int64(st.Bavail) * int64(st.Bsize)
	return total, free, nil
}
