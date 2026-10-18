package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Exercise the actual MyGo request and JSON decoder in a writable app bundle.
// MyGo embeds its feed at build time and does not expose a test/client setter.
func TestMyGoUpdateNetwork(t *testing.T) {
	var calls atomic.Int32
	var respond atomic.Value
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		respond.Load().(func(http.ResponseWriter, *http.Request))(w, r)
	}))
	defer server.Close()
	name := "check"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	executable := filepath.Join(t.TempDir(), "Rivu.app", "Contents", "MacOS", name)
	feed := server.URL + "/update-windows-amd64.json"
	flags := "-X github.com/egoist/mygo.packageUpdateFeed=" + feed +
		" -X github.com/egoist/mygo.packageUpdateKey=Bq7SOwF0NSwMvjBZQ9ZTGMMQA94AL8FyvLDnbqmtx94=" +
		" -X github.com/egoist/mygo.production=1"
	build := exec.Command("go", "build", "-ldflags", flags, "-o", executable, "./testdata/check")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build updater harness: %v\n%s", err, output)
	}
	manifest := `{"version":"999.0.0","url":"https://example.test/update.tar.gz","size":1,"signature":"test"}`
	for _, test := range []struct {
		name      string
		first     func(http.ResponseWriter, *http.Request)
		always    bool
		wantCalls int32
		wantState string
		timeout   string
	}{
		{"EOF before headers", func(w http.ResponseWriter, _ *http.Request) {
			connection, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				_ = connection.Close()
			}
		}, false, 2, "available", "5s"},
		{"truncated body", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "1000")
			fmt.Fprint(w, `{"version":`)
		}, false, 2, "available", "5s"},
		{"incomplete JSON without content length", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"version":`)
		}, false, 2, "available", "5s"},
		{"persistent EOF stops after three attempts", func(w http.ResponseWriter, _ *http.Request) {
			connection, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				_ = connection.Close()
			}
		}, true, 3, "failed", "5s"},
		{"missing manifest is not retried", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}, true, 1, "failed", "5s"},
		{"invalid JSON is not retried", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, "not JSON")
		}, true, 1, "failed", "5s"},
		{"cancellation stops a stalled response", func(_ http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}, true, 1, "failed", "100ms"},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls.Store(0)
			respond.Store(func(w http.ResponseWriter, r *http.Request) {
				if calls.Load() == 1 || test.always {
					test.first(w, r)
					return
				}
				fmt.Fprint(w, manifest)
			})
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, filepath.Join(t.TempDir(), "settings.json"), test.timeout)
			command.Env = append(os.Environ(), "APPIMAGE=")
			output, err := command.Output()
			if err != nil && test.wantState != "failed" {
				t.Fatalf("transient update request must recover, got %v\n%s", err, output)
			}
			var snapshot Snapshot
			if err := json.Unmarshal(output, &snapshot); err != nil {
				t.Fatal(err)
			}
			if snapshot.State != test.wantState || calls.Load() != test.wantCalls {
				t.Fatalf("wrong recovery: calls=%d snapshot=%+v", calls.Load(), snapshot)
			}
			if test.wantState == "available" && (snapshot.Version != "999.0.0" || strings.Contains(snapshot.Error, "EOF") || snapshot.LastSuccess.IsZero()) {
				t.Fatal("stale failure after recovery")
			}
			if test.wantState == "failed" && (snapshot.Error == "" || !snapshot.LastSuccess.IsZero()) {
				t.Fatal("failed check recorded as successful")
			}
		})
	}
}
