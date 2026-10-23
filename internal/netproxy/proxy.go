package netproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/http/httpproxy"
)

// WinHTTP's synchronous PAC lookup cannot be interrupted by Go's context.
// Bound callers' wait and permit at most one outstanding native lookup.
var proxyLookupSlot = make(chan struct{}, 1)

func lookupProxy(ctx context.Context, target *url.URL, resolve func(*url.URL) (*url.URL, error)) (*url.URL, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case proxyLookupSlot <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	type result struct {
		proxy *url.URL
		err   error
	}
	done := make(chan result, 1)
	go func() {
		defer func() { <-proxyLookupSlot }()
		proxy, err := resolve(target)
		done <- result{proxy, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-done:
		return result.proxy, result.err
	}
}

// ForRequest resolves each request afresh, preferring environment settings over
// the current Windows system proxy. A nil route uses the OS network (including TUN).
func ForRequest(request *http.Request) (*url.URL, error) {
	return proxyForRequest(request, httpproxy.FromEnvironment(), systemProxy)
}

func proxyForRequest(request *http.Request, environment *httpproxy.Config, system func(context.Context, *url.URL) (*url.URL, error)) (*url.URL, error) {
	// Wildcard listener addresses refer to this host, never a remote proxy's
	// host. The engine's local-client peering helper advertises these addresses.
	if address, err := netip.ParseAddr(request.URL.Hostname()); err == nil && address.IsUnspecified() {
		return nil, nil
	}
	// Explicit environment settings take precedence, including NO_PROXY.
	if (request.URL.Scheme == "http" && environment.HTTPProxy != "") ||
		(request.URL.Scheme == "https" && environment.HTTPSProxy != "") {
		return environment.ProxyFunc()(request.URL)
	}
	if bypassProxy(request.URL, environment.NoProxy) {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(request.Context(), 5*time.Second)
	defer cancel()
	return system(ctx, request.URL)
}

func bypassProxy(target *url.URL, bypass string) bool {
	var rules []string
	for _, rule := range strings.FieldsFunc(bypass, func(r rune) bool { return r == ',' || r == ';' }) {
		rule = strings.TrimSpace(rule)
		if strings.EqualFold(rule, "<local>") {
			if !strings.Contains(target.Hostname(), ".") && !strings.Contains(target.Hostname(), ":") {
				return true
			}
			continue
		}
		rules = append(rules, rule)
	}
	config := httpproxy.Config{
		HTTPProxy: "http://proxy.invalid", HTTPSProxy: "http://proxy.invalid",
		NoProxy: strings.Join(rules, ","),
	}
	proxy, _ := config.ProxyFunc()(target)
	return proxy == nil
}

func windowsProxy(target *url.URL, servers, bypass string) (*url.URL, error) {
	if bypassProxy(target, bypass) {
		return nil, nil
	}
	var address string
	for _, server := range strings.FieldsFunc(servers, func(r rune) bool { return r == ';' || r == ' ' }) {
		protocol, value, named := strings.Cut(server, "=")
		named = named && !strings.Contains(server, "://")
		if named {
			if strings.EqualFold(protocol, target.Scheme) {
				address = value
				break
			}
		} else if address == "" {
			address = server
		}
	}
	if address == "" {
		return nil, nil
	}
	if !strings.Contains(address, "://") {
		address = "http://" + address
	}
	proxy, err := url.Parse(address)
	if err != nil || proxy.Hostname() == "" ||
		(proxy.Scheme != "http" && proxy.Scheme != "https" && proxy.Scheme != "socks5" && proxy.Scheme != "socks5h") {
		return nil, fmt.Errorf("invalid Windows system proxy address")
	}
	return proxy, nil
}
