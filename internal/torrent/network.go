package torrent

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	bt "github.com/anacrolix/torrent"
	"github.com/ntnyq/rivu/internal/netproxy"
)

type proxyPeerDialer struct {
	bt.Dialer
	resolve        func(*http.Request) (*url.URL, error)
	proxyHostnames bool
	routing        *peerRouting
}

func (d proxyPeerDialer) Dial(ctx context.Context, address string) (net.Conn, error) {
	// Peer connections use the HTTPS/CONNECT route for the peer's authority.
	// Resolve on every dial so changing system proxy/TUN modes needs no restart.
	request, err := http.NewRequestWithContext(ctx, http.MethodConnect, "https://"+address, nil)
	if err != nil {
		return nil, err
	}
	eligible := d.canTunnel(address)
	route, err := d.resolve(request)
	if err != nil {
		if eligible && d.routing != nil && !errors.Is(err, context.Canceled) {
			d.routing.record(proxyLookupFailure(err))
		}
		return nil, fmt.Errorf("resolve peer proxy: %w", err)
	}
	if route == nil {
		return d.Dialer.Dial(ctx, address)
	}
	if !strings.HasPrefix(d.DialerNetwork(), "tcp") {
		// Do not race a direct uTP connection against a proxied TCP connection.
		return nil, fmt.Errorf("peer proxy requires TCP; UDP needs a system route")
	}
	if !eligible {
		return nil, fmt.Errorf("peer address does not match proxy dialer")
	}
	return d.dialTunnel(ctx, address, route)
}

// The engine races its IPv4 and IPv6 dialers. A proxy bypasses their normal
// address family check, so only one TCP dialer attempts a tunnel per address.
func (d proxyPeerDialer) canTunnel(address string) bool {
	network := d.DialerNetwork()
	if !strings.HasPrefix(network, "tcp") {
		return false
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return d.proxyHostnames
	}
	return (network != "tcp4" || ip.Unmap().Is4()) && (network != "tcp6" || !ip.Unmap().Is4())
}

func (d proxyPeerDialer) dialTunnel(ctx context.Context, address string, route *url.URL) (net.Conn, error) {
	proxyCtx := ctx
	cancel := func() {}
	if d.routing != nil && d.routing.fallback.Load() {
		if deadline, ok := ctx.Deadline(); ok {
			// Leave part of upstream's dial budget for the system route, even
			// when there are enough waiting peers to shorten that budget.
			reserve := min(time.Until(deadline)/2, 3*time.Second)
			proxyCtx, cancel = context.WithDeadline(ctx, deadline.Add(-reserve))
		}
	}
	conn, err := netproxy.DialTunnel(proxyCtx, address, route)
	cancel()
	if err == nil || d.routing == nil || errors.Is(err, context.Canceled) {
		return conn, err
	}
	status := proxyTunnelFailure(err)
	if !d.routing.fallback.Load() || ctx.Err() != nil {
		d.routing.record(status)
		return nil, err
	}

	conn, fallbackErr := d.Dialer.Dial(ctx, address)
	status.FallbackTried = true
	if fallbackErr != nil {
		status.FallbackFailed = peerDialCause(fallbackErr)
	}
	d.routing.record(status)
	if fallbackErr != nil {
		return nil, errors.Join(err, fallbackErr)
	}
	return conn, nil
}

func (r *peerRouting) configure(client *bt.Client) {
	// Retain the engine's bound IPv4/IPv6 TCP and uTP dialers for direct/TUN
	// routes, but route each outgoing connection through the current policy.
	var dialers []bt.Dialer
	for _, listener := range client.Listeners() {
		if dialer, ok := listener.(bt.Dialer); ok {
			dialers = append(dialers, dialer)
		}
	}
	for _, dialer := range peerRouteDialers(dialers, netproxy.ForRequest, r) {
		client.AddDialer(dialer)
	}
}

func peerRouteDialers(dialers []bt.Dialer, resolve func(*http.Request) (*url.URL, error), routing *peerRouting) []bt.Dialer {
	hostnameDialer := -1
	for i, dialer := range dialers {
		if strings.HasPrefix(dialer.DialerNetwork(), "tcp") && hostnameDialer == -1 {
			hostnameDialer = i
		}
	}
	routes := make([]bt.Dialer, 0, len(dialers))
	for i, dialer := range dialers {
		routes = append(routes, proxyPeerDialer{Dialer: dialer, resolve: resolve, proxyHostnames: i == hostnameDialer, routing: routing})
	}
	return routes
}
