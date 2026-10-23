package netproxy

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestConnectPreservesPeerBytesAndAuthenticatesProxy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:secret"))
		if r.Method != http.MethodConnect || r.Host != "peer.invalid:6881" || r.Header.Get("Proxy-Authorization") != wantAuth {
			t.Errorf("wrong CONNECT method, authority or proxy authentication")
			w.WriteHeader(http.StatusProxyAuthRequired)
			return
		}
		conn, buffered, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_, _ = buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\npeer handshake")
		_ = buffered.Flush()
		_, _ = io.Copy(conn, buffered)
	}))
	defer server.Close()
	route, _ := url.Parse(server.URL)
	route.User = url.UserPassword("user", "secret")
	conn, err := DialTunnel(t.Context(), "peer.invalid:6881", route)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(time.Second))
	data := make([]byte, len("peer handshake"))
	if _, err = io.ReadFull(conn, data); err != nil || string(data) != "peer handshake" {
		t.Fatalf("lost buffered peer handshake: %q %v", data, err)
	}
	if _, err = conn.Write([]byte("piece")); err != nil {
		t.Fatal(err)
	}
	data = make([]byte, 5)
	if _, err = io.ReadFull(conn, data); err != nil || string(data) != "piece" {
		t.Fatalf("tunnel did not remain usable after setup context ended: %q %v", data, err)
	}
}

func TestConnectRefusalIsRedacted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusProxyAuthRequired)
		_, _ = io.WriteString(w, "secret credentials echoed by proxy")
	}))
	defer server.Close()
	route, _ := url.Parse(server.URL)
	route.User = url.UserPassword("private-user", "private-password")
	_, err := DialTunnel(t.Context(), "peer.invalid:6881", route)
	if err == nil || err.Error() != "proxy CONNECT failed: HTTP 407" {
		t.Fatalf("wrong or unredacted proxy error: %v", err)
	}
}

func TestConnectResponseHeaderLimit(t *testing.T) {
	const limit = 64 << 10
	for _, size := range []int{limit - 1, limit, limit + 1, 16 << 20} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			prefix, suffix := "HTTP/1.1 200 OK\r\nX-Padding: ", "\r\n\r\n"
			header := prefix + strings.Repeat("x", size-len(prefix)-len(suffix)) + suffix
			payload := strings.Repeat("peer data", limit)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
				_, _ = io.WriteString(conn, header+payload)
			}))
			defer server.Close()
			route, _ := url.Parse(server.URL)
			conn, err := DialTunnel(t.Context(), "peer.invalid:6881", route)
			if conn != nil {
				defer conn.Close()
			}
			if size > limit {
				if err == nil {
					t.Fatal("accepted oversized CONNECT response header")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
			got, err := io.ReadAll(conn)
			if err != nil || string(got) != payload {
				t.Fatalf("header limit truncated tunnel payload: bytes=%d err=%v", len(got), err)
			}
		})
	}
}

func TestConnectCancellationClosesPendingHandshake(t *testing.T) {
	closed := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_, _ = io.Copy(io.Discard, conn)
		close(closed)
	}))
	defer server.Close()
	route, _ := url.Parse(server.URL)
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	_, err := DialTunnel(ctx, "peer.invalid:6881", route)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("blocked handshake did not cancel: %v", err)
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("canceled tunnel left a connection open")
	}
}

func TestHTTPSProxyVerifiesCertificate(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("untrusted proxy certificate was accepted")
	}))
	defer server.Close()
	route, _ := url.Parse(server.URL)
	_, err := DialTunnel(t.Context(), "peer.invalid:6881", route)
	if err == nil || !strings.Contains(err.Error(), "proxy TLS handshake") {
		t.Fatalf("expected certificate rejection: %v", err)
	}
}

func TestSOCKSProxyResolvesPeerRemotely(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		var header [2]byte
		if _, err := io.ReadFull(conn, header[:]); err != nil {
			t.Error(err)
			return
		}
		methods := make([]byte, int(header[1]))
		if _, err := io.ReadFull(conn, methods); err != nil {
			t.Error(err)
			return
		}
		_, _ = conn.Write([]byte{5, 0})
		var request [5]byte
		if _, err := io.ReadFull(conn, request[:]); err != nil {
			t.Error(err)
			return
		}
		if request[0] != 5 || request[1] != 1 || request[3] != 3 {
			t.Errorf("expected SOCKS5 CONNECT with remote DNS: %v", request)
			return
		}
		target := make([]byte, int(request[4])+2)
		if _, err := io.ReadFull(conn, target); err != nil {
			t.Error(err)
			return
		}
		if string(target[:len(target)-2]) != "peer.invalid" || target[len(target)-2] != 0x1a || target[len(target)-1] != 0xe1 {
			t.Error("wrong SOCKS peer address")
			return
		}
		_, _ = conn.Write(append([]byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 1}, []byte("peer")...))
	}()
	route, _ := url.Parse("socks5://" + listener.Addr().String())
	conn, err := DialTunnel(t.Context(), "peer.invalid:6881", route)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	data, err := io.ReadAll(conn)
	if err != nil || string(data) != "peer" {
		t.Fatalf("SOCKS tunnel failed: %q %v", data, err)
	}
	<-done
}
