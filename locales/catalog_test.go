package locales

import (
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/i18n"
)

func TestCatalogCompleteness(t *testing.T) {
	want := []string{"en", "ja", "ko", "zh-Hans", "zh-Hant"}
	if !slices.Equal(Catalog.Languages(), want) {
		t.Fatalf("languages = %v", Catalog.Languages())
	}
	read := func(tag string) map[string]json.RawMessage {
		t.Helper()
		data, err := files.ReadFile(tag + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var messages map[string]json.RawMessage
		if err := json.Unmarshal(data, &messages); err != nil {
			t.Fatal(err)
		}
		return messages
	}
	source := read("en")
	keyPattern := regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)
	for _, tag := range want {
		messages := read(tag)
		for key := range source {
			if strings.HasPrefix(key, "@") {
				continue
			}
			var text string
			if err := json.Unmarshal(messages[key], &text); err != nil || strings.TrimSpace(text) == "" {
				t.Errorf("%s is missing %s", tag, key)
			}
		}
		for key := range messages {
			if strings.HasPrefix(key, "@") {
				continue
			}
			if !strings.HasPrefix(key, "mygo.") && !keyPattern.MatchString(key) {
				t.Errorf("%s has invalid message key %q", tag, key)
			}
			if _, ok := source[key]; !ok {
				t.Errorf("%s has unknown message %q", tag, key)
			}
		}
	}
}

func TestLanguageMatchingAndFormatting(t *testing.T) {
	for _, tc := range []struct{ preferred, matched, settings string }{
		{"en-GB", "en", "Settings"},
		{"zh-CN", "zh-Hans", "设置"},
		{"zh_SG.UTF-8", "zh-Hans", "设置"},
		{"zh-TW", "zh-Hant", "設定"},
		{"zh-HK", "zh-Hant", "設定"},
		{"zh-Hant-CN", "zh-Hant", "設定"},
		{"zh-Hans-TW", "zh-Hans", "设置"},
		{"ja-JP", "ja", "設定"},
		{"ko_KR", "ko", "설정"},
	} {
		if got := i18n.Match([]string{tc.preferred}, Catalog.Languages()); got != tc.matched {
			t.Errorf("%s matched %s", tc.preferred, got)
		}
		if got := In(tc.preferred).SettingsLabel(); got != tc.settings {
			t.Errorf("%s settings = %s", tc.preferred, got)
		}
	}
	if got := In("de").MenuPauseAll(); got != "Pause All" {
		t.Fatal("fallback:", got)
	}
	if got := In("en").TransferEtaHours(2, 3); got != "About 2 hr 3 min" {
		t.Fatal("ordered arguments:", got)
	}
	if got := In("ja").TransferEtaSeconds(5); got != "約 5 秒" {
		t.Fatal("Japanese argument:", got)
	}
	if got := In("en").SeedingRatioProgress(1.236, 2); got != "Ratio 1.24 / 2.0" {
		t.Fatal("precision:", got)
	}
	if got := In("en").TransferCount(1234); got != "Transfers: 1234" {
		t.Fatal("ungrouped count:", got)
	}
	if got := In("en").SettingsLoginStatus(true); got != "Enabled in system: true" {
		t.Fatal("boolean:", got)
	}
	for _, tag := range Catalog.Languages() {
		if got := In(tag).FilesOwnershipChanged("资料/{file}%s.txt"); !strings.Contains(got, "资料/{file}%s.txt") {
			t.Errorf("%s path was interpreted as a template: %s", tag, got)
		}
		if got := In(tag).TorrentReadFailed(errors.New("permission denied")); !strings.Contains(got, "permission denied") {
			t.Errorf("%s lost error details: %s", tag, got)
		}
	}
}

// Typed call sites are checked by the compiler; retain the application's
// unused-message check for both direct accessors and deferred Messages methods.
func TestMessageReferences(t *testing.T) {
	used := map[string]bool{}
	err := filepath.WalkDir("..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".mygo", "build", "node_modules", "locales":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok {
				used[selector.Sel.Name] = true
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := files.ReadFile("en.json")
	if err != nil {
		t.Fatal(err)
	}
	var messages map[string]json.RawMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		t.Fatal(err)
	}
	for key := range messages {
		if strings.HasPrefix(key, "@") || strings.HasPrefix(key, "mygo.") {
			continue
		}
		name := strings.ToUpper(key[:1]) + key[1:]
		if !used[name] {
			t.Errorf("unused message %q", key)
		}
	}
}
