package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
	"path/filepath"
	"strings"
)

func fileKind(t session.Task) string {
	name := strings.ToLower(t.Source.Name)
	if strings.Contains(name, "movie") || strings.HasSuffix(name, ".mp4") || strings.HasSuffix(name, ".mkv") {
		return "movie"
	}
	if strings.Contains(name, "archive") || strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tar") {
		return "archive"
	}
	return "file"
}

func fileSymbol(c *ui.Context, t session.Task, large bool) {
	kind := fileKind(t)
	color := c.Theme().Accent
	if kind == "movie" {
		color = ui.Hex("#9263b7")
	}
	if kind == "archive" {
		color = c.Theme().TextMuted
	}
	width, height := float32(32), float32(39)
	if large {
		width, height = 40, 48
	}
	ui.Row(c).Size(width, height).Shrink(0).Radius(5).Border(1, color.Alpha(.17)).Background(color.Alpha(.06)).Justify(ui.Center).Children(func() { glyph(c, kind, 20, color) })
}

func (w *mainWindow) taskRow(c *ui.Context, t session.Task, visible []session.Task) {
	p := designTheme(c)
	row := ui.Row(c).Key(t.ID).Padding(17, 17, 15, 17).Gap(12).AlignItems(ui.Start).Height(102).Shrink(0).BorderWidth(0, 0, 1, 0).BorderColor(p.line).Label(t.Source.Name).Focusable()
	if w.selected[t.ID] {
		row.Background(p.soft)
	} else if row.Hovered() {
		row.Background(p.alt)
	}
	row.Children(func() {
		ui.Box(c).Margin(3, 0, 0, 0).Shrink(0).Children(func() { fileSymbol(c, t, false) })
		ui.Column(c).Grow(1).MinWidth(0).Gap(6).Children(func() {
			ui.Row(c).Gap(10).Children(func() {
				ui.Text(c, t.Source.Name).FontSize(12).FontWeight(600).Grow(1).MinWidth(0).SingleLine()
				ui.Text(c, taskPercent(t)).FontSize(11).TextColor(p.muted)
			})
			ui.Row(c).Gap(10).Children(func() {
				text := bytes(t.Done) + " / " + bytes(t.Size)
				if t.Metadata == "" {
					text = tr("transfer.waitingFileList")
				}
				ui.Text(c, text).FontSize(10).TextColor(p.muted)
				ui.Spacer(c)
				color := p.muted
				if t.Completed || t.Error != "" {
					color = taskColor(c, t.Phase)
				}
				ui.Text(c, phaseLabel(t.Phase)).FontSize(10).TextColor(color)
			})
			ui.Box(c).Margin(4, 0, 1, 0).Children(func() {
				progressTrack(c, progress(t), taskColor(c, t.Phase), p.track, 3)
			})
			ui.Row(c).Gap(10).Children(func() {
				left, right := "↓ "+rate(t.DownRate)+"   ↑ "+rate(t.UpRate), eta(t)
				color := p.blue
				switch t.Phase {
				case "paused":
					left, right, color = tr("transfer.paused"), tr("transfer.waitingResume"), p.muted
				case "completed":
					left, right, color = tr("transfer.filesReady"), tr("seeding.goalReached"), p.muted
				case "seeding":
					left = "↑ " + rate(t.UpRate)
					right = fmt.Sprintf(tr("seeding.ratioProgress"), float64(t.Uploaded)/float64(max(1, t.TotalSize)), t.SeedRatio)
					color = p.muted
				case "error":
					left, right, color = taskError(t), tr("transfer.retryHint"), p.muted
				case "metadata":
					left, right, color = tr("transfer.fetchingMetadata"), "", p.muted
				}
				ui.Text(c, left).FontSize(10).TextColor(color).SingleLine().Grow(1).MinWidth(0)
				ui.Text(c, right).FontSize(10).TextColor(p.muted)
			})
		})
		icon := "pause"
		if !t.Run {
			icon = "play"
		}
		if t.Phase == "completed" {
			icon = "folder"
		}
		if t.Error != "" {
			icon = "alert"
		}
		if compactIconButton(c, icon, fmt.Sprintf(tr("transfer.actions"), t.ID), 23, 15, p.muted) && w.app != nil {
			if t.Phase == "completed" {
				w.reveal(t.Destination)
			} else {
				kind := session.CommandResume
				if t.Run && t.Error == "" {
					kind = session.CommandPause
				}
				w.app.execute(session.Command{Kind: kind, IDs: []string{t.ID}})
			}
		}
	})
	if row.Clicked() {
		mods := row.ClickModifiers()
		if mods&ui.Shift != 0 && w.anchor != "" {
			start, end := -1, -1
			for i, v := range visible {
				if v.ID == w.anchor {
					start = i
				}
				if v.ID == t.ID {
					end = i
				}
			}
			if start >= 0 && end >= 0 {
				for _, v := range visible[min(start, end) : max(start, end)+1] {
					w.selected[v.ID] = true
				}
			}
		} else if mods&ui.Cmd != 0 {
			w.selected[t.ID] = !w.selected[t.ID]
			w.anchor = t.ID
		} else {
			w.selectOnly(t.ID)
		}
	}
	if row.DoubleClicked() {
		w.selectOnly(t.ID)
		w.inspector = true
	}
	if row.Shortcut(0, ui.KeySpace) {
		w.toggleSelected()
	}
	row.ContextMenu(func(m *ui.Menu) {
		kind, label := session.CommandResume, tr("transfer.resume")
		if t.Run {
			kind, label = session.CommandPause, tr("transfer.pause")
		}
		if m.Item(label).Chosen() {
			w.app.execute(session.Command{Kind: kind, IDs: []string{t.ID}})
		}
		if m.Item(tr("files.showInFinder")).Chosen() {
			w.reveal(t.Destination)
		}
		if m.Item(tr("transfer.copyMagnet")).Chosen() {
			w.copy(magnet(t))
		}
		if m.Item(tr("transfer.verify")).Chosen() {
			w.app.execute(session.Command{Kind: session.CommandVerify, IDs: []string{t.ID}})
		}
		if m.Item(tr("transfer.startFirst")).Chosen() {
			w.app.execute(session.Command{Kind: session.CommandFirst, IDs: []string{t.ID}})
		}
		m.Separator()
		if m.Item(tr("transfer.removeMenu")).Chosen() {
			w.selectOnly(t.ID)
			w.remove()
		}
	})
}

func compactPath(path string) string {
	if path == "" {
		return tr("transfer.notSelected")
	}
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) > 3 && parts[1] == "Users" {
		return "~/" + strings.Join(parts[3:], "/")
	}
	return path
}
