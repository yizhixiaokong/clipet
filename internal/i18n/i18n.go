// Package i18n provides internationalization support for clipet.
package i18n

import (
	"fmt"
	"log"
	"sync"
	"text/template"
)

// Manager manages translations and language settings.
type Manager struct {
	mu       sync.RWMutex
	language string
	bundle   *Bundle
	fallback string
}

// NewManager creates a new i18n manager.
func NewManager(language, fallback string, bundle *Bundle) *Manager {
	return &Manager{
		language: language,
		bundle:   bundle,
		fallback: fallback,
	}
}

// T translates a key with the given template variables.
// Variables are passed as alternating key-value pairs:
//
//	i18n.T("ui.home.feed_success", "oldHunger", 50, "newHunger", 75)
func (m *Manager) T(key string, args ...interface{}) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Parse key-value pairs
	vars := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := fmt.Sprintf("%v", args[i])
			vars[key] = args[i+1]
		}
	}

	// Try primary language
	translation, err := m.bundle.Lookup(m.language, key, vars)
	if err == nil {
		return translation
	}

	// Try fallback language
	if m.fallback != m.language {
		translation, err = m.bundle.Lookup(m.fallback, key, vars)
		if err == nil {
			return translation
		}
	}

	// Log missing translation
	log.Printf("[i18n] Missing translation for key: %s (lang: %s)", key, m.language)

	// Return key as fallback
	return key
}

// TN translates a key with plural support based on count.
func (m *Manager) TN(key string, count int, args ...interface{}) string {
	// Add count to variables
	args = append(args, "count", count)

	// Select plural form based on count
	var pluralKey string
	if count == 1 {
		pluralKey = key + "_one"
	} else {
		pluralKey = key + "_other"
	}

	return m.T(pluralKey, args...)
}

// SetLanguage changes the active language at runtime.
func (m *Manager) SetLanguage(lang string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.language = lang
}

// GetLanguage returns the current active language.
func (m *Manager) GetLanguage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.language
}

// compileTemplate compiles a translation string with variables.
func compileTemplate(text string, vars map[string]interface{}) (string, error) {
	tmpl, err := template.New("translation").Parse(text)
	if err != nil {
		return "", err
	}

	var result string
	writer := &stringWriter{&result}
	err = tmpl.Execute(writer, vars)
	return result, err
}

// stringWriter is a simple io.Writer wrapper for strings.
type stringWriter struct {
	s *string
}

func (w *stringWriter) Write(p []byte) (n int, err error) {
	*w.s += string(p)
	return len(p), nil
}
