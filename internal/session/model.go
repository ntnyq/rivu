// Package session owns durable user intent and download scheduling.
package session

import "time"

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
	ID           string    `json:"id"`
	Source       Source    `json:"source"`
	Metadata     string    `json:"metadata,omitempty"`
	Destination  string    `json:"destination"`
	Run          bool      `json:"run"`
	Reason       string    `json:"reason,omitempty"`
	Batch        bool      `json:"batch"`
	SeedRatio    float64   `json:"seedRatio"`
	SeedOverride bool      `json:"seedOverride"`
	Uploaded     int64     `json:"uploaded"`
	Added        time.Time `json:"added"`
	Order        int64     `json:"order"`
	Completed    bool      `json:"completed"`
	Epoch        uint64    `json:"epoch"`
	Notified     uint64    `json:"notified"`
	Removing     bool      `json:"removing"`
	Trash        bool      `json:"trash"`
	Trashed      []string  `json:"trashed,omitempty"`
}
type Observation struct {
	Metadata             *Source
	Files                []File
	Phase                string
	Downloaded, Uploaded int64
	Peers                int
	Error                string
	Trackers             []string
}
type Task struct {
	Record
	Phase                 string
	Done, Size, TotalSize int64
	DownRate, UpRate      float64
	Peers                 int
	Error                 string
	Files                 []File
	Trackers              []string
}
type Snapshot struct {
	Revision         uint64
	Tasks            []Task
	Error            string
	DownRate, UpRate float64
	SlowMode         bool
	Quiesced         bool
}
type Command struct {
	Kind                string
	IDs                 []string
	FileIndex, Priority int
	Ratio               float64
	Trash               bool
}
type Import struct {
	Source      Source
	Destination string
	Run         bool
	SeedRatio   float64
}
type Engine interface {
	Attach(Record) error
	Start(string, []File) error
	Stop(string) error
	Observe(string) Observation
	Limits(int64, int64)
	Close() error
}
