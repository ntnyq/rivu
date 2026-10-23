//go:build !windows

package app

import "time"

func trayDoubleClickInterval() time.Duration { return 0 }
