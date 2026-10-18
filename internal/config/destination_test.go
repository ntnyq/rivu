//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo"
)

func TestSystemDownloadDestination(t *testing.T) {
	previous, err := mygo.App.Path(mygo.PathDownloads)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { mygo.App.SetPath(mygo.PathDownloads, previous) })
	downloads := filepath.Join(t.TempDir(), "下载目录")
	mygo.App.SetPath(mygo.PathDownloads, downloads)
	want := filepath.Join(downloads, "Rivu")
	if got := Default().Destination; got != want || !IsDefaultDestination(got) {
		t.Fatalf("system downloads ignored: %q", got)
	}
	if _, err := os.Stat(downloads); !os.IsNotExist(err) {
		t.Fatalf("default discovery created a directory: %v", err)
	}
	home, _ := os.UserHomeDir()
	if !IsDefaultDestination(filepath.Join(home, "Downloads", "Rivu")) {
		t.Fatal("legacy default no longer recognized")
	}
	custom := filepath.Join(t.TempDir(), "custom")
	if IsDefaultDestination(custom) {
		t.Fatal("custom location can be created automatically")
	}
	path := filepath.Join(t.TempDir(), "settings.json")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Update(func(s *Settings) { s.Destination = custom }); err != nil {
		t.Fatal(err)
	}
	mygo.App.SetPath(mygo.PathDownloads, filepath.Join(t.TempDir(), "moved"))
	restored, err := Open(path)
	if err != nil || restored.Snapshot().Destination != custom {
		t.Fatalf("saved choice changed after OS directory move: %v", err)
	}
	mygo.App.SetPath(mygo.PathDownloads, "relative")
	if DefaultDestination() != filepath.Join(home, "Downloads", "Rivu") {
		t.Fatal("invalid system directory did not fall back")
	}
}
