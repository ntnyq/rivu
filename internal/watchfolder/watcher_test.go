package watchfolder

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInvalidFileDoesNotBlockOtherImports(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.torrent", "b.torrent"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	w := Watcher{Journal: filepath.Join(t.TempDir(), "ledger.json")}
	count := 0
	submit := func(_ context.Context, path string, data []byte) error {
		if filepath.Base(path) == "a.torrent" {
			return errors.New("invalid torrent")
		}
		if string(data) != "b.torrent" {
			t.Fatal("unexpected stable bytes")
		}
		count++
		return nil
	}
	if err := w.Scan(t.Context(), dir, submit); err != nil {
		t.Fatal(err)
	}
	if w.Scan(t.Context(), dir, submit) == nil || count != 1 {
		t.Fatal("valid import blocked by invalid file")
	}
}

func TestStableImportLedgerAndRetry(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "test.torrent")
	if err := os.WriteFile(source, []byte("first"), 0600); err != nil {
		t.Fatal(err)
	}
	w := Watcher{Journal: filepath.Join(t.TempDir(), "ledger.json")}
	count := 0
	fail := false
	submit := func(context.Context, string, []byte) error {
		if fail {
			return errors.New("save failed")
		}
		count++
		return nil
	}
	scan := func() {
		t.Helper()
		if err := w.Scan(t.Context(), dir, submit); err != nil {
			t.Fatal(err)
		}
	}
	scan()
	if count != 0 {
		t.Fatal("imported before stable")
	}
	scan()
	if count != 1 {
		t.Fatal("not imported")
	}
	w = Watcher{Journal: w.Journal}
	scan()
	scan()
	if count != 1 {
		t.Fatal("reimported after restart")
	}
	if err := os.WriteFile(source, []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}
	scan()
	fail = true
	if w.Scan(t.Context(), dir, submit) == nil {
		t.Fatal("failure hidden")
	}
	fail = false
	scan()
	if count != 2 {
		t.Fatal("failed import not retried")
	}
	if err := os.WriteFile(w.Journal, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if w.Scan(t.Context(), dir, submit) == nil {
		t.Fatal("corrupt journal overwritten")
	}
}
