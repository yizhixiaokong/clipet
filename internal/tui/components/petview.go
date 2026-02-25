// Package components provides reusable TUI components.
package components

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"strings"
)

// PetView manages the animated ASCII display of a pet.
type PetView struct {
	pet      *game.Pet
	registry *plugin.Registry

	frameIndex int
}

// NewPetView creates a new PetView.
func NewPetView(pet *game.Pet, reg *plugin.Registry) *PetView {
	return &PetView{
		pet:      pet,
		registry: reg,
	}
}

// Tick advances the animation by one frame.
func (pv *PetView) Tick() {
	pv.frameIndex++
}

// SetPet updates the pet reference (after load/save).
func (pv *PetView) SetPet(pet *game.Pet) {
	pv.pet = pet
}

// Render returns the current ASCII art frame as a string.
func (pv *PetView) Render() string {
	animState := string(pv.pet.CurrentAnimation)

	frame := pv.registry.GetFrames(pv.pet.Species, pv.pet.StageID, animState)
	if frame == nil || len(frame.Frames) == 0 {
		return pv.fallbackArt()
	}

	idx := pv.frameIndex % len(frame.Frames)
	return strings.TrimRight(frame.Frames[idx], "\n")
}

func (pv *PetView) fallbackArt() string {
	return "  ?\n ?\n  ?"
}
