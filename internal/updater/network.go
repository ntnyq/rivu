package updater

import (
	"context"
	"net"
	"net/http"
	"time"
)

type updateTransportKey struct{}

type updateTransportRouter struct{ fallback http.RoundTripper }

func (r updateTransportRouter) RoundTrip(request *http.Request) (*http.Response, error) {
	if transport, ok := request.Context().Value(updateTransportKey{}).(http.RoundTripper); ok {
		return transport.RoundTrip(request)
	}
	return r.fallback.RoundTrip(request)
}

func (r updateTransportRouter) CloseIdleConnections() {
	if transport, ok := r.fallback.(interface{ CloseIdleConnections() }); ok {
		transport.CloseIdleConnections()
	}
}

func init() {
	// MyGo v0.5.5 uses http.DefaultClient with no client option. Install this
	// adapter before workers start; only explicitly marked update contexts
	// use our transport. Other HTTP requests retain their original behavior,
	// and http.DefaultTransport stays available to clients that clone it.
	client := *http.DefaultClient
	fallback := client.Transport
	if fallback == nil {
		fallback = http.DefaultTransport
	}
	client.Transport = updateTransportRouter{fallback: fallback}
	http.DefaultClient = &client
}

func updateNetworkContext(ctx context.Context) (context.Context, func()) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = updateProxy
	transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext
	transport.TLSHandshakeTimeout = 5 * time.Second
	transport.ResponseHeaderTimeout = 10 * time.Second
	return context.WithValue(ctx, updateTransportKey{}, http.RoundTripper(transport)), transport.CloseIdleConnections
}
