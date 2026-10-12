package updater

import (
	"github.com/ntnyq/rivu/internal/config"
	"testing"
	"time"
)

func TestConsentAndDailyBackoff(t *testing.T) {
	p := config.Default()
	now := time.Now()
	if Due(p, now) {
		t.Fatal("checks without consent")
	}
	p.AutoCheck = true
	if !Due(p, now) {
		t.Fatal("initial check blocked")
	}
	p.LastAttempt = now.Add(-time.Hour)
	if Due(p, now) {
		t.Fatal("failure backoff skipped")
	}
	p.LastAttempt = time.Time{}
	p.LastCheck = now
	if Due(p, now) {
		t.Fatal("daily limit skipped")
	}
}
