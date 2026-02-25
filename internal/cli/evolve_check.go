package cli

import (
	"clipet/internal/game"
	"fmt"
)

// checkAndReportEvolution checks if the pet qualifies for evolution
// and automatically evolves using the best candidate.
// This is the CLI-mode evolution (auto-pick best match).
func checkAndReportEvolution(pet *game.Pet) {
	candidates := game.CheckEvolution(pet, registry)
	if len(candidates) == 0 {
		return
	}

	best := game.BestCandidate(candidates)
	if best == nil {
		return
	}

	oldStageID := pet.StageID
	game.DoEvolve(pet, *best)
	_ = petStore.Save(pet)

	fmt.Printf("\n✨ %s 进化了！%s → %s（%s）\n", pet.Name, oldStageID, best.ToStage.Name, best.ToStage.Phase)
}
