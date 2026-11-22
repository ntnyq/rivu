package torrent

import (
	"context"
	"errors"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ntnyq/rivu/internal/netproxy"
	"github.com/ntnyq/rivu/internal/session"
)

type peerRouting struct {
	fallback atomic.Bool
	mu       sync.Mutex
	failure  session.PeerProxyStatus
}

func (r *peerRouting) snapshot() session.PeerProxyStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	if time.Since(r.failure.At) > 5*time.Minute {
		return session.PeerProxyStatus{}
	}
	return r.failure
}

func (r *peerRouting) record(status session.PeerProxyStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	status.At = time.Now()
	r.failure = status
}

func proxyLookupFailure(err error) session.PeerProxyStatus {
	status := session.PeerProxyStatus{Failure: session.PeerProxyLookup, Cause: peerDialCause(err)}
	var code syscall.Errno
	if errors.As(err, &code) {
		status.Code = int(code)
	}
	return status
}

func proxyTunnelFailure(err error) session.PeerProxyStatus {
	status := session.PeerProxyStatus{Failure: session.PeerProxyHandshake, Cause: peerDialCause(err)}
	var tunnel *netproxy.TunnelError
	if !errors.As(err, &tunnel) {
		return status
	}
	switch tunnel.Failure {
	case netproxy.TunnelConfig:
		status.Failure = session.PeerProxyConfig
	case netproxy.TunnelConnect:
		status.Failure = session.PeerProxyConnect
	case netproxy.TunnelTLS:
		status.Failure = session.PeerProxyTLS
	case netproxy.TunnelSOCKS:
		status.Failure = session.PeerProxySOCKS
	case netproxy.TunnelResponse:
		status.Failure = session.PeerProxyResponse
		status.Code = tunnel.StatusCode
	}
	return status
}

func peerDialCause(err error) session.PeerDialCause {
	var dns *net.DNSError
	var network net.Error
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &network) && network.Timeout():
		return session.PeerDialTimeout
	case errors.As(err, &dns):
		return session.PeerDialDNS
	case errors.Is(err, syscall.ECONNREFUSED):
		return session.PeerDialRefused
	case errors.Is(err, syscall.ENETUNREACH), errors.Is(err, syscall.EHOSTUNREACH):
		return session.PeerDialUnreachable
	}
	if runtime.GOOS == "windows" {
		// Winsock returns these codes in socket errors. Go's syscall package
		// does not map them to their POSIX counterparts.
		switch {
		case errors.Is(err, syscall.Errno(10060)): // WSAETIMEDOUT
			return session.PeerDialTimeout
		case errors.Is(err, syscall.Errno(10061)): // WSAECONNREFUSED
			return session.PeerDialRefused
		case errors.Is(err, syscall.Errno(10051)), errors.Is(err, syscall.Errno(10065)): // WSAENETUNREACH, WSAEHOSTUNREACH
			return session.PeerDialUnreachable
		}
	}
	return session.PeerDialOther
}
