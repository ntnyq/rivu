package torrent

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	bt "github.com/anacrolix/torrent"
)

type observedPeerDialer struct {
	network string
	calls   int
}

func TestPeerProxyDoesNotDuplicateAddressFamilies(t *testing.T) {
	for _, address := range []string{"192.0.2.1:6881", "[2001:db8::1]:6881", "[::ffff:192.0.2.1]:6881", "peer.invalid:6881"} {
		t.Run(address, func(t *testing.T) {
			var connects atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				connects.Add(1)
				w.WriteHeader(http.StatusProxyAuthRequired)
			}))
			defer server.Close()
			proxy, _ := url.Parse(server.URL)
			dialers := []bt.Dialer{&observedPeerDialer{network: "tcp4"}, &observedPeerDialer{network: "tcp6"}, &observedPeerDialer{network: "udp4"}, &observedPeerDialer{network: "udp6"}}
			for i, d := range peerRouteDialers(dialers, http.ProxyURL(proxy)) {
				direct := dialers[i].(*observedPeerDialer)
				conn, err := d.Dial(t.Context(), address)
				if conn != nil {
					conn.Close()
				}
				if err == nil || direct.calls != 0 {
					t.Fatalf("unexpected direct dial or successful proxy: %s %v", direct.network, err)
				}
			}
			if got := connects.Load(); got != 1 {
				t.Fatalf("got %d proxy connections for one peer, want 1", got)
			}
		})
	}
}

func TestPeerProxyHostnameWithIPv6Only(t *testing.T) {
	var connects atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connects.Add(1)
		w.WriteHeader(http.StatusProxyAuthRequired)
	}))
	defer server.Close()
	proxy, _ := url.Parse(server.URL)
	dialers := []bt.Dialer{&observedPeerDialer{network: "udp6"}, &observedPeerDialer{network: "tcp6"}}
	for _, d := range peerRouteDialers(dialers, http.ProxyURL(proxy)) {
		conn, err := d.Dial(t.Context(), "peer.invalid:6881")
		if conn != nil {
			conn.Close()
		}
		if err == nil {
			t.Fatal("expected proxy rejection")
		}
	}
	if connects.Load() != 1 {
		t.Fatal("IPv6-only client did not route hostname through proxy")
	}
}

var directPeerRoute = errors.New("direct peer route")

func (d *observedPeerDialer) DialerNetwork() string { return d.network }
func (d *observedPeerDialer) Dial(context.Context, string) (net.Conn, error) {
	d.calls++
	return nil, directPeerRoute
}

func TestPeerRouteChangesAndProxyFailureDoesNotBypass(t *testing.T) {
	var connects atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connects.Add(1)
		w.WriteHeader(http.StatusProxyAuthRequired)
	}))
	defer server.Close()
	proxy, _ := url.Parse(server.URL)
	var route *url.URL
	var lookupErr error
	for _, network := range []string{"tcp4", "tcp6", "udp4", "udp6"} {
		direct := &observedPeerDialer{network: network}
		dialer := proxyPeerDialer{Dialer: direct, resolve: func(r *http.Request) (*url.URL, error) {
			if r.URL.String() != "https://192.0.2.1:6881" {
				t.Errorf("wrong peer proxy lookup: %s", r.URL)
			}
			return route, lookupErr
		}}
		route, lookupErr = nil, nil
		if _, err := dialer.Dial(t.Context(), "192.0.2.1:6881"); !errors.Is(err, directPeerRoute) {
			t.Fatalf("%s did not use OS/TUN route: %v", network, err)
		}
		route = proxy
		before := connects.Load()
		if _, err := dialer.Dial(t.Context(), "192.0.2.1:6881"); err == nil || errors.Is(err, directPeerRoute) || direct.calls != 1 {
			t.Fatalf("%s bypassed proxy failure: %v", network, err)
		}
		if network[:3] == "udp" && connects.Load() != before {
			t.Fatal("tried to tunnel uTP over HTTP")
		}
		lookupErr = errors.New("PAC unavailable")
		if _, err := dialer.Dial(t.Context(), "192.0.2.1:6881"); !errors.Is(err, lookupErr) || direct.calls != 1 {
			t.Fatalf("%s bypassed proxy lookup failure: %v", network, err)
		}
		route, lookupErr = nil, nil
		if _, err := dialer.Dial(t.Context(), "192.0.2.1:6881"); !errors.Is(err, directPeerRoute) || direct.calls != 2 {
			t.Fatalf("%s did not return to OS/TUN route: %v", network, err)
		}
	}
}
