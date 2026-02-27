package game

import (
	"clipet/internal/game/capabilities"
	"clipet/internal/plugin"
	"time"
)

// DefaultActionConfigs returns hardcoded default action configurations
// These are used when species pack doesn't define custom actions
func DefaultActionConfigs() map[string]struct {
	Cooldown   time.Duration
	EnergyCost int
	Hunger     int
	Happiness  int
	Health     int
	Energy     int
} {
	return map[string]struct {
		Cooldown   time.Duration
		EnergyCost int
		Hunger     int
		Happiness  int
		Health     int
		Energy     int
	}{
		"feed": {
			Cooldown:  CooldownFeed,
			Hunger:    25,
			Happiness: 5,
		},
		"play": {
			Cooldown:  CooldownPlay,
			Energy:    -10,
			Happiness: 20,
		},
		"rest": {
			Cooldown: CooldownRest,
			Energy:   30,
			Health:   5,
			Happiness: -5,
		},
		"heal": {
			Cooldown:  CooldownHeal,
			EnergyCost: 15,
			Health:    25,
		},
		"talk": {
			Cooldown: CooldownTalk,
			Happiness: 5,
		},
	}
}

// GetActionCooldown returns the cooldown for an action from plugin or defaults
func GetActionCooldown(registry *plugin.Registry, speciesID, actionID string) time.Duration {
	// Try to get from plugin config
	if registry != nil {
		if action := registry.GetAction(speciesID, actionID); action != nil {
			return action.Cooldown
		}
	}

	// Fallback to defaults
	defaults := DefaultActionConfigs()
	if cfg, ok := defaults[actionID]; ok {
		return cfg.Cooldown
	}

	return 0
}

// GetActionEffects returns the attribute effects for an action from plugin or defaults
func GetActionEffects(registry *plugin.Registry, speciesID, actionID string) (hunger, happiness, health, energy int) {
	// Try to get from plugin config
	if registry != nil {
		if action := registry.GetAction(speciesID, actionID); action != nil {
			return action.Effects.Hunger,
				action.Effects.Happiness,
				action.Effects.Health,
				action.Effects.Energy
		}
	}

	// Fallback to defaults
	defaults := DefaultActionConfigs()
	if cfg, ok := defaults[actionID]; ok {
		return cfg.Hunger, cfg.Happiness, cfg.Health, cfg.Energy
	}

	return 0, 0, 0, 0
}

// GetActionEnergyCost returns the energy cost for an action from plugin or defaults
func GetActionEnergyCost(registry *plugin.Registry, speciesID, actionID string) int {
	// Try to get from plugin config
	if registry != nil {
		if action := registry.GetAction(speciesID, actionID); action != nil {
			return action.EnergyCost
		}
	}

	// Fallback to defaults
	defaults := DefaultActionConfigs()
	if cfg, ok := defaults[actionID]; ok {
		return cfg.EnergyCost
	}

	return 0
}

// CalculateDynamicCooldown calculates the actual cooldown based on attribute urgency
// This implements the plugin-controlled dynamic cooldown system
func CalculateDynamicCooldown(
	registry *plugin.Registry,
	speciesID, actionID string,
	attrValue int,
) time.Duration {
	// Get base cooldown from plugin or defaults
	baseCooldown := GetActionCooldown(registry, speciesID, actionID)

	// Get dynamic cooldown config from plugin
	var dcc capabilities.DynamicCooldownConfig
	if registry != nil {
		dcc = registry.GetDynamicCooldownConfig(speciesID)
	} else {
		dcc = capabilities.DynamicCooldownConfig{}.Defaults()
	}

	// Apply multiplier based on urgency
	multiplier := dcc.GetMultiplier(attrValue)
	return time.Duration(float64(baseCooldown) * multiplier)
}
