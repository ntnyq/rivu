package files

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRejectUnsafeMembers(t *testing.T) {
	for _, names := range [][]string{{"../escape"}, {"/absolute"}, {"a", "A"}, {"a", "a/b"}, {"a\\b"}, {"a/../b"}} {
		if ValidateMembers(names) == nil {
			t.Fatalf("accepted %v", names)
		}
	}
}
func TestConflictsAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "existing"), nil, 0o600)
	if CheckConflicts(dir, []string{"existing"}) == nil {
		t.Fatal("accepted overwrite")
	}
	t.Run("symlink", func(t *testing.T) {
		if err := os.Symlink(t.TempDir(), filepath.Join(dir, "link")); err != nil {
			if runtime.GOOS == "windows" {
				t.Skipf("symlink privilege unavailable: %v", err)
			}
			t.Fatal(err)
		}
		if CheckConflicts(dir, []string{"link/file"}) == nil {
			t.Fatal("accepted symlink")
		}
	})
	if err := CheckConflicts(dir, []string{"safe/file"}); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsMemberNames(t *testing.T) {
	for _, name := range []string{"CON", "con.txt", "NUL.iso", "AUX", "PRN", "COM1", "com9.log", "LPT1", "LPT9.ext", "COM¹.txt", "LPT²", "LPT³", "CONIN$", "CONOUT$", "CON .txt", "a<", "a>", `a"`, "a|", "a?", "a*", "a\x1f", "a.", "a "} {
		if validWindowsComponent(name) {
			t.Errorf("accepted Windows filename %q", name)
		}
		if runtime.GOOS == "windows" && ValidateMembers([]string{"bundle/" + name}) == nil {
			t.Errorf("accepted Windows member %q", name)
		}
	}
	for _, name := range []string{"console.txt", "COM0", "COM10", "LPT10", "file.iso", "下载.iso"} {
		if !validWindowsComponent(name) {
			t.Errorf("rejected valid filename %q", name)
		}
	}
	if runtime.GOOS != "windows" {
		for _, name := range []string{"CON", "a?", "a*"} {
			if err := ValidateMembers([]string{name}); err != nil {
				t.Errorf("Windows rules leaked to this platform: %v", err)
			}
		}
	}
}

func TestCreateDownloadDirectory(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "Rivu")
	for range 2 {
		if err := CreateDownloadDirectory(dir); err != nil {
			t.Fatal(err)
		}
	}
	missingParent := filepath.Join(parent, "unmounted", "Downloads")
	if err := CreateDownloadDirectory(filepath.Join(missingParent, "Rivu")); !os.IsNotExist(err) {
		t.Fatalf("recreated missing downloads parent: %v", err)
	}
	if _, err := os.Stat(filepath.Join(parent, "unmounted")); !os.IsNotExist(err) {
		t.Fatalf("created an absent volume directory: %v", err)
	}
	file := filepath.Join(parent, "file")
	if err := os.WriteFile(file, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := CreateDownloadDirectory(file); err == nil {
		t.Fatal("accepted a file as the download directory")
	}
}
