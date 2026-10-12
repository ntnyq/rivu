package files

import (
	"os"
	"path/filepath"
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
	_ = os.Symlink(t.TempDir(), filepath.Join(dir, "link"))
	if CheckConflicts(dir, []string{"link/file"}) == nil {
		t.Fatal("accepted symlink")
	}
	if err := CheckConflicts(dir, []string{"safe/file"}); err != nil {
		t.Fatal(err)
	}
}
