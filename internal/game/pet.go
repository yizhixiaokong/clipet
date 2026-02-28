// Package game contains the core game logic, independent of any UI framework.
package game

import (
	"clipet/internal/game/capabilities"
	"clipet/internal/plugin"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// globalTimeManager is the package-level singleton for time evolution management
var globalTimeManager = NewTimeManager()

// RegisterTimeHook registers a time hook with the global time manager
func RegisterTimeHook(hook TimeHook, priority TimeHookPriority) {
	globalTimeManager.RegisterHook(hook, priority)
}

// PetStage represents the current life phase of a pet.
type PetStage string

const (
	StageEgg    PetStage = "egg"
	StageBaby   PetStage = "baby"
	StageChild  PetStage = "child"
	StageAdult  PetStage = "adult"
	StageLegend PetStage = "legend"
)

// AnimState represents the current animation state of a pet.
type AnimState string

const (
	AnimIdle     AnimState = "idle"
	AnimEating   AnimState = "eating"
	AnimSleeping AnimState = "sleeping"
	AnimPlaying  AnimState = "playing"
	AnimSad      AnimState = "sad"
	AnimHappy    AnimState = "happy"
)

// Default action configuration constants.
const (
	// Cooldown durations
	DefaultCooldownFeed = 10 * time.Minute
	DefaultCooldownPlay = 5 * time.Minute
	DefaultCooldownRest = 10 * time.Minute
	DefaultCooldownHeal = 20 * time.Minute
	DefaultCooldownTalk = 10 * time.Second

	// Feed action
	DefaultFeedHunger    = 35
	DefaultFeedHappiness = 5

	// Play action
	DefaultPlayHappiness = 20
	DefaultPlayEnergy    = -8

	// Rest action
	DefaultRestEnergy    = 40
	DefaultRestHealth    = 5
	DefaultRestHappiness = -5

	// Heal action
	DefaultHealHealth    = 25
	DefaultHealEnergyCost = 15

	// Talk action
	DefaultTalkHappiness = 1
)

// Error type constants for ActionResult
const (
	ErrEnergyLow      = "energy_low"
	ErrHealthLow      = "health_low"
	ErrCooldown       = "cooldown"
	ErrDead           = "dead"
	ErrInvalidAction  = "invalid_action"
	ErrFullHunger     = "full_hunger"
	ErrFullEnergy     = "full_energy"
	ErrSkillSystem    = "skill_system"
	ErrSkillUnknown   = "skill_unknown"
	ErrSkillNotActive = "skill_not_active"
)

// ActionResult holds the outcome of a pet action.
type ActionResult struct {
	OK                bool              // whether the action succeeded
	ErrorType         string            // standardized error type for i18n (empty if OK)
	Message           string            // human-readable feedback (for internal logs)
	Changes           map[string][2]int // attr name -> {old, new}
	Animation         AnimState         // animation to play (empty = no change)
	AnimationDuration time.Duration     // how long the animation should last
}

// diminish calculates a diminishing-return gain.
// As 'current' approaches 100 the effective gain shrinks toward 1.
func diminish(base, current int) int {
	gain := base * (100 - current) / 100
	if gain < 1 {
		gain = 1
	}
	return gain
}

// failResult is a convenience helper for failed actions (without ErrorType).
func failResult(msg string) ActionResult {
	return ActionResult{OK: false, Message: msg}
}

// failResultWithType is a convenience helper for failed actions with ErrorType.
func failResultWithType(errorType, msg string) ActionResult {
	return ActionResult{OK: false, ErrorType: errorType, Message: msg}
}

// cooldownLeft returns a human-readable remaining cooldown string.
func cooldownLeft(last time.Time, cd time.Duration) string {
	remaining := cd - time.Since(last)
	if remaining <= 0 {
		return ""
	}
	if remaining < time.Minute {
		return fmt.Sprintf("%d秒", int(remaining.Seconds()))
	}
	return fmt.Sprintf("%d分%d秒", int(remaining.Minutes()), int(remaining.Seconds())%60)
}

// Pet is the central game entity representing the player's virtual pet.
type Pet struct {
	// Basic info
	Name    string   `json:"name"`
	Species string   `json:"species"`  // species pack ID, e.g. "cat"
	Stage   PetStage `json:"stage"`    // current life phase
	StageID string   `json:"stage_id"` // current evolution node ID, e.g. "baby"

	Birthday time.Time `json:"birthday"`

	// Accumulated offline time (from natural offline + dev timeskip)
	// Applied when TUI starts, then cleared
	AccumulatedOfflineDuration time.Duration `json:"accumulated_offline_duration,omitempty"`

	// Attributes (0-100)
	Hunger    int `json:"hunger"` // fullness, higher = less hungry
	Happiness int `json:"happiness"`
	Health    int `json:"health"`
	Energy    int `json:"energy"`

	// Timestamps
	LastFedAt     time.Time `json:"last_fed_at"`
	LastPlayedAt  time.Time `json:"last_played_at"`
	LastRestedAt  time.Time `json:"last_rested_at"`
	LastHealedAt  time.Time `json:"last_healed_at"`
	LastTalkedAt     time.Time `json:"last_talked_at"`
	LastCheckedAt    time.Time `json:"last_checked_at"`
	LastAdventureAt  time.Time `json:"last_adventure_at"`
	LastSkillUsedAt  time.Time `json:"last_skill_used_at"` // NEW: skill cooldown tracking

	// Statistics
	TotalInteractions   int `json:"total_interactions"`
	GamesWon            int `json:"games_won"`
	AdventuresCompleted int `json:"adventures_completed"`
	DialogueCount       int `json:"dialogue_count"`

	// Evolution accumulation scores
	AccHappiness      int     `json:"acc_happiness"`
	AccHealth         int     `json:"acc_health"`
	AccPlayful        int     `json:"acc_playful"`
	NightInteractions int     `json:"night_interactions"`
	DayInteractions   int     `json:"day_interactions"`
	FeedRegularity    float64 `json:"feed_regularity"`
	FeedCount         int     `json:"feed_count"`
	FeedExpectedCount int     `json:"feed_expected_count"`

	// State
	Alive                 bool          `json:"alive"`
	CurrentAnimation      AnimState     `json:"current_animation"`
	AnimationEndTime      time.Time     `json:"animation_end_time"`      // when current animation should end
	LifecycleWarningShown bool          `json:"lifecycle_warning_shown"` // NEW: lifecycle tracking

	// Ending information
	EndingType    string `json:"ending_type,omitempty"`    // Ending type for i18n lookup
	EndingMessage string `json:"ending_message,omitempty"` // Plugin-provided message (optional)

	// Custom attributes (Phase 3)
	CustomAttributes map[string]int `json:"custom_attributes,omitempty"` // NEW: custom attribute storage

	// Plugin registry (not serialized)
	registry *plugin.Registry `json:"-"`

	// Capabilities registry (not serialized)
	capabilitiesReg *capabilities.Registry `json:"-"`
}

// NewPet creates a new pet with the given name and species.
// It sets initial attributes from the provided base stats.
func NewPet(name, species, eggStageID string, hunger, happiness, health, energy int, registry *plugin.Registry) *Pet {
	now := time.Now()
	return &Pet{
		Name:             name,
		Species:          species,
		Stage:            StageEgg,
		StageID:          eggStageID,
		Birthday:         now,
		Hunger:           hunger,
		Happiness:        happiness,
		Health:           health,
		Energy:           energy,
		LastFedAt:        now,
		LastPlayedAt:     now,
		LastRestedAt:     now,
		LastHealedAt:     now,
		LastTalkedAt:     now,
		LastCheckedAt:    now,
		LastAdventureAt:  now,
		LastSkillUsedAt:  now,
		Alive:            true,
		CurrentAnimation: AnimIdle,
		registry:         registry,
	}
}

// SetRegistry sets the plugin registry for the pet.
// This is needed after loading a pet from save file, as registry is not serialized.
func (p *Pet) SetRegistry(registry *plugin.Registry) {
	p.registry = registry
}

// SetCapabilitiesRegistry sets the capabilities registry for the pet.
func (p *Pet) SetCapabilitiesRegistry(capReg *capabilities.Registry) {
	p.capabilitiesReg = capReg
}

// Registry returns the plugin registry for the pet.
func (p *Pet) Registry() *plugin.Registry {
	return p.registry
}

// CapabilitiesRegistry returns the capabilities registry for the pet.
func (p *Pet) CapabilitiesRegistry() *capabilities.Registry {
	return p.capabilitiesReg
}

// Feed increases the pet's hunger (fullness) level.
// Dynamic cooldown based on urgency. Prerequisite: hunger < 95. Diminishing returns on gain.
func (p *Pet) Feed() ActionResult {
	if !p.Alive {
		return failResultWithType(ErrDead, "宠物已经不在了...")
	}

	// Calculate dynamic cooldown based on current hunger
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "feed", p.Hunger)
	if left := cooldownLeft(p.LastFedAt, cooldown); left != "" {
		return failResultWithType(ErrCooldown, fmt.Sprintf("宠物还不饿，%s后可以再喂", left))
	}
	if p.Hunger >= 95 {
		return failResultWithType(ErrFullHunger, "宠物已经很饱了！")
	}

	// Get effects from plugin or defaults
	hungerGain, happinessGain, _, _ := GetActionEffects(p.registry, p.Species, "feed")

	// Apply passive trait effects
	if p.capabilitiesReg != nil {
		hungerGain, happinessGain, _, _ = p.capabilitiesReg.ApplyPassiveEffects(
			p.Species, "feed", hungerGain, happinessGain, 0, 0)
	}

	ch := make(map[string][2]int)
	oldH := p.Hunger
	oldHp := p.Happiness
	p.Hunger = clamp(p.Hunger+diminish(hungerGain, p.Hunger), 0, 100)
	p.Happiness = clamp(p.Happiness+diminish(happinessGain, p.Happiness), 0, 100)
	ch["hunger"] = [2]int{oldH, p.Hunger}
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	p.LastFedAt = time.Now()
	p.TotalInteractions++
	p.FeedCount++
	p.trackTimeOfDay()
	// Evolution modifiers are applied in evolution checks, not here
	return ActionResult{
		OK:                true,
		Message:           "喂食成功！",
		Changes:           ch,
		Animation:         AnimEating,
		AnimationDuration: 2 * time.Second,
	}
}

