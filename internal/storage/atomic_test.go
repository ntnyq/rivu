package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReplaceAndCleanup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	for _, text := range []string{"old", "new complete contents"} {
		if err := WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "new complete contents" {
		t.Fatalf("%q, %v", data, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("leaked temporary files: %v, %v", entries, err)
	}
	if err := WriteFile(dir, []byte("bad target"), 0o600); err == nil {
		t.Fatal("replaced directory with a file")
	}
}

func TestExportReplacesSymlinkInsteadOfTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks requires Windows privileges")
	}
	dir := t.TempDir()
	target, link := filepath.Join(dir, "target"), filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(link, []byte("export"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "original" {
		t.Fatalf("modified symlink target: %q, %v", data, err)
	}
}
