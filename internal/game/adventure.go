package game

import (
	"clipet/internal/plugin"
	"math/rand"
	"time"
)

// Adventure cooldown.
const CooldownAdventure = 10 * time.Minute

// AdventureCheckResult holds the result of CanAdventure check.
type AdventureCheckResult struct {
	OK        bool
	ErrorType string // standardized error type for i18n
	Message   string // human-readable feedback (for internal logs)
}

// AdventureResult holds the outcome of a completed adventure.
type AdventureResult struct {
	Adventure plugin.Adventure        // the adventure that was played
	Choice    plugin.AdventureChoice  // the player's chosen option
	Outcome   plugin.AdventureOutcome // the weighted random outcome
	Changes   map[string][2]int       // attr name -> {old, new}
}

// CanAdventure checks whether the pet is in a state to go on an adventure.
func CanAdventure(pet *Pet) AdventureCheckResult {
	if !pet.Alive {
		return AdventureCheckResult{
			OK:        false,
			ErrorType: ErrDead,
			Message:   "宠物已经不在了...",
		}
	}
	if pet.Energy < 15 {
		return AdventureCheckResult{
			OK:        false,
			ErrorType: ErrEnergyLow,
			Message:   "精力不足，需要至少15点精力才能冒险！",
		}
	}
	return AdventureCheckResult{OK: true}
}

// PickAdventure selects a random adventure available for the pet's current stage.
// Returns nil if no adventures are available.
func PickAdventure(pet *Pet, reg *plugin.Registry) *plugin.Adventure {
	adventures := reg.GetAdventures(pet.Species, pet.StageID)
	if len(adventures) == 0 {
		return nil
	}
	picked := adventures[rand.Intn(len(adventures))]
	return &picked
}

// ResolveOutcome picks a weighted random outcome from a choice.
func ResolveOutcome(choice plugin.AdventureChoice) plugin.AdventureOutcome {
	if len(choice.Outcomes) == 0 {
		return plugin.AdventureOutcome{Text: "什么都没发生...", Effects: nil}
	}

	totalWeight := 0
	for _, o := range choice.Outcomes {
		totalWeight += o.Weight
	}
	if totalWeight <= 0 {
		return choice.Outcomes[0]
	}

	roll := rand.Intn(totalWeight)
	cumulative := 0
	for _, o := range choice.Outcomes {
		cumulative += o.Weight
		if roll < cumulative {
			return o
		}
	}
	return choice.Outcomes[len(choice.Outcomes)-1]
}

// ApplyAdventureOutcome applies the outcome effects to the pet and returns
// the changes map. Energy cost (10) is always deducted.
func ApplyAdventureOutcome(pet *Pet, outcome plugin.AdventureOutcome) map[string][2]int {
	changes := make(map[string][2]int)
	const energyCost = 10

	// Record old values
	oldH := pet.Hunger
	oldHp := pet.Happiness
	oldHl := pet.Health
	oldE := pet.Energy

	// Deduct base energy cost
	pet.Energy = Clamp(pet.Energy-energyCost, 0, 100)

	// Apply effects from outcome
	for attr, delta := range outcome.Effects {
		switch attr {
		case "hunger":
			pet.Hunger = Clamp(pet.Hunger+delta, 0, 100)
		case "happiness":
			pet.Happiness = Clamp(pet.Happiness+delta, 0, 100)
		case "health":
			pet.Health = Clamp(pet.Health+delta, 0, 100)
		case "energy":
			pet.Energy = Clamp(pet.Energy+delta, 0, 100)
		default:
			// Apply to custom attributes (supports custom accumulators)
			// Record old value before modification
			oldCustom := pet.GetCustomAcc(attr)
			pet.AddCustomAcc(attr, delta)
			newCustom := pet.GetCustomAcc(attr)
			if newCustom != oldCustom {
				changes[attr] = [2]int{oldCustom, newCustom}
			}
		}
	}

	// Record changes (only attributes that actually changed)
	if pet.Hunger != oldH {
		changes["hunger"] = [2]int{oldH, pet.Hunger}
	}
	if pet.Happiness != oldHp {
		changes["happiness"] = [2]int{oldHp, pet.Happiness}
	}
	if pet.Health != oldHl {
		changes["health"] = [2]int{oldHl, pet.Health}
	}
	if pet.Energy != oldE {
		changes["energy"] = [2]int{oldE, pet.Energy}
	}

	// Update stats
	pet.AdventuresCompleted++
	pet.TotalInteractions++

	return changes
}
