// Package i18n provides Rivu's application messages. Simplified Chinese source
// messages are stable catalog keys; missing translations fall back to English.
package i18n

import (
	"embed"
	"encoding/json"
	"strings"
	"sync/atomic"
)

//go:embed locales/*.json
var resources embed.FS

var catalogs = loadCatalogs()
var current atomic.Value

func loadCatalogs() map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, code := range []string{"zh-Hans", "zh-Hant", "en", "ja", "ko"} {
		data, err := resources.ReadFile("locales/" + code + ".json")
		if err != nil {
			panic(err)
		}
		var messages map[string]string
		if err := json.Unmarshal(data, &messages); err != nil {
			panic(err)
		}
		result[code] = messages
	}
	return result
}

// Valid reports whether a persisted language preference is supported.
func Valid(preference string) bool {
	return preference == "system" || catalogs[preference] != nil
}

// Resolve maps the native locale to a supported language. Explicit script tags
// take precedence over regions (for example zh-Hans-TW remains simplified).
func Resolve(preference, systemLocale string) string {
	if preference != "system" && catalogs[preference] != nil {
		return preference
	}
	locale := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(systemLocale), "_", "-"))
	locale = strings.SplitN(strings.SplitN(locale, ".", 2)[0], "@", 2)[0]
	parts := strings.Split(locale, "-")
	switch parts[0] {
	case "en", "ja", "ko":
		return parts[0]
	case "zh":
		for _, part := range parts[1:] {
			if part == "hans" {
				return "zh-Hans"
			}
			if part == "hant" {
				return "zh-Hant"
			}
		}
		for _, part := range parts[1:] {
			if part == "tw" || part == "hk" || part == "mo" {
				return "zh-Hant"
			}
		}
		return "zh-Hans"
	default:
		return "en"
	}
}

// SetLanguage is safe while background tasks are generating notifications.
// The caller reads the native system locale on the UI thread.
func SetLanguage(preference, systemLocale string) {
	current.Store(Resolve(preference, systemLocale))
}

func Language() string {
	if code, ok := current.Load().(string); ok {
		return code
	}
	return "en"
}

func T(message string) string { return Translate(Language(), message) }

func Translate(language, message string) string {
	if text := catalogs[language][message]; text != "" {
		return text
	}
	if text := catalogs["en"][message]; text != "" {
		return text
	}
	return message
}
