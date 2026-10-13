package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

type bandwidthInput struct {
	value string
	unit  int64
}

func bandwidthValue(bytes int64) bandwidthInput {
	unit := int64(1000)
	if bytes >= 1000000 && bytes%1000000 == 0 {
		unit = 1000000
	}
	return bandwidthInput{strconv.FormatFloat(float64(bytes)/float64(unit), 'f', -1, 64), unit}
}
func parseScaled(text string, scale int64) (int64, error) {
	v, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v*float64(scale) >= float64(math.MaxInt64) {
		return 0, fmt.Errorf("invalid numeric value %q", text)
	}
	return int64(math.Round(v * float64(scale))), nil
}
func (v bandwidthInput) bytes() (int64, error) { return parseScaled(v.value, v.unit) }
func clockMinute(value string) (int, error) {
	t, err := time.Parse("15:04", value)
	return t.Hour()*60 + t.Minute(), err
}

func bandwidthField(c *ui.Context, label string, value *bandwidthInput, line bool) {
	row := ui.Row(c).Padding(12, 0).Gap(8)
	if line {
		row.BorderWidth(1, 0, 0, 0).BorderColor(c.Theme().Border)
	}
	row.Children(func() {
		ui.Text(c, label).FontSize(11).Grow(1)
		ui.TextInput(c, &value.value).Label(label).Width(80).FontSize(11)
		unit := "KB/s"
		if value.unit == 1000000 {
			unit = "MB/s"
		}
		menuControl(c, label+" "+tr("单位"), unit).Width(75).Menu(func(m *ui.Menu) {
			for _, choice := range []struct {
				label string
				scale int64
			}{{"KB/s", 1000}, {"MB/s", 1000000}} {
				if m.Item(choice.label).Checked(value.unit == choice.scale).Chosen() {
					if bytes, err := value.bytes(); err == nil {
						value.value = strconv.FormatFloat(float64(bytes)/float64(choice.scale), 'f', -1, 64)
					}
					value.unit = choice.scale
				}
			}
		})
	})
}
func seedingControl(c *ui.Context, mode *int) {
	labels := []string{tr("完成后停止"), tr("达到分享率"), tr("无限做种")}
	ui.Row(c).Padding(12, 0).Children(func() {
		ui.Text(c, tr("做种策略")).FontSize(12)
		ui.Spacer(c)
		menuControl(c, tr("做种策略"), labels[*mode]).Width(145).Menu(func(m *ui.Menu) {
			for i, label := range labels {
				if m.Item(label).Checked(*mode == i).Chosen() {
					*mode = i
				}
			}
		})
	})
}
func (w *settingsWindow) automationView(c *ui.Context) {
	settingGroup(c, func() {
		switchRow(c, tr("定时限速"), tr("每天按本地时间启用临时限速，支持跨午夜。"), &w.settings.ScheduleEnabled, false, true)
		settingField(c, tr("开始时间 HH:MM"), &w.scheduleStart, true)
		settingField(c, tr("结束时间 HH:MM"), &w.scheduleEnd, true)
	})
	settingGroup(c, func() {
		switchRow(c, tr("监视文件夹"), tr("稳定的种子文件会自动导入默认下载目录；原文件保留。"), &w.settings.WatchEnabled, false, true)
		settingField(c, tr("种子文件夹绝对路径"), &w.settings.WatchFolder, true)
		settingField(c, tr("自动添加标签"), &w.settings.WatchTag, true)
	})
	ui.Text(c, tr("自动策略不会恢复手动暂停的任务。做种时长只累计实际做种时间。")).FontSize(11)
}
