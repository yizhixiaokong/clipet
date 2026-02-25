package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"clipet/internal/game"
)

// JSONStore implements Store using a JSON file.
type JSONStore struct {
	path string
}

// NewJSONStore creates a new JSONStore.
// If dir is empty, it defaults to ~/.local/share/clipet/.
func NewJSONStore(dir string) (*JSONStore, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		dir = filepath.Join(home, ".local", "share", "clipet")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	return &JSONStore{
		path: filepath.Join(dir, "save.json"),
	}, nil
}

// Save writes the pet state to a JSON file atomically.
func (s *JSONStore) Save(pet *game.Pet) error {
	data, err := json.MarshalIndent(pet, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal pet: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("rename save file: %w", err)
	}

	return nil
}

// Load reads the pet state from the JSON file.
func (s *JSONStore) Load() (*game.Pet, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read save file: %w", err)
	}

	var pet game.Pet
	if err := json.Unmarshal(data, &pet); err != nil {
		return nil, fmt.Errorf("unmarshal pet: %w", err)
	}

	return &pet, nil
}

// Exists returns true if the save file exists.
func (s *JSONStore) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

// Path returns the save file path.
func (s *JSONStore) Path() string {
	return s.path
}
