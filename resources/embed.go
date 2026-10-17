// Package resources embeds the artwork shipped with the native application.
package resources

import "embed"

//go:embed tray.png icon.svg ui/*.svg
var files embed.FS

// ReadFile reads artwork by its path relative to the resources directory.
func ReadFile(name string) ([]byte, error) {
	return files.ReadFile(name)
}
