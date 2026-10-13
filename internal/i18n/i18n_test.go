package i18n

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestResolve(t *testing.T) {
	for _, tc := range []struct{ preference, system, want string }{
		{"system", "zh-CN", "zh-Hans"}, {"system", "zh_SG.UTF-8", "zh-Hans"},
		{"system", "zh-TW", "zh-Hant"}, {"system", "zh-HK", "zh-Hant"}, {"system", "zh-MO", "zh-Hant"},
		{"system", "zh-Hant-CN", "zh-Hant"}, {"system", "zh-Hans-TW", "zh-Hans"}, {"system", "zh", "zh-Hans"},
		{"system", "en-GB", "en"}, {"system", "ja-JP", "ja"}, {"system", "ko_KR", "ko"},
		{"system", "de-DE", "en"}, {"system", "", "en"}, {"system", "C", "en"},
		{"ja", "zh-CN", "ja"}, {"zh-Hant", "en-US", "zh-Hant"},
	} {
		if got := Resolve(tc.preference, tc.system); got != tc.want {
			t.Errorf("Resolve(%q, %q) = %q; want %q", tc.preference, tc.system, got, tc.want)
		}
	}
}

func TestCatalogs(t *testing.T) {
	format := regexp.MustCompile(`%[+#\- .0-9]*[a-zA-Z%]`)
	keyPattern := regexp.MustCompile(`^[a-z][a-zA-Z0-9]*(\.[a-z][a-zA-Z0-9]*)+$`)
	for code, catalog := range catalogs {
		for key := range catalog {
			if !keyPattern.MatchString(key) {
				t.Errorf("%s: invalid semantic key: %q", code, key)
			}
			if _, ok := catalogs["en"][key]; !ok {
				t.Errorf("%s: unknown key: %q", code, key)
			}
		}
		if len(catalog) != len(catalogs["en"]) {
			t.Errorf("%s: catalog size differs", code)
		}
		for key, english := range catalogs["en"] {
			text, ok := catalog[key]
			if !ok || strings.TrimSpace(text) == "" {
				t.Errorf("%s missing %q", code, key)
				continue
			}
			if !reflect.DeepEqual(format.FindAllString(english, -1), format.FindAllString(text, -1)) {
				t.Errorf("%s: format mismatch for %q: %q", code, key, text)
			}
		}
	}
	if got := Translate("unsupported", "menu.pauseAll"); got != "Pause All" {
		t.Fatal(got)
	}
	if got := Translate("en", "missing.key"); got != "Translation unavailable" {
		t.Fatal(got)
	}
}

func TestMissingTranslationFallsBackToEnglish(t *testing.T) {
	const key = "settings.save"
	original := catalogs["ja"][key]
	t.Cleanup(func() { catalogs["ja"][key] = original })
	delete(catalogs["ja"], key)
	if got := Translate("ja", key); got != "Save Settings" {
		t.Fatal(got)
	}
	catalogs["ja"][key] = ""
	if got := Translate("ja", key); got != "Save Settings" {
		t.Fatal(got)
	}
}

