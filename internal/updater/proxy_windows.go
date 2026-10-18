package updater

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	winHTTP            = windows.NewLazySystemDLL("winhttp.dll")
	getUserProxyConfig = winHTTP.NewProc("WinHttpGetIEProxyConfigForCurrentUser")
	openProxySession   = winHTTP.NewProc("WinHttpOpen")
	closeProxySession  = winHTTP.NewProc("WinHttpCloseHandle")
	setProxyTimeouts   = winHTTP.NewProc("WinHttpSetTimeouts")
	getAutomaticProxy  = winHTTP.NewProc("WinHttpGetProxyForUrl")
	globalFree         = windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalFree")
)

// These layouts use Win32 BOOL/DWORD (32 bits), including on 64-bit Windows.
type userProxyConfig struct {
	autoDetect    uint32
	autoConfigURL *uint16
	proxy         *uint16
	bypass        *uint16
}

type autoProxyOptions struct {
	flags, detectFlags uint32
	configURL          *uint16
	reserved           uintptr
	reservedFlags      uint32
	autoLogon          uint32
}

type proxyInfo struct {
	accessType uint32
	proxy      *uint16
	bypass     *uint16
}

type windowsProxySettings struct {
	autoDetect            bool
	pacURL, proxy, bypass string
}

type windowsProxyResolver struct {
	mu              sync.Mutex
	autoDetectRetry time.Time
}

var currentWindowsProxy windowsProxyResolver

func systemProxy(ctx context.Context, target *url.URL) (*url.URL, error) {
	settings, err := readWindowsProxySettings()
	if err != nil {
		return nil, err
	}
	return currentWindowsProxy.resolve(ctx, target, settings, resolveWindowsProxy)
}

func (r *windowsProxyResolver) resolve(ctx context.Context, target *url.URL, settings windowsProxySettings, resolve func(*url.URL, windowsProxySettings) (*url.URL, error)) (*url.URL, error) {
	if settings.pacURL == "" {
		if settings.proxy != "" || !settings.autoDetect {
			return windowsProxy(target, settings.proxy, settings.bypass)
		}
		r.mu.Lock()
		deferDetection := time.Now().Before(r.autoDetectRetry)
		r.mu.Unlock()
		if deferDetection {
			return nil, nil
		}
	}
	proxy, err := lookupProxy(ctx, target, func(target *url.URL) (*url.URL, error) { return resolve(target, settings) })
	if err != nil && settings.pacURL == "" && !errors.Is(err, context.Canceled) {
		// Auto-detection is enabled by default on many home networks without
		// WPAD. Its absence/timeout must not disable a working direct route or
		// repeat discovery for every GitHub redirect. Explicit PAC stays strict.
		r.mu.Lock()
		r.autoDetectRetry = time.Now().Add(time.Minute)
		r.mu.Unlock()
		return windowsProxy(target, settings.proxy, settings.bypass)
	}
	return proxy, err
}

func freeProxyString(value *uint16) {
	if value != nil {
		_, _, _ = globalFree.Call(uintptr(unsafe.Pointer(value)))
	}
}

func readWindowsProxySettings() (windowsProxySettings, error) {
	var config userProxyConfig
	ok, _, err := getUserProxyConfig.Call(uintptr(unsafe.Pointer(&config)))
	defer freeProxyString(config.autoConfigURL)
	defer freeProxyString(config.proxy)
	defer freeProxyString(config.bypass)
	if ok == 0 {
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			return windowsProxySettings{}, nil
		}
		return windowsProxySettings{}, fmt.Errorf("read Windows system proxy: %w", err)
	}
	return windowsProxySettings{
		autoDetect: config.autoDetect != 0,
		pacURL:     windows.UTF16PtrToString(config.autoConfigURL),
		proxy:      windows.UTF16PtrToString(config.proxy),
		bypass:     windows.UTF16PtrToString(config.bypass),
	}, nil
}

func resolveWindowsProxy(target *url.URL, config windowsProxySettings) (*url.URL, error) {
	hasPAC := config.pacURL != ""
	// A configured manual proxy should not wait for WPAD discovery on every
	// request. An explicit PAC script, when present, still decides the route.
	if !hasPAC && (!config.autoDetect || config.proxy != "") {
		return windowsProxy(target, config.proxy, config.bypass)
	}

	// Resolve PAC/WPAD using Windows itself, never by evaluating scripts in Go.
	session, _, err := openProxySession.Call(0, 1, 0, 0, 0) // WINHTTP_ACCESS_TYPE_NO_PROXY
	if session == 0 {
		return nil, fmt.Errorf("open Windows proxy resolver: %w", err)
	}
	defer closeProxySession.Call(session)
	ok, _, err := setProxyTimeouts.Call(session, 5000, 5000, 5000, 5000)
	if ok == 0 {
		return nil, fmt.Errorf("set Windows proxy timeouts: %w", err)
	}
	pacURL, err := windows.UTF16PtrFromString(config.pacURL)
	if err != nil {
		return nil, err
	}
	options := autoProxyOptions{flags: 2, configURL: pacURL} // WINHTTP_AUTOPROXY_CONFIG_URL
	if !hasPAC {
		options.flags = 1       // WINHTTP_AUTOPROXY_AUTO_DETECT
		options.detectFlags = 3 // DHCP and DNS-A
		options.configURL = nil
	}
	address, err := windows.UTF16PtrFromString(target.String())
	if err != nil {
		return nil, err
	}
	var info proxyInfo
	ok, _, err = getAutomaticProxy.Call(session, uintptr(unsafe.Pointer(address)), uintptr(unsafe.Pointer(&options)), uintptr(unsafe.Pointer(&info)))
	defer freeProxyString(info.proxy)
	defer freeProxyString(info.bypass)
	if ok == 0 {
		if errors.Is(err, windows.Errno(12002)) { // ERROR_WINHTTP_TIMEOUT
			err = errors.Join(err, context.DeadlineExceeded)
		}
		return nil, fmt.Errorf("resolve Windows automatic proxy: %w", err)
	}
	if info.accessType == 1 { // PAC returned DIRECT: do not fall back to static proxy.
		return nil, nil
	}
	if info.accessType != 3 || info.proxy == nil { // WINHTTP_ACCESS_TYPE_NAMED_PROXY
		return nil, fmt.Errorf("Windows automatic proxy returned no route")
	}
	return windowsProxy(target, windows.UTF16PtrToString(info.proxy), windows.UTF16PtrToString(info.bypass))
}
