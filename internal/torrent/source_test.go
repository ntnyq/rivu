package torrent

import (
	"bytes"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func fixture(t *testing.T) (string, []byte, []byte) {
	t.Helper()
	dir := t.TempDir()
	payload := bytes.Repeat([]byte("Rivu local integration fixture.\n"), 10000)
	name := filepath.Join(dir, "payload.txt")
	if err := os.WriteFile(name, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	info := metainfo.Info{PieceLength: 16384}
	if err := info.BuildFromFilePath(name); err != nil {
		t.Fatal(err)
	}
	encoded, err := bencode.Marshal(&info)
	if err != nil {
		t.Fatal(err)
	}
	mi := metainfo.MetaInfo{InfoBytes: encoded}
	var b bytes.Buffer
	if err := mi.Write(&b); err != nil {
		t.Fatal(err)
	}
	return dir, b.Bytes(), payload
}
func TestParseTorrentAndIdentity(t *testing.T) {
	_, meta, _ := fixture(t)
	s, err := parseMeta(meta)
	if err != nil {
		t.Fatal(err)
	}
	if s.Identity.V1 == "" || len(s.Files) != 1 || s.Files[0].Path != "payload.txt" {
		t.Fatal(s)
	}
	magnet := "magnet:?xt=urn:btih:" + s.Identity.V1 + "&tr=http%3A%2F%2F127.0.0.1%3A8000%2Fannounce"
	m, err := (Parser{}).Parse(magnet)
	if err != nil || !s.Identity.Matches(m.Identity) {
		t.Fatal(err, m)
	}
}
func TestMalformedAndTraversal(t *testing.T) {
	for _, data := range [][]byte{[]byte("garbage"), []byte("de")} {
		if _, err := parseMeta(data); err == nil {
			t.Fatal("accepted invalid metainfo")
		}
	}
	info := metainfo.Info{Name: "..", PieceLength: 16384, Length: 1, Pieces: make([]byte, 20)}
	encoded, _ := bencode.Marshal(&info)
	mi := metainfo.MetaInfo{InfoBytes: encoded}
	var b bytes.Buffer
	_ = mi.Write(&b)
	if _, err := parseMeta(b.Bytes()); err == nil {
		t.Fatal("accepted traversal")
	}
}
func TestTrackerRedaction(t *testing.T) {
	if got := RedactTracker("https://user:pass@tracker.example/passkey/announce?token=secret"); got != "https://tracker.example" {
		t.Fatal(got)
	}
}
func TestRootedStorageRejectsSymlinkEscape(t *testing.T) {
	dir, meta, _ := fixture(t)
	s, err := parseMeta(meta)
	if err != nil {
		t.Fatal(err)
	}
	_ = s
	mi, _ := metainfo.Load(bytes.NewReader(meta))
	info, _ := mi.UnmarshalInfo()
	outside := t.TempDir()
	target := filepath.Join(outside, "outside")
	_ = os.WriteFile(target, []byte("safe"), 0o600)
	rootDir := t.TempDir()
	if err := os.Symlink(target, filepath.Join(rootDir, "payload.txt")); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink privilege unavailable: %v", err)
		}
		t.Fatal(err)
	}
	root, _ := os.OpenRoot(rootDir)
	defer root.Close()
	st := &rootedStorage{root: root, allowed: true}
	impl, err := st.OpenTorrent(t.Context(), &info, mi.HashInfoBytes())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = impl.Piece(info.Piece(0)).WriteAt([]byte("evil"), 0); err == nil {
		t.Fatal("escaped root")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "safe" {
		t.Fatal("external file changed")
	}
	_ = dir
}

func TestV2MetadataRejectsInvalidRoot(t *testing.T) {
	info := metainfo.Info{Name: "payload", PieceLength: 16384, MetaVersion: 2, FileTree: metainfo.FileTree{File: metainfo.FileTreeFile{Length: 10, PiecesRoot: "short"}}}
	encoded, _ := bencode.Marshal(&info)
	mi := metainfo.MetaInfo{InfoBytes: encoded}
	var b bytes.Buffer
	_ = mi.Write(&b)
	if _, err := parseMeta(b.Bytes()); err == nil {
		t.Fatal("accepted truncated pieces root")
	}
}

func TestMetadataGeometryValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		length int64
		root   string
		hybrid bool
		hashes int
		valid  bool
	}{
		{"empty", 0, "", false, 0, true},
		{"empty short root", 0, "short", false, 0, false},
		{"empty long root", 0, string(make([]byte, 33)), false, 0, false},
		{"missing root", 16384, "", false, 0, false},
		{"v2", 16384, string(make([]byte, 32)), false, 0, true},
		{"hybrid missing hashes", 16384, string(make([]byte, 32)), true, 0, false},
		{"hybrid too few hashes", 32768, string(make([]byte, 32)), true, 1, false},
		{"hybrid too many hashes", 16384, string(make([]byte, 32)), true, 2, false},
		{"hybrid", 16384, string(make([]byte, 32)), true, 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			info := metainfo.Info{Name: "payload", MetaVersion: 2, PieceLength: 16384,
				FileTree: metainfo.FileTree{File: metainfo.FileTreeFile{Length: tc.length, PiecesRoot: tc.root}}}
			if tc.hybrid {
				info.Length = tc.length
				info.Pieces = make([]byte, tc.hashes*20)
			}
			encoded, err := bencode.Marshal(&info)
			if err != nil {
				t.Fatal(err)
			}
			mi := metainfo.MetaInfo{InfoBytes: encoded}
			var data bytes.Buffer
			if err = mi.Write(&data); err != nil {
				t.Fatal(err)
			}
			// Both imported files and metadata arriving through an engine handle
			// must reject these structures without relying on panic recovery.
			if _, err = parseMeta(data.Bytes()); (err == nil) != tc.valid {
				t.Fatalf("parseMeta: valid=%t, error=%v", tc.valid, err)
			}
			st := &rootedStorage{}
			if _, err = st.OpenTorrent(t.Context(), &info, mi.HashInfoBytes()); (err == nil) != tc.valid {
				t.Fatalf("OpenTorrent: valid=%t, error=%v", tc.valid, err)
			}
		})
	}
}
