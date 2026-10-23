package app

import "time"

// All gesture state and callbacks run on the main thread. Delay the first
// click so opening a focusable flyout cannot interfere with the second click.
type trayClickGesture struct {
	interval       time.Duration
	single, double func()
	after          func(time.Duration, func()) func()
	stop           func()
	generation     uint64
}

func (g *trayClickGesture) click() {
	if g.stop != nil {
		g.cancel()
		g.double()
		return
	}
	g.generation++
	generation := g.generation
	g.stop = g.after(g.interval, func() {
		// Stopping a timer cannot retract a callback already queued on main.
		if g.generation != generation {
			return
		}
		g.stop = nil
		g.single()
	})
}

func (g *trayClickGesture) cancel() {
	g.generation++
	if g.stop != nil {
		g.stop()
		g.stop = nil
	}
}
