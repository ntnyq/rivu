package updater

import (
	"errors"

	"golang.org/x/sys/windows"
)

func platformRetryableCheckError(err error) bool {
	// syscall.ECONNRESET/ECONNABORTED are synthetic Unix-compatible values on
	// Windows; real socket failures carry the distinct Winsock error numbers.
	return errors.Is(err, windows.WSAECONNRESET) || errors.Is(err, windows.WSAECONNABORTED) ||
		errors.Is(err, windows.WSAENETRESET) || errors.Is(err, windows.WSAETIMEDOUT) ||
		errors.Is(err, windows.ERROR_NETNAME_DELETED) || errors.Is(err, windows.ERROR_CONNECTION_ABORTED)
}
