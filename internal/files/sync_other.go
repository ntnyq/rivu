//go:build !windows

package files

import "os"

func SyncDirectory(dir *os.File) error { return dir.Sync() }
