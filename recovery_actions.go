package main

import (
	"context"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/internal/session"
	"github.com/ntnyq/rivu/locales"
)

func problemText(issue *problem.Error, fallback string) string {
	if issue == nil {
		return fallback
	}
	labels := map[problem.Code]string{
		problem.DiskFull:        locales.RecoveryDiskFull(),
		problem.Destination:     locales.RecoveryDestination(),
		problem.Metadata:        locales.RecoveryMetadata(),
		problem.Persistence:     locales.RecoveryPersistence(),
		problem.RestoreDisabled: locales.RecoveryRestoreDisabled(),
		problem.Conflict:        locales.FilesDestinationConflict(),
		problem.Engine:          locales.RecoveryEngine(),
	}
	text := labels[issue.Code]
	if issue.Message != nil {
		return text + "\n" + issue.Message(locales.In(i18n.Locale()))
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
	w.run(locales.FilesChooseDownloadFolder(), func(context.Context) (string, error) {
		var err error
		destination, err = w.desktop.PickFolder()
		return locales.FilesFolderSelected(), err
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
			ui.Text(c, locales.RelocationConfirm()).Bold()
			ui.Text(c, w.relocatePath).MaxLines(3)
			ui.Text(c, locales.RelocationHint())
			if ui.Button(c, locales.CommonCancel()).Clicked() {
				w.relocateOpen = false
			}
			if ui.Button(c, locales.RelocationConfirm()).Clicked() {
				w.app.execute(session.Command{Kind: session.CommandRelocate, IDs: []string{w.relocateID}, Destination: w.relocatePath})
				w.relocateOpen = false
			}
		})
	})
}
