package plugin

import (
	"fmt"
	"strings"
)

// ValidationError holds details about a plugin validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks a SpeciesPack for correctness and completeness.
// Returns a list of validation errors (empty if valid).
func Validate(pack *SpeciesPack) []ValidationError {
	var errs []ValidationError

	// Species metadata
	if pack.Species.ID == "" {
		errs = append(errs, ValidationError{"species.id", "required"})
	}
	if pack.Species.Name == "" {
		errs = append(errs, ValidationError{"species.name", "required"})
	}
	if pack.Species.Version == "" {
		errs = append(errs, ValidationError{"species.version", "required"})
	}

	// Stages
	if len(pack.Stages) == 0 {
		errs = append(errs, ValidationError{"stages", "at least one stage required"})
	}

	stageIDs := make(map[string]bool)
	hasEgg := false
	for i, stage := range pack.Stages {
		prefix := fmt.Sprintf("stages[%d]", i)
		if stage.ID == "" {
			errs = append(errs, ValidationError{prefix + ".id", "required"})
		}
		if stage.Name == "" {
			errs = append(errs, ValidationError{prefix + ".name", "required"})
		}
		if !ValidPhases[stage.Phase] {
			errs = append(errs, ValidationError{prefix + ".phase", fmt.Sprintf("invalid phase %q, must be one of: egg, baby, child, adult, legend", stage.Phase)})
		}
		if stageIDs[stage.ID] {
			errs = append(errs, ValidationError{prefix + ".id", fmt.Sprintf("duplicate stage ID %q", stage.ID)})
		}
		stageIDs[stage.ID] = true
		if stage.Phase == PhaseEgg {
			hasEgg = true
		}
	}

	if !hasEgg {
		errs = append(errs, ValidationError{"stages", "must have at least one egg stage"})
	}

	// Evolutions
	for i, evo := range pack.Evolutions {
		prefix := fmt.Sprintf("evolutions[%d]", i)
		if evo.From == "" {
			errs = append(errs, ValidationError{prefix + ".from", "required"})
		} else if !stageIDs[evo.From] {
			errs = append(errs, ValidationError{prefix + ".from", fmt.Sprintf("references unknown stage %q", evo.From)})
		}
		if evo.To == "" {
			errs = append(errs, ValidationError{prefix + ".to", "required"})
		} else if !stageIDs[evo.To] {
			errs = append(errs, ValidationError{prefix + ".to", fmt.Sprintf("references unknown stage %q", evo.To)})
		}
	}

	// Check evolution chain connectivity: every non-egg stage should be reachable
	reachable := make(map[string]bool)
	for _, stage := range pack.Stages {
		if stage.Phase == PhaseEgg {
			reachable[stage.ID] = true
		}
	}
	changed := true
	for changed {
		changed = false
		for _, evo := range pack.Evolutions {
			if reachable[evo.From] && !reachable[evo.To] {
				reachable[evo.To] = true
				changed = true
			}
		}
	}
	for _, stage := range pack.Stages {
		if !reachable[stage.ID] {
			errs = append(errs, ValidationError{"evolutions", fmt.Sprintf("stage %q is not reachable from any egg stage", stage.ID)})
		}
	}

	// Dialogues (optional but validate structure if present)
	for i, dg := range pack.Dialogues {
		prefix := fmt.Sprintf("dialogues[%d]", i)
		if len(dg.Lines) == 0 {
			errs = append(errs, ValidationError{prefix + ".lines", "must have at least one line"})
		}
		// Validate stage references (skip wildcards)
		for _, s := range dg.Stage {
			if s == "*" || strings.Contains(s, "*") {
				continue
			}
			if !stageIDs[s] {
				errs = append(errs, ValidationError{prefix + ".stage", fmt.Sprintf("references unknown stage %q", s)})
			}
		}
	}

	// Adventures (optional but validate structure if present)
	for i, adv := range pack.Adventures {
		prefix := fmt.Sprintf("adventures[%d]", i)
		if adv.ID == "" {
			errs = append(errs, ValidationError{prefix + ".id", "required"})
		}
		if len(adv.Choices) == 0 {
			errs = append(errs, ValidationError{prefix + ".choices", "must have at least one choice"})
		}
		for j, choice := range adv.Choices {
			cPrefix := fmt.Sprintf("%s.choices[%d]", prefix, j)
			if choice.Text == "" {
				errs = append(errs, ValidationError{cPrefix + ".text", "required"})
			}
			if len(choice.Outcomes) == 0 {
				errs = append(errs, ValidationError{cPrefix + ".outcomes", "must have at least one outcome"})
			}
		}
	}

	// Frames: check that at least egg idle frames exist
	eggStageID := ""
	for _, stage := range pack.Stages {
		if stage.Phase == PhaseEgg {
			eggStageID = stage.ID
			break
		}
	}
	if eggStageID != "" {
		key := FrameKey(eggStageID, "idle")
		if frame, ok := pack.Frames[key]; !ok || len(frame.Frames) == 0 {
			errs = append(errs, ValidationError{"frames", fmt.Sprintf("missing idle frames for egg stage %q", eggStageID)})
		}
	}

	return errs
}
