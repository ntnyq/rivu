package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/internal/session"
)

func problemText(issue *problem.Error, fallback string) string {
	if issue == nil {
		return fallback
	}
	labels := map[problem.Code]string{
		problem.DiskFull:        tr("磁盘空间不足，请释放空间或重新定位目录。"),
		problem.Destination:     tr("下载目录不可用，请连接磁盘或重新定位目录。"),
		problem.Metadata:        tr("元数据获取失败，请检查来源后重试。"),
		problem.Persistence:     tr("保存失败，请修复目录权限或空间后重新读取。"),
		problem.RestoreDisabled: tr("恢复已关闭，请手动继续"),
		problem.Conflict:        tr("目标路径已由另一个任务使用"),
		problem.Engine:          tr("任务异常，请检查文件和连接后重试。"),
	}
	text := labels[issue.Code]
	if issue.Key != "" {
		return text + "\n" + fmt.Sprintf(strings.ReplaceAll(tr(issue.Key), "%w", "%v"), issue.Args...)
	}
	if issue.Context != "" {
		text += "\n" + issue.Context
	}
	if issue.Cause != nil {
		text += "\n" + issue.Cause.Error()
	}
	return text
}
func taskError(task session.Task) string { return problemText(task.Issue, task.Error) }

func (w *mainWindow) chooseRelocation(id string) {
	var destination string
	w.run(tr("选择下载目录"), func(context.Context) (string, error) {
		var err error
		destination, err = w.desktop.PickFolder()
		return tr("已选择目录"), err
	}, func() {
		w.relocateID, w.relocatePath, w.relocateOpen = id, destination, true
	})
}
func (w *mainWindow) relocationConfirmation(c *ui.Context) {
	if !w.relocateOpen {
		return
	}
	ui.DialogBase(c, &w.relocateOpen, func(_, panel ui.Element) {
		panel.Width(460).Padding(24).Gap(12).Children(func() {
			ui.Text(c, tr("确认重新定位")).Bold()
			ui.Text(c, w.relocatePath).MaxLines(3)
			ui.Text(c, tr("不会移动原目录文件。新目录中的已有数据将被校验并保留删除权限；缺失或损坏的数据将在继续任务后下载。"))
			if ui.Button(c, tr("取消")).Clicked() {
				w.relocateOpen = false
			}
			if ui.Button(c, tr("确认重新定位")).Clicked() {
				w.app.execute(session.Command{Kind: session.CommandRelocate, IDs: []string{w.relocateID}, Destination: w.relocatePath})
				w.relocateOpen = false
			}
		})
	})
}
