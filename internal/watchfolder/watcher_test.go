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
	var w Watcher
	count := 0
	submit := func(_ context.Context, path, _ string, data []byte) error {
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

func TestStableContentAndRetry(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "test.torrent")
	if err := os.WriteFile(source, []byte("first"), 0600); err != nil {
		t.Fatal(err)
	}
	var w Watcher
	var keys []string
	fail := false
	submit := func(_ context.Context, path, key string, _ []byte) error {
		if fail {
			return errors.New("save failed")
		}
		if path != source || key == "" {
			t.Fatalf("invalid watched input %q, %q", path, key)
		}
		keys = append(keys, key)
		return nil
	}
	scan := func() {
		t.Helper()
		if err := w.Scan(t.Context(), dir, submit); err != nil {
			t.Fatal(err)
		}
	}
	scan()
	if len(keys) != 0 {
		t.Fatal("imported before stable")
	}
	scan()
	if len(keys) != 1 {
		t.Fatal("not imported")
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
	if len(keys) != 2 || keys[0] == keys[1] {
		t.Fatal("changed content did not get a distinct key after retry")
	}
}
