package app

import (
	"testing"
	"time"
)

func TestTrayClickGesture(t *testing.T) {
	var queued []func()
	singles, doubles, stopped := 0, 0, 0
	g := trayClickGesture{
		interval: 700 * time.Millisecond,
		single:   func() { singles++ },
		double:   func() { doubles++ },
		after: func(delay time.Duration, fn func()) func() {
			if delay != 700*time.Millisecond {
				t.Fatalf("ignored system double-click interval: %v", delay)
			}
			queued = append(queued, fn)
			return func() { stopped++ }
		},
	}
	g.click()
	if singles != 0 || doubles != 0 {
		t.Fatal("first click opened a window before the double-click interval")
	}
	queued[0]()
	if singles != 1 || doubles != 0 {
		t.Fatal("single click did not open the flyout")
	}
	g.click()
	g.click()
	// Simulate a timer that fired before Stop, but whose main-thread callback
	// runs after the double click. It must not reopen the flyout.
	queued[1]()
	if singles != 1 || doubles != 1 || stopped != 1 {
		t.Fatalf("double click: single=%d double=%d stopped=%d", singles, doubles, stopped)
	}
	g.click()
	g.cancel() // Right-click, tray recreation, disable, or application exit.
	g.click()
	queued[2]()
	if singles != 1 || doubles != 1 {
		t.Fatal("canceled callback affected a subsequent click")
	}
	queued[3]()
	if singles != 2 || doubles != 1 {
		t.Fatal("gesture did not recover after cancellation")
	}
}
