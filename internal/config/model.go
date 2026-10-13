package config

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ntnyq/rivu/internal/i18n"
)

type Settings struct {
	ScheduleEnabled  bool   `json:"scheduleEnabled"`
	ScheduleStart    int    `json:"scheduleStart"`
	ScheduleEnd      int    `json:"scheduleEnd"`
	SeedLimitSeconds int64  `json:"seedLimitSeconds"`
	WatchFolder      string `json:"watchFolder,omitempty"`
	WatchTag         string `json:"watchTag,omitempty"`
	WatchEnabled     bool   `json:"watchEnabled"`

	Version         int       `json:"version"`
	Theme           string    `json:"theme"`
	Language        string    `json:"language"`
	Destination     string    `json:"destination"`
	Concurrent      int       `json:"concurrent"`
	Restore         bool      `json:"restore"`
	Tray            bool      `json:"tray"`
	Login           bool      `json:"login"`
	BackgroundLogin bool      `json:"backgroundLogin"`
	NotifyComplete  bool      `json:"notifyComplete"`
	NotifyError     bool      `json:"notifyError"`
	ConfirmQuit     bool      `json:"confirmQuit"`
	SeedRatio       float64   `json:"seedRatio"` // -1 unlimited, 0 stop on completion
	DownloadLimit   int64     `json:"downloadLimit"`
	UploadLimit     int64     `json:"uploadLimit"`
	SlowDownload    int64     `json:"slowDownload"`
	SlowUpload      int64     `json:"slowUpload"`
	SlowMode        bool      `json:"slowMode"`
	ListenPort      int       `json:"listenPort"`
	DHT             bool      `json:"dht"`
	PEX             bool      `json:"pex"`
	AutoCheck       bool      `json:"autoCheck"`
	AutoInstall     bool      `json:"autoInstall"`
	LastCheck       time.Time `json:"lastCheck"`
	LastAttempt     time.Time `json:"lastAttempt"`
}

func Default() Settings {
	home, _ := os.UserHomeDir()
	return Settings{ScheduleStart: 22 * 60, ScheduleEnd: 8 * 60, Version: 2, Theme: "system", Language: "system", Destination: filepath.Join(home, "Downloads", "Rivu"), Concurrent: 3, Restore: true, Tray: true, BackgroundLogin: true, NotifyComplete: true, NotifyError: true, ConfirmQuit: true, SeedRatio: 2, SlowDownload: 2000000, SlowUpload: 512000, DHT: true, PEX: true}
}
func (s Settings) Validate() error {
	if s.ScheduleStart < 0 || s.ScheduleStart >= 1440 || s.ScheduleEnd < 0 || s.ScheduleEnd >= 1440 || s.ScheduleEnabled && s.ScheduleStart == s.ScheduleEnd {
		return fmt.Errorf("invalid schedule interval")
	}
	if s.SeedLimitSeconds < 0 || s.SeedLimitSeconds > 315360000 {
		return fmt.Errorf("invalid seeding duration")
	}
	if s.WatchEnabled && !filepath.IsAbs(s.WatchFolder) {
		return fmt.Errorf("watch folder must be absolute")
	}
	if len(s.WatchTag) > 80 || strings.ContainsAny(s.WatchTag, ",\n\r\x00") || strings.TrimSpace(s.WatchTag) != s.WatchTag {
		return fmt.Errorf("invalid watch folder tag")
	}

	if s.Version != 2 {
		return fmt.Errorf(i18n.T("config.unsupportedVersion"), s.Version)
	}
	if !i18n.Valid(s.Language) {
		return fmt.Errorf(i18n.T("config.unknownLanguage"), s.Language)
	}
	if s.Theme != "system" && s.Theme != "light" && s.Theme != "dark" {
		return fmt.Errorf(i18n.T("config.unknownTheme"), s.Theme)
	}
	if !filepath.IsAbs(s.Destination) {
		return fmt.Errorf("%s", i18n.T("config.absoluteDestination"))
	}
	if s.Concurrent < 1 || s.Concurrent > 32 {
		return fmt.Errorf("%s", i18n.T("config.concurrency"))
	}
	if math.IsNaN(s.SeedRatio) || math.IsInf(s.SeedRatio, 0) || s.SeedRatio < -1 {
		return fmt.Errorf("%s", i18n.T("config.seedRatio"))
	}
	if s.DownloadLimit < 0 || s.UploadLimit < 0 || s.SlowDownload <= 0 || s.SlowUpload <= 0 {
		return fmt.Errorf("%s", i18n.T("config.bandwidth"))
	}
	if s.ListenPort < 0 || s.ListenPort > 65535 {
		return fmt.Errorf("%s", i18n.T("config.listenPort"))
	}
	if s.AutoInstall && !s.AutoCheck {
		return fmt.Errorf("%s", i18n.T("config.autoInstallRequiresCheck"))
	}
	return nil
}
func (s Settings) Limits() (int64, int64) {
	if s.SlowMode {
		return s.SlowDownload, s.SlowUpload
	}
	return s.DownloadLimit, s.UploadLimit
}
