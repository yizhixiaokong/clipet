package game

import (
	"clipet/internal/game/capabilities"
	"clipet/internal/plugin"
)

// LifecycleManager manages lifecycle events for pets
type LifecycleManager struct {
	pluginRegistry *plugin.Registry
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(pluginRegistry *plugin.Registry) *LifecycleManager {
	return &LifecycleManager{
		pluginRegistry: pluginRegistry,
	}
}

// CheckLifecycle checks the pet's lifecycle state
func (m *LifecycleManager) CheckLifecycle(pet *Pet) capabilities.LifecycleState {
	species := m.pluginRegistry.GetSpecies(pet.Species)
	if species == nil {
		// Fallback to defaults if species not found
		return capabilities.LifecycleState{
			NearEnd:    false,
			AgePercent: 0.0,
		}
	}

	maxAge := species.Lifecycle.MaxAgeHours
	agePercent := pet.AgeHours() / maxAge

	return capabilities.LifecycleState{
		NearEnd:    agePercent >= species.Lifecycle.WarningThreshold,
		AgePercent: agePercent,
	}
}

// TriggerEnding triggers an appropriate ending based on the pet's life
func (m *LifecycleManager) TriggerEnding(pet *Pet) capabilities.EndingResult {
	species := m.pluginRegistry.GetSpecies(pet.Species)
	if species == nil {
		return capabilities.EndingResult{
			Type:    "peaceful_rest",
			Message: "平静地度过了这一生，它已经离开了...",
		}
	}

	// Check custom endings defined in the species pack
	for _, ending := range species.Endings {
		if m.matchesEndingCondition(pet, ending.Condition) {
			return capabilities.EndingResult{
				Type:    ending.Type,
				Message: ending.Message,
			}
		}
	}

	// Default endings based on pet's life quality
	happiness := pet.MoodScore()

	switch {
	case happiness > 90 && pet.TotalInteractions > 500:
		return capabilities.EndingResult{
			Type:    "blissful_passing",
			Message: "带着满满的幸福，你的宠物安详地离开了...",
		}
	case pet.AdventuresCompleted > 30:
		return capabilities.EndingResult{
			Type:    "heroic_tale",
			Message: "它度过了充满冒险的一生，成为了传奇...",
		}
	default:
		return capabilities.EndingResult{
			Type:    "peaceful_rest",
			Message: "平静地度过了这一生，它已经离开了...",
		}
	}
}

// matchesEndingCondition checks if a pet meets the ending condition
func (m *LifecycleManager) matchesEndingCondition(pet *Pet, condition capabilities.EndingCondition) bool {
	if condition.MinHappiness > 0 && pet.MoodScore() < condition.MinHappiness {
		return false
	}
	if condition.MinAgeHours > 0 && pet.AgeHours() < condition.MinAgeHours {
		return false
	}
	if condition.MinAdventures > 0 && pet.AdventuresCompleted < condition.MinAdventures {
		return false
	}
	return true
}
