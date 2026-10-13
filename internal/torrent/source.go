// Package torrent adapts anacrolix/torrent to Rivu's session contracts.
package torrent

import (
	"bytes"
	"fmt"
	"github.com/ntnyq/rivu/internal/problem"
	"io"
	"net/url"
	"os"
	"strings"

	bt "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/session"

	"github.com/ntnyq/rivu/internal/i18n"
)

const maxMetadata = 16 << 20

type Parser struct{}

func (Parser) Parse(raw string) (session.Source, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) > 65536 {
		return session.Source{}, problem.Message(problem.Metadata, "torrent.sourceTooLong")
	}
	if strings.HasPrefix(strings.ToLower(raw), "magnet:") {
		spec, err := bt.TorrentSpecFromMagnetUri(raw)
		if err != nil {
			return session.Source{}, problem.Message(problem.Metadata, "torrent.invalidMagnet")
		}
		s := session.Source{Raw: raw, Name: spec.DisplayName, Identity: identity(spec)}
		if s.Identity.V1 == "" && s.Identity.V2 == "" {
			return s, problem.Message(problem.Metadata, "torrent.missingInfohash")
		}
		if s.Name == "" {
			s.Name = i18n.T("torrent.waitingMetadata")
		}
		for _, tier := range spec.Trackers {
			s.Trackers = append(s.Trackers, tier...)
		}
		if len(s.Trackers) == 0 && len(spec.PeerAddrs) == 0 {
			return s, problem.Message(problem.Metadata, "torrent.bootstrapRequired")
		}
		return s, nil
	}
	f, err := os.Open(raw)
	if err != nil {
		return session.Source{}, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxMetadata+1))
	if err != nil {
		return session.Source{}, err
	}
	if len(data) > maxMetadata {
		return session.Source{}, problem.Message(problem.Metadata, "torrent.metadataTooLarge")
	}
	return (Parser{}).ParseBytes(raw, data)
}

// ParseBytes accepts the stable content already read by a folder watcher.
func (Parser) ParseBytes(raw string, data []byte) (session.Source, error) {
	if len(data) > maxMetadata {
		return session.Source{}, fmt.Errorf("metadata exceeds 16 MiB")
	}
	s, err := parseMeta(data)
	s.Raw = raw
	return s, err
}
func identity(spec *bt.TorrentSpec) session.Identity {
	var i session.Identity
	if spec.InfoHash != (metainfo.Hash{}) {
		i.V1 = spec.InfoHash.HexString()
	}
	if spec.InfoHashV2.Ok {
		i.V2 = fmt.Sprintf("%x", spec.InfoHashV2.Value)
	}
	return i
}
func parseMeta(data []byte) (session.Source, error) {
	mi, err := metainfo.Load(bytes.NewReader(data))
	if err != nil {
		return session.Source{}, problem.Message(problem.Metadata, "torrent.readFailed", err)
	}
	info, err := mi.UnmarshalInfo()
	if err != nil {
		return session.Source{}, err
	}
	if err := validateInfo(&info); err != nil {
		return session.Source{}, err
	}
	members := info.UpvertedFiles()
	spec, err := bt.TorrentSpecFromMetaInfoErr(mi)
	if err != nil {
		return session.Source{}, err
	}
	s := session.Source{PieceLength: info.PieceLength, Name: info.BestName(), Identity: identity(spec), Meta: data, Private: info.Private != nil && *info.Private}
	var names []string
	var total int64
	for _, f := range members {
		parts := append([]string{info.BestName()}, f.BestPath()...)
		for _, part := range parts {
			if strings.ContainsAny(part, "/\\") || part == "" || part == "." || part == ".." {
				return s, problem.Message(problem.Metadata, "torrent.unsafeFilename")
			}
		}
		name := strings.Join(parts, "/")
		if f.Length < 0 || f.Length > (1<<50)-total {
			return s, problem.Message(problem.Metadata, "torrent.invalidFileSize")
		}
		total += f.Length
		names = append(names, name)
		s.Files = append(s.Files, session.File{Path: name, Size: f.Length, Offset: f.TorrentOffset, Priority: 1})
	}
	if err := files.ValidateMembers(names); err != nil {
		return s, err
	}
	for _, tier := range spec.Trackers {
		s.Trackers = append(s.Trackers, tier...)
	}
	return s, nil
}
func RedactTracker(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return i18n.T("torrent.invalidTracker")
	}
	return u.Scheme + "://" + u.Hostname()
}

func validateInfo(info *metainfo.Info) error {
	if info.PieceLength <= 0 || info.PieceLength > 64<<20 || len(info.Pieces)%20 != 0 {
		return problem.Message(problem.Metadata, "torrent.unsupportedStructure")
	}
	if info.MetaVersion == 2 {
		valid := info.PieceLength >= 16384 && info.PieceLength&(info.PieceLength-1) == 0
		info.FileTree.Walk(nil, func(_ []string, ft *metainfo.FileTree) {
			if !ft.IsDir() && (ft.File.Length < 0 || ft.File.Length > 1<<50 ||
				(ft.File.Length > 0 || ft.File.PiecesRoot != "") && len(ft.File.PiecesRoot) != 32) {
				valid = false
			}
		})
		if !valid {
			return problem.Message(problem.Metadata, "torrent.invalidFileTree")
		}
	}
	members := info.UpvertedFiles()
	if len(members) == 0 || len(members) > 100000 {
		return problem.Message(problem.Metadata, "torrent.unsupportedFileCount")
	}
	var total int64
	for _, f := range members {
		if f.Length < 0 || f.Length > (1<<50)-total {
			return problem.Message(problem.Metadata, "torrent.invalidFileSize")
		}
		total += f.Length
	}
	if info.HasV1() {
		pieces := (total + info.PieceLength - 1) / info.PieceLength
		if info.HasV2() {
			// Hybrid storage uses the v2 geometry, but the engine indexes the
			// v1 hash array once for every v2 piece.
			pieces = int64(info.NumPieces())
		}
		if int64(len(info.Pieces)/20) != pieces {
			return problem.Message(problem.Metadata, "torrent.unsupportedStructure")
		}
	}
	return nil
}
