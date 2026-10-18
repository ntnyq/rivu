//go:build integration

package torrent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bt "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/session"
)

func TestLocalPreviewAndAdoptedRepair(t *testing.T) {
	dir, meta, payload := fixture(t)
	cfg := bt.NewDefaultClientConfig()
	cfg.DataDir = dir
	cfg.ListenPort = 0
	cfg.NoDHT = true
	cfg.DisablePEX = true
	cfg.DisableTrackers = true
	cfg.NoDefaultPortForwarding = true
	cfg.Seed = true
	seed, err := bt.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer seed.Close()
	mi, err := metainfo.Load(bytes.NewReader(meta))
	if err != nil {
		t.Fatal(err)
	}
	seeding, err := seed.AddTorrent(mi)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	if err = seeding.VerifyDataContext(ctx); err != nil {
		t.Fatal(err)
	}
	raw := fmt.Sprintf("magnet:?xt=urn:btih:%s&x.pe=127.0.0.1:%d", mi.HashInfoBytes().HexString(), seed.LocalPort())
	prefs := config.Default()
	prefs.DHT = false
	prefs.PEX = false
	tempDir := t.TempDir()
	// Unix uses TMPDIR; Windows uses TMP or TEMP.
	for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(key, tempDir)
	}
	source, err := Preview(ctx, raw, prefs)
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Files) != 1 || len(source.Meta) == 0 {
		t.Fatal("preview did not resolve files")
	}
	if entries, err := os.ReadDir(os.TempDir()); err != nil || len(entries) != 0 {
		t.Fatal("preview storage leaked", err)
	}
	canceled, stop := context.WithCancel(ctx)
	stop()
	if _, err := Preview(canceled, raw, prefs); !errors.Is(err, context.Canceled) {
		t.Fatal("preview cancellation ignored", err)
	}
	destination := t.TempDir()
	name := filepath.Join(destination, "payload.txt")
	if err = os.WriteFile(name, []byte("damaged"), 0600); err != nil {
		t.Fatal(err)
	}
	adopted, err := files.InspectExisting(destination, []string{"payload.txt"}, true)
	if err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(t.TempDir(), "source.torrent")
	if err = os.WriteFile(metadata, meta, 0600); err != nil {
		t.Fatal(err)
	}
	engine := New(prefs)
	defer engine.Close()
	r := session.Record{ID: "adopted", Source: source, Metadata: metadata, Destination: destination, Adopted: adopted}
	if err = engine.Attach(r); err != nil {
		t.Fatal(err)
	}
	if err = engine.Start(r.ID, source.Files); err != nil {
		t.Fatal(err)
	}
	for {
		obs := engine.Observe(r.ID)
		if obs.Error != "" {
			t.Fatal(obs.Error)
		}
		if len(obs.Files) == 1 && obs.Files[0].Done == int64(len(payload)) {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("adopted repair timed out")
		case <-time.After(20 * time.Millisecond):
		}
	}
	if err = engine.Stop(r.ID); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(name)
	if err != nil || !bytes.Equal(data, payload) {
		t.Fatal("adopted payload not repaired", err)
	}
	owners, err := files.ReadOwnership(metadata + ".files.json")
	if err != nil || owners["payload.txt"] != "" {
		t.Fatal("adopted file gained Trash authority", err)
	}
}

func TestLocalHTTPTrackerDiagnostics(t *testing.T) {
	_, meta, _ := fixture(t)
	tracker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := bencode.Marshal(map[string]any{"interval": int64(60), "peers": ""})
		_, _ = w.Write(data)
	}))
	defer tracker.Close()
	mi, err := metainfo.Load(bytes.NewReader(meta))
	if err != nil {
		t.Fatal(err)
	}
	mi.Announce = tracker.URL + "/private-passkey?token=secret"
	mi.AnnounceList = nil
	var data bytes.Buffer
	if err = mi.Write(&data); err != nil {
		t.Fatal(err)
	}
	source, err := parseMeta(data.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "meta.torrent")
	if err = os.WriteFile(path, data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	prefs := config.Default()
	prefs.DHT = false
	prefs.PEX = false
	engine := New(prefs)
	defer engine.Close()
	r := session.Record{ID: "tracker", Source: source, Metadata: path, Destination: t.TempDir()}
	if err = engine.Attach(r); err != nil {
		t.Fatal(err)
	}
	if err = engine.Start(r.ID, source.Files); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	for {
		engine.diagnosticAt = time.Time{}
		obs := engine.Observe(r.ID)
		for _, row := range obs.Diagnostics {
			if row.State == "ok" {
				if row.Next.IsZero() || strings.Contains(fmt.Sprint(row), "secret") || strings.Contains(fmt.Sprint(row), "passkey") {
					t.Fatal("invalid diagnostic", row)
				}
				return
			}
		}
		select {
		case <-ctx.Done():
			var raw bytes.Buffer
			engine.client.WriteStatus(&raw)
			t.Fatalf("tracker diagnostics absent: %s", raw.String())
		case <-time.After(20 * time.Millisecond):
		}
	}
}