// Play increases the pet's happiness and decreases energy.
// Dynamic cooldown based on urgency. Prerequisite: energy >= cost. Diminishing returns on happiness gain.
func (p *Pet) Play() ActionResult {
	if !p.Alive {
		return failResultWithType(ErrDead, "宠物已经不在了...")
	}

	// Get energy cost from plugin or default
	energyCost := GetActionEnergyCost(p.registry, p.Species, "play")
	if energyCost == 0 {
		energyCost = 10 // fallback
	}

	// Calculate dynamic cooldown based on current happiness (urgency)
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "play", p.Happiness)
	if left := cooldownLeft(p.LastPlayedAt, cooldown); left != "" {
		return failResultWithType(ErrCooldown, fmt.Sprintf("宠物还在喘气，%s后可以再玩", left))
	}
	if p.Energy < energyCost {
		return failResultWithType(ErrEnergyLow, "宠物太累了，先休息一下吧！")
	}

	// Get effects from plugin or defaults
	_, happinessGain, _, energyLoss := GetActionEffects(p.registry, p.Species, "play")

	// Apply passive trait effects
	if p.capabilitiesReg != nil {
		_, happinessGain, _, energyLoss = p.capabilitiesReg.ApplyPassiveEffects(
			p.Species, "play", 0, happinessGain, 0, energyLoss)
	}

	ch := make(map[string][2]int)
	oldHp := p.Happiness
	oldE := p.Energy
	p.Happiness = clamp(p.Happiness+diminish(happinessGain, p.Happiness), 0, 100)
	p.Energy = clamp(p.Energy+energyLoss, 0, 100) // energyLoss is negative
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	ch["energy"] = [2]int{oldE, p.Energy}
	p.AccPlayful += p.addEvolutionPoints(1, "play")
	p.LastPlayedAt = time.Now()
	p.TotalInteractions++
	p.trackTimeOfDay()
	return ActionResult{
		OK:                true,
		Message:           "玩耍愉快！",
		Changes:           ch,
		Animation:         AnimPlaying,
		AnimationDuration: 2 * time.Second,
	}
}

