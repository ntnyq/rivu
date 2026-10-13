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
		problem.DiskFull:        tr("recovery.diskFull"),
		problem.Destination:     tr("recovery.destination"),
		problem.Metadata:        tr("recovery.metadata"),
		problem.Persistence:     tr("recovery.persistence"),
		problem.RestoreDisabled: tr("recovery.restoreDisabled"),
		problem.Conflict:        tr("files.destinationConflict"),
		problem.Engine:          tr("recovery.engine"),
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
	w.run(tr("files.chooseDownloadFolder"), func(context.Context) (string, error) {
		var err error
		destination, err = w.desktop.PickFolder()
		return tr("files.folderSelected"), err
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
			ui.Text(c, tr("relocation.confirm")).Bold()
			ui.Text(c, w.relocatePath).MaxLines(3)
			ui.Text(c, tr("relocation.hint"))
			if ui.Button(c, tr("common.cancel")).Clicked() {
				w.relocateOpen = false
			}
			if ui.Button(c, tr("relocation.confirm")).Clicked() {
				w.app.execute(session.Command{Kind: session.CommandRelocate, IDs: []string{w.relocateID}, Destination: w.relocatePath})
				w.relocateOpen = false
			}
		})
	})
}
