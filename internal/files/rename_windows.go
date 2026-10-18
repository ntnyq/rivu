package files

import (
	"errors"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// RenameExclusive uses directory-relative NT operations so a renamed parent
// cannot redirect publication. ReplaceIfExists remains false, including for links.
func RenameExclusive(dir *os.File, old, new string) error {
	defer runtime.KeepAlive(dir)
	for _, name := range []string{old, new} {
		if !fs.ValidPath(name) || strings.ContainsAny(name, "/\\:") || !validWindowsComponent(name) {
			return &os.LinkError{Op: "rename", Old: old, New: new, Err: os.ErrInvalid}
		}
	}
	fail := func(err error) error {
		var status windows.NTStatus
		if errors.As(err, &status) {
			err = status.Errno()
		}
		return &os.LinkError{Op: "rename", Old: old, New: new, Err: err}
	}
	name, err := windows.NewNTUnicodeString(old)
	if err != nil {
		return fail(err)
	}
	attrs := windows.OBJECT_ATTRIBUTES{RootDirectory: windows.Handle(dir.Fd()), ObjectName: name, Attributes: windows.OBJ_DONT_REPARSE}
	attrs.Length = uint32(unsafe.Sizeof(attrs))
	var handle windows.Handle
	var iosb windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&handle, windows.DELETE|windows.SYNCHRONIZE, &attrs, &iosb, nil, 0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_OPEN,
		windows.FILE_NON_DIRECTORY_FILE|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	if err != nil {
		return fail(err)
	}
	defer windows.CloseHandle(handle)
	target, err := windows.UTF16FromString(new)
	if err != nil {
		return fail(err)
	}
	var layout fileRenameInformation
	size := int(unsafe.Offsetof(layout.FileName)) + 2*(len(target)-1)
	buffer := make([]byte, max(size, int(unsafe.Sizeof(layout))))
	info := (*fileRenameInformation)(unsafe.Pointer(&buffer[0]))
	info.RootDirectory = windows.Handle(dir.Fd())
	info.FileNameLength = uint32(2 * (len(target) - 1))
	copy(unsafe.Slice(&info.FileName[0], len(target)-1), target)
	if err = windows.NtSetInformationFile(handle, &iosb, &buffer[0], uint32(len(buffer)), windows.FileRenameInformation); err != nil {
		return fail(err)
	}
	return nil
}

// FILE_RENAME_INFORMATION has pointer-sized alignment on both Windows architectures.
type fileRenameInformation struct {
	ReplaceIfExists uint32
	RootDirectory   windows.Handle
	FileNameLength  uint32
	FileName        [1]uint16
}
