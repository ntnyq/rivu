package main

import (
	"context"
	"errors"
	"github.com/ntnyq/rivu/internal/problem"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/desktop"
	"github.com/ntnyq/rivu/internal/tasks"
)

// windowBase owns only native lifetime, status, and cancellable work.
type windowBase struct {
	app            *application
	win            *mygo.Window
	desktop        desktop.Service
	task           tasks.Task
	status, job    string
	failed, closed bool
}

func (w *windowBase) close() {
	if w.closed {
		return
	}
	w.closed = true
	w.task.Close()
}

func (w *windowBase) dispatch(fn func()) {
	if w.win == nil {
		if !w.closed {
			fn()
		}
		return
	}
	w.win.Update(func() {
		if !w.closed {
			fn()
		}
	})
}

func (w *windowBase) report(message string, err error) {
	w.failed = err != nil
	w.status = message
	if err != nil {
		detail := err.Error()
		var issue *problem.Error
		if errors.As(err, &issue) {
			detail = problemText(issue, detail)
		}
		w.status += "：" + detail
	}
}

// run expects work to return display text, already translated if appropriate.
func (w *windowBase) run(label string, work func(context.Context) (string, error), apply func()) {
	if w.task.Running() {
		return
	}
	w.job = label
	w.report(label+"…", nil)
	w.task.Start(w.dispatch, func(ctx context.Context, _ func(float64)) (string, error) { return work(ctx) }, func(float64) {}, func(message string, err error) {
		w.job = ""
		if errors.Is(err, desktop.ErrCanceled) || errors.Is(err, context.Canceled) {
			w.report(tr("common.canceled"), nil)
			return
		}
		if err != nil {
			w.report(label, err)
			return
		}
		if apply != nil {
			apply()
		}
		w.report(message, nil)
	})
}

func (w *windowBase) reveal(path string) {
	w.run(tr("files.reveal"), func(context.Context) (string, error) { return tr("files.revealed"), w.desktop.Reveal(path) }, nil)
}

func (w *windowBase) copy(text string) {
	w.run(tr("common.copy"), func(context.Context) (string, error) { return tr("common.copied"), w.desktop.Copy(text) }, nil)
}

func (w *windowBase) openURL(raw string) {
	w.run(tr("common.openWebsite"), func(context.Context) (string, error) { return tr("common.browserOpened"), w.desktop.OpenURL(raw) }, nil)
}
func applyTheme(theme string) { mygo.Theme.SetSource(mygo.ThemeSource(theme)) }
