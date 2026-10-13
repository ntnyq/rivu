package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/session"
	torrentengine "github.com/ntnyq/rivu/internal/torrent"
	"github.com/ntnyq/rivu/internal/watchfolder"
)

func (a *application) watchFolder(ctx context.Context) {
	w := watchfolder.Watcher{Journal: filepath.Join(filepath.Dir(a.store.Path()), "watch-folder.json")}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	lastError := ""
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s := a.store.Snapshot()
			if !s.WatchEnabled {
				continue
			}
			err := w.Scan(ctx, s.WatchFolder, func(ctx context.Context, path string, data []byte) error {
				home, _ := os.UserHomeDir()
				if s.Destination == filepath.Join(home, "Downloads", "Rivu") {
					if err := os.MkdirAll(s.Destination, 0755); err != nil {
						return err
					}
				}
				source, err := (torrentengine.Parser{}).ParseBytes(path, data)
				if err != nil {
					return err
				}
				var tags []string
				if s.WatchTag != "" {
					tags = []string{s.WatchTag}
				}
				_, err = a.session.Import(ctx, session.Import{Source: source, Destination: s.Destination, Run: true, SeedRatio: s.SeedRatio, SeedLimitSeconds: s.SeedLimitSeconds, Tags: tags})
				return err
			})
			if err == nil {
				lastError = ""
			} else if err.Error() != lastError {
				lastError = err.Error()
				mygo.RunOnMain(func() { a.report(err) })
			}
		}
	}
}
