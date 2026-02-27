package capabilities

import "time"

// LifecycleConfig defines the lifecycle parameters for a species
type LifecycleConfig struct {
	MaxAgeHours      float64 `toml:"max_age_hours"`      // Maximum lifespan in hours
	EndingType       string  `toml:"ending_type"`        // death | ascend | eternal
	WarningThreshold float64 `toml:"warning_threshold"`  // Warning threshold (0.0-1.0)
}

// Defaults returns a LifecycleConfig with sensible defaults
func (lc LifecycleConfig) Defaults() LifecycleConfig {
	if lc.MaxAgeHours == 0 {
		lc.MaxAgeHours = 720.0 // Default: 30 days
	}
	if lc.EndingType == "" {
		lc.EndingType = "death"
	}
	if lc.WarningThreshold == 0 {
		lc.WarningThreshold = 0.8 // Warn at 80% of lifespan
	}
	return lc
}

// PassiveEffect defines a passive personality trait effect
type PassiveEffect struct {
	// Attribute modifiers (multipliers, e.g., 0.8 = 80%, 1.2 = 120%)
	FeedHungerBonus     float64 `toml:"feed_hunger_bonus"`      // Feed hunger gain multiplier
	FeedHappinessBonus  float64 `toml:"feed_happiness_bonus"`   // Feed happiness gain multiplier
	PlayHappinessBonus  float64 `toml:"play_happiness_bonus"`   // Play happiness gain multiplier
	SleepEnergyBonus    float64 `toml:"sleep_energy_bonus"`     // Sleep energy gain multiplier

	// Special effects
	ResurrectChance         float64 `toml:"resurrect_chance"`          // Chance to resurrect on death (0.0-1.0)
	HealthRestorePercent    float64 `toml:"health_restore_percent"`    // Health restored on resurrection (%)
	HealthRegenMultiplier   string  `toml:"health_regen_multiplier"`   // Expression for health regen (e.g., "magic * 0.01")
}

// ActiveEffect defines an active ability that the player can trigger
type ActiveEffect struct {
	EnergyCost     int           `toml:"energy_cost"`     // Energy cost to activate
	HealthRestore  int           `toml:"health_restore"`  // Health restored
	HungerRestore  int           `toml:"hunger_restore"`  // Hunger restored
	HappinessBoost int           `toml:"happiness_boost"` // Happiness boost
	Cooldown       time.Duration `toml:"cooldown"`        // Cooldown duration
}

// EvolutionModifier defines modifications to evolution point accumulation
type EvolutionModifier struct {
	NightInteractionBonus float64 `toml:"night_interaction_bonus"` // Night interaction points multiplier (e.g., 1.5 = +50%)
	DayInteractionBonus   float64 `toml:"day_interaction_bonus"`   // Day interaction points multiplier
	FeedBonus             float64 `toml:"feed_bonus"`              // Feed action points multiplier
	PlayBonus             float64 `toml:"play_bonus"`              // Play action points multiplier
	AdventureBonus        float64 `toml:"adventure_bonus"`         // Adventure points multiplier
}

// PersonalityTrait represents a personality characteristic (not a combat ability)
type PersonalityTrait struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Type        string `toml:"type"` // passive | active | modifier

	// Passive traits (e.g., "picky eater": -20% feed hunger, +10% feed happiness)
	PassiveEffect *PassiveEffect `toml:"passive_effect"`

	// Active skills (e.g., "purr heal": consume energy to restore health)
	ActiveEffect *ActiveEffect `toml:"active_effect"`

	// Evolution modifiers (e.g., "night owl": +50% night interaction evolution points)
	EvolutionModifier *EvolutionModifier `toml:"evolution_modifier"`
}

// EndingCondition defines when a specific ending should trigger
type EndingCondition struct {
	MinHappiness  int     `toml:"min_happiness"`   // Minimum happiness score
	MinAgeHours   float64 `toml:"min_age_hours"`   // Minimum age in hours
	MinAdventures int     `toml:"min_adventures"`  // Minimum completed adventures
}

// Ending represents a possible ending for the pet's life
type Ending struct {
	Type       string          `toml:"type"`       // blissful_passing | adventurous_life | peaceful_rest
	Name       string          `toml:"name"`       // Display name
	Condition  EndingCondition `toml:"condition"`  // Trigger condition
	Message    string          `toml:"message"`    // Ending message
}

// LifecycleState represents the current lifecycle state of a pet
type LifecycleState struct {
	NearEnd    bool    // Is pet near end of life?
	AgePercent float64 // Age as percentage of max lifespan (0.0-1.0)
}

// EndingResult represents the result of triggering an ending
type EndingResult struct {
	Type    string // Ending type
	Message string // Ending message
}
