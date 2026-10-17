package main

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

type fileTreeState struct {
	search       string
	collapsed    map[string]bool
	page         int
	paths        []string
	cachedSearch string
	rows         []fileRow
}

func (s *fileTreeState) visible(files []session.File) []fileRow {
	changed := s.rows == nil || s.cachedSearch != s.search || len(s.paths) != len(files)
	if !changed {
		for i, file := range files {
			if s.paths[i] != file.Path {
				changed = true
				break
			}
		}
	}
	if changed {
		s.paths = make([]string, len(files))
		for i, file := range files {
			s.paths[i] = file.Path
		}
		s.rows = fileRows(files, s.search, s.collapsed)
		s.cachedSearch = s.search
		s.page = 0
	}
	return s.rows
}

type fileRow struct {
	name    string
	indices []int
	folder  bool
	depth   int
}

func fileRows(files []session.File, search string, collapsed map[string]bool) []fileRow {
	groups := map[string][]int{}
	query := strings.ToLower(strings.TrimSpace(search))
	for i, f := range files {
		if !strings.Contains(strings.ToLower(f.Path), query) {
			continue
		}
		groups[f.Path] = []int{i}
		for dir := path.Dir(f.Path); dir != "."; dir = path.Dir(dir) {
			groups[dir] = append(groups[dir], i)
		}
	}
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	rows := make([]fileRow, 0, len(names))
	for _, name := range names {
		hidden := false
		if query == "" {
			for dir := path.Dir(name); dir != "."; dir = path.Dir(dir) {
				hidden = hidden || collapsed[dir]
			}
		}
		if hidden {
			continue
		}
		indices := groups[name]
		folder := len(indices) != 1 || files[indices[0]].Path != name
		rows = append(rows, fileRow{name, indices, folder, strings.Count(name, "/")})
	}
	return rows
}

// Render only a page of file rows so large torrents do not build thousands of
// native elements. Directory actions apply atomically to all matching descendants.
func fileTree(c *ui.Context, files []session.File, state *fileTreeState, enabled bool, change func([]int, int)) {
	if state.collapsed == nil {
		state.collapsed = map[string]bool{}
	}
	ui.TextInput(c, &state.search).Label(locales.FilesSearch()).Placeholder(locales.FilesSearch())
	rows := state.visible(files)
	const pageSize = 60
	pages := max(1, (len(rows)+pageSize-1)/pageSize)
	state.page = min(state.page, pages-1)
	if pages > 1 {
		ui.Row(c).Gap(6).Children(func() {
			if ui.Button(c, "←").Label(locales.FilesPreviousPage()).Disabled(state.page == 0).Clicked() {
				state.page--
			}
			ui.Text(c, fmt.Sprintf("%d / %d", state.page+1, pages)).FontSize(10)
			if ui.Button(c, "→").Label(locales.FilesNextPage()).Disabled(state.page+1 == pages).Clicked() {
				state.page++
			}
		})
	}
	for _, row := range rows[state.page*pageSize : min(len(rows), (state.page+1)*pageSize)] {
		ui.Column(c).Key(row.name).Padding(5, 0, 5, float32(min(row.depth, 4)*8)).Gap(5).Children(func() {
			ui.Row(c).Gap(5).Children(func() {
				if row.folder {
					label := "−"
					if state.collapsed[row.name] {
						label = "+"
					}
					if ui.Button(c, label).Label(row.name).Clicked() {
						state.collapsed[row.name] = !state.collapsed[row.name]
						state.rows = nil
					}
				}
				ui.Text(c, path.Base(row.name)).FontSize(11).MaxLines(2).Grow(1).MinWidth(0)
			})
			var size int64
			priority := files[row.indices[0]].Priority
			for _, i := range row.indices {
				size += files[i].Size
				if files[i].Priority != priority {
					priority = -1
				}
			}
			labels := []string{locales.FilesPrioritySkip(), locales.FilesPriorityNormal(), locales.FilesPriorityHigh()}
			label := locales.FilesPriorityMixed()
			if priority >= 0 {
				label = labels[priority]
			}
			ui.Row(c).Gap(6).Children(func() {
				ui.Text(c, bytes(size)).FontSize(10).Grow(1)
				menuControl(c, row.name+" "+locales.FilesPriorityLabel(), label).Disabled(!enabled).Menu(func(m *ui.Menu) {
					for i, label := range labels {
						if m.Item(label).Checked(priority == i).Chosen() {
							change(row.indices, i)
						}
					}
				})
			})
		})
	}
}

func parseTags(raw string) []string {
	var tags []string
	seen := map[string]bool{}
	for _, tag := range strings.Split(raw, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" && !seen[tag] {
			tags = append(tags, tag)
			seen[tag] = true
		}
	}
	return tags
}
