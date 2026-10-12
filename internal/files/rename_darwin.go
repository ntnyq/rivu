//go:build darwin

package files

import (
	"golang.org/x/sys/unix"
	"os"
)

// RenameExclusive publishes a temporary file within an open directory without
// replacing an existing destination. Both names must be basenames.
func RenameExclusive(dir *os.File, old, new string) error {
	return unix.RenameatxNp(int(dir.Fd()), old, int(dir.Fd()), new, unix.RENAME_EXCL)
}
