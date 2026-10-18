package files

import (
	"math"

	"golang.org/x/sys/windows"
)

func FreeSpace(dir string) (int64, error) {
	free, _, err := DiskSpace(dir)
	return free, err
}

func DiskSpace(dir string) (free, total int64, err error) {
	name, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return 0, 0, err
	}
	var available, capacity uint64
	if err = windows.GetDiskFreeSpaceEx(name, &available, &capacity, nil); err != nil {
		return 0, 0, err
	}
	return int64(min(available, math.MaxInt64)), int64(min(capacity, math.MaxInt64)), nil
}
