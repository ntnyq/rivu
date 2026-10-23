package netproxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

const maxConnectHeaderBytes = 64 << 10

// DialTunnel connects a TCP peer through the selected proxy. Failures never
// fall back to a direct connection. UDP needs an OS route such as a TUN device.
func DialTunnel(ctx context.Context, address string, route *url.URL) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if route == nil || route.Hostname() == "" {
		return nil, fmt.Errorf("missing proxy address")
	}
	port := route.Port()
	if port == "" {
		switch route.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		case "socks5", "socks5h":
			port = "1080"
		}
	}
	endpoint := net.JoinHostPort(route.Hostname(), port)
	dialer := &net.Dialer{}
	if route.Scheme == "socks5" || route.Scheme == "socks5h" {
		var auth *proxy.Auth
		if route.User != nil {
			password, _ := route.User.Password()
			auth = &proxy.Auth{User: route.User.Username(), Password: password}
		}
		d, err := proxy.SOCKS5("tcp", endpoint, auth, dialer)
		if err != nil {
			return nil, err
		}
		return d.(proxy.ContextDialer).DialContext(ctx, "tcp", address)
	}
	if route.Scheme != "http" && route.Scheme != "https" {
		return nil, fmt.Errorf("unsupported proxy protocol")
	}
	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return nil, fmt.Errorf("connect proxy: %w", err)
	}
	if route.Scheme == "https" {
		secure := tls.Client(conn, &tls.Config{ServerName: route.Hostname(), MinVersion: tls.VersionTLS12})
		if err = secure.HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, fmt.Errorf("proxy TLS handshake: %w", err)
		}
		conn = secure
	}
	// Context cancellation must interrupt both writing CONNECT and reading its
	// response. Wait for an already-running cancellation callback before return.
	canceled := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { _ = conn.Close(); close(canceled) })
	defer func() {
		if !stop() {
			<-canceled
		}
	}()
	request := &http.Request{Method: http.MethodConnect, URL: &url.URL{Opaque: address}, Host: address, Header: make(http.Header)}
	if route.User != nil {
		password, _ := route.User.Password()
		credentials := route.User.Username() + ":" + password
		request.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credentials)))
	}
	if err = request.Write(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("write proxy CONNECT: %w", err)
	}
	// ReadResponse has no header size limit of its own. Limit only the proxy
	// handshake; the peer stream must remain unrestricted after CONNECT.
	reader := bufio.NewReader(io.LimitReader(conn, maxConnectHeaderBytes))
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		conn.Close()
		// A proxy can return arbitrary text, including credentials. Do not
		// propagate malformed status/header contents into diagnostics.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("invalid proxy CONNECT response")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: HTTP %d", response.StatusCode)
	}
	if err = ctx.Err(); err != nil {
		conn.Close()
		return nil, err
	}
	return &tunnelConn{Conn: conn, reader: io.MultiReader(reader, conn)}, nil
}

// The proxy can send peer bytes in the same packet as the CONNECT response.
type tunnelConn struct {
	net.Conn
	reader io.Reader
}

func (c *tunnelConn) Read(p []byte) (int, error) { return c.reader.Read(p) }
