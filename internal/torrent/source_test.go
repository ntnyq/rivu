package torrent

import (
	"bytes"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"os"
	"path/filepath"
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
	_ = os.Symlink(target, filepath.Join(rootDir, "payload.txt"))
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
