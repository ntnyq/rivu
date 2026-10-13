package config

import "time"

// SlowAt evaluates a daily local-time interval, including overnight schedules.
// Manual slow mode is an independent override; scheduling never changes Run.
func (s Settings) SlowAt(now time.Time) bool {
	if s.SlowMode {
		return true
	}
	if !s.ScheduleEnabled {
		return false
	}
	minute := now.Hour()*60 + now.Minute()
	if s.ScheduleStart < s.ScheduleEnd {
		return minute >= s.ScheduleStart && minute < s.ScheduleEnd
	}
	return minute >= s.ScheduleStart || minute < s.ScheduleEnd
}

func (s Settings) LimitsAt(now time.Time) (int64, int64) {
	if s.SlowAt(now) {
		return s.SlowDownload, s.SlowUpload
	}
	return s.DownloadLimit, s.UploadLimit
}
