// Package game contains the core game logic, independent of any UI framework.
package game

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
	Alive            bool      `json:"alive"`
	CurrentAnimation AnimState `json:"current_animation"`
}

// NewPet creates a new pet with the given name and species.
// It sets initial attributes from the provided base stats.
func NewPet(name, species, eggStageID string, hunger, happiness, health, energy int) *Pet {
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
		LastCheckedAt:    now,
		Alive:            true,
		CurrentAnimation: AnimIdle,
	}
}

// Feed increases the pet's hunger (fullness) level.
// Cooldown: 10min. Prerequisite: hunger < 95. Diminishing returns on gain.
func (p *Pet) Feed() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}
	if left := cooldownLeft(p.LastFedAt, CooldownFeed); left != "" {
		return failResult(fmt.Sprintf("宠物还不饿，%s后可以再喂", left))
	}
	if p.Hunger >= 95 {
		return failResult("宠物已经很饱了！")
	}
	ch := make(map[string][2]int)
	oldH := p.Hunger
	oldHp := p.Happiness
	p.Hunger = clamp(p.Hunger+diminish(25, p.Hunger), 0, 100)
	p.Happiness = clamp(p.Happiness+diminish(5, p.Happiness), 0, 100)
	ch["hunger"] = [2]int{oldH, p.Hunger}
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	p.LastFedAt = time.Now()
	p.TotalInteractions++
	p.FeedCount++
	p.trackTimeOfDay()
	return ActionResult{OK: true, Message: "喂食成功！", Changes: ch}
}

// Play increases the pet's happiness and decreases energy.
// Cooldown: 5min. Prerequisite: energy >= 10. Diminishing returns on happiness gain.
func (p *Pet) Play() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}
	if left := cooldownLeft(p.LastPlayedAt, CooldownPlay); left != "" {
		return failResult(fmt.Sprintf("宠物还在喘气，%s后可以再玩", left))
	}
	if p.Energy < 10 {
		return failResult("宠物太累了，先休息一下吧！")
	}
	ch := make(map[string][2]int)
	oldHp := p.Happiness
	oldE := p.Energy
	p.Happiness = clamp(p.Happiness+diminish(20, p.Happiness), 0, 100)
	p.Energy = clamp(p.Energy-10, 0, 100)
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	ch["energy"] = [2]int{oldE, p.Energy}
	p.AccPlayful++
	p.LastPlayedAt = time.Now()
	p.TotalInteractions++
	p.trackTimeOfDay()
	return ActionResult{OK: true, Message: "玩耍愉快！", Changes: ch}
}

// Talk records a dialogue interaction.
// Cooldown: 2min. Diminishing returns on happiness gain.
func (p *Pet) Talk() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}
	if left := cooldownLeft(p.LastTalkedAt, CooldownTalk); left != "" {
		return failResult(fmt.Sprintf("宠物需要消化一下，%s后可以再聊", left))
	}
	ch := make(map[string][2]int)
	oldHp := p.Happiness
	p.Happiness = clamp(p.Happiness+diminish(5, p.Happiness), 0, 100)
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	p.DialogueCount++
	p.TotalInteractions++
	p.AccHappiness++
	p.LastTalkedAt = time.Now()
	p.trackTimeOfDay()
	return ActionResult{OK: true, Message: "聊天愉快！", Changes: ch}
}

// Rest lets the pet sleep/rest, recovering energy and a small amount of health.
// Cooldown: 15min. Prerequisite: energy < 90. Diminishing returns on energy gain.
func (p *Pet) Rest() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}
	if left := cooldownLeft(p.LastRestedAt, CooldownRest); left != "" {
		return failResult(fmt.Sprintf("宠物还不困，%s后可以再休息", left))
	}
	if p.Energy >= 90 {
		return failResult("宠物精力充沛，不需要休息！")
	}
	ch := make(map[string][2]int)
	oldE := p.Energy
	oldH := p.Health
	oldHp := p.Happiness
	p.Energy = clamp(p.Energy+diminish(30, p.Energy), 0, 100)
	p.Health = clamp(p.Health+diminish(5, p.Health), 0, 100)
	p.Happiness = clamp(p.Happiness-5, 0, 100)
	ch["energy"] = [2]int{oldE, p.Energy}
	ch["health"] = [2]int{oldH, p.Health}
	ch["happiness"] = [2]int{oldHp, p.Happiness}
	p.LastRestedAt = time.Now()
	p.TotalInteractions++
	p.trackTimeOfDay()
	return ActionResult{OK: true, Message: "休息一下～", Changes: ch}
}

// Heal treats the pet, recovering health but costing energy.
// Cooldown: 20min. Prerequisite: energy >= 10. Diminishing returns on health gain.
func (p *Pet) Heal() ActionResult {
	if !p.Alive {
		return failResult("宠物已经不在了...")
	}
	if left := cooldownLeft(p.LastHealedAt, CooldownHeal); left != "" {
		return failResult(fmt.Sprintf("刚治疗过，%s后可以再治疗", left))
	}
	if p.Energy < 10 {
		return failResult("宠物精力不足，需要先休息！")
	}
	ch := make(map[string][2]int)
	oldH := p.Health
	oldE := p.Energy
	p.Health = clamp(p.Health+diminish(25, p.Health), 0, 100)
	p.Energy = clamp(p.Energy-15, 0, 100)
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

// SimulateDecay applies time-based attribute decay over the given duration.
// Decay rates per hour: hunger -3, happiness -2, energy -1.
// If hunger drops below 20, health decays at -0.5/hr.
// If health reaches 0, the pet dies.
func (p *Pet) SimulateDecay(elapsed time.Duration) {
	if !p.Alive {
		return
	}
	hours := elapsed.Hours()
	p.Hunger = clamp(p.Hunger-int(3*hours), 0, 100)
	p.Happiness = clamp(p.Happiness-int(2*hours), 0, 100)
	p.Energy = clamp(p.Energy-int(1*hours), 0, 100)
	if p.Hunger < 20 {
		p.Health = clamp(p.Health-int(0.5*hours), 0, 100)
	}
	if p.Health <= 0 {
		p.Alive = false
	}

	// Advance all cooldown timestamps
	p.LastFedAt = p.LastFedAt.Add(elapsed)
	p.LastPlayedAt = p.LastPlayedAt.Add(elapsed)
	p.LastRestedAt = p.LastRestedAt.Add(elapsed)
	p.LastHealedAt = p.LastHealedAt.Add(elapsed)
	p.LastTalkedAt = p.LastTalkedAt.Add(elapsed)
	p.LastAdventureAt = p.LastAdventureAt.Add(elapsed)
	p.LastCheckedAt = time.Now()
}

// SetField sets a pet field by name from a raw string value.
// Returns the previous value as a string for display purposes.
func (p *Pet) SetField(field string, raw string) (old string, err error) {
	switch strings.ToLower(field) {
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
	default:
		return "", fmt.Errorf("unknown field %q; valid: hunger, happiness, health, energy, name, species, stage_id, alive", field)
	}
	return old, nil
}

// GetAttr returns a named attribute value (hunger, happiness, health, energy).
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
		return 0
	}
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
