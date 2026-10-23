package torrent

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	bt "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/netproxy"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
	"golang.org/x/time/rate"
)

type handle struct {
	torrent        *bt.Torrent
	storage        *rootedStorage
	raw            string
	started        bool
	metadata       bool
	diagnosticHash string
}
type Engine struct {
	settings         config.Settings
	diagnostics      trackerDiagnostics
	client, metadata *bt.Client
	handles          map[string]*handle
	down, up         *rate.Limiter
}

func New(s config.Settings) *Engine {
	e := &Engine{settings: s, handles: map[string]*handle{}, down: rate.NewLimiter(rate.Inf, 1<<20), up: rate.NewLimiter(rate.Inf, 1<<20)}
	d, u := s.LimitsAt(time.Now())
	e.Limits(d, u)
	return e
}
func (e *Engine) getClient(metadata bool) (*bt.Client, error) {
	ptr := &e.client
	if metadata {
		ptr = &e.metadata
	}
	if *ptr != nil {
		return *ptr, nil
	}
	cfg := bt.NewDefaultClientConfig()
	cfg.ListenPort = e.settings.ListenPort
	cfg.Seed = true
	cfg.NoDefaultPortForwarding = true
	cfg.DisableWebtorrent = true
	cfg.DisableWebseeds = true
	cfg.HTTPProxy = netproxy.ForRequest
	cfg.DialForPeerConns = false // Install route-aware wrappers after construction.
	cfg.NoDHT = !e.settings.DHT || metadata
	cfg.DisablePEX = !e.settings.PEX || metadata
	if metadata {
		cfg.ListenPort = 0
	}
	cfg.DownloadRateLimiter = e.down
	cfg.UploadRateLimiter = e.up
	cfg.Slogger = slog.New(trackerDiagnosticHandler{store: &e.diagnostics})
	c, err := bt.NewClient(cfg)
	if err == nil {
		configurePeerRoutes(c)
		*ptr = c
	}
	return c, err
}
func (e *Engine) Attach(r session.Record) error {
	if _, ok := e.handles[r.ID]; ok {
		return nil
	}
	if err := files.ValidateDestination(r.Destination); err != nil {
		return err
	}
	var spec *bt.TorrentSpec
	var err error
	metadata := r.Metadata == ""
	if metadata {
		spec, err = bt.TorrentSpecFromMagnetUri(r.Source.Raw)
	} else {
		data, er := os.ReadFile(r.Metadata)
		if er != nil {
			return er
		}
		parsed, er := parseMeta(data)
		if er != nil {
			return er
		}
		if !parsed.Identity.Matches(r.Source.Identity) {
			return problem.Message(problem.Engine, locales.Messages.TorrentIdentityMismatch, nil)
		}
		mi, er := metainfo.Load(bytes.NewReader(data))
		if er != nil {
			return er
		}
		spec, err = bt.TorrentSpecFromMetaInfoErr(mi)
	}
	if err != nil {
		return err
	}
	if !metadata && strings.HasPrefix(strings.ToLower(r.Source.Raw), "magnet:") {
		original, err := bt.TorrentSpecFromMagnetUri(r.Source.Raw)
		if err != nil {
			return err
		}
		if !identity(original).Matches(r.Source.Identity) {
			return problem.Message(problem.Engine, locales.Messages.TorrentIdentityMismatch, nil)
		}
		// Metainfo retains trackers, but not the magnet's explicit peers.
		spec.PeerAddrs = original.PeerAddrs
	}
	c, err := e.getClient(metadata)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(r.Destination)
	if err != nil {
		return err
	}
	st := &rootedStorage{root: root, manifest: r.Metadata + ".files.json", adopted: r.Adopted}
	spec.Storage = st
	spec.DisallowDataDownload = true
	spec.DisallowDataUpload = true
	spec.Sources = nil
	spec.Webseeds = nil // Never fetch arbitrary metadata URLs from a magnet.
	hash := spec.InfoHash
	if hash == (metainfo.Hash{}) && spec.InfoHashV2.Ok {
		hash = *spec.InfoHashV2.Value.ToShort()
	}
	diagnosticHash := hash.HexString()
	e.diagnostics.retain(diagnosticHash)
	t, fresh, err := addSpec(c, spec)
	if err != nil {
		e.diagnostics.release(diagnosticHash)
		root.Close()
		return err
	}
	if !fresh {
		e.diagnostics.release(diagnosticHash)
		root.Close()
		return problem.Message(problem.Engine, locales.Messages.TorrentAlreadyRunning, nil)
	}
	e.handles[r.ID] = &handle{torrent: t, storage: st, raw: r.Source.Raw, metadata: metadata, diagnosticHash: diagnosticHash}
	return nil
}
func (e *Engine) Start(id string, policy []session.File) error {
	h := e.handles[id]
	if h == nil {
		return problem.Message(problem.Engine, locales.Messages.TorrentNotAdded, nil)
	}
	if h.started {
		return nil
	}
	if h.torrent.Info() == nil {
		return nil
	}
	fs := h.torrent.Files()
	if len(fs) != len(policy) {
		return problem.Message(problem.Engine, locales.Messages.TorrentSelectionMismatch, nil)
	}
	if err := h.storage.admit(policy); err != nil {
		return err
	}
	for i, f := range fs {
		priority := bt.PiecePriorityNone
		// Empty files are already created by admission. Upstream can otherwise
		// prioritize the preceding shared piece for a zero-length file.
		if f.Length() > 0 && policy[i].Priority == 1 {
			priority = bt.PiecePriorityNormal
		}
		if f.Length() > 0 && policy[i].Priority == 2 {
			priority = bt.PiecePriorityHigh
		}
		f.SetPriority(priority)
	}
	h.torrent.AllowDataDownload()
	h.torrent.AllowDataUpload()
	h.started = true
	return nil
}
func (e *Engine) Stop(id string) error {
	h := e.handles[id]
	if h == nil {
		return nil
	}
	h.storage.setAllowed(false)
	h.torrent.DisallowDataUpload()
	h.torrent.DisallowDataDownload()
	h.torrent.Drop()
	e.diagnostics.release(h.diagnosticHash)
	delete(e.handles, id)
	return h.storage.root.Close()
}
func (e *Engine) Observe(id string) session.Observation {
	h := e.handles[id]
	if h == nil {
		return session.Observation{}
	}
	t := h.torrent
	o := session.Observation{Phase: session.PhaseMetadata, Diagnostics: e.diagnostics.snapshot(h.diagnosticHash)}
	if err := h.storage.failure(); err != nil {
		o.Failure = err
		o.Error = err.Error()
		return o
	}
	stats := t.Stats()
	o.Uploaded = stats.BytesWrittenData.Int64()
	o.Downloaded = stats.BytesReadUsefulData.Int64()
	o.Peers = stats.ActivePeers
	if t.Info() == nil {
		return o
	}
	mi := t.Metainfo()
	var b bytes.Buffer
	if err := mi.Write(&b); err != nil {
		o.Error = err.Error()
		return o
	}
	if h.metadata {
		s, err := parseMeta(b.Bytes())
		if err != nil {
			o.Error = err.Error()
			return o
		}
		s.Raw = h.raw
		o.Metadata = &s
		return o
	}
	o.Phase = "waiting"
	if o.Peers > 0 {
		o.Phase = "downloading"
	}
	for _, p := range t.PieceStateRuns() {
		if p.Checking || p.Hashing || p.QueuedForHash {
			o.Phase = "verifying"
			break
		}
	}
	info := t.Info()
	members := info.UpvertedFiles()
	for i, f := range t.Files() {
		mf := members[i]
		var done int64
		for piece := mf.BeginPieceIndex(info.PieceLength); piece < mf.EndPieceIndex(info.PieceLength); piece++ {
			if t.PieceState(piece).Complete {
				begin := max(mf.TorrentOffset, int64(piece)*info.PieceLength)
				end := min(mf.TorrentOffset+mf.Length, int64(piece+1)*info.PieceLength)
				done += max(0, end-begin)
			}
		}
		o.Files = append(o.Files, session.File{Path: f.Path(), Size: f.Length(), Done: done})
	}
	for _, tier := range mi.UpvertedAnnounceList() {
		for _, tr := range tier {
			o.Trackers = append(o.Trackers, RedactTracker(tr))
		}
	}
	return o
}
func (e *Engine) Limits(d, u int64) {
	set := func(l *rate.Limiter, v int64) {
		if v == 0 {
			l.SetLimit(rate.Inf)
		} else {
			l.SetLimit(rate.Limit(v))
		}
		l.SetBurst(int(max(65536, min(v, 1<<20))))
	}
	set(e.down, d)
	set(e.up, u)
}
func (e *Engine) Close() error {
	var errs []error
	for id := range e.handles {
		errs = append(errs, e.Stop(id))
	}
	if e.client != nil {
		errs = append(errs, e.client.Close()...)
		e.client = nil
	}
	if e.metadata != nil {
		errs = append(errs, e.metadata.Close()...)
		e.metadata = nil
	}
	return errors.Join(errs...)
}

// v1.61.0 requires a nonzero short hash even for v2-only input. Its public
// short-hash path upgrades the identity when metadata is authenticated. Keep
// the full v2 identity in the session and compare it before payload admission.
func addSpec(c *bt.Client, spec *bt.TorrentSpec) (*bt.Torrent, bool, error) {
	copy := *spec
	if copy.InfoHash == (metainfo.Hash{}) && copy.InfoHashV2.Ok {
		copy.InfoHash = *copy.InfoHashV2.Value.ToShort()
		copy.InfoHashV2.SetNone()
	}
	if copy.InfoHash == (metainfo.Hash{}) {
		return nil, false, problem.Message(problem.Engine, locales.Messages.TorrentInvalidIdentity, nil)
	}
	return c.AddTorrentSpec(&copy)
}
