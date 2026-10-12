package config

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

type Settings struct {
	Version         int       `json:"version"`
	Theme           string    `json:"theme"`
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
	return Settings{Version: 2, Theme: "system", Destination: filepath.Join(home, "Downloads", "Rivu"), Concurrent: 3, Restore: true, Tray: true, BackgroundLogin: true, NotifyComplete: true, NotifyError: true, ConfirmQuit: true, SeedRatio: 2, SlowDownload: 2000000, SlowUpload: 512000, DHT: true, PEX: true}
}
func (s Settings) Validate() error {
	if s.Version != 2 {
		return fmt.Errorf("unsupported settings version %d", s.Version)
	}
	if s.Theme != "system" && s.Theme != "light" && s.Theme != "dark" {
		return fmt.Errorf("unknown theme %q", s.Theme)
	}
	if !filepath.IsAbs(s.Destination) {
		return fmt.Errorf("download destination must be absolute")
	}
	if s.Concurrent < 1 || s.Concurrent > 32 {
		return fmt.Errorf("concurrent downloads must be 1–32")
	}
	if math.IsNaN(s.SeedRatio) || math.IsInf(s.SeedRatio, 0) || s.SeedRatio < -1 {
		return fmt.Errorf("invalid seed ratio")
	}
	if s.DownloadLimit < 0 || s.UploadLimit < 0 || s.SlowDownload <= 0 || s.SlowUpload <= 0 {
		return fmt.Errorf("invalid bandwidth limit")
	}
	if s.ListenPort < 0 || s.ListenPort > 65535 {
		return fmt.Errorf("invalid listening port")
	}
	if s.AutoInstall && !s.AutoCheck {
		return fmt.Errorf("automatic installation requires automatic checks")
	}
	return nil
}
func (s Settings) Limits() (int64, int64) {
	if s.SlowMode {
		return s.SlowDownload, s.SlowUpload
	}
	return s.DownloadLimit, s.UploadLimit
}
