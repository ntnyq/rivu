package app

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

func (w *mainWindow) inspectorView(c *ui.Context) {
	p := designTheme(c)
	theme := c.Theme()
	ui.Column(c).Width(271).Shrink(0).FillHeight().Padding(22, 18).Gap(0).Background(p.alt).BorderWidth(0, 0, 0, 1).BorderColor(p.line).Children(func() {
		t := w.chosen()
		if t == nil {
			ui.Text(c, locales.TransferDetails()).FontSize(14).FontWeight(600)
			ui.Text(c, locales.TransferSelectHint()).FontSize(11).TextColor(p.muted).Padding(20, 0)
			return
		}
		ui.Row(c).Height(48).Children(func() {
			fileSymbol(c, *t, true)
			ui.Spacer(c)
			ui.Text(c, phaseLabel(t.Phase)).FontSize(10).TextColor(taskColor(c, t.Phase)).Padding(4, 7).Radius(5).Background(p.soft)
		})
		ui.Text(c, t.Source.Name).FontSize(14).FontWeight(600).MaxLines(3).Padding(17, 0, 6, 0)
		subtitle := locales.TransferFileSummary(bytes(t.TotalSize))
		if t.Metadata == "" {
			subtitle = locales.TransferWaitingMetadata()
		}
		ui.Text(c, subtitle).FontSize(10).TextColor(p.muted)
		ui.Row(c).Height(28).Shrink(0).Padding(2).Gap(3).Radius(6).Background(p.track).Margin(22, 0).Children(func() {
			for i, label := range []string{locales.TransferOverview(), locales.FilesTitle(), locales.CommonConnections()} {
				tab := ui.Row(c).Grow(1).Height(24).Radius(4).Justify(ui.Center).Focusable().Label(label)
				if w.tab == i {
					tab.Background(p.surface).Border(1, p.line)
				}
				tab.Children(func() { ui.Text(c, label).FontSize(10) })
				if tab.Clicked() {
					w.tab = i
				}
			}
		})
		ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).Gap(0).Children(func() {
				switch w.tab {
				case 0:
					ui.Text(c, locales.TransferProgress()).FontSize(10).TextColor(p.muted)
					ui.Row(c).Gap(2).AlignItems(ui.End).Padding(8, 0, 9, 0).Children(func() {
						ui.Text(c, fmt.Sprintf("%.1f", progress(*t)*100)).FontSize(26).FontWeight(600)
						ui.Text(c, "%").FontSize(13).TextColor(p.muted)
					})
					progressTrack(c, progress(*t), taskColor(c, t.Phase), p.track, 4)
					ui.Text(c, bytes(t.Done)+" / "+bytes(t.Size)).FontSize(10).TextColor(p.muted).Padding(9, 0, 0, 0)
					ui.Row(c).Gap(20).Padding(22, 0, 18, 0).Children(func() {
						for i, v := range []struct {
							label string
							value float64
						}{{locales.TransferDownloadSpeed(), t.DownRate}, {locales.TransferUploadSpeed(), t.UpRate}} {
							ui.Column(c).Grow(1).Gap(7).Children(func() {
								ui.Text(c, v.label).FontSize(10).TextColor(p.muted)
								number, unit := rateParts(v.value)
								color := p.text
								if i == 0 {
									color = p.blue
								}
								ui.Row(c).Gap(3).AlignItems(ui.End).Children(func() {
									ui.Text(c, number).FontSize(16).Bold().TextColor(color)
									ui.Text(c, unit).FontSize(10).TextColor(p.muted).Padding(0, 0, 1, 0)
								})
							})
						}
					})
					ui.Column(c).Padding(14, 0, 0, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Gap(12).Children(func() {
						for _, v := range [][2]string{{locales.TransferTimeRemaining(), eta(*t)}, {locales.TransferUploaded(), bytes(t.Uploaded)}, {locales.TransferShareRatio(), fmt.Sprintf("%.2f", float64(t.Uploaded)/float64(max(1, t.TotalSize)))}, {locales.TransferConnectedPeers(), fmt.Sprint(t.Peers)}, {locales.TransferDateAdded(), t.Added.Format("01-02 15:04")}} {
							ui.Row(c).Gap(8).Children(func() { ui.Text(c, v[0]).FontSize(10).TextColor(p.muted); ui.Spacer(c); ui.Text(c, v[1]).FontSize(10) })
						}
					})
					ui.Column(c).Margin(17, 0, 0, 0).Padding(17, 0, 0, 0).BorderWidth(1, 0, 0, 0).BorderColor(p.line).Gap(10).Children(func() {
						ui.Text(c, locales.FilesSaveLocation()).FontSize(10).TextColor(p.muted)
						loc := ui.Row(c).Gap(7).Focusable().Label(locales.FilesShowInFinder())
						loc.Children(func() {
							glyph(c, "folder", 15, p.blue)
							ui.Text(c, compactPath(t.Destination)).FontSize(10).SingleLine().Grow(1).MinWidth(0)
							glyph(c, "external", 15, p.blue)
						})
						if loc.Clicked() {
							w.reveal(t.Destination)
						}
					})
					seed := locales.SeedingKeep()
					if t.SeedRatio == 0 {
						seed = locales.SeedingStopAfterDownload()
					}
					note := locales.SeedingForever()
					if t.SeedRatio == 0 {
						note = ""
					}
					if t.SeedRatio > 0 {
						note = locales.SeedingRatioGoal(t.SeedRatio)
					}
					ui.Text(c, seed+"\n"+note).FontSize(10).TextColor(p.muted).Padding(18, 0, 0, 0)
					change := actionButton(c, "settings", locales.SeedingPolicy(), false).FontSize(10).TextColor(p.blue).Margin(10, 0).Disabled(w.app == nil)
					change.Menu(func(m *ui.Menu) {
						for _, v := range []struct {
							label string
							ratio float64
						}{{locales.SeedingStopOnCompletion(), 0}, {locales.SeedingRatioTwo(), 2}, {locales.SeedingForever(), -1}} {
							if m.Item(v.label).Chosen() {
								w.app.execute(session.Command{Kind: session.CommandSeed, IDs: []string{t.ID}, Ratio: v.ratio, SeedLimitSeconds: t.SeedLimitSeconds})
							}
						}
					})
					ui.Text(c, locales.SeedingElapsed(float64(t.SeedSeconds)/3600)).FontSize(10).TextColor(p.muted)
					durationLabel := locales.SeedingUnlimitedDuration()
					if t.SeedLimitSeconds > 0 {
						durationLabel = locales.SeedingDurationLimit(float64(t.SeedLimitSeconds) / 3600)
					}
					menuControl(c, locales.SeedingDuration(), durationLabel).Disabled(w.app == nil).Menu(func(m *ui.Menu) {
						for _, hours := range []int64{0, 1, 24, 72} {
							label := locales.SeedingUnlimitedDuration()
							if hours > 0 {
								label = locales.SeedingDurationLimit(float64(hours))
							}
							if m.Item(label).Checked(t.SeedLimitSeconds == hours*3600).Chosen() {
								w.app.execute(session.Command{Kind: session.CommandSeed, IDs: []string{t.ID}, Ratio: t.SeedRatio, SeedLimitSeconds: hours * 3600})
							}
						}
					})
					if w.tagsTask != t.ID {
						w.tagsTask = t.ID
						w.tagsDraft = strings.Join(t.Tags, ", ")
					}
					field(c, locales.TransferTags(), &w.tagsDraft)
					if ui.Button(c, locales.TransferSaveTags()).Disabled(w.app == nil).Clicked() {
						w.app.execute(session.Command{Kind: session.CommandTags, IDs: []string{t.ID}, Tags: parseTags(w.tagsDraft)})
					}
					if ui.Button(c, locales.TransferChangeLocation()).Disabled(w.app == nil || t.Metadata == "").Clicked() {
						w.chooseRelocation(t.ID)
					}
					if t.Error != "" {
						ui.Text(c, taskError(*t)).FontSize(11).TextColor(p.orange).Padding(12).Radius(7).Background(p.orange.Alpha(.09))
						if ui.Button(c, locales.TransferRetry()).Disabled(w.app == nil).Clicked() {
							w.app.execute(session.Command{Kind: session.CommandResume, IDs: []string{t.ID}})
						}
					}

				case 1:
					if t.Metadata == "" {
						ui.Text(c, locales.FilesPendingMetadata())
						return
					}
					ui.Text(c, locales.FilesBoundaryHint()).FontSize(11).TextColor(theme.TextMuted)
					fileTree(c, t.Source.Files, &w.fileState, w.app != nil, func(indices []int, priority int) {
						w.app.execute(session.Command{Kind: session.CommandPriority, IDs: []string{t.ID}, FileIndices: indices, Priority: priority})
					})
				case 2:
					ui.Text(c, locales.TransferPeerCount(t.Peers))
					ui.Text(c, "Tracker").Bold()
					if len(t.Trackers) == 0 {
						ui.Text(c, locales.TrackerNoObservations()).TextColor(theme.TextMuted)
					}
					for _, tr := range t.Trackers {
						ui.Text(c, tr).FontSize(12).MaxLines(2)
					}
					for _, tracker := range t.Diagnostics {
						state := map[string]string{"waiting": locales.TrackerWaiting(), "announcing": locales.TrackerAnnouncing(), "ok": locales.TrackerSucceeded(), "failed": locales.TrackerFailed()}[tracker.State]
						detail := map[string]string{"timeout": locales.TrackerTimeout(), "dns": locales.TrackerDnsFailed(), "refused": locales.TrackerRefused(), "authorization": locales.TrackerUnauthorized(), "announce": locales.TrackerError()}[tracker.Detail]
						ui.Text(c, tracker.Host+" · "+state).FontSize(11).MaxLines(2)
						if detail != "" {
							ui.Text(c, detail).FontSize(10).TextColor(p.orange)
						}
						if !tracker.Next.IsZero() {
							ui.Text(c, locales.TrackerNextAnnounce(tracker.Next.Local().Format("15:04:05"))).FontSize(10)
						}
					}
					if t.Peers == 0 {
						ui.Text(c, locales.TrackerWaitingHint()).FontSize(11)
					}
					ui.Text(c, locales.TrackerPrivacyHint()).FontSize(10).TextColor(theme.TextMuted)
				}
			})
		})
	})
}
