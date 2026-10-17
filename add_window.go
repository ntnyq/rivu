package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/desktop"
	"github.com/ntnyq/rivu/internal/files"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/internal/session"
	torrentengine "github.com/ntnyq/rivu/internal/torrent"
	"github.com/ntnyq/rivu/locales"
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
	d.win = mygo.NewWindow(mygo.WindowOptions{Title: locales.TransferAdd(), Width: 550, Height: 660, MinWidth: 500, MinHeight: 500, Parent: parent.win, Modal: true, Content: ui.View(d.view)})
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
	d.run(locales.ImportReadSources(), func(ctx context.Context) (string, error) {
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
					item.source.Name = locales.ImportMagnet()
				}
				item.error = err.Error()
				item.issue = problem.From(err)
			}
			items = append(items, item)
		}
		if len(items) == 0 {
			return "", fmt.Errorf("%s", locales.ImportSourceRequired())
		}
		return locales.ImportReviewHint(), nil
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
	d.run(locales.ImportChooseFiles(), func(context.Context) (string, error) {
		var err error
		paths, err = mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: d.win, Title: locales.ImportChooseFiles(), Multiple: true})
		if err != nil {
			return "", err
		}
		if len(paths) == 0 {
			return "", desktop.ErrCanceled
		}
		return locales.ImportFilesSelected(), nil
	}, func() { d.input = strings.Join(paths, "\n"); d.items = nil })
}
func (d *addWindow) paste() {
	var value string
	d.run(locales.ImportPasteMagnet(), func(context.Context) (string, error) {
		var err error
		value, err = d.desktop.Paste()
		return locales.ImportPasted(), err
	}, func() { d.input = value; d.items = nil })
}
func (d *addWindow) submit() {
	if d.input != d.preparedInput {
		d.report(locales.ImportReadAgain(), fmt.Errorf("%s", locales.ImportSourcesChanged()))
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
	d.run(locales.TransferAdd(), func(ctx context.Context) (string, error) {
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
		return locales.ImportProcessed(), nil
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
	d.run(locales.FilesChooseDownloadFolder(), func(context.Context) (string, error) {
		var err error
		folder, err = d.desktop.PickFolder()
		return locales.FilesFolderSelected(), err
	}, func() { d.destination = filepath.Clean(folder) })
}
