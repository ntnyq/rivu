package torrent

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/ntnyq/rivu/internal/session"
)

// The pinned engine's WriteStatus reads its announce timer without a lock.
// Consume its structured announcing/announced events instead. This adapter is
// version-sensitive and covered by real HTTP/UDP tracker integration tests.
// Only redacted rows for attached torrents are retained; no logs reach disk.
type trackerDiagnostics struct {
	mu      sync.Mutex
	entries map[string]*trackerDiagnosticEntry
}

type trackerDiagnosticEntry struct {
	refs int
	rows map[[32]byte]session.TrackerStatus
}

func (d *trackerDiagnostics) retain(hash string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.entries == nil {
		d.entries = make(map[string]*trackerDiagnosticEntry)
	}
	if d.entries[hash] == nil {
		d.entries[hash] = &trackerDiagnosticEntry{rows: make(map[[32]byte]session.TrackerStatus)}
	}
	d.entries[hash].refs++
}

func (d *trackerDiagnostics) release(hash string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if entry := d.entries[hash]; entry != nil {
		entry.refs--
		if entry.refs == 0 {
			delete(d.entries, hash)
		}
	}
}

func (d *trackerDiagnostics) snapshot(hash string) []session.TrackerStatus {
	d.mu.Lock()
	defer d.mu.Unlock()
	var rows []session.TrackerStatus
	if entry := d.entries[hash]; entry != nil {
		for _, row := range entry.rows {
			rows = append(rows, row)
		}
	}
	slices.SortFunc(rows, func(a, b session.TrackerStatus) int { return strings.Compare(a.Host, b.Host) })
	return rows
}

type trackerDiagnosticHandler struct {
	store      *trackerDiagnostics
	hash, host string
	key        [32]byte
	grouped    bool
}

func (h trackerDiagnosticHandler) Enabled(context.Context, slog.Level) bool {
	return !h.grouped && h.hash != "" && h.host != ""
}

func (h trackerDiagnosticHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.grouped {
		return h
	}
	for _, attr := range attrs {
		switch attr.Key {
		case "short infohash":
			if hash, ok := attr.Value.Any().(metainfo.Hash); ok {
				h.hash = hash.HexString()
			}
		case "url":
			raw := fmt.Sprint(attr.Value.Any())
			if u, err := url.Parse(raw); err == nil && u.Hostname() != "" {
				h.host, h.key = RedactTracker(raw), sha256.Sum256([]byte(raw))
			}
		}
	}
	return h
}

func (h trackerDiagnosticHandler) WithGroup(name string) slog.Handler {
	if name != "" {
		h.grouped = true
	}
	return h
}

func (h trackerDiagnosticHandler) Handle(_ context.Context, record slog.Record) error {
	if !h.Enabled(context.Background(), record.Level) {
		return nil
	}
	row := session.TrackerStatus{Host: h.host}
	switch record.Message {
	case "announcing":
		row.State = "announcing"
	case "announced":
		var response tracker.AnnounceResponse
		var responseOK bool
		var failure error
		record.Attrs(func(attr slog.Attr) bool {
			switch attr.Key {
			case "resp":
				response, responseOK = attr.Value.Any().(tracker.AnnounceResponse)
			case "err":
				failure, _ = attr.Value.Any().(error)
			}
			return true
		})
		if failure != nil {
			row.State, row.Detail = "failed", trackerFailureCategory(failure)
		} else if responseOK {
			row.State, row.Detail = "ok", fmt.Sprintf("%d peers", len(response.Peers))
			// The tracker supplies an interval, not the engine's exact schedule. The
			// latter also depends on peer demand and is deliberately not inspected.
			if response.Interval > 0 {
				row.Next = record.Time.Add(time.Duration(response.Interval) * time.Second)
			}
		} else {
			return nil
		}
	default:
		return nil
	}
	h.store.mu.Lock()
	defer h.store.mu.Unlock()
	if entry := h.store.entries[h.hash]; entry != nil {
		entry.rows[h.key] = row
	}
	return nil
}

func trackerFailureCategory(err error) string {
	// Remote errors may contain credentials, passkeys, and arbitrary text.
	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "timeout"), strings.Contains(lower, "deadline"):
		return "timeout"
	case strings.Contains(lower, "no such host"):
		return "dns"
	case strings.Contains(lower, "refused"):
		return "refused"
	case strings.Contains(lower, "401"), strings.Contains(lower, "403"), strings.Contains(lower, "unauthor"):
		return "authorization"
	default:
		return "announce"
	}
}
