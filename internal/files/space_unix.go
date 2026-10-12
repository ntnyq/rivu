//go:build darwin || linux

package files

import "golang.org/x/sys/unix"

func FreeSpace(dir string) (int64, error) {
	var s unix.Statfs_t
	err := unix.Statfs(dir, &s)
	return int64(s.Bavail) * int64(s.Bsize), err
}

func DiskSpace(dir string) (free, total int64, err error) {
	var stat unix.Statfs_t
	err = unix.Statfs(dir, &stat)
	return int64(stat.Bavail) * int64(stat.Bsize), int64(stat.Blocks) * int64(stat.Bsize), err
}
