// Package watchfolder imports stable local torrent files without moving or deleting them.
package watchfolder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Watcher struct {
	previous map[string]string
}

// Scan submits stable files with a path and content key for durable deduplication.
func (w *Watcher) Scan(ctx context.Context, folder string, submit func(context.Context, string, string, []byte) error) error {
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
			if !errors.Is(err, os.ErrNotExist) {
				failures = append(failures, err)
			}
			continue
		}
		if !stat.Mode().IsRegular() || stat.Size() > 16<<20 {
			continue
		}
		file, err := root.Open(entry.Name())
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				failures = append(failures, err)
			}
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
		if w.previous[path] != fingerprint {
			continue
		}
		if err := submit(ctx, path, key, data); err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", entry.Name(), err))
		}
	}
	return errors.Join(failures...)
}
