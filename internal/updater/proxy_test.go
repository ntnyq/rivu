package updater

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/http/httpproxy"
)

func TestWindowsProxySelection(t *testing.T) {
	for _, test := range []struct {
		name, target, servers, bypass, want string
	}{
		{"shared proxy", "https://github.com/manifest", "127.0.0.1:7890", "", "http://127.0.0.1:7890"},
		{"HTTPS-specific proxy", "https://github.com/manifest", "http=127.0.0.1:8080;https=127.0.0.1:7890", "", "http://127.0.0.1:7890"},
		{"HTTP-specific proxy", "http://example.com/manifest", "http=127.0.0.1:8080;https=127.0.0.1:7890", "", "http://127.0.0.1:8080"},
		{"no proxy for this protocol", "https://github.com/manifest", "http=127.0.0.1:8080", "", ""},
		{"domain bypass", "https://api.example.com/manifest", "127.0.0.1:7890", "*.example.com;localhost", ""},
		{"local bypass", "http://intranet/manifest", "127.0.0.1:7890", "<local>", ""},
		{"loopback bypass", "http://127.0.0.1/manifest", "127.0.0.1:7890", "", ""},
		{"redirect host keeps proxy", "https://release-assets.githubusercontent.com/archive", "127.0.0.1:7890", "*.example.com;<local>", "http://127.0.0.1:7890"},
		{"IPv6 proxy", "https://github.com/manifest", "[::1]:7890", "", "http://[::1]:7890"},
	} {
		t.Run(test.name, func(t *testing.T) {
			target, err := url.Parse(test.target)
			if err != nil {
				t.Fatal(err)
			}
			proxy, err := windowsProxy(target, test.servers, test.bypass)
			if err != nil {
				t.Fatal(err)
			}
			got := ""
			if proxy != nil {
				got = proxy.String()
			}
			if got != test.want {
				t.Fatalf("proxy=%q want %q", got, test.want)
			}
		})
	}
}

func TestUpdateProxyPrecedence(t *testing.T) {
	for _, test := range []struct {
		name        string
		environment httpproxy.Config
		want        string
		systemCalls int
	}{
		{"system proxy without environment", httpproxy.Config{}, "http://127.0.0.1:7890", 1},
		{"explicit environment proxy", httpproxy.Config{HTTPSProxy: "http://127.0.0.1:8080"}, "http://127.0.0.1:8080", 0},
		{"HTTP environment does not override HTTPS system proxy", httpproxy.Config{HTTPProxy: "http://127.0.0.1:8080"}, "http://127.0.0.1:7890", 1},
		{"environment bypass with proxy", httpproxy.Config{HTTPSProxy: "http://127.0.0.1:8080", NoProxy: "github.com"}, "", 0},
		{"environment bypass without proxy", httpproxy.Config{NoProxy: "github.com"}, "", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, "https://github.com/manifest", nil)
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			proxy, err := proxyForRequest(request, &test.environment, func(context.Context, *url.URL) (*url.URL, error) {
				calls++
				return url.Parse("http://127.0.0.1:7890")
			})
			if err != nil {
				t.Fatal(err)
			}
			got := ""
			if proxy != nil {
				got = proxy.String()
			}
			if got != test.want || calls != test.systemCalls {
				t.Fatalf("proxy=%q system calls=%d", got, calls)
			}
		})
	}
}

func TestUpdateUsesSystemProxy(t *testing.T) {
	var calls atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Host == "update.invalid" {
			w.Header().Set("Location", "http://assets.invalid/manifest")
			w.WriteHeader(http.StatusFound)
			return
		}
		if r.URL.Host != "assets.invalid" {
			t.Errorf("proxy received wrong target: %s", r.URL)
		}
		_, _ = io.WriteString(w, "manifest from proxy")
	}))
	defer proxy.Close()
	ctx, closeNetwork := updateNetworkContext(context.Background())
	defer closeNetwork()
	transport := ctx.Value(updateTransportKey{}).(*http.Transport)
	transport.Proxy = func(request *http.Request) (*url.URL, error) {
		return proxyForRequest(request, &httpproxy.Config{}, func(context.Context, *url.URL) (*url.URL, error) {
			return url.Parse(proxy.URL)
		})
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://update.invalid/manifest", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil || string(body) != "manifest from proxy" || calls.Load() != 2 {
		t.Fatalf("proxy request failed: body=%q error=%v calls=%d", body, err, calls.Load())
	}
}

func TestProxyLookupCancellationAndConcurrency(t *testing.T) {
	target, _ := url.Parse("https://github.com/manifest")
	block := make(chan struct{})
	finished := make(chan struct{})
	defer close(block)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := lookupProxy(ctx, target, func(*url.URL) (*url.URL, error) {
		defer close(finished)
		<-block
		return nil, nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("blocked lookup did not cancel: %v", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel2()
	_, err = lookupProxy(ctx2, target, func(*url.URL) (*url.URL, error) {
		t.Error("started another native lookup while first was blocked")
		return nil, nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("queued lookup did not cancel: %v", err)
	}
}
