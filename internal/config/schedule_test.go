package config

import (
	"testing"
	"time"
)

func TestScheduledLimits(t *testing.T) {
	s := Default()
	s.ScheduleEnabled = true
	s.ScheduleStart = 22 * 60
	s.ScheduleEnd = 8 * 60
	for _, tc := range []struct {
		hour, minute int
		slow         bool
	}{{21, 59, false}, {22, 0, true}, {0, 0, true}, {7, 59, true}, {8, 0, false}} {
		now := time.Date(2026, 10, 13, tc.hour, tc.minute, 0, 0, time.FixedZone("test", 8*3600))
		if s.SlowAt(now) != tc.slow {
			t.Fatal("overnight boundary", tc)
		}
		d, u := s.LimitsAt(now)
		if tc.slow && (d != s.SlowDownload || u != s.SlowUpload) {
			t.Fatal("wrong limits")
		}
	}
	s.ScheduleStart = 8 * 60
	s.ScheduleEnd = 18 * 60
	if !s.SlowAt(time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)) {
		t.Fatal("day interval")
	}
	s.ScheduleEnabled = false
	s.SlowMode = true
	if !s.SlowAt(time.Now()) {
		t.Fatal("manual override lost")
	}
	s.ScheduleEnabled = true
	s.ScheduleEnd = s.ScheduleStart
	if s.Validate() == nil {
		t.Fatal("ambiguous all-day interval accepted")
	}
}
