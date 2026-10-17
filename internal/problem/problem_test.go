package problem

import (
	"errors"
	"testing"

	"github.com/egoist/mygo/i18n"
	"github.com/ntnyq/rivu/locales"
)

func TestMessageDiagnosticPreservesEnglishAndCause(t *testing.T) {
	old := i18n.Locale()
	t.Cleanup(func() { i18n.SetLocaleFunc(func() string { return old }) })
	i18n.SetLocaleFunc(func() string { return "ja" })
	cause := errors.New("permission denied")
	issue := Message(Metadata, func(m locales.Messages) string { return m.TorrentReadFailed(cause) }, cause)
	if got := issue.Error(); got != "Cannot read torrent metadata: permission denied" {
		t.Fatal(got)
	}
	if !errors.Is(issue, cause) {
		t.Fatal("wrapped cause was lost")
	}
	if issue.Message == nil || issue.Message(locales.In("ja")) == issue.Error() {
		t.Fatal("deferred presentation data was lost")
	}
	i18n.SetLocaleFunc(func() string { return "zh-Hans" })
	if got := issue.Error(); got != "Cannot read torrent metadata: permission denied" {
		t.Fatal("diagnostic changed with interface language:", got)
	}
}
