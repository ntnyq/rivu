//go:build !darwin && !linux

package files

import (
	"fmt"
	"os"
)

func RenameExclusive(_ *os.File, _, _ string) error {
	return fmt.Errorf("safe payload publication is unavailable on this platform")
}
