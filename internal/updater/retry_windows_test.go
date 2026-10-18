package updater

import (
	"net"
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func TestWinsockCheckRetries(t *testing.T) {
	for _, errno := range []windows.Errno{
		windows.WSAECONNRESET, windows.WSAECONNABORTED, windows.WSAENETRESET,
		windows.WSAETIMEDOUT, windows.ERROR_NETNAME_DELETED, windows.ERROR_CONNECTION_ABORTED,
	} {
		err := &net.OpError{Op: "read", Net: "tcp", Err: &os.SyscallError{Syscall: "wsarecv", Err: errno}}
		if !retryableCheckError(err) {
			t.Fatalf("did not retry actual Windows socket error: %v", err)
		}
	}
}