// Talk records a dialogue interaction.
// Dynamic cooldown based on urgency. Diminishing returns on happiness gain.
func (p *Pet) Talk() ActionResult {
	if !p.Alive {
		return failResultWithType(ErrDead, "宠物已经不在了...")
	}

	// Calculate dynamic cooldown based on current happiness (urgency)
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "talk", p.Happiness)
	if left := cooldownLeft(p.LastTalkedAt, cooldown); left != "" {
		return failResultWithType(ErrCooldown, fmt.Sprintf("宠物需要消化一下，%s后可以再聊", left))
	}

	// Get effects from plugin or defaults
	_, happinessGain, _, _ := GetActionEffects(p.registry, p.Species, "talk")

	ch := make(map[string][2]int)
	oldHp := p.Happiness
	p.Happiness = clamp(p.Happiness+diminish(happinessGain, p.Happiness), 0, 100)
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	p.DialogueCount++
	p.TotalInteractions++
	p.AccHappiness += p.addEvolutionPoints(1, "happiness")
	p.LastTalkedAt = time.Now()
	p.trackTimeOfDay()
	return ActionResult{OK: true, Message: "聊天愉快！", Changes: ch}
}

// Rest lets the pet sleep/rest, recovering energy and a small amount of health.
// Dynamic cooldown based on urgency. Prerequisite: energy < 90. Diminishing returns on energy gain.
func (p *Pet) Rest() ActionResult {
	if !p.Alive {
		return failResultWithType(ErrDead, "宠物已经不在了...")
	}

	// Calculate dynamic cooldown based on current energy (urgency)
	// Low energy = urgent (short cooldown), high energy = not urgent (long cooldown)
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "rest", p.Energy)
	if left := cooldownLeft(p.LastRestedAt, cooldown); left != "" {
		return failResultWithType(ErrCooldown, fmt.Sprintf("宠物还不困，%s后可以再休息", left))
	}
	if p.Energy >= 90 {
		return failResultWithType(ErrFullEnergy, "宠物精力充沛，不需要休息！")
	}

	// Get effects from plugin or defaults
	_, happinessLoss, healthGain, energyGain := GetActionEffects(p.registry, p.Species, "rest")

	// Apply passive trait effects (use "sleep" for rest action)
	if p.capabilitiesReg != nil {
		_, happinessLoss, healthGain, energyGain = p.capabilitiesReg.ApplyPassiveEffects(
			p.Species, "sleep", 0, happinessLoss, healthGain, energyGain)
	}

	ch := make(map[string][2]int)
	oldE := p.Energy
	oldH := p.Health
	oldHp := p.Happiness
	p.Energy = clamp(p.Energy+diminish(energyGain, p.Energy), 0, 100)
	p.Health = clamp(p.Health+diminish(healthGain, p.Health), 0, 100)
	p.Happiness = clamp(p.Happiness+happinessLoss, 0, 100) // happinessLoss is negative
	ch["energy"] = [2]int{oldE, p.Energy}
	ch["health"] = [2]int{oldH, p.Health}
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	p.LastRestedAt = time.Now()
	p.TotalInteractions++
	p.trackTimeOfDay()
	return ActionResult{
		OK:                true,
		Message:           "休息一下～",
		Changes:           ch,
		Animation:         AnimSleeping,
		AnimationDuration: 2 * time.Second,
	}
}

