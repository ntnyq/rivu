package app

import (
	"context"
	"errors"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/desktop"
	"github.com/ntnyq/rivu/internal/problem"
	"github.com/ntnyq/rivu/internal/windowtask"
	"github.com/ntnyq/rivu/locales"
)

// windowBase owns only native lifetime, status, and cancellable work.
type windowBase struct {
	app            *application
	win            *mygo.Window
	desktop        desktop.Service
	task           windowtask.Task
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
			w.report(locales.CommonCanceled(), nil)
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

// runCommitted tracks status for admitted application work. Closing this
// window dismisses its status without canceling a global change.
func (w *windowBase) runCommitted(label string, work func() (string, error), apply func()) bool {
	if w.closed || w.task.Running() {
		return false
	}
	if w.app == nil {
		w.run(label, func(context.Context) (string, error) { return work() }, apply)
		return true
	}
	if w.app.quitting || w.app.shutdownReady {
		return false
	}
	type result struct {
		message string
		err     error
	}
	completed := make(chan result, 1)
	w.app.work.enqueue(func() {
		message, err := work()
		completed <- result{message, err}
	})
	w.run(label, func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case r := <-completed:
			return r.message, r.err
		}
	}, apply)
	return true
}

func (w *windowBase) reveal(path string) {
	w.run(locales.FilesReveal(), func(context.Context) (string, error) { return locales.FilesRevealed(), w.desktop.Reveal(path) }, nil)
}

func (w *windowBase) copy(text string) {
	w.run(locales.CommonCopy(), func(context.Context) (string, error) { return locales.CommonCopied(), w.desktop.Copy(text) }, nil)
}

func (w *windowBase) openURL(raw string) {
	w.run(locales.CommonOpenWebsite(), func(context.Context) (string, error) { return locales.CommonBrowserOpened(), w.desktop.OpenURL(raw) }, nil)
}
func applyTheme(theme string) { mygo.Theme.SetSource(mygo.ThemeSource(theme)) }
