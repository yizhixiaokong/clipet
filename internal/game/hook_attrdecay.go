package game

import (
	"clipet/internal/game/capabilities"
	"clipet/internal/plugin"
	"time"
)

// AttrDecayHook handles attribute decay over time
type AttrDecayHook struct {
	registry *plugin.Registry
}

func NewAttrDecayHook(registry *plugin.Registry) *AttrDecayHook {
	return &AttrDecayHook{registry: registry}
}

func (h *AttrDecayHook) Name() string {
	return "AttrDecay"
}

func (h *AttrDecayHook) OnTimeAdvance(elapsed time.Duration, pet *Pet) {
	// Get decay config from plugin
	var decayConfig capabilities.DecayConfig
	if h.registry != nil {
		decayConfig = h.registry.GetDecayConfig(pet.Species)
	} else {
		decayConfig = capabilities.DecayConfig{}.Defaults()
	}

	hours := elapsed.Hours()

	// Attribute decay using plugin-controlled rates
	pet.Hunger = clamp(pet.Hunger-int(decayConfig.Hunger*hours), 0, 100)
	pet.Happiness = clamp(pet.Happiness-int(decayConfig.Happiness*hours), 0, 100)
	pet.Energy = clamp(pet.Energy-int(decayConfig.Energy*hours), 0, 100)

	// Health decay due to hunger
	if pet.Hunger < 20 {
		pet.Health = clamp(pet.Health-int(decayConfig.Health*hours), 0, 100)
	}
}
