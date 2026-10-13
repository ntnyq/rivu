package main

import (
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"

	"github.com/ntnyq/rivu/internal/session"
)

func filterLabels() []string {
	return []string{tr("全部任务"), tr("下载中"), tr("做种中"), tr("已完成"), tr("已暂停"), tr("需要处理")}
}

var phases = map[session.Phase]string{"queued": "等待队列", "metadata": "获取元数据", "verifying": "校验中", "waiting": "等待节点", "downloading": "下载中", "seeding": "做种中", "completed": "已完成", "paused": "已暂停", "error": "需要处理", "removing": "正在移除"}

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
		return fmt.Sprintf(tr("约 %d 秒"), seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf(tr("约 %d 分钟"), (seconds+30)/60)
	}
	return fmt.Sprintf(tr("约 %d 小时 %d 分钟"), seconds/3600, seconds%3600/60)
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
	summary := tr("无限做种")
	if ratio == 0 {
		return tr("下载完成后停止做种")
	}
	if ratio > 0 {
		summary = fmt.Sprintf(tr("分享率达到 %.1f 时自动停止"), ratio)
	}
	if seconds > 0 {
		summary += " · " + fmt.Sprintf(tr("最长 %.1f 小时"), float64(seconds)/3600)
	}
	return summary
}
