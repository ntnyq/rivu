// Package torrent adapts anacrolix/torrent to Rivu's session contracts.
package torrent

import (
	"bytes"
	"fmt"
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
		return session.Source{}, fmt.Errorf("%s", i18n.T("下载来源过长"))
	}
	if strings.HasPrefix(strings.ToLower(raw), "magnet:") {
		spec, err := bt.TorrentSpecFromMagnetUri(raw)
		if err != nil {
			return session.Source{}, fmt.Errorf("%s", i18n.T("无效的磁力链接"))
		}
		s := session.Source{Raw: raw, Name: spec.DisplayName, Identity: identity(spec)}
		if s.Identity.V1 == "" && s.Identity.V2 == "" {
			return s, fmt.Errorf("%s", i18n.T("磁力链接缺少有效 infohash"))
		}
		if s.Name == "" {
			s.Name = i18n.T("等待获取元数据")
		}
		for _, tier := range spec.Trackers {
			s.Trackers = append(s.Trackers, tier...)
		}
		if len(s.Trackers) == 0 && len(spec.PeerAddrs) == 0 {
			return s, fmt.Errorf("%s", i18n.T("未解析的磁力链接需要 Tracker 或明确的节点地址；获取元数据前禁用 DHT/PEX"))
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
		return session.Source{}, fmt.Errorf("%s", i18n.T("种子元数据超过 16 MB"))
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
		return session.Source{}, fmt.Errorf(i18n.T("无法读取种子元数据: %w"), err)
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
				return s, fmt.Errorf("%s", i18n.T("不安全的种子文件名"))
			}
		}
		name := strings.Join(parts, "/")
		if f.Length < 0 || f.Length > (1<<50)-total {
			return s, fmt.Errorf("%s", i18n.T("无效的文件大小"))
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
		return i18n.T("无效 Tracker")
	}
	return u.Scheme + "://" + u.Hostname()
}

func validateInfo(info *metainfo.Info) error {
	if info.PieceLength <= 0 || info.PieceLength > 64<<20 || len(info.Pieces)%20 != 0 {
		return fmt.Errorf("%s", i18n.T("不支持的种子结构"))
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
			return fmt.Errorf("%s", i18n.T("无效的 v2 文件树"))
		}
	}
	members := info.UpvertedFiles()
	if len(members) == 0 || len(members) > 100000 {
		return fmt.Errorf("%s", i18n.T("不支持的文件数量"))
	}
	var total int64
	for _, f := range members {
		if f.Length < 0 || f.Length > (1<<50)-total {
			return fmt.Errorf("%s", i18n.T("无效的文件大小"))
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
			return fmt.Errorf("%s", i18n.T("不支持的种子结构"))
		}
	}
	return nil
}
