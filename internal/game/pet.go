// Package game contains the core game logic, independent of any UI framework.
package game

import (
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

// Action cooldown durations.
const (
	CooldownFeed = 10 * time.Minute
	CooldownPlay = 5 * time.Minute
	CooldownRest = 15 * time.Minute
	CooldownHeal = 20 * time.Minute
	CooldownTalk = 2 * time.Minute
)

// ActionResult holds the outcome of a pet action.
type ActionResult struct {
	OK      bool              // whether the action succeeded
	Message string            // human-readable feedback
	Changes map[string][2]int // attr name -> {old, new}
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

// failResult is a convenience helper for failed actions.
func failResult(msg string) ActionResult {
	return ActionResult{OK: false, Message: msg}
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
	Alive                 bool      `json:"alive"`
	CurrentAnimation      AnimState `json:"current_animation"`
	LifecycleWarningShown bool      `json:"lifecycle_warning_shown"` // NEW: lifecycle tracking

	// Ending information
	EndingMessage string `json:"ending_message,omitempty"` // Final message when pet dies

	// Custom attributes (Phase 3)
	CustomAttributes map[string]int `json:"custom_attributes,omitempty"` // NEW: custom attribute storage

	// Plugin registry (not serialized)
	registry *plugin.Registry `json:"-"`
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

// Registry returns the plugin registry for the pet.
func (p *Pet) Registry() *plugin.Registry {
	return p.registry
}

// Feed increases the pet's hunger (fullness) level.
// Dynamic cooldown based on urgency. Prerequisite: hunger < 95. Diminishing returns on gain.
func (p *Pet) Feed() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}

	// Calculate dynamic cooldown based on current hunger
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "feed", p.Hunger)
	if left := cooldownLeft(p.LastFedAt, cooldown); left != "" {
		return failResult(fmt.Sprintf("宠物还不饿，%s后可以再喂", left))
	}
	if p.Hunger >= 95 {
		return failResult("宠物已经很饱了！")
	}

	// Get effects from plugin or defaults
	hungerGain, happinessGain, _, _ := GetActionEffects(p.registry, p.Species, "feed")

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
	return ActionResult{OK: true, Message: "喂食成功！", Changes: ch}
}

// Play increases the pet's happiness and decreases energy.
// Dynamic cooldown based on urgency. Prerequisite: energy >= cost. Diminishing returns on happiness gain.
func (p *Pet) Play() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}

	// Get energy cost from plugin or default
	energyCost := GetActionEnergyCost(p.registry, p.Species, "play")
	if energyCost == 0 {
		energyCost = 10 // fallback
	}

	// Calculate dynamic cooldown based on current happiness (urgency)
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "play", p.Happiness)
	if left := cooldownLeft(p.LastPlayedAt, cooldown); left != "" {
		return failResult(fmt.Sprintf("宠物还在喘气，%s后可以再玩", left))
	}
	if p.Energy < energyCost {
		return failResult("宠物太累了，先休息一下吧！")
	}

	// Get effects from plugin or defaults
	_, happinessGain, _, energyLoss := GetActionEffects(p.registry, p.Species, "play")

	ch := make(map[string][2]int)
	oldHp := p.Happiness
	oldE := p.Energy
	p.Happiness = clamp(p.Happiness+diminish(happinessGain, p.Happiness), 0, 100)
	p.Energy = clamp(p.Energy+energyLoss, 0, 100) // energyLoss is negative
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	ch["energy"] = [2]int{oldE, p.Energy}
	p.AccPlayful++
	p.LastPlayedAt = time.Now()
	p.TotalInteractions++
	p.trackTimeOfDay()
	return ActionResult{OK: true, Message: "玩耍愉快！", Changes: ch}
}

// Talk records a dialogue interaction.
// Dynamic cooldown based on urgency. Diminishing returns on happiness gain.
func (p *Pet) Talk() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}

	// Calculate dynamic cooldown based on current happiness (urgency)
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "talk", p.Happiness)
	if left := cooldownLeft(p.LastTalkedAt, cooldown); left != "" {
		return failResult(fmt.Sprintf("宠物需要消化一下，%s后可以再聊", left))
	}

	// Get effects from plugin or defaults
	_, happinessGain, _, _ := GetActionEffects(p.registry, p.Species, "talk")

	ch := make(map[string][2]int)
	oldHp := p.Happiness
	p.Happiness = clamp(p.Happiness+diminish(happinessGain, p.Happiness), 0, 100)
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	p.DialogueCount++
	p.TotalInteractions++
	p.AccHappiness++
	p.LastTalkedAt = time.Now()
	p.trackTimeOfDay()
	return ActionResult{OK: true, Message: "聊天愉快！", Changes: ch}
}

// Rest lets the pet sleep/rest, recovering energy and a small amount of health.
// Dynamic cooldown based on urgency. Prerequisite: energy < 90. Diminishing returns on energy gain.
func (p *Pet) Rest() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}

	// Calculate dynamic cooldown based on current energy (urgency)
	// Note: for energy, higher value = less urgent, so we use (100 - Energy)
	urgencyValue := 100 - p.Energy
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "rest", urgencyValue)
	if left := cooldownLeft(p.LastRestedAt, cooldown); left != "" {
		return failResult(fmt.Sprintf("宠物还不困，%s后可以再休息", left))
	}
	if p.Energy >= 90 {
		return failResult("宠物精力充沛，不需要休息！")
	}

	// Get effects from plugin or defaults
	_, happinessLoss, healthGain, energyGain := GetActionEffects(p.registry, p.Species, "rest")

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
	return ActionResult{OK: true, Message: "休息一下～", Changes: ch}
}

// Heal treats the pet, recovering health but costing energy.
// Dynamic cooldown based on urgency. Prerequisite: energy >= cost. Diminishing returns on health gain.
func (p *Pet) Heal() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}

	// Get energy cost from plugin or default
	energyCost := GetActionEnergyCost(p.registry, p.Species, "heal")
	if energyCost == 0 {
		energyCost = 15 // fallback
	}

	// Calculate dynamic cooldown based on current health (urgency)
	// Note: for health, higher value = less urgent, so we use (100 - Health)
	urgencyValue := 100 - p.Health
	cooldown := CalculateDynamicCooldown(p.registry, p.Species, "heal", urgencyValue)
	if left := cooldownLeft(p.LastHealedAt, cooldown); left != "" {
		return failResult(fmt.Sprintf("刚治疗过，%s后可以再治疗", left))
	}
	if p.Energy < energyCost {
		return failResult("宠物精力不足，需要先休息！")
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
	p.AccHealth++
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

	default:
		return "", fmt.Errorf("unknown field %q\n\nValid fields:\n  Attributes: hunger, happiness, health, energy (0-100)\n  Info: name, species, stage_id, alive, age_hours\n  Stats: interactions, feeds, dialogues, adventures, wins\n  Evolution: acc_happiness, acc_health, acc_playful, night, day, feed_regularity\n  Lifecycle: lifecycle_warning", field)
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

// ApplyOfflineDecay calculates the time elapsed since LastCheckedAt
// and applies the corresponding attribute decay. Should be called
// when loading a pet from a save file.
func (p *Pet) ApplyOfflineDecay() {
	if !p.Alive {
		return
	}
	elapsed := time.Since(p.LastCheckedAt)
	if elapsed < time.Minute {
		return
	}
	p.SimulateDecay(elapsed)
}