// Check application call sites as well as catalogs, including deferred errors.
// Dynamic forwarding is confined to these small, audited presentation adapters;
// new dynamic calls must not silently bypass reference checking.
func messageReferences(path string, source any) ([]string, []string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
	if err != nil {
		return nil, nil, err
	}
	imports := map[string]string{}
	for _, spec := range file.Imports {
		name, _ := strconv.Unquote(spec.Path.Value)
		alias := filepath.Base(name)
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		imports[alias] = name
	}
	var keys, diagnostics []string
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		function := ""
		if ok {
			function = fn.Name.Name
		}
		ast.Inspect(decl, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			argument := -1
			forwarding := false
			switch callee := call.Fun.(type) {
			case *ast.Ident:
				if file.Name.Name == "main" && callee.Name == "tr" {
					argument = 0
					if path == "recovery_actions.go" && function == "problemText" && len(call.Args) == 1 {
						if selector, ok := call.Args[0].(*ast.SelectorExpr); ok {
							owner, ok := selector.X.(*ast.Ident)
							forwarding = ok && owner.Name == "issue" && selector.Sel.Name == "Key"
						}
					}
				}
				if file.Name.Name == "i18n" && callee.Name == "Translate" {
					argument = 1
					if path == "internal/i18n/i18n.go" && function == "T" && len(call.Args) == 2 {
						key, ok := call.Args[1].(*ast.Ident)
						forwarding = ok && key.Name == "key"
					}
				}
			case *ast.SelectorExpr:
				owner, ok := callee.X.(*ast.Ident)
				if !ok {
					break
				}
				switch imports[owner.Name] {
				case "github.com/ntnyq/rivu/internal/i18n":
					switch callee.Sel.Name {
					case "T":
						argument = 0
						if path == "localization.go" && function == "tr" && len(call.Args) == 1 {
							key, ok := call.Args[0].(*ast.Ident)
							forwarding = ok && key.Name == "key"
						}
					case "Translate":
						argument = 1
						if path == "internal/problem/problem.go" && function == "Error" && len(call.Args) == 2 {
							if selector, ok := call.Args[1].(*ast.SelectorExpr); ok {
								owner, ok := selector.X.(*ast.Ident)
								forwarding = ok && owner.Name == "e" && selector.Sel.Name == "Key"
							}
						}
					}
				case "github.com/ntnyq/rivu/internal/problem":
					if callee.Sel.Name == "Message" {
						argument = 1
					}
				}
			}
			if argument < 0 || argument >= len(call.Args) {
				return true
			}
			if literal, ok := call.Args[argument].(*ast.BasicLit); ok && literal.Kind == token.STRING {
				key, _ := strconv.Unquote(literal.Value)
				keys = append(keys, key)
			} else if !forwarding {
				diagnostics = append(diagnostics, "translation key must be a literal")
			}
			return true
		})
	}
	return keys, diagnostics, nil
}

func TestMessageCoverage(t *testing.T) {
	used := map[string]bool{}
	err := filepath.WalkDir("../..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "build" || entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, err := filepath.Rel("../..", path)
		if err != nil {
			return err
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		keys, diagnostics, err := messageReferences(filepath.ToSlash(relative), source)
		if err != nil {
			return err
		}
		for _, diagnostic := range diagnostics {
			t.Errorf("%s: %s", path, diagnostic)
		}
		for _, key := range keys {
			used[key] = true
			if _, ok := catalogs["en"][key]; !ok {
				t.Errorf("%s: missing message %q", path, key)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for key := range catalogs["en"] {
		if !used[key] {
			t.Errorf("unused message %q", key)
		}
	}
}

func TestMessageReferences(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		keys         []string
		diagnostics  int
	}{
		{"literal", `package main; func view() { tr("settings.save") }`, []string{"settings.save"}, 0},
		{"dynamic", `package main; func view(key string) { tr(key) }`, nil, 1},
		{"alias", `package view; import messages "github.com/ntnyq/rivu/internal/i18n"; func view() { messages.T("settings.save"); messages.Translate("en", "menu.settings") }`, []string{"settings.save", "menu.settings"}, 0},
		{"deferred", `package engine; import p "github.com/ntnyq/rivu/internal/problem"; func run(key string) { p.Message(p.Engine, "torrent.invalidIdentity"); p.Message(p.Engine, key) }`, []string{"torrent.invalidIdentity"}, 1},
		{"unrelated", `package main; func view() { other.T("not a key") }`, nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			keys, diagnostics, err := messageReferences("example.go", tc.source)
			if err != nil || !reflect.DeepEqual(keys, tc.keys) || len(diagnostics) != tc.diagnostics {
				t.Fatalf("keys=%v diagnostics=%v err=%v", keys, diagnostics, err)
			}
		})
	}
}

func TestConcurrentSwitching(t *testing.T) {
	old := Language()
	defer SetLanguage(old, "")
	var workers sync.WaitGroup
	for code := range catalogs {
		workers.Go(func() {
			for range 100 {
				SetLanguage(code, "")
				if T("settings.save") == "" {
					t.Error("empty translation")
				}
			}
		})
	}
	workers.Wait()
	SetLanguage("ja", "en-US")
	if got := fmt.Sprintf(T("transfer.eta.seconds"), 5); got != "約 5 秒" {
		t.Fatal(got)
	}
}
