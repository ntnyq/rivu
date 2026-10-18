//go:build !windows

package updater

func platformRetryableCheckError(error) bool { return false }
