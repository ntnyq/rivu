package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiskSpace(t *testing.T) {
	dir := t.TempDir()
	free, total, err := DiskSpace(dir)
	if err != nil || total <= 0 || free < 0 || free > total {
		t.Fatalf("invalid disk space: %d/%d, %v", free, total, err)
	}
	available, err := FreeSpace(dir)
	if err != nil || available < 0 || available > total {
		t.Fatalf("invalid available space: %d, %v", available, err)
	}
	if _, _, err := DiskSpace(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing directory accepted")
	}
}

func TestFileIdentityAndExclusivePublication(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	directory, err := root.Open(".")
	if err != nil {
		t.Fatal(err)
	}
	defer directory.Close()
	file, err := root.OpenFile("temporary", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	before, err := FileIdentity(file)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Write([]byte("payload")); err != nil {
		t.Fatal(err)
	}
	if err = file.Sync(); err != nil {
		t.Fatal(err)
	}
	if err = RenameExclusive(directory, "temporary", "下载.iso"); err != nil {
		t.Fatal(err)
	}
	if err = SyncDirectory(directory); err != nil {
		t.Fatal(err)
	}
	published, err := root.Open("下载.iso")
	if err != nil {
		t.Fatal(err)
	}
	defer published.Close()
	after, err := FileIdentity(published)
	if err != nil || before != after {
		t.Fatalf("identity changed after write/rename: %s -> %s, %v", before, after, err)
	}
	if _, err := root.Stat("temporary"); !os.IsNotExist(err) {
		t.Fatalf("source survived rename: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "another"), []byte("other"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := RenameExclusive(directory, "another", "下载.iso"); !os.IsExist(err) {
		t.Fatalf("publication must reject existing target: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "下载.iso"))
	if err != nil || string(data) != "payload" {
		t.Fatalf("existing payload changed: %q, %v", data, err)
	}
	identities, err := InspectExisting(dir, []string{"下载.iso"}, true)
	if err != nil || identities["下载.iso"] != before {
		t.Fatalf("adoption identity differs: %v, %v", identities, err)
	}
	other, err := root.Open("another")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	different, err := FileIdentity(other)
	if err != nil || different == before {
		t.Fatalf("unrelated file has same identity: %v", err)
	}
}

func TestExclusivePublicationUsesOpenDirectory(t *testing.T) {
	base := t.TempDir()
	original, moved := filepath.Join(base, "original"), filepath.Join(base, "moved")
	if err := os.Mkdir(original, 0700); err != nil {
		t.Fatal(err)
	}
	// OpenRoot's handle does not share delete access on Windows. Keep it on
	// the parent so only the directory-relative handle follows the rename.
	root, err := os.OpenRoot(base)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	directory, err := root.Open("original")
	if err != nil {
		t.Fatal(err)
	}
	defer directory.Close()
	if err := os.WriteFile(filepath.Join(original, "temporary"), []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(original, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(original, 0700); err != nil {
		t.Fatal(err)
	}
	if err := RenameExclusive(directory, "temporary", "payload"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(moved, "payload"))
	if err != nil || string(data) != "payload" {
		t.Fatalf("publication did not stay anchored: %q, %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(original, "payload")); !os.IsNotExist(err) {
		t.Fatalf("publication followed replaced parent: %v", err)
	}
}
