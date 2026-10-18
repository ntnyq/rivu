package config

import (
	"os"
	"path/filepath"
)

// DefaultDestination follows the OS downloads folder. Saved choices are never migrated.
func DefaultDestination() string {
	dir, err := downloadsDirectory()
	if err == nil && filepath.IsAbs(dir) {
		return filepath.Join(dir, "Rivu")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Downloads", "Rivu")
}

// IsDefaultDestination also recognizes the default used before OS folder discovery.
// Only these application-owned defaults may be created automatically.
func IsDefaultDestination(dir string) bool {
	if dir == DefaultDestination() {
		return true
	}
	home, err := os.UserHomeDir()
	return err == nil && dir == filepath.Join(home, "Downloads", "Rivu")
}