// Heal treats the pet, recovering health but costing energy.
// Dynamic cooldown based on urgency. Prerequisite: energy >= cost. Diminishing returns on health gain.
func (p *Pet) Heal() ActionResult {
	if !p.Alive {
		return failResultWithType(ErrDead, "宠物已经不在了...")
	}

	// Get energy cost from plugin or default
	energyCost := GetActionEnergyCost(p.registry, p.Species, "heal")
	if energyCost == 0 {
		energyCost = 15 // fallback
	}

	// Calculate dynamic cooldown based on current health (urgency)
	// Low health = urgent (short cooldown), high health = not urgent (long cooldown)
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "heal", p.Health)
	if left := cooldownLeft(p.LastHealedAt, cooldown); left != "" {
		return failResultWithType(ErrCooldown, fmt.Sprintf("刚治疗过，%s后可以再治疗", left))
	}
	if p.Energy < energyCost {
		return failResultWithType(ErrEnergyLow, "宠物精力不足，需要先休息！")
	}

	// Get effects from plugin or defaults
	_, _, healthGain, energyLoss := GetActionEffects(p.registry, p.Species, "heal")

	ch := make(map[string][2]int)
	oldH := p.Health
	oldE := p.Energy
	p.Health = clamp(p.Health+diminish(healthGain, p.Health), 0, 100)
	p.Energy = clamp(p.Energy+energyLoss, 0, 100) // energyLoss is negative
	ch["health"] = [2]int{oldH, p.Health}
	ch["energy"] = [2]int{oldE, p.Energy}
	p.AccHealth += p.addEvolutionPoints(1, "health")
	p.LastHealedAt = time.Now()
	p.TotalInteractions++
	p.trackTimeOfDay()
	return ActionResult{OK: true, Message: "治疗完成！", Changes: ch}
}

// MoodScore calculates the composite mood score (0-100).
func (p *Pet) MoodScore() int {
	score := float64(p.Hunger)*0.25 +
		float64(p.Happiness)*0.35 +
		float64(p.Health)*0.25 +
		float64(p.Energy)*0.15
	return clamp(int(score), 0, 100)
}

// MoodName returns a human-readable mood string.
func (p *Pet) MoodName() string {
	score := p.MoodScore()
	switch {
	case score > 80:
		return "happy"
	case score > 60:
		return "normal"
	case score > 40:
		return "unhappy"
	case score > 20:
		return "sad"
	default:
		return "miserable"
	}
}

// AgeHours returns the pet's age in hours.
func (p *Pet) AgeHours() float64 {
	return time.Since(p.Birthday).Hours()
}

// IsAlive checks if the pet is still alive.
func (p *Pet) IsAlive() bool {
	return p.Alive
}

// UpdateAnimation sets the appropriate animation based on current state.
func (p *Pet) UpdateAnimation() {
	// If we're in a timed animation and it hasn't expired, keep it
	if !p.AnimationEndTime.IsZero() && time.Now().Before(p.AnimationEndTime) {
		return
	}

	// Clear animation end time when expired
	p.AnimationEndTime = time.Time{}

	if !p.Alive {
		p.CurrentAnimation = AnimSad
		return
	}

	mood := p.MoodName()
	switch {
	case p.Energy < 15:
		p.CurrentAnimation = AnimSleeping
	case mood == "sad" || mood == "miserable":
		p.CurrentAnimation = AnimSad
	case mood == "happy":
		p.CurrentAnimation = AnimHappy
	default:
		p.CurrentAnimation = AnimIdle
	}
}

