package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectSignedRelease(t *testing.T) {
	for _, target := range []string{"darwin-universal", "windows-amd64", "linux-amd64"} {
		t.Run(target, func(t *testing.T) {
			config, source := releaseFixture(t, target)
			out := filepath.Join(t.TempDir(), "release")
			if err := collectAssets(config, target, source, out); err != nil {
				t.Fatal(err)
			}
			// The combined upload is checked again in place before creating a tag.
			if err := collectAssets(config, target, out, out); err != nil {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(out)
			if err != nil {
				t.Fatal(err)
			}
			expected := 3
			if target == "linux-amd64" {
				expected = 4
			}
			if len(entries) != expected {
				t.Fatalf("collected %d files, want %d", len(entries), expected)
			}
		})
	}
}

func TestRejectInvalidRelease(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*testing.T, *releaseConfig, string)
		want   string
	}{
		{"tampered archive", func(t *testing.T, _ *releaseConfig, source string) {
			writeFixture(t, filepath.Join(source, "rivu-0.0.2-darwin-universal.tar.gz"), []byte("changed"))
		}, "invalid update signature"},
		{"missing manifest", func(t *testing.T, _ *releaseConfig, source string) {
			if err := os.Remove(filepath.Join(source, "update-darwin-universal.json")); err != nil {
				t.Fatal(err)
			}
		}, "update-darwin-universal.json"},
		{"missing installer", func(t *testing.T, _ *releaseConfig, source string) {
			if err := os.Remove(filepath.Join(source, "Rivu 0.0.2.dmg")); err != nil {
				t.Fatal(err)
			}
		}, "Rivu 0.0.2.dmg"},
		{"wrong public key", func(t *testing.T, config *releaseConfig, _ string) {
			publicKey, _, err := ed25519.GenerateKey(nil)
			if err != nil {
				t.Fatal(err)
			}
			config.Updates.PublicKey = base64.StdEncoding.EncodeToString(publicKey)
		}, "invalid update signature"},
		{"wrong version", func(t *testing.T, _ *releaseConfig, source string) {
			changeManifest(t, source, func(m *updateManifest) { m.Version = "0.0.1" })
		}, "wrong version"},
		{"wrong architecture", func(t *testing.T, _ *releaseConfig, source string) {
			changeManifest(t, source, func(m *updateManifest) { m.URL = strings.ReplaceAll(m.URL, "darwin-universal", "windows-amd64") })
		}, "archive URL"},
		{"wrong size", func(t *testing.T, _ *releaseConfig, source string) {
			changeManifest(t, source, func(m *updateManifest) { m.Size++ })
		}, "archive size"},
		{"unsigned", func(t *testing.T, _ *releaseConfig, source string) {
			changeManifest(t, source, func(m *updateManifest) { m.Signature = "" })
		}, "invalid update signature"},
		{"unexpected deltas", func(t *testing.T, _ *releaseConfig, source string) {
			changeManifest(t, source, func(m *updateManifest) { m.Deltas = []json.RawMessage{json.RawMessage(`{}`)} })
		}, "unsupported deltas"},
		{"invalid key", func(_ *testing.T, config *releaseConfig, _ string) {
			config.Updates.PublicKey = "invalid"
		}, "invalid updates.publicKey"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config, source := releaseFixture(t, "darwin-universal")
			test.change(t, &config, source)
			out := filepath.Join(t.TempDir(), "release")
			if err := collectAssets(config, "darwin-universal", source, out); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("want error containing %q, got %v", test.want, err)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatal("invalid assets were collected")
			}
		})
	}
}

func releaseFixture(t *testing.T, target string) (releaseConfig, string) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	config := releaseConfig{Version: "0.0.2"}
	config.Updates.GitHub = "ntnyq/rivu"
	config.Updates.TagPrefix = "v"
	config.Updates.PublicKey = base64.StdEncoding.EncodeToString(publicKey)
	config.Updates.Deltas = new(int)
	source := t.TempDir()
	archive := []byte("archive")
	digest := sha256.Sum256(archive)
	manifest := updateManifest{
		Version:   config.Version,
		URL:       "https://github.com/ntnyq/rivu/releases/download/v0.0.2/rivu-0.0.2-" + target + ".tar.gz",
		Size:      int64(len(archive)),
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, digest[:])),
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(source, "rivu-0.0.2-"+target+".tar.gz"), archive)
	writeFixture(t, filepath.Join(source, "update-"+target+".json"), manifestJSON)
	for _, name := range []string{"Rivu 0.0.2.dmg", "Rivu Setup 0.0.2 amd64.exe", "rivu_0.0.2_amd64.deb", "install.sh", "unrelated-file"} {
		writeFixture(t, filepath.Join(source, name), []byte("installer"))
	}
	return config, source
}

func changeManifest(t *testing.T, source string, change func(*updateManifest)) {
	t.Helper()
	path := filepath.Join(source, "update-darwin-universal.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var manifest updateManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	change(&manifest)
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, path, data)
}

func writeFixture(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
