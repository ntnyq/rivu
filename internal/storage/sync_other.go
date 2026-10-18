//go:build !darwin && !linux

package storage

// Windows cannot fsync a directory through an unprivileged handle. File
// contents are flushed before replacement; metadata durability depends on the
// filesystem. Other systems have compilation support only.
func syncDirectory(string) error { return nil }
