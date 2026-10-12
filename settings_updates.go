package main

import (
	"context"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/updater"
)

func (w *window) updateSettingsView(c *ui.Context, p palette) {
	ui.Row(c).Gap(13).Children(func() {
		ui.Image(c, appMark).Size(48, 48)
		ui.Column(c).Gap(7).Children(func() {
			ui.Text(c, "Rivu").FontSize(16).Bold()
			ui.Text(c, "版本 "+version+" · 为你的 Mac 而设计").FontSize(11).TextColor(p.muted)
		})
	})
	s := updater.Snapshot{State: "idle"}
	available := w.app != nil && w.app.updater != nil
	if available {
		s = w.app.updater.Snapshot()
	}
	labels := map[string]string{"idle": "尚未检查更新", "checking": "正在检查更新…", "unavailable": "更新不可用", "upToDate": "已是最新版本", "available": "Rivu " + s.Version + " 可供更新", "installing": "正在下载并安装…", "restartRequired": "更新已安装，需要重启", "failed": "检查失败", "installFailed": "安装失败"}
	settingGroup(c, func() {
		ui.Column(c).Padding(17, 0).Gap(11).Children(func() {
			ui.Text(c, labels[s.State]).FontSize(12).Bold()
			if s.Error != "" {
				ui.Text(c, s.Error).FontSize(11).TextColor(p.muted)
			}
			if s.Notes != "" {
				ui.Text(c, s.Notes).FontSize(11).TextColor(p.muted).MaxLines(5)
			}
			if s.State == "idle" {
				ui.Text(c, "检查是否有新版本，让 Rivu 保持最新。").FontSize(11).TextColor(p.muted)
			}
			if !s.LastSuccess.IsZero() {
				ui.Text(c, "上次检查："+s.LastSuccess.Format("2006-01-02 15:04")).FontSize(10).TextColor(p.muted)
			}
			if s.State == "installing" && s.Total > 0 {
				progressTrack(c, float64(s.Downloaded)/float64(s.Total), p.blue, p.track, 3)
			}
			if (s.State == "available" || s.State == "installFailed") && actionButton(c, "download", "下载并安装更新…", true).FontSize(11).Disabled(w.task.Running()).Clicked() {
				w.run("安装更新", func(ctx context.Context) (string, error) {
					return "安装完成，请重启", w.app.updater.Install(ctx, false)
				}, nil)
			}
			if s.State == "restartRequired" && actionButton(c, "refresh", "保存并重启", true).FontSize(11).Clicked() {
				w.app.relaunch = true
				mygo.App.Quit()
			}
		})
	})
	settingGroup(c, func() {
		switchRow(c, "自动检查更新", "获得同意后，每天检查一次", &w.settings.AutoCheck, false, true)
		switchRow(c, "空闲时自动安装更新", "无下载、校验或做种任务时安装", &w.settings.AutoInstall, true, w.settings.AutoCheck)
	})
	if !w.settings.AutoCheck {
		w.settings.AutoInstall = false
	}
	ui.Row(c).Gap(8).Children(func() {
		if actionButton(c, "refresh", "检查更新", false).FontSize(11).Disabled(!available || w.task.Running()).Clicked() {
			w.run("检查更新", func(ctx context.Context) (string, error) { return "检查完成", w.app.updater.Check(ctx) }, nil)
		}
		ui.Spacer(c)
		if actionButton(c, "globe", "产品网站", false).FontSize(11).Clicked() {
			w.openURL(website)
		}
		if actionButton(c, "github", "GitHub", false).FontSize(11).Clicked() {
			w.openURL(repository)
		}
	})
	ui.Text(c, "自动安装默认关闭。安装前会保存会话，完成后由你决定重启。").FontSize(11).TextColor(p.muted)
}
