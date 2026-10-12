//go:build !darwin && !linux

package files

import (
	"fmt"
	"os"
)

func FileIdentity(info os.FileInfo) (string, error) {
	return "", fmt.Errorf("safe payload ownership is unavailable on this platform")
}
