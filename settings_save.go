package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/config"
)

func (w *settingsWindow) saveSettings() {
	s := w.settings
	s.Destination = w.destination
	var err error
	s.Concurrent, err = strconv.Atoi(w.concurrent)
	if err != nil {
		w.report(tr("并行数量无效"), err)
		return
	}
	s.SeedRatio = 0
	if w.seedMode == 2 {
		s.SeedRatio = -1
	}
	if w.seedMode == 1 {
		s.SeedRatio, err = strconv.ParseFloat(w.seedRatio, 64)
		if err == nil && s.SeedRatio <= 0 {
			err = fmt.Errorf("ratio must be positive")
		}
	}
	if err != nil {
		w.report(tr("分享率无效"), err)
		return
	}
	s.ListenPort, err = strconv.Atoi(w.listenPort)
	if err != nil {
		w.report(tr("端口无效"), err)
		return
	}
	for _, v := range []struct {
		input bandwidthInput
		value *int64
	}{{w.downloadLimit, &s.DownloadLimit}, {w.uploadLimit, &s.UploadLimit}, {w.slowDownload, &s.SlowDownload}, {w.slowUpload, &s.SlowUpload}} {
		*v.value, err = v.input.bytes()
		if err != nil {
			w.report(tr("限速数值无效"), err)
			return
		}
	}
	s.SeedLimitSeconds, err = parseScaled(w.seedHours, 3600)
	if err != nil {
		w.report(tr("设置无效"), err)
		return
	}
	s.ScheduleStart, err = clockMinute(w.scheduleStart)
	if err != nil {
		w.report(tr("设置无效"), err)
		return
	}
	s.ScheduleEnd, err = clockMinute(w.scheduleEnd)
	if err != nil {
		w.report(tr("设置无效"), err)
		return
	}
	if err = s.Validate(); err != nil {
		w.report(tr("设置无效"), err)
		return
	}
	w.run(tr("保存设置"), func(context.Context) (string, error) {
		if w.app != nil {
			w.app.settingsMu.Lock()
			defer w.app.settingsMu.Unlock()
		}
		old := w.store.Snapshot()
		if old.Login != s.Login {
			if err := mygo.App.SetOpenAtLogin(s.Login); err != nil {
				return "", fmt.Errorf(tr("登录项注册失败: %w"), err)
			}
		}
		if err := w.store.SavePreferences(s); err != nil {
			if old.Login != s.Login {
				_ = mygo.App.SetOpenAtLogin(old.Login)
			}
			return "", err
		}
		if w.app != nil {
			if err := w.app.session.Configure(w.store.Snapshot()); err != nil {
				return "", err
			}
		}
		return "设置已保存；连接设置将在下次启动生效", nil
	}, func() {
		w.settings = w.store.Snapshot()
		applyTheme(w.settings.Theme)
		if w.app != nil {
			w.app.applyLanguage()
		}
		if w.app != nil {
			if main := w.app.window; main != nil && !main.closed {
				main.dispatch(func() { main.settings = w.settings; main.refreshDisk() })
			}
		}
		w.loginStatus = fmt.Sprintf(tr("系统当前启用：%t"), mygo.App.OpenAtLogin())
	})
}
func (w *mainWindow) saveQuitPreference() {
	w.run(tr("保存退出偏好"), func(context.Context) (string, error) {
		return tr("已保存"), w.store.Update(func(p *config.Settings) { p.ConfirmQuit = false })
	}, w.app.shutdown)
}
