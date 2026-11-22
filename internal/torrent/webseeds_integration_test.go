//go:build integration

package torrent

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
)

func TestLocalWebSeedRequiresTorrentOptIn(t *testing.T) {
	_, data, payload := fixture(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.ServeContent(w, r, "payload.txt", time.Time{}, bytes.NewReader(payload))
	}))
	defer server.Close()
	mi, err := metainfo.Load(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	mi.UrlList = metainfo.UrlList{server.URL + "/payload.txt"}
	var metadata bytes.Buffer
	if err := mi.Write(&metadata); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "sample.torrent")
	if err := os.WriteFile(path, metadata.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	source, err := (Parser{}).Parse(path)
	if err != nil || !source.HasWebSeeds {
		t.Fatalf("web seed missing in review: %v", err)
	}
	settings := config.Default()
	settings.DHT = false
	settings.PEX = false
	for _, test := range []struct {
		name, raw string
		optIn     bool
	}{
		{"default", path, false},
		{"resolved-magnet", "magnet:?xt=urn:btih:" + source.Identity.V1, true},
		{"allowed", path, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			engine := New(settings)
			defer engine.Close()
			destination := t.TempDir()
			src := source
			src.Raw = test.raw
			record := session.Record{ID: "sample", Source: src, Metadata: path, Destination: destination, AllowWebSeeds: test.optIn}
			if err := engine.Attach(record); err != nil {
				t.Fatal(err)
			}
			if err := engine.Start(record.ID, src.Files); err != nil {
				t.Fatal(err)
			}
			if !test.optIn || strings.HasPrefix(test.raw, "magnet:") {
				if got := engine.handles[record.ID].torrent.WebseedPeerConns(); len(got) != 0 || requests.Load() != 0 {
					t.Fatal("an unauthorized web seed was registered or requested")
				}
				return
			}
			deadline := time.After(10 * time.Second)
			ticker := time.NewTicker(25 * time.Millisecond)
			defer ticker.Stop()
			for {
				obs := engine.Observe(record.ID)
				if obs.Error != "" {
					t.Fatal(obs.Error)
				}
				if len(obs.Files) == 1 && obs.Files[0].Done == int64(len(payload)) {
					break
				}
				select {
				case <-deadline:
					t.Fatalf("web seed transfer stalled: downloaded=%d requests=%d", obs.Downloaded, requests.Load())
				case <-ticker.C:
				}
			}
			got, err := os.ReadFile(filepath.Join(destination, "payload.txt"))
			if err != nil || !bytes.Equal(got, payload) || requests.Load() == 0 {
				t.Fatalf("web seed download integrity failed: %v", err)
			}
		})
	}
}
