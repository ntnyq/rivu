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

func (w *windowBase) run(label string, work func(context.Context) (string, error), apply func()) {
	if w.task.Running() {
		return
	}
	w.job = label
	w.report(label+"…", nil)
	w.task.Start(w.dispatch, func(ctx context.Context, _ func(float64)) (string, error) { return work(ctx) }, func(float64) {}, func(message string, err error) {
		w.job = ""
		if errors.Is(err, desktop.ErrCanceled) || errors.Is(err, context.Canceled) {
			w.report(tr("已取消"), nil)
			return
		}
		if err != nil {
			w.report(label, err)
			return
		}
		if apply != nil {
			apply()
		}
		w.report(tr(message), nil)
	})
}

func (w *windowBase) reveal(path string) {
	w.run(tr("在文件管理器中显示"), func(context.Context) (string, error) { return tr("已显示文件"), w.desktop.Reveal(path) }, nil)
}

func (w *windowBase) copy(text string) {
	w.run(tr("复制"), func(context.Context) (string, error) { return tr("已复制"), w.desktop.Copy(text) }, nil)
}

func (w *windowBase) openURL(raw string) {
	w.run(tr("打开网页"), func(context.Context) (string, error) { return tr("已打开浏览器"), w.desktop.OpenURL(raw) }, nil)
}
func applyTheme(theme string) { mygo.Theme.SetSource(mygo.ThemeSource(theme)) }
