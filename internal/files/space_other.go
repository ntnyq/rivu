//go:build !darwin && !linux && !windows

package files

import "errors"

func FreeSpace(string) (int64, error) {
	return 0, errors.New("free-space query unavailable on this platform")
}

func DiskSpace(string) (int64, int64, error) {
	return 0, 0, errors.New("disk-space query unavailable on this platform")
}
