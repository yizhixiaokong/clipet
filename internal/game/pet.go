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

// Pet is the central game entity representing the player's virtual pet.
type Pet struct {
	// Basic info
	Name    string   `json:"name"`
	Species string   `json:"species"`  // species pack ID, e.g. "cat"
	Stage   PetStage `json:"stage"`    // current life phase
	StageID string   `json:"stage_id"` // current evolution node ID, e.g. "baby_cat"

	Birthday time.Time `json:"birthday"`

	// Attributes (0-100)
	Hunger    int `json:"hunger"` // fullness, higher = less hungry
	Happiness int `json:"happiness"`
	Health    int `json:"health"`
	Energy    int `json:"energy"`

	// Timestamps
	LastFedAt     time.Time `json:"last_fed_at"`
	LastPlayedAt  time.Time `json:"last_played_at"`
	LastCheckedAt time.Time `json:"last_checked_at"`

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
func (p *Pet) Feed() {
	if !p.Alive {
		return
	}
	p.Hunger = clamp(p.Hunger+25, 0, 100)
	p.Happiness = clamp(p.Happiness+5, 0, 100)
	p.LastFedAt = time.Now()
	p.TotalInteractions++
	p.FeedCount++
	p.trackTimeOfDay()
}

// Play increases the pet's happiness and decreases energy.
func (p *Pet) Play() {
	if !p.Alive {
		return
	}
	p.Happiness = clamp(p.Happiness+20, 0, 100)
	p.Energy = clamp(p.Energy-10, 0, 100)
	p.AccPlayful++
	p.LastPlayedAt = time.Now()
	p.TotalInteractions++
	p.trackTimeOfDay()
}

// Talk records a dialogue interaction.
func (p *Pet) Talk() {
	if !p.Alive {
		return
	}
	p.Happiness = clamp(p.Happiness+5, 0, 100)
	p.DialogueCount++
	p.TotalInteractions++
	p.AccHappiness++
	p.trackTimeOfDay()
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

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

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