// trackTimeOfDay records whether an interaction happened during day or night.
func (p *Pet) trackTimeOfDay() {
	hour := time.Now().Hour()
	if hour >= 6 && hour < 18 {
		p.DayInteractions++
	} else {
		p.NightInteractions++
	}
}

// Clamp constrains val to the range [min, max].
// Exported for use by sub-packages (e.g. games).
func Clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// clamp is the internal shorthand.
func clamp(val, min, max int) int { return Clamp(val, min, max) }

// AdvanceTime executes time evolution by delegating to the global TimeManager.
func (p *Pet) AdvanceTime(elapsed time.Duration) {
	if !p.Alive {
		return
	}
	globalTimeManager.AdvanceTime(elapsed, p)
}

// SimulateDecay applies time-based attribute decay over the given duration.
// Decay rates per hour: hunger -3, happiness -2, energy -1.
// If hunger drops below 20, health decays at -0.5/hr.
// If health reaches 0, the pet dies.
// Deprecated: Use AdvanceTime instead. This method now delegates to AdvanceTime
// for backward compatibility.
func (p *Pet) SimulateDecay(elapsed time.Duration) {
	p.AdvanceTime(elapsed)
}

// SetField sets a pet field by name from a raw string value.
// Returns the previous value as a string for display purposes.
func (p *Pet) SetField(field string, raw string) (old string, err error) {
	switch strings.ToLower(field) {
	// Core attributes (0-100)
	case "hunger":
		old = strconv.Itoa(p.Hunger)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.Hunger = clamp(v, 0, 100)
	case "happiness":
		old = strconv.Itoa(p.Happiness)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.Happiness = clamp(v, 0, 100)
	case "health":
		old = strconv.Itoa(p.Health)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.Health = clamp(v, 0, 100)
	case "energy":
		old = strconv.Itoa(p.Energy)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.Energy = clamp(v, 0, 100)

	// Basic info
	case "name":
		old = p.Name
		p.Name = raw
	case "species":
		old = p.Species
		p.Species = raw
	case "stage_id":
		old = p.StageID
		p.StageID = raw
	case "alive":
		old = strconv.FormatBool(p.Alive)
		b, e := strconv.ParseBool(raw)
		if e != nil {
			return "", e
		}
		p.Alive = b

	// Statistics
	case "total_interactions", "interactions":
		old = strconv.Itoa(p.TotalInteractions)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.TotalInteractions = v
	case "feed_count", "feeds":
		old = strconv.Itoa(p.FeedCount)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.FeedCount = v
	case "dialogue_count", "dialogues":
		old = strconv.Itoa(p.DialogueCount)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.DialogueCount = v
	case "adventures_completed", "adventures":
		old = strconv.Itoa(p.AdventuresCompleted)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.AdventuresCompleted = v
	case "games_won", "wins":
		old = strconv.Itoa(p.GamesWon)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.GamesWon = v

	// Evolution accumulators
	case "acc_happiness":
		old = strconv.Itoa(p.AccHappiness)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.AccHappiness = v
	case "acc_health":
		old = strconv.Itoa(p.AccHealth)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.AccHealth = v
	case "acc_playful":
		old = strconv.Itoa(p.AccPlayful)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.AccPlayful = v
	case "night_interactions", "night":
		old = strconv.Itoa(p.NightInteractions)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.NightInteractions = v
	case "day_interactions", "day":
		old = strconv.Itoa(p.DayInteractions)
		v, e := strconv.Atoi(raw)
		if e != nil {
			return "", e
		}
		p.DayInteractions = v
	case "feed_regularity":
		old = strconv.FormatFloat(p.FeedRegularity, 'f', 2, 64)
		v, e := strconv.ParseFloat(raw, 64)
		if e != nil {
			return "", e
		}
		p.FeedRegularity = v

	// Lifecycle (M7)
	case "lifecycle_warning", "warning_shown":
		old = strconv.FormatBool(p.LifecycleWarningShown)
		b, e := strconv.ParseBool(raw)
		if e != nil {
			return "", e
		}
		p.LifecycleWarningShown = b

	// Age manipulation
	case "age_hours", "age":
		currentAge := p.AgeHours()
		old = strconv.FormatFloat(currentAge, 'f', 1, 64)
		v, e := strconv.ParseFloat(raw, 64)
		if e != nil {
			return "", e
		}
		if v < 0 {
			return "", fmt.Errorf("age cannot be negative")
		}
		// Adjust birthday to achieve desired age
		p.Birthday = time.Now().Add(-time.Duration(v * float64(time.Hour)))

	// Custom attributes (Phase 1: custom:attr_name or attr_name)
	case "custom":
		// Format: "custom:attr_name" or just "attr_name" for custom attributes
		return "", fmt.Errorf("invalid custom attribute format, use 'custom:attr_name' or 'set custom:attr_name value'")
	default:
		// Check for custom: prefix
		if strings.HasPrefix(field, "custom:") {
			attrName := strings.TrimPrefix(field, "custom:")
			old = strconv.Itoa(p.GetAttr(attrName))
			v, e := strconv.Atoi(raw)
			if e != nil {
				return "", e
			}
			p.SetAttr(attrName, v)
			return old, nil
		}
		return "", fmt.Errorf("unknown field %q\n\nValid fields:\n  Attributes: hunger, happiness, health, energy (0-100)\n  Info: name, species, stage_id, alive, age_hours\n  Stats: interactions, feeds, dialogues, adventures, wins\n  Evolution: acc_happiness, acc_health, acc_playful, night, day, feed_regularity\n  Lifecycle: lifecycle_warning\n  Custom: custom:<attr_name>", field)
	}
	return old, nil
}

