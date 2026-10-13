package torrent

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ntnyq/rivu/internal/problem"
	"os"
	"time"

	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
)

// Preview has its own client and empty storage root. It never admits payload,
// persists a task, or opens the user's destination. Closing the draft cancels it.
func Preview(ctx context.Context, raw string, settings config.Settings) (session.Source, error) {
	source, err := (Parser{}).Parse(raw)
	if err != nil || len(source.Meta) > 0 {
		return source, err
	}
	if err := ctx.Err(); err != nil {
		return source, err
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	dir, err := os.MkdirTemp("", "rivu-preview-")
	if err != nil {
		return source, err
	}
	defer os.RemoveAll(dir)
	engine := New(settings)
	defer engine.Close()
	if err = engine.Attach(session.Record{ID: "preview", Source: source, Destination: dir}); err != nil {
		return source, err
	}
	h := engine.handles["preview"]
	select {
	case <-ctx.Done():
		return source, problem.New(problem.Metadata, "preview canceled or timed out", ctx.Err())
	case <-h.torrent.GotInfo():
	}
	var data bytes.Buffer
	mi := h.torrent.Metainfo()
	if err = mi.Write(&data); err != nil {
		return source, err
	}
	resolved, err := parseMeta(data.Bytes())
	if err != nil {
		return source, err
	}
	if !resolved.Identity.Matches(source.Identity) {
		return source, fmt.Errorf("metadata identity mismatch")
	}
	resolved.Raw = raw
	return resolved, nil
}
