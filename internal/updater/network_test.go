package updater

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestUpdateNetworkUsesSharedProxyAcrossRedirects(t *testing.T) {
	var calls atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Host == "update.invalid" {
			w.Header().Set("Location", "http://assets.invalid/manifest")
			w.WriteHeader(http.StatusFound)
			return
		}
		if r.URL.Host != "assets.invalid" {
			t.Errorf("unexpected update destination %s", r.URL.Host)
		}
		_, _ = io.WriteString(w, "manifest from proxy")
	}))
	defer proxy.Close()
	for _, name := range []string{"http_proxy", "NO_PROXY", "no_proxy"} {
		t.Setenv(name, "")
	}
	t.Setenv("HTTP_PROXY", proxy.URL)
	ctx, closeNetwork := updateNetworkContext(context.Background())
	defer closeNetwork()
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
		t.Fatalf("update proxy routing failed: body=%q error=%v calls=%d", body, err, calls.Load())
	}
}