// GetAttr returns a named attribute value (hunger, happiness, health, energy, or custom).
func (p *Pet) GetAttr(name string) int {
	switch strings.ToLower(name) {
	case "hunger":
		return p.Hunger
	case "happiness":
		return p.Happiness
	case "health":
		return p.Health
	case "energy":
		return p.Energy
	default:
		// Check custom attributes
		if p.CustomAttributes != nil {
			if val, ok := p.CustomAttributes[strings.ToLower(name)]; ok {
				return val
			}
		}
		return 0
	}
}

// SetAttr sets a custom attribute value (not for core attributes).
// Core attributes (hunger, happiness, health, energy) should be set directly.
func (p *Pet) SetAttr(name string, value int) {
	if p.CustomAttributes == nil {
		p.CustomAttributes = make(map[string]int)
	}
	p.CustomAttributes[strings.ToLower(name)] = value
}

// UpdateFeedRegularity recalculates the feed regularity based on age.
// Expected feed count: ~3 feeds per 24 hours of age.
func (p *Pet) UpdateFeedRegularity() {
	ageHours := p.AgeHours()
	if ageHours < 1 {
		p.FeedExpectedCount = 1
	} else {
		p.FeedExpectedCount = int(ageHours / 8) // ~3 per day
		if p.FeedExpectedCount < 1 {
			p.FeedExpectedCount = 1
		}
	}
	if p.FeedCount >= p.FeedExpectedCount {
		p.FeedRegularity = 1.0
	} else {
		p.FeedRegularity = float64(p.FeedCount) / float64(p.FeedExpectedCount)
	}
}

// AccumulateOfflineTime calculates the time elapsed since LastCheckedAt
// and accumulates it to AccumulatedOfflineDuration for later application.
// This should be called when loading a pet from a save file.
func (p *Pet) AccumulateOfflineTime() {
	if !p.Alive {
		return
	}
	elapsed := time.Since(p.LastCheckedAt)
	if elapsed < time.Minute {
		return
	}
	// Accumulate offline time for later application (when TUI starts)
	p.AccumulatedOfflineDuration += elapsed
	// Update LastCheckedAt to now so we don't double-count this time
	p.LastCheckedAt = time.Now()
}

// MarkAsChecked updates LastCheckedAt to current time.
// This should be called before saving the pet during online play
// to prevent counting online time as offline time.
func (p *Pet) MarkAsChecked() {
	p.LastCheckedAt = time.Now()
}

// UseSkill uses an active skill/ability.
// The skill must be defined in the species traits with type="active".
// Returns an ActionResult indicating success or failure.
func (p *Pet) UseSkill(skillID string) ActionResult {
	if !p.Alive {
		return failResultWithType(ErrDead, "宠物已经不在了...")
	}

	// Get the skill from capabilities registry
	if p.capabilitiesReg == nil {
		return failResultWithType(ErrSkillSystem, "技能系统未初始化")
	}

	trait, exists := p.capabilitiesReg.GetTrait(p.Species, skillID)
	if !exists {
		return failResultWithType(ErrSkillUnknown, "未知技能")
	}

	if trait.Type != "active" || trait.ActiveEffect == nil {
		return failResultWithType(ErrSkillNotActive, "这不是一个主动技能")
	}

	effect := trait.ActiveEffect

	// Cooldown is already parsed as time.Duration
	cooldown := effect.Cooldown
	if cooldown == 0 {
		cooldown = 30 * time.Minute // fallback
	}

	// Check cooldown
	if left := cooldownLeft(p.LastSkillUsedAt, cooldown); left != "" {
		return failResultWithType(ErrCooldown, fmt.Sprintf("技能冷却中，%s后可再次使用", left))
	}

	// Check energy cost
	if p.Energy < effect.EnergyCost {
		return failResultWithType(ErrEnergyLow, fmt.Sprintf("精力不足，需要 %d 精力", effect.EnergyCost))
	}

	// Apply skill effect
	ch := make(map[string][2]int)
	oldHealth := p.Health
	oldEnergy := p.Energy

	p.Energy -= effect.EnergyCost
	p.Health = clamp(p.Health+effect.HealthRestore, 0, 100)

	ch["health"] = [2]int{oldHealth, p.Health}
	ch["energy"] = [2]int{oldEnergy, p.Energy}

	p.LastSkillUsedAt = time.Now()
	p.TotalInteractions++

	return ActionResult{
		OK:                true,
		Message:           fmt.Sprintf("使用技能「%s」！", trait.Name),
		Changes:           ch,
		Animation:         AnimHappy,
		AnimationDuration: 2 * time.Second,
	}
}

