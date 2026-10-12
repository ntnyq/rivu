package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/storage"

	"github.com/ntnyq/rivu/internal/i18n"
)

type diskState struct {
	Version  int      `json:"version"`
	Revision uint64   `json:"revision"`
	Tasks    []Record `json:"tasks"`
}
type sample struct {
	at       time.Time
	down, up int64
	task     Task
	since    time.Time
	retry    time.Time
}

// mu serializes mutations and engine calls on background workers. Snapshot uses
// atomic publication, so disk/engine latency never blocks native rendering.
type Manager struct {
	mu                  sync.Mutex
	path                string
	state               diskState
	engine              Engine
	settings            config.Settings
	active              map[string]bool
	holds               map[string]string
	samples             map[string]sample
	watchers            map[chan struct{}]bool
	snapshot            atomic.Pointer[Snapshot]
	loadErr, errorState error
	barrier, closed     bool
	cancel              context.CancelFunc
	done                chan struct{}
	lastSave            time.Time
	trash               func(string) error
}

func Open(dir string, e Engine, s config.Settings, trash func(string) error) (*Manager, error) {
	m := &Manager{path: filepath.Join(dir, "session.json"), engine: e, settings: s, state: diskState{Version: 1}, active: map[string]bool{}, holds: map[string]string{}, samples: map[string]sample{}, watchers: map[chan struct{}]bool{}, done: make(chan struct{}), trash: trash, lastSave: time.Now()}
	data, err := os.ReadFile(m.path)
	if err == nil {
		err = json.Unmarshal(data, &m.state)
		if err == nil {
			err = m.validate()
		}
	} else if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		m.state = diskState{Version: 1}
		m.loadErr = err
		m.errorState = fmt.Errorf(i18n.T("会话无法读取，原文件已保留: %w"), err)
	}
	for _, r := range m.state.Tasks {
		if !s.Restore && r.Run {
			m.holds[r.ID] = i18n.T("恢复已关闭，请手动继续")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.publish()
	go m.loop(ctx)
	return m, err
}
func (m *Manager) validate() error {
	if m.state.Version != 1 {
		return fmt.Errorf("unsupported session schema %d", m.state.Version)
	}
	seen := map[string]bool{}
	for _, r := range m.state.Tasks {
		id, err := hex.DecodeString(r.ID)
		if err != nil || len(id) != 16 || seen[r.ID] {
			return fmt.Errorf("invalid task ID")
		}
		seen[r.ID] = true
		if !filepath.IsAbs(r.Destination) || r.Source.Identity == (Identity{}) || math.IsNaN(r.SeedRatio) || r.SeedRatio < -1 {
			return fmt.Errorf("invalid task record")
		}
		if r.Metadata != "" && r.Metadata != filepath.Join(filepath.Dir(m.path), "metadata", r.ID+".torrent") {
			return fmt.Errorf("invalid metadata reference")
		}
		if err := files.ValidateMembers(names(r.Source.Files)); err != nil {
			return err
		}
		if len(r.Source.Files) > 0 {
			if err := validateSelection(r.Source.Files); err != nil {
				return err
			}
		}
		for _, p := range r.Trashed {
			if !slices.Contains(names(r.Source.Files), p) {
				return fmt.Errorf("invalid trash journal")
			}
		}
	}
	for i, r := range m.state.Tasks {
		for _, other := range m.state.Tasks[:i] {
			if r.Source.Identity.Matches(other.Source.Identity) {
				return fmt.Errorf("duplicate torrent identity")
			}
		}
	}
	return nil
}
func (m *Manager) loop(ctx context.Context) {
	defer close(m.done)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			if !m.closed && !m.barrier && m.loadErr == nil {
				m.tick()
			}
			m.mu.Unlock()
		}
	}
}
func (m *Manager) save(next diskState) error {
	if m.loadErr != nil {
		return m.loadErr
	}
	next.Revision = m.state.Revision + 1
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(m.path), 0o700); err != nil {
		return err
	}
	if err = storage.WriteFile(m.path, append(data, '\n'), 0o600); err != nil {
		var uncertain *storage.CommitError
		if errors.As(err, &uncertain) {
			m.loadErr = err
			m.errorState = err
			m.barrier = true
			for id := range m.active {
				_ = m.stop(id)
			}
		}
		return err
	}
	m.state = next
	m.lastSave = time.Now()
	m.errorState = nil
	return nil
}
func cloneState(s diskState) diskState {
	data, _ := json.Marshal(s)
	var out diskState
	_ = json.Unmarshal(data, &out)
	return out
}
func names(fs []File) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.Path)
	}
	return out
}
func validateSelection(fs []File) error {
	selected := false
	for _, f := range fs {
		if f.Size < 0 || f.Size > 1<<50 || f.Priority < 0 || f.Priority > 2 {
			return fmt.Errorf("invalid priority")
		}
		selected = selected || f.Priority > 0
	}
	if !selected {
		return fmt.Errorf("%s", i18n.T("至少选择一个文件"))
	}
	return nil
}
func (m *Manager) checkDestination(s Source, dir string, except string) error {
	if err := files.CheckConflicts(dir, names(s.Files)); err != nil {
		return err
	}
	for _, r := range m.state.Tasks {
		if r.ID == except {
			continue
		}
		for _, a := range r.Source.Files {
			for _, b := range s.Files {
				left := strings.ToLower(filepath.Join(r.Destination, filepath.FromSlash(a.Path)))
				right := strings.ToLower(filepath.Join(dir, filepath.FromSlash(b.Path)))
				if left == right || strings.HasPrefix(left, right+string(filepath.Separator)) || strings.HasPrefix(right, left+string(filepath.Separator)) {
					return fmt.Errorf("%s", i18n.T("目标路径已由另一个任务使用"))
				}
			}
		}
	}
	free, err := files.FreeSpace(dir)
	if err != nil {
		return err
	}
	size := s.requiredSpace()
	if size > free {
		return fmt.Errorf(i18n.T("可用空间不足（需要 %d 字节）"), size)
	}
	return nil
}
func (m *Manager) metadata(r *Record, s Source) error {
	dir := filepath.Join(filepath.Dir(m.path), "metadata")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	r.Metadata = filepath.Join(dir, r.ID+".torrent")
	if err := storage.WriteFile(r.Metadata, s.Meta, 0o600); err != nil {
		return err
	}
	s.Meta = nil
	r.Source = s
	return nil
}
func (m *Manager) Import(ctx context.Context, in Import) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.publish()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if m.closed || m.barrier {
		return "", fmt.Errorf("%s", i18n.T("正在保存会话，请稍后重试"))
	}
	if m.loadErr != nil {
		return "", m.loadErr
	}
	for _, r := range m.state.Tasks {
		if r.Source.Identity.Matches(in.Source.Identity) {
			return r.ID, nil
		}
	}
	if in.Source.Identity == (Identity{}) {
		return "", fmt.Errorf("%s", i18n.T("无效的下载来源"))
	}
	if len(in.Source.Files) > 0 {
		if err := validateSelection(in.Source.Files); err != nil {
			return "", err
		}
	}
	absolute, err := filepath.Abs(in.Destination)
	if err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	in.Destination = canonical
	if err := m.checkDestination(in.Source, in.Destination, ""); err != nil {
		return "", err
	}
	if math.IsNaN(in.SeedRatio) || math.IsInf(in.SeedRatio, 0) || in.SeedRatio < -1 {
		return "", fmt.Errorf("invalid seed ratio")
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	r := Record{ID: hex.EncodeToString(id), Source: in.Source, Destination: in.Destination, Run: in.Run, SeedRatio: in.SeedRatio, Added: time.Now(), Order: time.Now().UnixNano()}
	if len(in.Source.Meta) > 0 {
		if err := m.metadata(&r, in.Source); err != nil {
			return "", err
		}
	}
	next := cloneState(m.state)
	next.Tasks = append(next.Tasks, r)
	if err := m.save(next); err != nil {
		return "", err
	}
	m.tick()
	return r.ID, nil
}
func (m *Manager) Execute(ctx context.Context, cmd Command) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.publish()
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.closed || m.barrier {
		return fmt.Errorf("%s", i18n.T("会话正在保存"))
	}
	if m.loadErr != nil {
		return m.loadErr
	}
	switch cmd.Kind {
	case "pause", "resume", "pauseAll", "resumeBatch", "verify", "priority", "remove", "seed", "first", "notified":
	default:
		return fmt.Errorf("unknown command %q", cmd.Kind)
	}
	next := cloneState(m.state)
	stopIDs := []string{}
	selectedIDs := []string{}
	for i := range next.Tasks {
		r := &next.Tasks[i]
		selected := slices.Contains(cmd.IDs, r.ID)
		if cmd.Kind == "pauseAll" {
			selected = r.Run
			r.Batch = false // Only the newest Pause All owns batch membership.
		}
		if cmd.Kind == "resumeBatch" {
			selected = r.Batch
		}
		if !selected {
			continue
		}
		selectedIDs = append(selectedIDs, r.ID)
		switch cmd.Kind {
		case "pause", "pauseAll":
			r.Run = false
			r.Reason = "manual"
			r.Batch = cmd.Kind == "pauseAll"
			stopIDs = append(stopIDs, r.ID)
		case "resume", "resumeBatch":
			if r.Completed && r.Reason == "ratio" && cmd.Kind == "resume" {
				r.SeedOverride = true
			}
			r.Run = true
			r.Reason = ""
			r.Batch = false
		case "verify":
			r.Completed = false
			stopIDs = append(stopIDs, r.ID)
			r.Run = true
			r.Reason = "verify"
		case "first":
			r.Order = 0
			r.Run = true
		case "priority":
			if cmd.FileIndex < 0 || cmd.FileIndex >= len(r.Source.Files) {
				return fmt.Errorf("invalid file")
			}
			r.Source.Files[cmd.FileIndex].Priority = cmd.Priority
			if err := validateSelection(r.Source.Files); err != nil {
				return err
			}
			r.Completed = false
			stopIDs = append(stopIDs, r.ID)
		case "seed":
			if math.IsNaN(cmd.Ratio) || math.IsInf(cmd.Ratio, 0) || cmd.Ratio < -1 {
				return fmt.Errorf("invalid ratio")
			}
			r.SeedRatio = cmd.Ratio
			r.SeedOverride = false
		case "remove":
			r.Removing = true
			r.Trash = cmd.Trash
			r.Run = false
			r.Batch = false
			stopIDs = append(stopIDs, r.ID)
		case "notified":
			r.Notified = r.Epoch
		}
	}
	// A pause must remain safe even when its intent cannot be saved.
	for _, id := range stopIDs {
		if err := m.stop(id); err != nil {
			return err
		}
	}
	if err := m.save(next); err != nil {
		m.errorState = fmt.Errorf(i18n.T("会话保存失败，请重试: %w"), err)
		for _, id := range stopIDs {
			m.holds[id] = i18n.T("暂停尚未保存，请重试")
		}
		return err
	}
	for _, id := range selectedIDs {
		delete(m.holds, id)
	}
	m.tick()
	return m.errorState
}
func (m *Manager) stop(id string) error {
	err := m.engine.Stop(id)
	delete(m.active, id)
	s := m.samples[id]
	s.down = 0
	s.up = 0
	s.at = time.Time{}
	s.task.DownRate = 0
	s.task.UpRate = 0
	m.samples[id] = s
	return err
}
func (m *Manager) tick() {
	if m.closed || m.barrier || m.loadErr != nil {
		return
	}
	changed := false
	now := time.Now()
	for i := range m.state.Tasks {
		r := &m.state.Tasks[i]
		if r.Removing {
			continue
		}
		if !m.active[r.ID] {
			continue
		}
		o := m.engine.Observe(r.ID)
		prev := m.samples[r.ID]
		if o.Error != "" {
			m.holds[r.ID] = o.Error
			_ = m.stop(r.ID)
			continue
		}
		if o.Metadata != nil {
			s := *o.Metadata
			if !s.Identity.Matches(r.Source.Identity) {
				m.holds[r.ID] = i18n.T("元数据标识不匹配")
				_ = m.stop(r.ID)
				continue
			}
			for _, other := range m.state.Tasks {
				if other.ID != r.ID && s.Identity.Matches(other.Source.Identity) {
					m.holds[r.ID] = i18n.T("元数据与现有任务重复，请移除此任务")
				}
			}
			if m.holds[r.ID] == "" {
				if err := m.checkDestination(s, r.Destination, r.ID); err != nil {
					m.holds[r.ID] = err.Error()
				} else if err = m.metadata(r, s); err != nil {
					m.holds[r.ID] = err.Error()
				} else {
					changed = true
				}
			}
			_ = m.stop(r.ID)
			continue
		}
		task := Task{Record: *r, Phase: o.Phase, Files: o.Files, Peers: o.Peers, Trackers: o.Trackers}
		for j, f := range r.Source.Files {
			task.TotalSize += f.Size
			if f.Priority > 0 {
				task.Size += f.Size
				if j < len(o.Files) {
					task.Done += o.Files[j].Done
				}
			}
		}
		r.Uploaded += max(0, o.Uploaded-prev.up)
		task.Uploaded = r.Uploaded
		if !prev.at.IsZero() {
			seconds := now.Sub(prev.at).Seconds()
			task.DownRate = float64(max(0, o.Downloaded-prev.down)) / seconds
			task.UpRate = float64(max(0, o.Uploaded-prev.up)) / seconds
		}
		if r.Completed && len(o.Files) > 0 && task.Done < task.Size && task.Phase != "verifying" {
			r.Completed = false
			changed = true
		}
		if len(o.Files) > 0 && task.Done == task.Size && task.Phase != "verifying" {
			task.Phase = "seeding"
			if !r.Completed {
				r.Completed = true
				r.Epoch++
				changed = true
			}
			ratio := float64(r.Uploaded) / float64(max(1, task.TotalSize))
			if !r.SeedOverride && r.SeedRatio >= 0 && (r.SeedRatio == 0 || task.TotalSize > 0 && ratio >= r.SeedRatio) {
				r.Run = false
				r.Reason = "ratio"
				changed = true
				_ = m.stop(r.ID)
			}
		}
		since := prev.since
		if since.IsZero() || o.Downloaded > prev.down || task.Phase == "verifying" {
			since = now
		}
		m.samples[r.ID] = sample{at: now, down: o.Downloaded, up: o.Uploaded, task: task, since: since, retry: prev.retry}
		if !m.active[r.ID] {
			_ = m.stop(r.ID)
		}
		// Yield stalled slots; retry after a minute without changing durable intent.
		stalledPhase := task.Phase == "waiting" || task.Phase == "downloading" || task.Phase == "metadata"
		if stalledPhase && now.Sub(since) > 90*time.Second {
			_ = m.stop(r.ID)
			s := m.samples[r.ID]
			s.retry = now.Add(time.Minute)
			s.since = now
			m.samples[r.ID] = s
		}
	}
	if changed || now.Sub(m.lastSave) >= 15*time.Second {
		if err := m.save(m.state); err != nil {
			m.errorState = err
			m.barrier = true
			for id := range m.active {
				_ = m.stop(id)
			}
			m.publish()
			return
		}
	}
	// Recovery of committed removals is safe and idempotent.
	for i := len(m.state.Tasks) - 1; i >= 0; i-- {
		if m.state.Tasks[i].Removing {
			if err := m.finishRemoval(i); err != nil {
				m.holds[m.state.Tasks[i].ID] = err.Error()
				if m.barrier || m.loadErr != nil {
					break
				}
			}
		}
	}
	if m.barrier || m.loadErr != nil {
		m.publish()
		return
	}
	m.schedule(now)
	m.publish()
}
func (m *Manager) finishRemoval(i int) error {
	r := m.state.Tasks[i]
	if err := m.stop(r.ID); err != nil {
		return err
	}
	if r.Trash {
		if m.trash == nil {
			return fmt.Errorf("%s", i18n.T("系统废纸篓不可用"))
		}
		if err := files.ValidateDestination(r.Destination); err != nil {
			return err
		}
		for _, f := range r.Source.Files {
			if slices.Contains(r.Trashed, f.Path) {
				continue
			}
			full := filepath.Join(r.Destination, filepath.FromSlash(f.Path))
			shared := false
			for _, other := range m.state.Tasks {
				if other.ID == r.ID {
					continue
				}
				for _, of := range other.Source.Files {
					if strings.EqualFold(full, filepath.Join(other.Destination, filepath.FromSlash(of.Path))) {
						shared = true
					}
				}
			}
			if shared {
				return fmt.Errorf("%s", i18n.T("文件由其他任务共享，无法移到废纸篓"))
			}
			real, err := filepath.EvalSymlinks(full)
			if err == nil {
				if real != full {
					return fmt.Errorf("%s", i18n.T("文件已被符号链接替换"))
				}
				st, er := os.Lstat(full)
				if er != nil {
					return er
				}
				if !st.Mode().IsRegular() {
					return fmt.Errorf("%s", i18n.T("只允许移除任务的普通文件"))
				}
				owners, er := files.ReadOwnership(r.Metadata + ".files.json")
				if er != nil {
					return er
				}
				identity, er := files.FileIdentity(st)
				if er != nil {
					return er
				}
				if owners[f.Path] != identity {
					return fmt.Errorf(i18n.T("文件归属已改变，已保留: %s"), f.Path)
				}
				if err = m.trash(full); err != nil {
					return err
				}
			} else if !os.IsNotExist(err) {
				return err
			}
			r.Trashed = append(r.Trashed, f.Path)
			next := cloneState(m.state)
			next.Tasks[i] = r
			if err := m.save(next); err != nil {
				return err
			}
		}
	}
	next := cloneState(m.state)
	next.Tasks = slices.Delete(next.Tasks, i, i+1)
	if err := m.save(next); err != nil {
		return err
	}
	delete(m.samples, r.ID)
	delete(m.holds, r.ID)
	return nil
}
func (m *Manager) publish() {
	snap := Snapshot{Revision: m.state.Revision, SlowMode: m.settings.SlowMode, Quiesced: m.barrier}
	if m.errorState != nil {
		snap.Error = m.errorState.Error()
	}
	for _, r := range m.state.Tasks {
		t := m.samples[r.ID].task
		t.Record = r
		if t.Phase == "verifying" {
			t.Completed = false
		}
		if len(t.Files) == 0 {
			t.Files = slices.Clone(r.Source.Files)
		}
		t.Size = 0
		t.TotalSize = 0
		for _, f := range r.Source.Files {
			t.TotalSize += f.Size
			if f.Priority > 0 {
				t.Size += f.Size
			}
		}
		if !m.active[r.ID] {
			t.DownRate = 0
			t.UpRate = 0
			t.Peers = 0
			t.Phase = "queued"
			if !r.Run {
				t.Phase = "paused"
				if t.Completed {
					t.Phase = "completed"
					if len(m.samples[r.ID].task.Files) == 0 {
						t.Done = t.Size
					}
				}
			}
		}
		if r.Removing {
			t.Phase = "removing"
		}
		if hold := m.holds[r.ID]; hold != "" {
			t.Error = hold
			t.Phase = "error"
		}
		snap.DownRate += t.DownRate
		snap.UpRate += t.UpRate
		snap.Tasks = append(snap.Tasks, t)
	}
	sort.SliceStable(snap.Tasks, func(i, j int) bool { return snap.Tasks[i].Added.After(snap.Tasks[j].Added) })
	// JSON cloning excludes Source.Meta and transient File.Done, so copy those
	// explicitly instead of sharing mutable file slices across consumers.
	for i := range snap.Tasks {
		t := &snap.Tasks[i]
		t.Source.Files = slices.Clone(t.Source.Files)
		t.Source.Trackers = slices.Clone(t.Source.Trackers)
		t.Files = slices.Clone(t.Files)
		t.Trackers = slices.Clone(t.Trackers)
		t.Trashed = slices.Clone(t.Trashed)
	}
	m.snapshot.Store(&snap)
	for ch := range m.watchers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
func (m *Manager) Snapshot() Snapshot {
	p := m.snapshot.Load()
	if p == nil {
		return Snapshot{}
	}
	s := *p
	s.Tasks = slices.Clone(s.Tasks)
	for i := range s.Tasks {
		t := &s.Tasks[i]
		t.Files = slices.Clone(t.Files)
		t.Source.Files = slices.Clone(t.Source.Files)
		t.Source.Trackers = slices.Clone(t.Source.Trackers)
		t.Trackers = slices.Clone(t.Trackers)
		t.Trashed = slices.Clone(t.Trashed)
	}
	return s
}
func (m *Manager) Watch() (<-chan struct{}, func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch := make(chan struct{}, 1)
	m.watchers[ch] = true
	var once sync.Once
	return ch, func() { once.Do(func() { m.mu.Lock(); delete(m.watchers, ch); close(ch); m.mu.Unlock() }) }
}
func (m *Manager) Configure(s config.Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := s.Validate(); err != nil {
		return err
	}
	m.settings = s
	d, u := s.Limits()
	m.engine.Limits(d, u)
	m.publish()
	return nil
}
func (m *Manager) Quiesce() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.barrier = true
	for id := range m.active {
		if err := m.stop(id); err != nil {
			return err
		}
	}
	err := m.save(m.state)
	if err != nil {
		m.errorState = err
	}
	m.publish()
	return err
}
func (m *Manager) Continue() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed && m.loadErr == nil {
		m.barrier = false
		m.tick()
	}
}
func (m *Manager) Close() error {
	if err := m.Quiesce(); err != nil {
		return err
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	err := m.engine.Close()
	if err == nil {
		m.closed = true
		m.cancel()
	}
	m.mu.Unlock()
	if err == nil {
		<-m.done
	}
	return err
}
func (m *Manager) Idle() bool {
	for _, t := range m.Snapshot().Tasks {
		if t.Run && !t.Removing {
			return false
		}
	}
	return true
}

// Reload is an explicit recovery action after a user repairs a damaged session.
// It never replaces a corrupt file with an empty session.
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer m.publish()
	if m.closed {
		return fmt.Errorf("session closed")
	}
	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	var next diskState
	if err = json.Unmarshal(data, &next); err != nil {
		return err
	}
	old := m.state
	m.state = next
	if err = m.validate(); err != nil {
		m.state = old
		return err
	}
	for id := range m.active {
		if err = m.stop(id); err != nil {
			m.state = old
			return err
		}
	}
	m.loadErr = nil
	m.errorState = nil
	m.barrier = false
	m.holds = map[string]string{}
	m.samples = map[string]sample{}
	if !m.settings.Restore {
		for _, r := range m.state.Tasks {
			if r.Run {
				m.holds[r.ID] = i18n.T("恢复已关闭，请手动继续")
			}
		}
	}
	m.tick()
	return nil
}
