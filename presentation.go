package main

import (
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"

	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

func filterLabels() []string {
	return []string{locales.FilterAll(), locales.PhaseDownloading(), locales.PhaseSeeding(), locales.PhaseCompleted(), locales.PhasePaused(), locales.PhaseError()}
}

func phaseLabel(phase session.Phase) string {
	switch phase {
	case session.PhaseQueued:
		return locales.PhaseQueued()
	case session.PhaseMetadata:
		return locales.PhaseMetadata()
	case session.PhaseVerifying:
		return locales.PhaseVerifying()
	case session.PhaseWaiting:
		return locales.PhaseWaiting()
	case session.PhaseDownloading:
		return locales.PhaseDownloading()
	case session.PhaseSeeding:
		return locales.PhaseSeeding()
	case session.PhaseCompleted:
		return locales.PhaseCompleted()
	case session.PhasePaused:
		return locales.PhasePaused()
	case session.PhaseError:
		return locales.PhaseError()
	case session.PhaseRemoving:
		return locales.PhaseRemoving()
	default:
		return "—"
	}
}

func matches(t session.Task, filter int) bool {
	switch filter {
	case 1:
		return t.Run && !t.Completed && !t.Removing && t.Error == ""
	case 2:
		return t.Phase == "seeding"
	case 3:
		return t.Completed
	case 4:
		return !t.Run && !t.Removing && !t.Completed
	case 5:
		return t.Error != ""
	}
	return true
}
func (w *mainWindow) visible() []session.Task {
	var out []session.Task
	for _, t := range w.snapshot.Tasks {
		if (w.tagFilter == "" || slices.Contains(t.Tags, w.tagFilter)) && matches(t, w.filter) && strings.Contains(strings.ToLower(t.Source.Name), strings.ToLower(w.search)) {
			out = append(out, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		switch w.sort {
		case 1:
			return out[i].Source.Name < out[j].Source.Name
		case 2:
			return progress(out[i]) > progress(out[j])
		case 3:
			return out[i].Phase < out[j].Phase
		}
		return out[i].Added.After(out[j].Added)
	})
	return out
}
func bytes(v int64) string {
	if v < 1000 {
		return fmt.Sprintf("%d B", v)
	}
	size := float64(v)
	units := []string{"KB", "MB", "GB", "TB"}
	for _, u := range units {
		size /= 1000
		if size < 1000 || u == "TB" {
			return fmt.Sprintf("%.1f %s", size, u)
		}
	}
	return "—"
}
func rate(v float64) string { return bytes(int64(v)) + "/s" }
func progress(t session.Task) float64 {
	if t.Size <= 0 {
		if t.Completed {
			return 1
		}
		return 0
	}
	return min(1, float64(t.Done)/float64(t.Size))
}
func eta(t session.Task) string {
	if t.Phase != "downloading" || t.DownRate < 1 {
		return "—"
	}
	seconds := max(1, int(float64(t.Size-t.Done)/t.DownRate))
	if seconds < 60 {
		return locales.TransferEtaSeconds(seconds)
	}
	if seconds < 3600 {
		return locales.TransferEtaMinutes((seconds + 30) / 60)
	}
	return locales.TransferEtaHours(seconds/3600, seconds%3600/60)
}
func magnet(t session.Task) string {
	v := url.Values{}
	if t.Source.Identity.V1 != "" {
		v.Add("xt", "urn:btih:"+t.Source.Identity.V1)
	}
	if t.Source.Identity.V2 != "" {
		v.Add("xt", "urn:btmh:1220"+t.Source.Identity.V2)
	}
	v.Set("dn", t.Source.Name)
	for _, tr := range t.Source.Trackers {
		v.Add("tr", tr)
	}
	return "magnet:?" + v.Encode()
}

func taskPercent(t session.Task) string {
	if t.Metadata == "" {
		return "—"
	}
	if progress(t) >= 1 {
		return "100%"
	}
	return fmt.Sprintf("%.1f%%", progress(t)*100)
}

func seedingSummary(ratio float64, seconds int64) string {
	summary := locales.SeedingForever()
	if ratio == 0 {
		return locales.SeedingStopAfterDownload()
	}
	if ratio > 0 {
		summary = locales.SeedingRatioGoal(ratio)
	}
	if seconds > 0 {
		summary += " · " + locales.SeedingDurationLimit(float64(seconds)/3600)
	}
	return summary
}