// addEvolutionPoints adds evolution accumulation points with modifiers applied.
// basePoints is the number of points to add.
// interactionType is one of: "happiness", "health", "playful", "feed", "adventure"
func (p *Pet) addEvolutionPoints(basePoints int, interactionType string) int {
	if p.capabilitiesReg == nil {
		return basePoints
	}

	modifier := p.capabilitiesReg.GetEvolutionModifier(p.Species)
	if modifier == nil {
		return basePoints
	}

	points := float64(basePoints)

	// Apply time-based modifiers
	hour := time.Now().Hour()
	isNight := hour < 6 || hour >= 18

	if isNight && modifier.NightInteractionBonus > 0 {
		points *= modifier.NightInteractionBonus
	} else if !isNight && modifier.DayInteractionBonus > 0 {
		points *= modifier.DayInteractionBonus
	}

	// Apply interaction-type modifiers
	switch interactionType {
	case "feed":
		if modifier.FeedBonus > 0 {
			points *= modifier.FeedBonus
		}
	case "play":
		if modifier.PlayBonus > 0 {
			points *= modifier.PlayBonus
		}
	case "adventure":
		if modifier.AdventureBonus > 0 {
			points *= modifier.AdventureBonus
		}
	}

	return int(points)
}

// DecayRoundResult records the result of one round of decay settlement
type DecayRoundResult struct {
	Round         int           // Round number
	Duration      time.Duration // Duration of this round
	StartAttrs    [4]int        // Attributes at start [Hunger, Happiness, Health, Energy]
	EndAttrs      [4]int        // Attributes at end
	Effects       []string      // Effects triggered in this round
	CriticalState bool          // Whether critical state was triggered
}

// ApplyMultiStageDecay applies multi-stage offline time decay
// Each round is 6 hours, checking attribute states and applying dynamic effects
func (p *Pet) ApplyMultiStageDecay(totalDuration time.Duration) []DecayRoundResult {
	if p.registry == nil {
		return nil
	}

	// Get configurations
	decayConfig := p.registry.GetDecayConfig(p.Species)
	interactionConfig := p.registry.GetAttributeInteractionConfig(p.Species)

	// Settle in rounds (6 hours each)
	const roundDuration = 6 * time.Hour
	rounds := int(totalDuration / roundDuration)
	remainder := totalDuration % roundDuration

	results := make([]DecayRoundResult, 0, rounds+1)

	// Settle round by round
	for i := 0; i < rounds; i++ {
		result := p.applyOneDecayRound(roundDuration, decayConfig, interactionConfig, i+1)
		results = append(results, result)
	}

	// Handle remaining time (less than 6 hours)
	if remainder > 0 {
		result := p.applyOneDecayRound(remainder, decayConfig, interactionConfig, rounds+1)
		results = append(results, result)
	}

	return results
}

// applyOneDecayRound applies one round of decay
func (p *Pet) applyOneDecayRound(dur time.Duration, decayConfig capabilities.DecayConfig,
	interactionConfig capabilities.AttributeInteractionConfig, roundNum int) DecayRoundResult {

	result := DecayRoundResult{
		Round:      roundNum,
		Duration:   dur,
		StartAttrs: [4]int{p.Hunger, p.Happiness, p.Health, p.Energy},
	}

	hours := dur.Hours()

	// 1. Apply base decay
	baseHunger := int(decayConfig.Hunger * hours)
	baseHappiness := int(decayConfig.Happiness * hours)
	baseEnergy := int(decayConfig.Energy * hours)

	p.Hunger = clamp(p.Hunger-baseHunger, 0, 100)
	p.Happiness = clamp(p.Happiness-baseHappiness, 0, 100)
	p.Energy = clamp(p.Energy-baseEnergy, 0, 100)

	// 2. Apply attribute interactions
	p.applyAttributeInteractions(hours, decayConfig, interactionConfig, &result)

	// 3. Record results
	result.EndAttrs = [4]int{p.Hunger, p.Happiness, p.Health, p.Energy}

	return result
}

