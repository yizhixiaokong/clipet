package i18n

import (
	"encoding/json"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// Loader loads translation files from filesystem.
type Loader struct {
	fsys embed.FS
	root string
}

// NewLoader creates a new translation loader.
func NewLoader(fsys embed.FS, root string) *Loader {
	return &Loader{
		fsys: fsys,
		root: root,
	}
}

// LoadAll loads all translations from the locales directory.
func (l *Loader) LoadAll(bundle *Bundle) error {
	// Walk the locales directory
	err := fs.WalkDir(l.fsys, l.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process JSON files
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Extract language code from path (e.g., "locales/zh-CN/tui.json" -> "zh-CN")
		relPath := strings.TrimPrefix(path, l.root+"/")
		parts := strings.Split(relPath, "/")
		if len(parts) < 2 {
			return nil
		}
		lang := parts[0]

		// Read file
		data, err := fs.ReadFile(l.fsys, path)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", path, err)
		}

		// Parse and merge into bundle
		var translations map[string]interface{}
		if err := parseJSON(data, &translations); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %w", path, err)
		}

		// Merge with existing translations for this language
		if existing, ok := bundle.translations[lang]; ok {
			merged := mergeMaps(existing, translations)
			bundle.translations[lang] = merged
		} else {
			bundle.translations[lang] = translations
		}

		return nil
	})

	return err
}

// LoadLanguage loads translations for a specific language.
func (l *Loader) LoadLanguage(bundle *Bundle, lang string) error {
	langDir := filepath.Join(l.root, lang)

	// Check if directory exists
	entries, err := fs.ReadDir(l.fsys, langDir)
	if err != nil {
		return fmt.Errorf("language directory not found: %s", lang)
	}

	// Load all JSON files in the language directory
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(langDir, entry.Name())
		data, err := fs.ReadFile(l.fsys, path)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", path, err)
		}

		if err := bundle.LoadFromJSON(lang, data); err != nil {
			return err
		}
	}

	return nil
}

// parseJSON is a helper to parse JSON bytes.
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// mergeMaps recursively merges two maps.
func mergeMaps(base, overlay map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base map
	for k, v := range base {
		result[k] = v
	}

	// Merge overlay
	for k, v := range overlay {
		if existing, ok := result[k]; ok {
			// If both are maps, merge recursively
			if existingMap, ok1 := existing.(map[string]interface{}); ok1 {
				if newMap, ok2 := v.(map[string]interface{}); ok2 {
					result[k] = mergeMaps(existingMap, newMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}
