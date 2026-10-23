package torrent

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	bt "github.com/anacrolix/torrent"
	"github.com/ntnyq/rivu/internal/netproxy"
)

type proxyPeerDialer struct {
	bt.Dialer
	resolve        func(*http.Request) (*url.URL, error)
	proxyHostnames bool
}

func (d proxyPeerDialer) Dial(ctx context.Context, address string) (net.Conn, error) {
	// Peer connections use the HTTPS/CONNECT route for the peer's authority.
	// Resolve on every dial so changing system proxy/TUN modes needs no restart.
	request, err := http.NewRequestWithContext(ctx, http.MethodConnect, "https://"+address, nil)
	if err != nil {
		return nil, err
	}
	route, err := d.resolve(request)
	if err != nil {
		return nil, fmt.Errorf("resolve peer proxy: %w", err)
	}
	if route == nil {
		return d.Dialer.Dial(ctx, address)
	}
	if !strings.HasPrefix(d.DialerNetwork(), "tcp") {
		// Do not race a direct uTP connection against a proxied TCP connection.
		return nil, fmt.Errorf("peer proxy requires TCP; UDP needs a system route")
	}
	// The engine races its IPv4 and IPv6 dialers for every peer. A tunnel
	// bypasses net.Dialer's address-family check, so enforce it here as well.
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		// Let one dialer send hostnames to the proxy for remote DNS, including
		// when only IPv6 is enabled locally.
		if !d.proxyHostnames {
			return nil, fmt.Errorf("peer hostname belongs to another proxy dialer")
		}
	} else if (d.DialerNetwork() == "tcp4" && !ip.Unmap().Is4()) || (d.DialerNetwork() == "tcp6" && ip.Unmap().Is4()) {
		return nil, fmt.Errorf("peer address does not match dialer family")
	}
	return netproxy.DialTunnel(ctx, address, route)
}

func configurePeerRoutes(client *bt.Client) {
	// Retain the engine's bound IPv4/IPv6 TCP and uTP dialers for direct/TUN
	// routes, but route each outgoing connection through the current policy.
	var dialers []bt.Dialer
	for _, listener := range client.Listeners() {
		if dialer, ok := listener.(bt.Dialer); ok {
			dialers = append(dialers, dialer)
		}
	}
	for _, dialer := range peerRouteDialers(dialers, netproxy.ForRequest) {
		client.AddDialer(dialer)
	}
}

func peerRouteDialers(dialers []bt.Dialer, resolve func(*http.Request) (*url.URL, error)) []bt.Dialer {
	hostnameDialer := -1
	for i, dialer := range dialers {
		if strings.HasPrefix(dialer.DialerNetwork(), "tcp") && hostnameDialer == -1 {
			hostnameDialer = i
		}
	}
	routes := make([]bt.Dialer, 0, len(dialers))
	for i, dialer := range dialers {
		routes = append(routes, proxyPeerDialer{Dialer: dialer, resolve: resolve, proxyHostnames: i == hostnameDialer})
	}
	return routes
}
