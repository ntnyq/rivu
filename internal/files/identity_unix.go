//go:build darwin || linux

package files

import (
	"fmt"
	"os"
	"syscall"
)

func FileIdentity(file *os.File) (string, error) {
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("file identity unavailable")
	}
	return fmt.Sprintf("%d:%d", st.Dev, st.Ino), nil
}
