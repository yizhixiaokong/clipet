package game

import (
	"clipet/internal/plugin"
)

// EvolveCandidate represents a single eligible evolution path.
type EvolveCandidate struct {
	Evolution plugin.Evolution
	ToStage   plugin.Stage
	Score     int // how many non-trivial conditions were satisfied
}

// CheckEvolution evaluates all evolution edges from the pet's current stage
// and returns candidates whose conditions are fully met.
func CheckEvolution(pet *Pet, reg *plugin.Registry) []EvolveCandidate {
	if !pet.Alive {
		return nil
	}

	evos := reg.GetEvolutionsFrom(pet.Species, pet.StageID)
	if len(evos) == 0 {
		return nil
	}

	var candidates []EvolveCandidate
	for _, evo := range evos {
		ok, score := evaluateCondition(pet, evo.Condition)
		if !ok {
			continue
		}
		toStage := reg.GetStage(pet.Species, evo.To)
		if toStage == nil {
			continue
		}
		candidates = append(candidates, EvolveCandidate{
			Evolution: evo,
			ToStage:   *toStage,
			Score:     score,
		})
	}
	return candidates
}

// evaluateCondition checks whether a pet meets all evolution conditions.
// Returns (met, score) where score counts the number of non-trivial conditions satisfied.
func evaluateCondition(pet *Pet, cond plugin.EvolutionCondition) (bool, int) {
	score := 0

	// min_age_hours
	if cond.MinAgeHours > 0 {
		if pet.AgeHours() < cond.MinAgeHours {
			return false, 0
		}
		score++
	}

	// attr_bias - check the corresponding accumulator
	if cond.AttrBias != "" {
		switch cond.AttrBias {
		case "happiness":
			if pet.AccHappiness <= 0 {
				return false, 0
			}
		case "health":
			if pet.AccHealth <= 0 {
				return false, 0
			}
		case "playful":
			if pet.AccPlayful <= 0 {
				return false, 0
			}
		}
		score++
	}

	// min_dialogues
	if cond.MinDialogues > 0 {
		if pet.DialogueCount < cond.MinDialogues {
			return false, 0
		}
		score++
	}

	// min_adventures
	if cond.MinAdventures > 0 {
		if pet.AdventuresCompleted < cond.MinAdventures {
			return false, 0
		}
		score++
	}

	// min_feed_regularity
	if cond.MinFeedRegularity > 0 {
		pet.UpdateFeedRegularity()
		if pet.FeedRegularity < cond.MinFeedRegularity {
			return false, 0
		}
		score++
	}

	// night_interactions_bias
	if cond.NightBias {
		if pet.NightInteractions <= pet.DayInteractions {
			return false, 0
		}
		score++
	}

	// day_interactions_bias
	if cond.DayBias {
		if pet.DayInteractions <= pet.NightInteractions {
			return false, 0
		}
		score++
	}

	// min_interactions
	if cond.MinInteractions > 0 {
		if pet.TotalInteractions < cond.MinInteractions {
			return false, 0
		}
		score++
	}

	// min_attr
	for attr, minVal := range cond.MinAttr {
		val := pet.GetAttr(attr)
		if val < minVal {
			return false, 0
		}
		score++
	}

	// custom_acc - check custom accumulators
	for accName, minVal := range cond.CustomAcc {
		val := pet.GetCustomAcc(accName)
		if val < minVal {
			return false, 0
		}
		score++
	}

	return true, score
}

// BestCandidate returns the single best evolution candidate.
// When multiple candidates qualify, the one with the highest score wins.
// Returns nil if no candidates.
func BestCandidate(candidates []EvolveCandidate) *EvolveCandidate {
	if len(candidates) == 0 {
		return nil
	}
	best := &candidates[0]
	for i := 1; i < len(candidates); i++ {
		if candidates[i].Score > best.Score {
			best = &candidates[i]
		}
	}
	return best
}

// DoEvolve executes an evolution, updating the pet's stage fields.
func DoEvolve(pet *Pet, candidate EvolveCandidate) {
	pet.StageID = candidate.ToStage.ID
	pet.Stage = PetStage(candidate.ToStage.Phase)
	// Reset accumulators for the new stage
	pet.AccHappiness = 0
	pet.AccHealth = 0
	pet.AccPlayful = 0
}
