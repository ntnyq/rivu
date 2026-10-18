package updater

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestWindowsProxyAPI(t *testing.T) {
	for _, procedure := range []*windows.LazyProc{getUserProxyConfig, openProxySession, closeProxySession, setProxyTimeouts, getAutomaticProxy, globalFree} {
		if err := procedure.Find(); err != nil {
			t.Fatal(err)
		}
	}
	var config userProxyConfig
	ok, _, err := getUserProxyConfig.Call(uintptr(unsafe.Pointer(&config)))
	defer freeProxyString(config.autoConfigURL)
	defer freeProxyString(config.proxy)
	defer freeProxyString(config.bypass)
	if ok == 0 && !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		t.Fatalf("read current user's proxy configuration: %v", err)
	}
}

func TestWindowsPACRouting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proxy.pac" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		_, _ = w.Write([]byte(`function FindProxyForURL(url, host) {
			if (host == "direct.example.com") { return "DIRECT"; }
			return "PROXY 127.0.0.1:7890";
		}`))
	}))
	defer server.Close()
	for _, test := range []struct {
		name, target, script, want string
		wantError                  bool
	}{
		{"PAC proxy", "https://github.com/manifest", "/proxy.pac", "http://127.0.0.1:7890", false},
		{"PAC DIRECT overrides manual proxy", "https://direct.example.com/manifest", "/proxy.pac", "", false},
		{"PAC lookup on redirected host", "https://release-assets.githubusercontent.com/archive", "/proxy.pac", "http://127.0.0.1:7890", false},
		{"PAC error cannot bypass to manual proxy", "https://github.com/manifest", "/missing.pac", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			target, err := url.Parse(test.target)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			var resolver windowsProxyResolver
			proxy, err := resolver.resolve(ctx, target, windowsProxySettings{pacURL: server.URL + test.script, proxy: "127.0.0.1:9999"}, resolveWindowsProxy)
			if (err != nil) != test.wantError {
				t.Fatalf("PAC error=%v wantError=%v", err, test.wantError)
			}
			got := ""
			if proxy != nil {
				got = proxy.String()
			}
			if got != test.want {
				t.Fatalf("PAC proxy=%q want %q", got, test.want)
			}
		})
	}
}

func TestWindowsManualProxyBeforeAutoDetection(t *testing.T) {
	target, err := url.Parse("https://github.com/manifest")
	if err != nil {
		t.Fatal(err)
	}
	proxy, err := resolveWindowsProxy(target, windowsProxySettings{autoDetect: true, proxy: "127.0.0.1:7890"})
	if err != nil || proxy == nil || proxy.String() != "http://127.0.0.1:7890" {
		t.Fatalf("manual proxy not used: %v %v", proxy, err)
	}
}

func TestWindowsUnavailableAutoDetection(t *testing.T) {
	var resolver windowsProxyResolver
	target, _ := url.Parse("https://github.com/manifest")
	calls := 0
	resolve := func(*url.URL, windowsProxySettings) (*url.URL, error) {
		calls++
		return nil, context.DeadlineExceeded
	}
	for range 2 {
		proxy, err := resolver.resolve(context.Background(), target, windowsProxySettings{autoDetect: true}, resolve)
		if proxy != nil || err != nil {
			t.Fatalf("unavailable WPAD blocked direct route: %v %v", proxy, err)
		}
	}
	if calls != 1 {
		t.Fatalf("repeated unavailable discovery: %d", calls)
	}
	proxy, err := resolver.resolve(context.Background(), target, windowsProxySettings{proxy: "127.0.0.1:7890"}, resolve)
	if err != nil || proxy == nil {
		t.Fatalf("new manual proxy was ignored: %v %v", proxy, err)
	}
	_, err = resolver.resolve(context.Background(), target, windowsProxySettings{pacURL: "http://example.com/proxy.pac"}, resolve)
	if err == nil {
		t.Fatal("explicit PAC failure was hidden by WPAD fallback")
	}
}
