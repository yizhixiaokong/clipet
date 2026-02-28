package i18n

import (
	"encoding/json"
	"fmt"
)

// Bundle manages translation files for multiple languages.
type Bundle struct {
	translations map[string]map[string]interface{} // lang -> key -> value
}

// NewBundle creates a new translation bundle.
func NewBundle() *Bundle {
	return &Bundle{
		translations: make(map[string]map[string]interface{}),
	}
}

// AddLanguage adds translations for a language.
func (b *Bundle) AddLanguage(lang string, data map[string]interface{}) {
	b.translations[lang] = data
}

// LoadFromJSON loads translations from JSON bytes.
func (b *Bundle) LoadFromJSON(lang string, data []byte) error {
	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse translations for %s: %w", lang, err)
	}

	b.AddLanguage(lang, translations)
	return nil
}

// Lookup finds a translation by key and executes it with variables.
func (b *Bundle) Lookup(lang, key string, vars map[string]interface{}) (string, error) {
	langData, ok := b.translations[lang]
	if !ok {
		return "", fmt.Errorf("language not found: %s", lang)
	}

	// Navigate nested keys (e.g., "ui.home.feed_success")
	value := lookupNested(langData, key)
	if value == nil {
		return "", fmt.Errorf("translation key not found: %s", key)
	}

	// Convert to string
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("translation value is not a string: %s", key)
	}

	// Compile template with variables
	return compileTemplate(text, vars)
}

// lookupNested navigates a nested map using dot notation.
func lookupNested(data map[string]interface{}, key string) interface{} {
	keys := splitKey(key)
	current := interface{}(data)

	for _, k := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[k]
			if !ok {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

// splitKey splits a key by dots, respecting escaped dots.
func splitKey(key string) []string {
	// Simple implementation: split by dots
	// TODO: Handle escaped dots if needed
	var result []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			result = append(result, key[start:i])
			start = i + 1
		}
	}
	result = append(result, key[start:])
	return result
}
