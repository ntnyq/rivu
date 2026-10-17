package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

func TestStructuredErrorUsesCurrentLanguage(t *testing.T) {
	useTestLanguage(t, "zh-Hans")
	issue := problem.Message(problem.Conflict, func(m locales.Messages) string { return m.FilesOwnershipChanged("payload") }, nil)
	before := problemText(issue, "")
	i18n.SetLocaleFunc(func() string { return "en" })
	after := problemText(issue, "")
	if before == after || !strings.Contains(after, "payload") || strings.Contains(after, "文件归属") {
		t.Fatal("error retained previous language", after)
	}
}

func BenchmarkMain1000Tasks(b *testing.B) {
	store, err := config.Open(filepath.Join(b.TempDir(), "settings.json"))
	if err != nil {
		b.Fatal(err)
	}
	w := newWindow(store)
	w.inspector = false
	for i := range 1000 {
		w.snapshot.Tasks = append(w.snapshot.Tasks, session.Task{Record: session.Record{ID: fmt.Sprint(i), Source: session.Source{Name: fmt.Sprintf("Task %d", i)}}})
	}
	tt := ui.NewTester(w.view, 1180, 760)
	tt.Frame()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		tt.Frame()
	}
}

func TestBandwidthUnitsAndInvalidInputs(t *testing.T) {
	for _, value := range []int64{0, 1, 512000, 2000000, 123456789} {
		if got, err := bandwidthValue(value).bytes(); err != nil || got != value {
			t.Fatalf("round trip %d: %d %v", value, got, err)
		}
	}
	for _, value := range []string{"NaN", "+Inf", "-1", "1e30", "abc"} {
		if _, err := parseScaled(value, 1000); err == nil {
			t.Fatal("invalid number", value)
		}
	}
	if got, err := clockMinute("23:59"); err != nil || got != 1439 {
		t.Fatal(got, err)
	}
	if _, err := clockMinute("24:00"); err == nil {
		t.Fatal("invalid time accepted")
	}
}

func TestFileTreeSearchAndDirectoryActions(t *testing.T) {
	files := []session.File{{Path: "a/one"}, {Path: "a/two"}, {Path: "b/three"}}
	rows := fileRows(files, "", map[string]bool{"a": true})
	if len(rows) != 3 || !rows[0].folder || len(rows[0].indices) != 2 {
		t.Fatalf("collapsed directory: %+v", rows)
	}
	rows = fileRows(files, "two", map[string]bool{"a": true})
	if len(rows) != 2 || rows[1].indices[0] != 1 {
		t.Fatalf("search did not expand matching path: %+v", rows)
	}
}

func BenchmarkFileTree10000Files(b *testing.B) {
	files := make([]session.File, 10000)
	for i := range files {
		files[i].Path = fmt.Sprintf("folder-%03d/file-%05d", i/100, i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = fileRows(files, "", nil)
	}
}

func BenchmarkCachedFileTree10000Files(b *testing.B) {
	files := make([]session.File, 10000)
	for i := range files {
		files[i].Path = fmt.Sprintf("folder-%03d/file-%05d", i/100, i)
	}
	state := fileTreeState{}
	state.visible(files)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = state.visible(files)
	}
}
func TestVirtualTaskListReachesLastTask(t *testing.T) {
	store, err := config.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	w := newWindow(store)
	w.inspector = false
	for i := range 1000 {
		w.snapshot.Tasks = append(w.snapshot.Tasks, session.Task{Record: session.Record{ID: fmt.Sprint(i), Source: session.Source{Name: fmt.Sprintf("Task %d", i)}}})
	}
	tt := ui.NewTester(w.view, 1180, 760)
	tt.Frame()
	if tt.HasText("Task 999") {
		t.Fatal("offscreen task eagerly built")
	}
	w.listState.ScrollTo(999, ui.End)
	tt.Frame()
	tt.Frame()
	if !tt.HasText("Task 999") {
		t.Fatal("cannot reach end of virtual list")
	}
}
