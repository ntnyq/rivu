//go:build integration

package torrent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	bt "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
)

// Neither the tracker nor the advertised peer is directly reachable. A real
// download must discover the peer through the HTTP proxy and tunnel BitTorrent
// through CONNECT; checking a transport field alone would miss either path.
func TestLocalSwarmThroughSystemProxyRoute(t *testing.T) {
	seedDir, meta, payload := fixture(t)
	mi, err := metainfo.Load(bytes.NewReader(meta))
	if err != nil {
		t.Fatal(err)
	}
	cfg := bt.NewDefaultClientConfig()
	cfg.DataDir, cfg.ListenPort = seedDir, 0
	cfg.NoDHT, cfg.DisablePEX, cfg.DisableTrackers = true, true, true
	cfg.NoDefaultPortForwarding, cfg.Seed = true, true
	seed, err := bt.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer seed.Close()
	seeding, err := seed.AddTorrent(mi)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err = seeding.VerifyDataContext(ctx); err != nil {
		t.Fatal(err)
	}
	var announces, tunnels atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			if r.Host != "192.0.2.1:6881" {
				t.Errorf("unexpected tunnel destination %q", r.Host)
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			upstream, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", seed.LocalPort()))
			if err != nil {
				t.Error(err)
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			defer upstream.Close()
			conn, buffered, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close()
			tunnels.Add(1)
			_, _ = buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
			_ = buffered.Flush()
			done := make(chan struct{})
			go func() { defer close(done); _, _ = io.Copy(upstream, buffered); _ = upstream.Close() }()
			_, _ = io.Copy(conn, upstream)
			_ = conn.Close()
			<-done
			return
		}
		if r.URL.Host != "tracker.invalid" {
			t.Errorf("unexpected tracker %q", r.URL.Host)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		announces.Add(1)
		// 192.0.2.1:6881 is a documentation-only address, mapped by this proxy.
		response, _ := bencode.Marshal(map[string]any{"interval": 60, "peers": string([]byte{192, 0, 2, 1, 0x1a, 0xe1})})
		_, _ = w.Write(response)
	}))
	defer proxy.Close()
	for _, name := range []string{"http_proxy", "https_proxy", "NO_PROXY", "no_proxy"} {
		t.Setenv(name, "")
	}
	// Environment and Windows system configuration feed the same route resolver.
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	mi.Announce = "http://tracker.invalid/announce"
	mi.AnnounceList = nil
	var encoded bytes.Buffer
	if err = mi.Write(&encoded); err != nil {
		t.Fatal(err)
	}
	source, err := parseMeta(encoded.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	metadata := filepath.Join(t.TempDir(), "fixture.torrent")
	if err = os.WriteFile(metadata, encoded.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	settings := config.Default()
	settings.DHT, settings.PEX = false, false
	engine := New(settings)
	defer engine.Close()
	record := session.Record{ID: "proxy", Source: source, Metadata: metadata, Destination: t.TempDir()}
	if err = engine.Attach(record); err != nil {
		t.Fatal(err)
	}
	if err = engine.Start(record.ID, source.Files); err != nil {
		t.Fatal(err)
	}
	for {
		obs := engine.Observe(record.ID)
		if obs.Error != "" {
			t.Fatal(obs.Error)
		}
		if len(obs.Files) == 1 && obs.Files[0].Done == int64(len(payload)) {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("download stalled: announces=%d tunnels=%d observation=%+v", announces.Load(), tunnels.Load(), obs)
		case <-time.After(20 * time.Millisecond):
		}
	}
	if announces.Load() == 0 || tunnels.Load() == 0 {
		t.Fatal("download bypassed proxy")
	}
	if err = engine.Stop(record.ID); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(record.Destination, "payload.txt"))
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("proxied download contents differ: %v", err)
	}
}
