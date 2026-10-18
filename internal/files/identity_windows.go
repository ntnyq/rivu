package files

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// FileIdentity queries the open handle, including ReFS's full 128-bit file ID.
// A filesystem that cannot provide stable IDs is rejected rather than claiming ownership.
func FileIdentity(file *os.File) (string, error) {
	defer runtime.KeepAlive(file)
	var info struct {
		VolumeSerialNumber uint64
		FileID             [16]byte
	}
	if err := windows.GetFileInformationByHandleEx(windows.Handle(file.Fd()), windows.FileIdInfo, (*byte)(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		return "", fmt.Errorf("file identity: %w", err)
	}
	return fmt.Sprintf("%d:%x", info.VolumeSerialNumber, info.FileID), nil
}
