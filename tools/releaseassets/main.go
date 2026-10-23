// Command releaseassets verifies and collects the signed MyGo release files.
package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type releaseConfig struct {
	Version string `json:"version"`
	Updates struct {
		GitHub    string `json:"github"`
		TagPrefix string `json:"tagPrefix"`
		PublicKey string `json:"publicKey"`
		Deltas    *int   `json:"deltas"`
	} `json:"updates"`
}

type updateManifest struct {
	Version   string            `json:"version"`
	URL       string            `json:"url"`
	Size      int64             `json:"size"`
	Signature string            `json:"signature"`
	Deltas    []json.RawMessage `json:"deltas"`
}

func main() {
	target := flag.String("target", "", "release target: darwin-universal, windows-amd64 or linux-amd64")
	source := flag.String("source", "", "directory containing MyGo build output (default build/<target>)")
	out := flag.String("out", "build/release", "directory for verified release assets")
	flag.Parse()
	if *source == "" {
		*source = filepath.Join("build", *target)
	}
	err := run(*target, *source, *out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(target, source, out string) error {
	configJSON, err := os.ReadFile("mygo.json")
	if err != nil {
		return err
	}
	var config releaseConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return err
	}
	return collectAssets(config, target, source, out)
}

func collectAssets(config releaseConfig, target, source, out string) error {
	var installers []string
	switch target {
	case "darwin-universal":
		installers = []string{"Rivu " + config.Version + ".dmg"}
	case "windows-amd64":
		installers = []string{"Rivu Setup " + config.Version + " amd64.exe"}
	case "linux-amd64":
		installers = []string{"rivu_" + config.Version + "_amd64.deb", "install.sh"}
	default:
		return fmt.Errorf("unsupported release target %q", target)
	}
	publicKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(config.Updates.PublicKey))
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid updates.publicKey")
	}
	if config.Version == "" || config.Updates.GitHub != "ntnyq/rivu" || config.Updates.TagPrefix != "v" {
		return fmt.Errorf("release needs a version and updates.github=ntnyq/rivu with tagPrefix=v")
	}
	if config.Updates.Deltas == nil || *config.Updates.Deltas != 0 {
		return fmt.Errorf("release collection requires updates.deltas=0")
	}
	manifestName := "update-" + target + ".json"
	manifestJSON, err := os.ReadFile(filepath.Join(source, manifestName))
	if err != nil {
		return err
	}
	var manifest updateManifest
	if err := json.Unmarshal(manifestJSON, &manifest); err != nil {
		return fmt.Errorf("%s: %w", manifestName, err)
	}
	archiveName := "rivu-" + config.Version + "-" + target + ".tar.gz"
	expectedURL := "https://github.com/ntnyq/rivu/releases/download/v" + config.Version + "/" + archiveName
	if manifest.Version != config.Version || manifest.URL != expectedURL || len(manifest.Deltas) != 0 {
		return fmt.Errorf("%s: wrong version, archive URL or unsupported deltas", manifestName)
	}
	archive, err := os.Open(filepath.Join(source, archiveName))
	if err != nil {
		return err
	}
	digest := sha256.New()
	size, hashErr := io.Copy(digest, archive)
	closeErr := archive.Close()
	if hashErr != nil {
		return hashErr
	}
	if closeErr != nil {
		return closeErr
	}
	if size <= 0 || size != manifest.Size {
		return fmt.Errorf("%s: archive size %d does not match manifest size %d", archiveName, size, manifest.Size)
	}
	// MyGo v0.6.3 signs the archive's SHA-256 digest with Ed25519.
	signature, err := base64.StdEncoding.DecodeString(manifest.Signature)
	if err != nil || !ed25519.Verify(publicKey, digest.Sum(nil), signature) {
		return fmt.Errorf("%s: invalid update signature", archiveName)
	}
	names := append(installers, archiveName, manifestName)
	for _, name := range names {
		info, err := os.Stat(filepath.Join(source, name))
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("%s: expected a nonempty release file", name)
		}
	}
	sourcePath, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	outPath, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	if sourcePath == outPath {
		return nil
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, name := range names {
		if err := copyAsset(filepath.Join(source, name), filepath.Join(out, name)); err != nil {
			return err
		}
	}
	return nil
}

func copyAsset(source, out string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(out)
	if err != nil {
		return err
	}
	_, err = io.Copy(output, input)
	closeErr := output.Close()
	if err != nil {
		return err
	}
	return closeErr
}
