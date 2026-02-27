package capabilities

import (
	"fmt"
	"sync"
)

// Registry manages personality traits for all species
type Registry struct {
	mu     sync.RWMutex
	traits map[string]map[string]PersonalityTrait // species_id -> trait_id -> trait
}

// NewRegistry creates a new capability registry
func NewRegistry() *Registry {
	return &Registry{
		traits: make(map[string]map[string]PersonalityTrait),
	}
}

// RegisterTraits registers all traits for a species
func (r *Registry) RegisterTraits(speciesID string, traits []PersonalityTrait) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.traits[speciesID]; !exists {
		r.traits[speciesID] = make(map[string]PersonalityTrait)
	}

	for _, trait := range traits {
		if _, exists := r.traits[speciesID][trait.ID]; exists {
			return fmt.Errorf("duplicate trait ID %q for species %q", trait.ID, speciesID)
		}
		r.traits[speciesID][trait.ID] = trait
	}

	return nil
}

// GetTrait retrieves a specific trait for a species
func (r *Registry) GetTrait(speciesID, traitID string) (PersonalityTrait, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if speciesTraits, exists := r.traits[speciesID]; exists {
		trait, ok := speciesTraits[traitID]
		return trait, ok
	}
	return PersonalityTrait{}, false
}

// GetAllTraits retrieves all traits for a species
func (r *Registry) GetAllTraits(speciesID string) []PersonalityTrait {
	r.mu.RLock()
	defer r.mu.RUnlock()

	speciesTraits, exists := r.traits[speciesID]
	if !exists {
		return nil
	}

	traits := make([]PersonalityTrait, 0, len(speciesTraits))
	for _, trait := range speciesTraits {
		traits = append(traits, trait)
	}
	return traits
}

// ApplyPassiveEffects applies all passive trait effects to a game action
// Returns modified hunger, happiness, health, energy values
func (r *Registry) ApplyPassiveEffects(speciesID string, action string, hunger, happiness, health, energy int) (int, int, int, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	speciesTraits, exists := r.traits[speciesID]
	if !exists {
		return hunger, happiness, health, energy
	}

	for _, trait := range speciesTraits {
		if trait.Type != "passive" || trait.PassiveEffect == nil {
			continue
		}

		effect := trait.PassiveEffect

		// Apply action-specific bonuses
		switch action {
		case "feed":
			if effect.FeedHungerBonus != 0 {
				hunger = int(float64(hunger) * (1.0 + effect.FeedHungerBonus))
			}
			if effect.FeedHappinessBonus != 0 {
				happiness = int(float64(happiness) * (1.0 + effect.FeedHappinessBonus))
			}
		case "play":
			if effect.PlayHappinessBonus != 0 {
				happiness = int(float64(happiness) * (1.0 + effect.PlayHappinessBonus))
			}
		case "sleep":
			if effect.SleepEnergyBonus != 0 {
				energy = int(float64(energy) * (1.0 + effect.SleepEnergyBonus))
			}
		}

		// Clamp values to valid range
		hunger = clamp(hunger, 0, 100)
		happiness = clamp(happiness, 0, 100)
		health = clamp(health, 0, 100)
		energy = clamp(energy, 0, 100)
	}

	return hunger, happiness, health, energy
}

// GetEvolutionModifier returns the combined evolution modifier for a species
func (r *Registry) GetEvolutionModifier(speciesID string) *EvolutionModifier {
	r.mu.RLock()
	defer r.mu.RUnlock()

	speciesTraits, exists := r.traits[speciesID]
	if !exists {
		return nil
	}

	// Combine all modifier traits
	var combined EvolutionModifier
	hasModifier := false

	for _, trait := range speciesTraits {
		if trait.Type != "modifier" || trait.EvolutionModifier == nil {
			continue
		}

		hasModifier = true
		mod := trait.EvolutionModifier

		if mod.NightInteractionBonus > combined.NightInteractionBonus {
			combined.NightInteractionBonus = mod.NightInteractionBonus
		}
		if mod.DayInteractionBonus > combined.DayInteractionBonus {
			combined.DayInteractionBonus = mod.DayInteractionBonus
		}
		if mod.FeedBonus > combined.FeedBonus {
			combined.FeedBonus = mod.FeedBonus
		}
		if mod.PlayBonus > combined.PlayBonus {
			combined.PlayBonus = mod.PlayBonus
		}
		if mod.AdventureBonus > combined.AdventureBonus {
			combined.AdventureBonus = mod.AdventureBonus
		}
	}

	if !hasModifier {
		return nil
	}

	return &combined
}

// GetActiveTraits returns all active abilities for a species
func (r *Registry) GetActiveTraits(speciesID string) []PersonalityTrait {
	r.mu.RLock()
	defer r.mu.RUnlock()

	speciesTraits, exists := r.traits[speciesID]
	if !exists {
		return nil
	}

	var activeTraits []PersonalityTrait
	for _, trait := range speciesTraits {
		if trait.Type == "active" && trait.ActiveEffect != nil {
			activeTraits = append(activeTraits, trait)
		}
	}
	return activeTraits
}

// clamp ensures a value is within the specified range
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
