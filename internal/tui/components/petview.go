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

// RenderInfo returns a status summary string.
func (pv *PetView) RenderInfo() string {
	p := pv.pet
	stageName := p.StageID
	if stage := pv.registry.GetStage(p.Species, p.StageID); stage != nil {
		stageName = stage.Name
	}

	var b strings.Builder
	b.WriteString(p.Name)
	b.WriteString(" — ")
	b.WriteString(stageName)
	b.WriteString("\n")
	b.WriteString(moodEmoji(p.MoodName()))
	b.WriteString(" ")
	b.WriteString(moodChinese(p.MoodName()))
	return b.String()
}

func (pv *PetView) fallbackArt() string {
	return "  ?\n ?\n  ?"
}

func moodEmoji(mood string) string {
	switch mood {
	case "happy":
		return "😊"
	case "normal":
		return "😐"
	case "unhappy":
		return "😕"
	case "sad":
		return "😢"
	case "miserable":
		return "😭"
	default:
		return "❓"
	}
}

func moodChinese(mood string) string {
	switch mood {
	case "happy":
		return "开心"
	case "normal":
		return "普通"
	case "unhappy":
		return "不太好"
	case "sad":
		return "难过"
	case "miserable":
		return "很差"
	default:
		return "未知"
	}
}
