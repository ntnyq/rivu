// Package session owns durable user intent and download scheduling.
package session

import "time"
import "github.com/ntnyq/rivu/internal/problem"

type Identity struct {
	V1 string `json:"v1,omitempty"`
	V2 string `json:"v2,omitempty"`
}

func (i Identity) Matches(j Identity) bool {
	return i.V1 != "" && i.V1 == j.V1 || i.V2 != "" && i.V2 == j.V2
}

type File struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Offset   int64  `json:"offset,omitempty"`
	Priority int    `json:"priority"` // 0 skip, 1 normal, 2 high
	Done     int64  `json:"-"`
}
type Source struct {
	Raw         string   `json:"raw"`
	Identity    Identity `json:"identity"`
	Name        string   `json:"name"`
	Files       []File   `json:"files,omitempty"`
	PieceLength int64    `json:"pieceLength,omitempty"`
	Private     bool     `json:"private"`
	Trackers    []string `json:"trackers,omitempty"`
	Meta        []byte   `json:"-"`
}
type Record struct {
	Adopted          map[string]string `json:"adopted,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	SeedSeconds      int64             `json:"seedSeconds,omitempty"`
	SeedLimitSeconds int64             `json:"seedLimitSeconds,omitempty"`
	ID               string            `json:"id"`
	Source           Source            `json:"source"`
	Metadata         string            `json:"metadata,omitempty"`
	Destination      string            `json:"destination"`
	Run              bool              `json:"run"`
	Reason           PauseReason       `json:"reason,omitempty"`
	Batch            bool              `json:"batch"`
	SeedRatio        float64           `json:"seedRatio"`
	SeedOverride     bool              `json:"seedOverride"`
	Uploaded         int64             `json:"uploaded"`
	Added            time.Time         `json:"added"`
	Order            int64             `json:"order"`
	Completed        bool              `json:"completed"`
	Epoch            uint64            `json:"epoch"`
	Notified         uint64            `json:"notified"`
	Removing         bool              `json:"removing"`
	Trash            bool              `json:"trash"`
	Trashed          []string          `json:"trashed,omitempty"`
}
type TrackerStatus struct {
	Host, State, Detail string
	Next                time.Time
}

type Observation struct {
	Diagnostics          []TrackerStatus
	Failure              error
	Metadata             *Source
	Files                []File
	Phase                Phase
	Downloaded, Uploaded int64
	Peers                int
	Error                string
	Trackers             []string
}
type Task struct {
	Diagnostics []TrackerStatus
	Issue       *problem.Error
	Record
	Phase                 Phase
	Done, Size, TotalSize int64
	DownRate, UpRate      float64
	Peers                 int
	Error                 string
	Files                 []File
	Trackers              []string
}
type Snapshot struct {
	Issue            *problem.Error
	Revision         uint64
	Tasks            []Task
	Error            string
	DownRate, UpRate float64
	SlowMode         bool
	Quiesced         bool
}
type Command struct {
	FileIndices         []int
	Destination         string
	Tags                []string
	SeedLimitSeconds    int64
	Kind                CommandKind
	IDs                 []string
	FileIndex, Priority int
	Ratio               float64
	Trash               bool
}
type Import struct {
	AdoptExisting    bool
	Tags             []string
	SeedLimitSeconds int64
	Source           Source
	Destination      string
	Run              bool
	SeedRatio        float64
}
type Engine interface {
	Attach(Record) error
	Start(string, []File) error
	Stop(string) error
	Observe(string) Observation
	Limits(int64, int64)
	Close() error
}
