package game

import (
	"clipet/internal/game/capabilities"
	"math/rand"
	"time"
)

// DeathCheckHook handles pet death checking and resurrection
type DeathCheckHook struct {
	capabilitiesReg *capabilities.Registry
}

func NewDeathCheckHook(capReg *capabilities.Registry) *DeathCheckHook {
	return &DeathCheckHook{
		capabilitiesReg: capReg,
	}
}

func (h *DeathCheckHook) Name() string {
	return "DeathCheck"
}

func (h *DeathCheckHook) OnTimeAdvance(elapsed time.Duration, pet *Pet) {
	// Only check if health drops to 0 or below
	if pet.Health > 0 {
		return
	}

	// Attempt resurrection if capabilities registry is available
	if h.capabilitiesReg != nil && h.attemptResurrection(pet) {
		// Resurrection succeeded, pet stays alive
		return
	}

	// No resurrection or failed, mark pet as dead
	pet.Alive = false
}

// attemptResurrection tries to resurrect the pet based on passive traits
// Returns true if resurrection succeeded
func (h *DeathCheckHook) attemptResurrection(pet *Pet) bool {
	traits := h.capabilitiesReg.GetAllTraits(pet.Species)
	if len(traits) == 0 {
		return false
	}

	// Check for resurrection traits
	for _, trait := range traits {
		if trait.Type != "passive" || trait.PassiveEffect == nil {
			continue
		}

		effect := trait.PassiveEffect
		if effect.ResurrectChance <= 0 {
			continue
		}

		// Roll for resurrection
		if rand.Float64() < effect.ResurrectChance {
			// Resurrection succeeded!
			restorePercent := effect.HealthRestorePercent
			if restorePercent <= 0 {
				restorePercent = 10.0 // Default 10% if not specified
			}

			// Restore health to percentage of max (100)
			pet.Health = int(restorePercent)
			if pet.Health < 1 {
				pet.Health = 1 // At least 1 HP
			}

			// Log resurrection
			// Note: In full implementation, this should emit a message to UI
			return true
		}
	}

	return false
}
