package torrent

import (
	"bytes"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ntnyq/rivu/internal/session"
)

// The pinned engine exposes HTTP/UDP announce state through WriteStatus, not
// StatusUpdated (those callbacks cover WebSocket trackers only). Keep this
// version-sensitive parser isolated and never expose its raw text or credentials.
var announceRow = regexp.MustCompile(`^\s*(\S+)\s+([0-9a-f]{40})\s+(true|false)\s+(true|false)\s+(\S+)\s+.*last ann: (.*)$`)

func parseDiagnostics(status string, now time.Time) map[string][]session.TrackerStatus {
	result := map[string][]session.TrackerStatus{}
	inTable := false
	for _, line := range strings.Split(status, "\n") {
		if strings.TrimSpace(line) == "Next announces:" {
			inTable = true
			continue
		}
		if strings.TrimSpace(line) == "Next announces" {
			break
		}
		if !inTable {
			continue
		}
		match := announceRow.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		raw, hash, active, wait, last := match[1], match[2], match[3], match[5], strings.TrimSpace(match[6])
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" {
			continue
		}
		item := session.TrackerStatus{Host: RedactTracker(raw), State: "waiting"}
		if active == "true" {
			item.State = "announcing"
		} else if strings.HasSuffix(last, " peers") {
			item.State = "ok"
			item.Detail = last
		} else if last != "never" {
			item.State = "failed"
			// Tracker errors can contain passkeys, URLs, and arbitrary remote text.
			// Preserve an actionable category, never their unredacted message.
			lower := strings.ToLower(last)
			switch {
			case strings.Contains(lower, "timeout"), strings.Contains(lower, "deadline"):
				item.Detail = "timeout"
			case strings.Contains(lower, "no such host"):
				item.Detail = "dns"
			case strings.Contains(lower, "refused"):
				item.Detail = "refused"
			case strings.Contains(lower, "401"), strings.Contains(lower, "403"), strings.Contains(lower, "unauthor"):
				item.Detail = "authorization"
			default:
				item.Detail = "announce"
			}
		}
		if duration, err := time.ParseDuration(wait); err == nil {
			item.Next = now.Add(max(0, duration))
		}
		result[hash] = append(result[hash], item)
	}
	return result
}

func (e *Engine) refreshDiagnostics() {
	now := time.Now()
	if now.Sub(e.diagnosticAt) < 5*time.Second {
		return
	}
	e.diagnosticAt = now
	e.diagnostics = map[string][]session.TrackerStatus{}
	var status bytes.Buffer
	if e.client != nil {
		e.client.WriteStatus(&status)
	}
	for hash, rows := range parseDiagnostics(status.String(), now) {
		e.diagnostics[hash] = rows
	}
	status.Reset()
	if e.metadata != nil {
		e.metadata.WriteStatus(&status)
	}
	for hash, rows := range parseDiagnostics(status.String(), now) {
		e.diagnostics[hash] = rows
	}
}
