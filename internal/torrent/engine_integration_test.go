//go:build integration

package torrent

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"github.com/anacrolix/torrent/bencode"
	"os"
	"path/filepath"
	"testing"
	"time"

	bt "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
)

func TestLocalSwarmDownloadPauseAndRestore(t *testing.T) {
	seedDir, meta, payload := fixture(t)
	testLocalTransfer(t, seedDir, meta, payload, false)
}
func testLocalTransfer(t *testing.T, seedDir string, meta, payload []byte, fromMagnet bool) {
	t.Helper()
	cfg := bt.NewDefaultClientConfig()
	cfg.DataDir = seedDir
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
	seedSpec, err := bt.TorrentSpecFromMetaInfoErr(mi)
	if err != nil {
		t.Fatal(err)
	}
	seeding, _, err := addSpec(seed, seedSpec)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	if err = seeding.VerifyDataContext(ctx); err != nil {
		t.Fatal(err)
	}
	s, err := parseMeta(meta)
	if err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	metadata := filepath.Join(t.TempDir(), "test.torrent")
	if err = os.WriteFile(metadata, meta, 0o600); err != nil {
		t.Fatal(err)
	}
	settings := config.Default()
	settings.DHT = false
	settings.PEX = false
	engine := New(settings)
	defer engine.Close()
	r := session.Record{ID: "fixture", Source: s, Metadata: metadata, Destination: destination}
	if fromMagnet {
		r.Metadata = ""
		r.Source = session.Source{Raw: "magnet:?xt=urn:btih:" + s.Identity.V1, Identity: s.Identity}
	}

	if err = engine.Attach(r); err != nil {
		t.Fatal(err)
	}
	h := engine.handles[r.ID]
	h.torrent.AddClientPeer(seed)
	if fromMagnet {
		for {
			obs := engine.Observe(r.ID)
			if obs.Error != "" {
				t.Fatal(obs.Error)
			}
			if obs.Metadata != nil {
				break
			}
			select {
			case <-ctx.Done():
				t.Fatal("metadata timed out")
			case <-time.After(20 * time.Millisecond):
			}
		}
		if entries, err := os.ReadDir(destination); err != nil || len(entries) != 0 {
			t.Fatal("metadata wrote payload", err)
		}
		if err = engine.Stop(r.ID); err != nil {
			t.Fatal(err)
		}
		r.Source = s
		r.Metadata = metadata
		if err = engine.Attach(r); err != nil {
			t.Fatal(err)
		}
		h = engine.handles[r.ID]
		h.torrent.AddClientPeer(seed)
	}

	time.Sleep(200 * time.Millisecond)
	if _, err = os.Stat(filepath.Join(destination, "payload.txt")); !os.IsNotExist(err) {
		t.Fatal("payload touched before admission", err)
	}
	if err = engine.Start(r.ID, s.Files); err != nil {
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
			t.Fatal("local transfer timed out", obs)
		case <-time.After(50 * time.Millisecond):
		}
	}
	if err = engine.Stop(r.ID); err != nil {
		t.Fatal(err)
	}
	if len(engine.handles) != 0 {
		t.Fatal("pause left a peer handle")
	}
	got, err := os.ReadFile(filepath.Join(destination, "payload.txt"))
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatal("download integrity failed", err)
	}
	if err = engine.Attach(r); err != nil {
		t.Fatal(err)
	}
	if err = engine.Start(r.ID, s.Files); err != nil {
		t.Fatal(err)
	}
	for {
		obs := engine.Observe(r.ID)
		if len(obs.Files) == 1 && obs.Files[0].Done == int64(len(payload)) {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("restore verification timed out")
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func TestLocalMagnetMetadataIsPayloadFree(t *testing.T) {
	dir, meta, data := fixture(t)
	testLocalTransfer(t, dir, meta, data, true)
}
func TestLocalV2AndHybridTransfer(t *testing.T) {
	for _, hybrid := range []bool{false, true} {
		name := "v2"
		if hybrid {
			name = "hybrid"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			data := bytes.Repeat([]byte{42}, 16384)
			if err := os.WriteFile(filepath.Join(dir, "payload.txt"), data, 0600); err != nil {
				t.Fatal(err)
			}
			root := sha256.Sum256(data)
			info := metainfo.Info{Name: "payload.txt", MetaVersion: 2, PieceLength: 16384, FileTree: metainfo.FileTree{File: metainfo.FileTreeFile{Length: int64(len(data)), PiecesRoot: string(root[:])}}}
			if hybrid {
				sum := sha1.Sum(data)
				info.Length = int64(len(data))
				info.Pieces = sum[:]
			}
			encoded, err := bencode.Marshal(&info)
			if err != nil {
				t.Fatal(err)
			}
			mi := metainfo.MetaInfo{InfoBytes: encoded}
			var b bytes.Buffer
			if err = mi.Write(&b); err != nil {
				t.Fatal(err)
			}
			testLocalTransfer(t, dir, b.Bytes(), data, false)
		})
	}
}

func TestSelectedFileAcrossPieceBoundary(t *testing.T) {
	seedDir := t.TempDir()
	bundle := filepath.Join(seedDir, "bundle")
	if err := os.Mkdir(bundle, 0700); err != nil {
		t.Fatal(err)
	}
	selected := bytes.Repeat([]byte("a"), 5000)
	skipped := bytes.Repeat([]byte("b"), 50000)
	if err := os.WriteFile(filepath.Join(bundle, "a.bin"), selected, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "b.bin"), skipped, 0600); err != nil {
		t.Fatal(err)
	}
	info := metainfo.Info{PieceLength: 16384}
	if err := info.BuildFromFilePath(bundle); err != nil {
		t.Fatal(err)
	}
	encoded, err := bencode.Marshal(&info)
	if err != nil {
		t.Fatal(err)
	}
	mi := &metainfo.MetaInfo{InfoBytes: encoded}
	var data bytes.Buffer
	if err = mi.Write(&data); err != nil {
		t.Fatal(err)
	}
	cfg := bt.NewDefaultClientConfig()
	cfg.DataDir = seedDir
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
	seeding, err := seed.AddTorrent(mi)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	if err = seeding.VerifyDataContext(ctx); err != nil {
		t.Fatal(err)
	}
	source, err := parseMeta(data.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if source.PieceLength != 16384 || source.Files[1].Offset != int64(len(selected)) {
		t.Fatalf("parser did not retain shared-piece geometry: %+v", source)
	}
	source.Files[1].Priority = 0
	metaPath := filepath.Join(t.TempDir(), "fixture.torrent")
	if err = os.WriteFile(metaPath, data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	prefs := config.Default()
	prefs.DHT = false
	prefs.PEX = false
	engine := New(prefs)
	defer engine.Close()
	record := session.Record{ID: "selection", Source: source, Metadata: metaPath, Destination: t.TempDir()}
	if err = engine.Attach(record); err != nil {
		t.Fatal(err)
	}
	engine.handles[record.ID].torrent.AddClientPeer(seed)
	if err = engine.Start(record.ID, source.Files); err != nil {
		t.Fatal(err)
	}
	for {
		obs := engine.Observe(record.ID)
		if obs.Error != "" {
			t.Fatal(obs.Error)
		}
		if len(obs.Files) == 2 && obs.Files[0].Done == int64(len(selected)) {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("selected transfer timed out", obs)
		case <-time.After(20 * time.Millisecond):
		}
	}
	if err = engine.Stop(record.ID); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(record.Destination, "bundle", "a.bin"))
	if err != nil || !bytes.Equal(got, selected) {
		t.Fatal("selected bytes differ", err)
	}
	overlap, err := os.ReadFile(filepath.Join(record.Destination, "bundle", "b.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if len(overlap) != 16384-len(selected) || !bytes.Equal(overlap, skipped[:len(overlap)]) {
		t.Fatalf("unexpected skipped overlap: %d", len(overlap))
	}
}
