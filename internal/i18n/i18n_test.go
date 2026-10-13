package i18n

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
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
	for code, catalog := range catalogs {
		if len(catalog) != len(catalogs["en"]) {
			t.Errorf("%s: catalog size differs", code)
		}
		for key := range catalogs["en"] {
			text, ok := catalog[key]
			if !ok || strings.TrimSpace(text) == "" {
				t.Errorf("%s missing %q", code, key)
				continue
			}
			if !reflect.DeepEqual(format.FindAllString(key, -1), format.FindAllString(text, -1)) {
				t.Errorf("%s: format mismatch for %q: %q", code, key, text)
			}
		}
	}
	if got := Translate("unsupported", "保存设置"); got != "Save Settings" {
		t.Fatal(got)
	}
	if got := Translate("en", "external diagnostic"); got != "external diagnostic" {
		t.Fatal(got)
	}
}

// Check application call sites as well as catalogs so an untranslated new UI
// message fails tests even when all five existing catalogs agree with each other.
func TestMessageCoverage(t *testing.T) {
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
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			name := ""
			argument := 0
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				name = fn.Name
			case *ast.SelectorExpr:
				if x, ok := fn.X.(*ast.Ident); ok {
					if x.Name == "i18n" {
						name = fn.Sel.Name
					}
					if x.Name == "problem" && fn.Sel.Name == "Message" {
						name = "T"
						argument = 1
					}
				}
			}
			if name != "tr" && name != "T" {
				return true
			}
			if len(call.Args) <= argument {
				return true
			}
			literal, ok := call.Args[argument].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			key, _ := strconv.Unquote(literal.Value)
			if _, ok := catalogs["en"][key]; !ok {
				t.Errorf("%s: missing message %q", path, key)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
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
				if T("保存设置") == "" {
					t.Error("empty translation")
				}
			}
		})
	}
	workers.Wait()
	SetLanguage("ja", "en-US")
	if got := fmt.Sprintf(T("约 %d 秒"), 5); got != "約 5 秒" {
		t.Fatal(got)
	}
}
