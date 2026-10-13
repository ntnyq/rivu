// Package watchfolder imports stable local torrent files without moving or deleting them.
package watchfolder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ntnyq/rivu/internal/storage"
)

type Watcher struct {
	Journal  string
	previous map[string]string
}

// Scan requires the same content in two scans. The durable ledger prevents a
// removed task from being re-imported on restart; changed torrent files are new input.
func (w *Watcher) Scan(ctx context.Context, folder string, submit func(context.Context, string, []byte) error) error {
	ledger := map[string]bool{}
	data, err := os.ReadFile(w.Journal)
	if err == nil {
		err = json.Unmarshal(data, &ledger)
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read watch journal: %w", err)
	}
	if ledger == nil {
		return fmt.Errorf("invalid watch journal")
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(folder)
	if err != nil {
		return err
	}
	defer root.Close()
	var failures []error
	next := map[string]string{}
	defer func() { w.previous = next }()
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".torrent") || entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			continue
		}
		path := filepath.Join(folder, entry.Name())
		stat, err := entry.Info()
		if err != nil {
			return err
		}
		if !stat.Mode().IsRegular() || stat.Size() > 16<<20 {
			continue
		}
		file, err := root.Open(entry.Name())
		if err != nil {
			failures = append(failures, err)
			continue
		}
		current, err := file.Stat()
		if err != nil || !current.Mode().IsRegular() {
			file.Close()
			failures = append(failures, fmt.Errorf("cannot read %s", entry.Name()))
			continue
		}
		data, err := io.ReadAll(io.LimitReader(file, (16<<20)+1))
		file.Close()
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if len(data) > 16<<20 {
			continue
		}
		hash := sha256.Sum256(data)
		fingerprint := hex.EncodeToString(hash[:])
		next[path] = fingerprint
		key := path + ":" + fingerprint
		if ledger[key] || w.previous[path] != fingerprint {
			continue
		}
		if err := submit(ctx, path, data); err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", entry.Name(), err))
			continue
		}
		ledger[key] = true
		encoded, err := json.Marshal(ledger)
		if err != nil {
			return err
		}
		if err = os.MkdirAll(filepath.Dir(w.Journal), 0700); err != nil {
			return err
		}
		if err = storage.WriteFile(w.Journal, encoded, 0600); err != nil {
			return err
		}
	}
	return errors.Join(failures...)
}
