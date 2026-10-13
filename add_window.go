package main

import (
	"context"
	"fmt"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/problem"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/desktop"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/session"
	torrentengine "github.com/ntnyq/rivu/internal/torrent"
)

type importItem struct {
	issue     *problem.Error
	tree      fileTreeState
	source    session.Source
	error     string
	submitted bool
}
type addWindow struct {
	*windowBase
	settings      config.Settings
	destination   string
	input         string
	preparedInput string
	items         []importItem
	start         bool
	adoptExisting bool
	tags          string
	free          int64
	spaceError    string
	generation    uint64
}

func (a *application) openAdd(sources []string) {
	if a.add != nil && !a.add.closed {
		a.add.win.Show()
		a.add.win.Focus()
		if len(sources) > 0 {
			a.add.input = strings.Join(sources, "\n")
			a.add.prepare()
		}
		return
	}
	parent := a.openWindow()
	d := &addWindow{windowBase: &windowBase{}, settings: a.store.Snapshot(), destination: a.store.Snapshot().Destination, start: true}
	d.app = a
	d.win = mygo.NewWindow(mygo.WindowOptions{Title: tr("添加下载"), Width: 550, Height: 660, MinWidth: 500, MinHeight: 500, Parent: parent.win, Modal: true, Content: ui.View(d.view)})
	d.desktop.Parent = d.win
	d.win.OnClosed(d.close)
	a.add = d
	d.input = strings.Join(sources, "\n")
	if len(sources) > 0 {
		d.prepare()
	}
}
func (d *addWindow) prepare() {
	if d.task.Running() {
		return
	}
	input := d.input
	destination := d.destination
	settings := d.settings
	d.generation++
	gen := d.generation
	var items []importItem
	var free int64
	var spaceError string
	d.run(tr("读取下载来源"), func(ctx context.Context) (string, error) {
		home, _ := os.UserHomeDir()
		if destination == filepath.Join(home, "Downloads", "Rivu") {
			if err := os.MkdirAll(destination, 0o755); err != nil {
				spaceError = err.Error()
			}
		}
		if err := files.ValidateDestination(destination); err != nil {
			spaceError = err.Error()
		} else if n, err := files.FreeSpace(destination); err != nil {
			spaceError = err.Error()
		} else {
			free = n
		}
		for _, line := range strings.Split(input, "\n") {
			if err := ctx.Err(); err != nil {
				return "", err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			s, err := torrentengine.Preview(ctx, line, settings)
			item := importItem{source: s}
			if err != nil {
				item.source.Name = filepath.Base(line)
				if strings.HasPrefix(line, "magnet:") {
					item.source.Name = tr("磁力链接")
				}
				item.error = err.Error()
				item.issue = problem.From(err)
			}
			items = append(items, item)
		}
		if len(items) == 0 {
			return "", fmt.Errorf("%s", tr("请添加种子文件或磁力链接"))
		}
		return tr("请确认保存位置和文件选择"), nil
	}, func() {
		if gen == d.generation {
			d.items = items
			d.preparedInput = input
			d.free = free
			d.spaceError = spaceError
		}
	})
}
func (d *addWindow) pickTorrents() {
	var paths []string
	d.run(tr("选择种子文件"), func(context.Context) (string, error) {
		var err error
		paths, err = mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: d.win, Title: tr("选择种子文件"), Multiple: true})
		if err != nil {
			return "", err
		}
		if len(paths) == 0 {
			return "", desktop.ErrCanceled
		}
		return tr("已选择种子文件，请读取并审阅"), nil
	}, func() { d.input = strings.Join(paths, "\n"); d.items = nil })
}
func (d *addWindow) paste() {
	var value string
	d.run(tr("粘贴磁力链接"), func(context.Context) (string, error) {
		var err error
		value, err = d.desktop.Paste()
		return tr("已粘贴，请读取并审阅"), err
	}, func() { d.input = value; d.items = nil })
}
func (d *addWindow) submit() {
	if d.input != d.preparedInput {
		d.report(tr("请重新读取已修改的下载来源"), fmt.Errorf("%s", tr("来源已更改")))
		return
	}
	items := make([]importItem, len(d.items))
	copy(items, d.items)
	for i := range items {
		items[i].source.Files = append([]session.File(nil), items[i].source.Files...)
	}
	destination, start, ratio := d.destination, d.start, d.settings.SeedRatio
	adopt, tags, seedLimit := d.adoptExisting, parseTags(d.tags), d.settings.SeedLimitSeconds
	var selected string
	var failed bool
	d.run(tr("添加下载"), func(ctx context.Context) (string, error) {
		for i := range items {
			item := &items[i]
			if item.error != "" || item.submitted {
				failed = failed || item.error != ""
				continue
			}
			id, err := d.app.session.Import(ctx, session.Import{AdoptExisting: adopt, Tags: tags, SeedLimitSeconds: seedLimit, Source: item.source, Destination: destination, Run: start, SeedRatio: ratio})
			if err != nil {
				item.error = err.Error()
				item.issue = problem.From(err)
				failed = true
			} else {
				item.submitted = true
				selected = id
			}
		}
		return tr("导入已处理；错误项目可修改后重新读取"), nil
	}, func() {
		d.items = items
		if selected != "" {
			w := d.app.openWindow()
			w.dispatch(func() { w.selectOnly(selected) })
		}
		if !failed {
			d.win.Close()
		}
	})
}

func (d *addWindow) pickFolder() {
	var folder string
	d.run(tr("选择下载目录"), func(context.Context) (string, error) {
		var err error
		folder, err = d.desktop.PickFolder()
		return tr("已选择目录"), err
	}, func() { d.destination = filepath.Clean(folder) })
}
