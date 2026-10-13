package torrent

import (
	"strings"
	"testing"
	"time"
)

func TestTrackerDiagnosticErrorRedaction(t *testing.T) {
	hash := strings.Repeat("a", 40)
	input := "Next announces:\n  https://user:secret@tracker.test/private?key=secret " + hash + " false false 1m0s false true true 0% false started next ann: anytime, last ann: timeout requesting secret  \nNext announces\n"
	rows := parseDiagnostics(input, time.Now())[hash]
	if len(rows) != 1 || rows[0].State != "failed" || rows[0].Detail != "timeout" || rows[0].Host != "https://tracker.test" || rows[0].Next.IsZero() {
		t.Fatalf("wrong diagnostic: %+v", rows)
	}
}
