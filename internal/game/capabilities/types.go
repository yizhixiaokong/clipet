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
	IsEternal  bool    // true for eternal pets (never age/die)
	IsLooping  bool    // true for looping lifecycle (resets age instead of ending)
}

// EndingResult represents the result of triggering an ending
type EndingResult struct {
	Type    string // Ending type
	Message string // Ending message
}

// DecayConfig defines attribute decay rates per hour
type DecayConfig struct {
	Hunger    float64 `toml:"hunger"`     // Hunger decay per hour (default: 1.0)
	Happiness float64 `toml:"happiness"`  // Happiness decay per hour (default: 0.5)
	Energy    float64 `toml:"energy"`     // Energy decay per hour (default: 0.3)
	Health    float64 `toml:"health"`     // Health decay per hour when hungry (default: 0.2)
}

// Defaults returns decay config with sensible defaults (slow unified decay)
func (dc DecayConfig) Defaults() DecayConfig {
	if dc.Hunger == 0 {
		dc.Hunger = 1.0 // Slow decay: ~1 point per hour
	}
	if dc.Happiness == 0 {
		dc.Happiness = 0.5
	}
	if dc.Energy == 0 {
		dc.Energy = 0.3
	}
	if dc.Health == 0 {
		dc.Health = 0.2
	}
	return dc
}

// DynamicCooldownConfig defines how cooldown scales with attribute urgency
type DynamicCooldownConfig struct {
	// Threshold tiers for urgency-based cooldown
	// When attribute is very low (0-30): very short cooldown
	// When attribute is medium (30-70): medium cooldown
	// When attribute is high (70-100): long cooldown

	// Low urgency multiplier (attribute < 30)
	LowUrgencyMultiplier float64 `toml:"low_urgency_multiplier"` // e.g., 0.1 (10% of base cooldown)

	// Medium urgency multiplier (30 <= attribute < 70)
	MediumUrgencyMultiplier float64 `toml:"medium_urgency_multiplier"` // e.g., 0.5 (50% of base cooldown)

	// High urgency multiplier (attribute >= 70)
	HighUrgencyMultiplier float64 `toml:"high_urgency_multiplier"` // e.g., 1.0 (100% of base cooldown)

	// Low/medium/high thresholds
	LowThreshold  int `toml:"low_threshold"`  // default: 30
	HighThreshold int `toml:"high_threshold"` // default: 70
}

// Defaults returns dynamic cooldown config with sensible defaults
func (dcc DynamicCooldownConfig) Defaults() DynamicCooldownConfig {
	if dcc.LowUrgencyMultiplier == 0 {
		dcc.LowUrgencyMultiplier = 0.1 // 10% cooldown when urgent
	}
	if dcc.MediumUrgencyMultiplier == 0 {
		dcc.MediumUrgencyMultiplier = 0.5 // 50% cooldown when medium
	}
	if dcc.HighUrgencyMultiplier == 0 {
		dcc.HighUrgencyMultiplier = 1.0 // 100% cooldown when full
	}
	if dcc.LowThreshold == 0 {
		dcc.LowThreshold = 30
	}
	if dcc.HighThreshold == 0 {
		dcc.HighThreshold = 70
	}
	return dcc
}

// GetMultiplier returns the cooldown multiplier based on attribute value
func (dcc DynamicCooldownConfig) GetMultiplier(attrValue int) float64 {
	if attrValue < dcc.LowThreshold {
		return dcc.LowUrgencyMultiplier
	} else if attrValue < dcc.HighThreshold {
		return dcc.MediumUrgencyMultiplier
	}
	return dcc.HighUrgencyMultiplier
}
