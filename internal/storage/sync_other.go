//go:build !darwin && !linux

package storage

// Other systems have compilation support only; see the release acceptance matrix.
func syncDirectory(string) error { return nil }
