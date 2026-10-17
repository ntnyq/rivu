package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/locales"
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
		menuControl(c, label+" "+locales.SettingsBandwidthUnit(), unit).Width(75).Menu(func(m *ui.Menu) {
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
	labels := []string{locales.SeedingStopOnCompletion(), locales.SeedingUntilRatio(), locales.SeedingForever()}
	ui.Row(c).Padding(12, 0).Children(func() {
		ui.Text(c, locales.SettingsSeedingPolicy()).FontSize(12)
		ui.Spacer(c)
		menuControl(c, locales.SettingsSeedingPolicy(), labels[*mode]).Width(145).Menu(func(m *ui.Menu) {
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
		switchRow(c, locales.SettingsSchedule(), locales.SettingsScheduleHint(), &w.settings.ScheduleEnabled, false, true)
		settingField(c, locales.SettingsScheduleStart(), &w.scheduleStart, true)
		settingField(c, locales.SettingsScheduleEnd(), &w.scheduleEnd, true)
	})
	settingGroup(c, func() {
		switchRow(c, locales.SettingsWatchFolder(), locales.SettingsWatchFolderHint(), &w.settings.WatchEnabled, false, true)
		settingField(c, locales.SettingsWatchFolderPath(), &w.settings.WatchFolder, true)
		settingField(c, locales.SettingsWatchTag(), &w.settings.WatchTag, true)
	})
	ui.Text(c, locales.SettingsAutomationHint()).FontSize(11)
}
