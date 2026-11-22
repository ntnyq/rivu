package app

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/session"
	torrentengine "github.com/ntnyq/rivu/internal/torrent"
	"github.com/ntnyq/rivu/internal/watchfolder"
)

func (a *application) watchFolder(ctx context.Context) {
	var w watchfolder.Watcher
	journal := filepath.Join(filepath.Dir(a.store.Path()), "watch-folder.json")
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
				w = watchfolder.Watcher{}
				lastError = ""
				continue
			}
			err := a.session.MigrateWatchJournal(journal)
			if err == nil {
				err = w.Scan(ctx, s.WatchFolder, func(ctx context.Context, path, key string, data []byte) error {
					return a.importWatched(ctx, s, path, key, data)
				})
			}
			if err == nil {
				lastError = ""
			} else if !errors.Is(err, session.ErrQuiesced) && !errors.Is(err, context.Canceled) && err.Error() != lastError {
				lastError = err.Error()
				mygo.RunOnMain(func() { a.report(err) })
			}
		}
	}
}

func (a *application) importWatched(ctx context.Context, s config.Settings, path, key string, data []byte) error {
	if a.session.Watched(key) {
		return nil
	}
	source, err := (torrentengine.Parser{}).ParseBytes(path, data)
	if err != nil {
		return err
	}
	if config.IsDefaultDestination(s.Destination) {
		if err := files.CreateDownloadDirectory(s.Destination); err != nil {
			return err
		}
	}
	var tags []string
	if s.WatchTag != "" {
		tags = []string{s.WatchTag}
	}
	_, err = a.session.Import(ctx, session.Import{WatchKey: key, Source: source, Destination: s.Destination, Run: true, SeedRatio: s.SeedRatio, SeedLimitSeconds: s.SeedLimitSeconds, Tags: tags})
	return err
}
