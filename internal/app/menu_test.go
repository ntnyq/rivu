package app

import (
	"fmt"
	"slices"
	"testing"
)

func TestOpenTorrentCommandDuringShutdown(t *testing.T) {
	for _, ready := range []bool{false, true} {
		for _, closedDraft := range []bool{false, true} {
			t.Run(fmt.Sprintf("shutdownReady=%t/closedDraft=%t", ready, closedDraft), func(t *testing.T) {
				a := newApplication(nil)
				a.ready = true
				a.quitting, a.shutdownReady = !ready, ready
				a.pending = []string{"queued.torrent"}
				a.post = func(func()) { t.Error("open command posted work during shutdown") }
				a.showAdd = func([]string) { t.Error("open command showed an import window during shutdown") }
				if closedDraft {
					a.add = &addWindow{windowBase: &windowBase{status: "unchanged"}}
					a.add.close()
				}
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("open command panicked during shutdown: %v", r)
					}
				}()
				a.command("open") // The menu, shortcut and command palette share this path.
				if !slices.Equal(a.pending, []string{"queued.torrent"}) {
					t.Fatal("open command changed the pending FIFO")
				}
				if a.add != nil && (a.add.job != "" || a.add.status != "unchanged" || a.add.task.Running()) {
					t.Fatal("open command started work on a closed import window")
				}
			})
		}
	}
}
