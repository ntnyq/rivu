package torrent

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
)

func TestTrackerDiagnosticEventsAndLifecycle(t *testing.T) {
	var store trackerDiagnostics
	hash := metainfo.Hash{1}
	key := hash.HexString()
	store.retain(key)
	logger := slog.New(trackerDiagnosticHandler{store: &store}).With("short infohash", hash, "url", "https://user:secret@tracker.test/private-passkey?token=secret")
	logger.Debug("announcing")
	if rows := store.snapshot(key); len(rows) != 1 || rows[0].State != "announcing" {
		t.Fatalf("announce event missing: %+v", rows)
	}
	logger.Debug("announced", "resp", tracker.AnnounceResponse{Interval: 60}, "err", nil)
	rows := store.snapshot(key)
	if len(rows) != 1 || rows[0].State != "ok" || rows[0].Detail != "0 peers" || rows[0].Next.IsZero() || rows[0].Host != "https://tracker.test" {
		t.Fatalf("wrong successful diagnostic: %+v", rows)
	}
	rows[0].State = "mutated snapshot"
	if store.snapshot(key)[0].State != "ok" {
		t.Fatal("snapshot aliases mutable state")
	}
	for message, category := range map[string]string{
		"timeout secret": "timeout", "deadline exceeded secret": "timeout", "no such host secret": "dns",
		"connection refused secret": "refused", "HTTP 403 secret": "authorization", "arbitrary remote secret": "announce",
	} {
		logger.Error("announced", "err", errors.New(message))
		rows = store.snapshot(key)
		if len(rows) != 1 || rows[0].State != "failed" || rows[0].Detail != category || !rows[0].Next.IsZero() || strings.Contains(fmt.Sprint(rows), "secret") {
			t.Fatalf("wrong or unredacted failure: %+v", rows)
		}
	}
	logger.Debug("announced", "err", nil) // Missing response is not a success.
	if store.snapshot(key)[0].State != "failed" {
		t.Fatal("accepted incomplete event")
	}
	// A rejected duplicate attachment must not erase the live torrent's rows.
	store.retain(key)
	store.release(key)
	if len(store.snapshot(key)) != 1 {
		t.Fatal("lost retained diagnostics")
	}
	store.release(key)
	logger.Debug("announcing") // A late announce must not resurrect a removed task.
	if len(store.snapshot(key)) != 0 {
		t.Fatal("retained removed task")
	}
}

func TestTrackerDiagnosticsConcurrentEvents(t *testing.T) {
	var store trackerDiagnostics
	hash := metainfo.Hash{2}
	key := hash.HexString()
	store.retain(key)
	logger := slog.New(trackerDiagnosticHandler{store: &store}).With("short infohash", hash)
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Go(func() {
			log := logger.With("url", fmt.Sprintf("udp4://tracker%d.test:80/announce", i))
			for range 100 {
				log.Debug("announcing")
				log.Debug("announced", "resp", tracker.AnnounceResponse{Interval: 60}, "err", nil)
				store.snapshot(key)
			}
		})
	}
	wg.Wait()
	if rows := store.snapshot(key); len(rows) != 4 {
		t.Fatalf("lost concurrent trackers: %+v", rows)
	}
	store.release(key)
}