// applyAttributeInteractions applies attribute interaction effects
func (p *Pet) applyAttributeInteractions(hours float64, decayConfig capabilities.DecayConfig,
	interactionConfig capabilities.AttributeInteractionConfig, result *DecayRoundResult) {

	var healthDecay float64
	var happinessDecay float64

	// Rule 1: Hunger → Health
	if p.Hunger < interactionConfig.HungerHealthThreshold {
		baseRate := decayConfig.Health * hours

		// Bonus when hunger=0
		if p.Hunger == 0 {
			healthDecay += baseRate * interactionConfig.HungerZeroHealthMultiplier
			result.Effects = append(result.Effects, "⚠️ 饥饿致死：健康大量衰减")
			result.CriticalState = true
		} else {
			healthDecay += baseRate
			result.Effects = append(result.Effects, "饿了：健康轻微衰减")
		}
	}

	// Rule 2: Energy → Health/Happiness
	if p.Energy < interactionConfig.EnergyCritThreshold {
		// Critical energy
		healthDecay += decayConfig.Health * hours * interactionConfig.EnergyCritHealthMultiplier
		happinessDecay += decayConfig.Happiness * hours * interactionConfig.EnergyCritHappinessMultiplier
		result.Effects = append(result.Effects, "🚨 精力耗尽：健康和快乐加速衰减")
		result.CriticalState = true
	} else if p.Energy < interactionConfig.EnergyLowThreshold {
		// Low energy
		healthDecay += interactionConfig.EnergyLowHealthRate * hours
		result.Effects = append(result.Effects, "疲惫：健康轻微衰减")
	}

	// Rule 3: Happiness → Health
	if p.Happiness < interactionConfig.HappinessLowThreshold {
		if p.Happiness == 0 {
			healthDecay += decayConfig.Health * hours * interactionConfig.HappinessZeroHealthMultiplier
			result.Effects = append(result.Effects, "💔 抑郁症：健康大量衰减")
			result.CriticalState = true
		} else {
			healthDecay += 0.15 * hours
			result.Effects = append(result.Effects, "不开心：健康轻微衰减")
		}
	}

	// Apply decay (with death protection)
	if healthDecay > 0 {
		// Rule 4: Death protection (health stops at 1)
		potentialHealth := p.Health - int(healthDecay)
		if potentialHealth <= 0 && p.Health > 1 {
			p.Health = 1 // Keep 1 HP
			result.Effects = append(result.Effects, "🛡️ 濒死保护：健康保持1点")
		} else {
			p.Health = clamp(p.Health-int(healthDecay), 0, 100)
		}
	}

	if happinessDecay > 0 {
		p.Happiness = clamp(p.Happiness-int(happinessDecay), 0, 100)
	}
}

// UpdateCooldowns updates all cooldown timestamps
func (p *Pet) UpdateCooldowns(elapsed time.Duration) {
	p.LastFedAt = p.LastFedAt.Add(-elapsed)
	p.LastPlayedAt = p.LastPlayedAt.Add(-elapsed)
	p.LastRestedAt = p.LastRestedAt.Add(-elapsed)
	p.LastHealedAt = p.LastHealedAt.Add(-elapsed)
	p.LastTalkedAt = p.LastTalkedAt.Add(-elapsed)
	p.LastAdventureAt = p.LastAdventureAt.Add(-elapsed)
	p.LastSkillUsedAt = p.LastSkillUsedAt.Add(-elapsed)
}

// DevOnlySimulateDecay applies time-based attribute decay WITHOUT triggering hooks.
// This is for dev tools (timeskip) to test decay without triggering death/evolution.
// Only applies attribute decay, does NOT check death or trigger lifecycle events.
// Uses multi-stage settlement algorithm for realistic decay simulation.
func (p *Pet) DevOnlySimulateDecay(elapsed time.Duration) {
	if p.registry == nil {
		return
	}

	// Use multi-stage settlement
	p.ApplyMultiStageDecay(elapsed)

	// Update cooldown timestamps
	p.UpdateCooldowns(elapsed)

	// Update last checked time (prevent double decay)
	p.LastCheckedAt = p.LastCheckedAt.Add(-elapsed)
}

// GetCustomAcc returns a custom accumulator value.
// This is an alias for GetAttr for semantic clarity when dealing with accumulators.
func (p *Pet) GetCustomAcc(name string) int {
	return p.GetAttr(name)
}

// AddCustomAcc adds a value to a custom accumulator.
// Initializes the accumulator to 0 if it doesn't exist.
func (p *Pet) AddCustomAcc(name string, delta int) {
	if p.CustomAttributes == nil {
		p.CustomAttributes = make(map[string]int)
	}
	key := strings.ToLower(name)
	current := p.CustomAttributes[key]
	p.CustomAttributes[key] = current + delta
}
