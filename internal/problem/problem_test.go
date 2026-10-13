package problem

import (
	"errors"
	"testing"

	"github.com/ntnyq/rivu/internal/i18n"
)

func TestMessageDiagnosticPreservesEnglishAndCause(t *testing.T) {
	old := i18n.Language()
	t.Cleanup(func() { i18n.SetLanguage(old, "") })
	i18n.SetLanguage("ja", "")
	cause := errors.New("permission denied")
	issue := Message(Metadata, "torrent.readFailed", cause)
	if got := issue.Error(); got != "Cannot read torrent metadata: permission denied" {
		t.Fatal(got)
	}
	if !errors.Is(issue, cause) {
		t.Fatal("wrapped cause was lost")
	}
	if issue.Key != "torrent.readFailed" || len(issue.Args) != 1 {
		t.Fatal("deferred presentation data was lost")
	}
	i18n.SetLanguage("zh-Hans", "")
	if got := issue.Error(); got != "Cannot read torrent metadata: permission denied" {
		t.Fatal("diagnostic changed with interface language:", got)
	}
}
