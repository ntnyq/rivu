//go:build !darwin && !linux && !windows

package files

import (
	"fmt"
	"os"
)

func FileIdentity(file *os.File) (string, error) {
	return "", fmt.Errorf("safe payload ownership is unavailable on this platform")
}
