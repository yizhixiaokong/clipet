// Package store provides persistence for pet state.
package store

import "clipet/internal/game"

// Store defines the interface for pet state persistence.
type Store interface {
	// Save persists the pet state.
	Save(pet *game.Pet) error
	// Load reads the pet state from storage.
	Load() (*game.Pet, error)
	// Exists returns true if a save file exists.
	Exists() bool
}
