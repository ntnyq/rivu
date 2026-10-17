package app

import (
	"io"
	"path/filepath"
	"testing"
)

func TestLaunchPathsAndMagnets(t *testing.T) {
	cwd := t.TempDir()
	magnet := "magnet:?xt=urn:btih:0123456789012345678901234567890123456789"
	o, err := parseLaunch([]string{"--config", "prefs.json", "--data-dir", "isolated", "one.torrent", magnet, "two.torrent"}, cwd, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if len(o.sources) != 3 || o.sources[0] != filepath.Join(cwd, "one.torrent") || o.sources[1] != magnet || o.dataDir != filepath.Join(cwd, "isolated") {
		t.Fatal(o)
	}
	for _, args := range [][]string{{"--unknown"}, {"--config"}} {
		if _, err := parseLaunch(args, cwd, io.Discard); err == nil {
			t.Fatal("accepted invalid flag")
		}
	}
}
