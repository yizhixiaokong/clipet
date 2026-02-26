// Package components provides reusable TUI components.
package components

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"strings"

	"github.com/mattn/go-runewidth"
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

// Render returns the current ASCII art frame as a normalized rectangular block.
func (pv *PetView) Render() string {
	animState := string(pv.pet.CurrentAnimation)

	frame := pv.registry.GetFrames(pv.pet.Species, pv.pet.StageID, animState)
	if frame == nil || len(frame.Frames) == 0 {
		return pv.fallbackArt()
	}

	idx := pv.frameIndex % len(frame.Frames)
	raw := strings.TrimRight(frame.Frames[idx], "\n")
	return normalizeArt(raw, frame.Width)
}

func (pv *PetView) fallbackArt() string {
	return "  ?\n ?\n  ?"
}

// normalizeArt pads every line to the same display width so the art forms
// a rectangular block that won't be skewed by per-line centering.
func normalizeArt(art string, minWidth int) string {
	lines := strings.Split(art, "\n")
	maxW := minWidth
	for _, l := range lines {
		w := displayWidth(l)
		if w > maxW {
			maxW = w
		}
	}
	for i, l := range lines {
		w := displayWidth(l)
		if w < maxW {
			lines[i] = l + strings.Repeat(" ", maxW-w)
		}
	}
	return strings.Join(lines, "\n")
}

// displayWidth returns the visible column width of a string,
// accounting for multi-byte UTF-8 characters and wide glyphs.
func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}
