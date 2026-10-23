package app

import (
	"time"

	"golang.org/x/sys/windows"
)

var getTrayDoubleClickTime = windows.NewLazySystemDLL("user32.dll").NewProc("GetDoubleClickTime")

func trayDoubleClickInterval() time.Duration {
	ms, _, _ := getTrayDoubleClickTime.Call()
	return time.Duration(ms) * time.Millisecond
}
