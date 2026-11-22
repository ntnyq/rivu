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
	"time"

	"github.com/ntnyq/rivu/internal/session"

	bt "github.com/anacrolix/torrent"
)

type observedPeerDialer struct {
	network string
	calls   int
	conn    net.Conn
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
			for i, d := range peerRouteDialers(dialers, http.ProxyURL(proxy), nil) {
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
	for _, d := range peerRouteDialers(dialers, http.ProxyURL(proxy), nil) {
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
	if d.conn != nil {
		return d.conn, nil
	}
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

func TestPeerProxyFallbackUsesOnlyMatchingTCPDialer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	proxy, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	routing := &peerRouting{}
	matched := &observedPeerDialer{network: "tcp4"}
	wrongFamily := &observedPeerDialer{network: "tcp6"}
	udp := &observedPeerDialer{network: "udp4"}
	dialers := peerRouteDialers([]bt.Dialer{matched, wrongFamily, udp}, http.ProxyURL(proxy), routing)
	const address = "192.0.2.1:6881"
	if _, err := dialers[0].Dial(t.Context(), address); err == nil || matched.calls != 0 {
		t.Fatalf("proxy refusal did not stop the dial by default: %v", err)
	}
	status := routing.snapshot()
	if status.Failure != session.PeerProxyResponse || status.Code != 403 || status.FallbackTried {
		t.Fatalf("missing strict proxy failure: %+v", status)
	}

	routing.fallback.Store(true)
	conn, remote := net.Pipe()
	defer conn.Close()
	defer remote.Close()
	matched.conn = conn
	got, err := dialers[0].Dial(t.Context(), address)
	if err != nil || got != conn || matched.calls != 1 {
		t.Fatalf("system route did not connect: conn=%v calls=%d err=%v", got, matched.calls, err)
	}
	for _, dialer := range dialers[1:] {
		if _, err := dialer.Dial(t.Context(), address); err == nil {
			t.Fatalf("%s incorrectly completed the dial", dialer.DialerNetwork())
		}
	}
	status = routing.snapshot()
	if wrongFamily.calls != 0 || udp.calls != 0 || status.Code != 403 || !status.FallbackTried || status.FallbackFailed != "" {
		t.Fatalf("fallback was duplicated or incorrectly reported: %+v", status)
	}

	matched.conn = nil
	if _, err := dialers[0].Dial(t.Context(), address); !errors.Is(err, directPeerRoute) {
		t.Fatalf("missing system route error: %v", err)
	}
	status = routing.snapshot()
	if status.Code != 403 || !status.FallbackTried || status.FallbackFailed != session.PeerDialOther {
		t.Fatalf("missing failed system route: %+v", status)
	}
}

func TestPeerProxyFallbackKeepsTimeForSystemRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	proxy, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	routing := &peerRouting{}
	routing.fallback.Store(true)
	conn, remote := net.Pipe()
	defer conn.Close()
	defer remote.Close()
	direct := &observedPeerDialer{network: "tcp4", conn: conn}
	dialer := proxyPeerDialer{Dialer: direct, resolve: http.ProxyURL(proxy), routing: routing}
	ctx, cancel := context.WithTimeout(t.Context(), 600*time.Millisecond)
	defer cancel()
	got, err := dialer.Dial(ctx, "192.0.2.1:6881")
	if err != nil || got != conn || ctx.Err() != nil {
		t.Fatalf("proxy timeout exhausted the system route budget: %v, context: %v", err, ctx.Err())
	}
	status := routing.snapshot()
	if status.Cause != session.PeerDialTimeout || !status.FallbackTried || status.FallbackFailed != "" {
		t.Fatalf("missing timeout or fallback result: %+v", status)
	}

	cancel()
	if _, err := dialer.Dial(ctx, "192.0.2.1:6881"); !errors.Is(err, context.Canceled) || direct.calls != 1 {
		t.Fatalf("canceled dial used the system route: %v", err)
	}
}

func TestPeerProxyLookupFailureRemainsStrict(t *testing.T) {
	routing := &peerRouting{}
	routing.fallback.Store(true)
	direct := &observedPeerDialer{network: "tcp4"}
	dialer := proxyPeerDialer{
		Dialer: direct, routing: routing,
		resolve: func(*http.Request) (*url.URL, error) { return nil, errors.New("private PAC URL") },
	}
	if _, err := dialer.Dial(t.Context(), "192.0.2.1:6881"); err == nil || direct.calls != 0 {
		t.Fatalf("lookup failure bypassed proxy configuration: %v", err)
	}
	status := routing.snapshot()
	if status.Failure != session.PeerProxyLookup || status.Cause != session.PeerDialOther || status.FallbackTried {
		t.Fatalf("wrong lookup diagnostic: %+v", status)
	}
}
