package capabilities

import (
	"fmt"
)

// PluginConstraints defines safety boundaries for species packs
type PluginConstraints struct {
	// Lifecycle boundaries
	MinLifespanHours float64 `toml:"min_lifespan_hours"` // Default: 24.0 (1 day)
	MaxLifespanHours float64 `toml:"max_lifespan_hours"` // Default: 87600.0 (10 years)

	// Attribute modifier limits
	MaxAttributeMultiplier float64 `toml:"max_attr_multiplier"` // Default: 3.0 (300%)
	MinAttributeMultiplier float64 `toml:"min_attr_multiplier"` // Default: 0.1 (10%)

	// Crisis event throttling
	MaxCrisisEventsPerHour int     `toml:"max_crisis_per_hour"` // Default: 2
	MinCrisisEventInterval float64 `toml:"min_crisis_interval"` // Default: 0.5 (30 minutes)

	// Adventure frequency limits
	MaxAdventureFrequency float64 `toml:"max_adventure_freq"` // Default: 6.0 (per hour)

	// Cooldown modifier limits
	MinCooldownMultiplier float64 `toml:"min_cooldown_mult"` // Default: 0.5 (50% faster)
	MaxCooldownMultiplier float64 `toml:"max_cooldown_mult"` // Default: 2.0 (200% slower)

	// Override justification (optional)
	Reason string `toml:"reason"` // Required if overriding defaults (min 50 chars)
}

// DefaultConstraints returns sensible default values
func DefaultConstraints() PluginConstraints {
	return PluginConstraints{
		MinLifespanHours:        24.0,     // 1 day minimum
		MaxLifespanHours:        87600.0,  // 10 years maximum
		MaxAttributeMultiplier:  3.0,      // 3x maximum bonus
		MinAttributeMultiplier:  0.1,      // 10% minimum
		MaxCrisisEventsPerHour:  2,
		MinCrisisEventInterval:  0.5,      // 30 minutes
		MaxAdventureFrequency:   6.0,      // Max 6 adventures per hour
		MinCooldownMultiplier:   0.5,
		MaxCooldownMultiplier:   2.0,
	}
}

// ValidateLifecycle checks if lifecycle configuration is within safe boundaries
func (c PluginConstraints) ValidateLifecycle(lc LifecycleConfig) []string {
	var errs []string

	// Skip validation for eternal
	if lc.EndingType == "eternal" {
		return nil
	}

	if lc.MaxAgeHours < c.MinLifespanHours {
		errs = append(errs, fmt.Sprintf("max_age_hours %.1f is below minimum %.1f",
			lc.MaxAgeHours, c.MinLifespanHours))
	}
	if lc.MaxAgeHours > c.MaxLifespanHours {
		errs = append(errs, fmt.Sprintf("max_age_hours %.1f exceeds maximum %.1f",
			lc.MaxAgeHours, c.MaxLifespanHours))
	}

	return errs
}

// ValidatePassiveEffect checks attribute modifiers
func (c PluginConstraints) ValidatePassiveEffect(effect PassiveEffect) []string {
	var errs []string

	multipliers := []struct {
		name string
		val  float64
	}{
		{"feed_hunger_bonus", effect.FeedHungerBonus},
		{"feed_happiness_bonus", effect.FeedHappinessBonus},
		{"play_happiness_bonus", effect.PlayHappinessBonus},
		{"sleep_energy_bonus", effect.SleepEnergyBonus},
	}

	for _, m := range multipliers {
		if m.val != 0 {
			mult := 1.0 + m.val
			if mult < c.MinAttributeMultiplier || mult > c.MaxAttributeMultiplier {
				errs = append(errs, fmt.Sprintf("%s %.2f is outside range [%.2f, %.2f]",
					m.name, mult, c.MinAttributeMultiplier, c.MaxAttributeMultiplier))
			}
		}
	}

	// Resurrection chance should be between 0.0-1.0
	if effect.ResurrectChance < 0 || effect.ResurrectChance > 1.0 {
		errs = append(errs, fmt.Sprintf("resurrect_chance %.2f is outside [0.0, 1.0]",
			effect.ResurrectChance))
	}

	return errs
}

// ValidateOverride checks if constraint override is justified
func (c PluginConstraints) ValidateOverride() []string {
	var errs []string

	// If any constraint is overridden from defaults, require justification
	defaults := DefaultConstraints()
	hasOverride := c.MinLifespanHours != defaults.MinLifespanHours ||
		c.MaxLifespanHours != defaults.MaxLifespanHours ||
		c.MaxAttributeMultiplier != defaults.MaxAttributeMultiplier ||
		c.MinAttributeMultiplier != defaults.MinAttributeMultiplier ||
		c.MaxCrisisEventsPerHour != defaults.MaxCrisisEventsPerHour ||
		c.MinCrisisEventInterval != defaults.MinCrisisEventInterval ||
		c.MaxAdventureFrequency != defaults.MaxAdventureFrequency ||
		c.MinCooldownMultiplier != defaults.MinCooldownMultiplier ||
		c.MaxCooldownMultiplier != defaults.MaxCooldownMultiplier

	if hasOverride {
		if len(c.Reason) < 50 {
			errs = append(errs, "constraint overrides require a 'reason' field with at least 50 characters")
		}
	}

	return errs
}
