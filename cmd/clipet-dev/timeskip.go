package main

import (
	"clipet/internal/game"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newTimeskipCmd() *cobra.Command {
	var hours float64
	var days float64

	cmd := &cobra.Command{
		Use:   "timeskip",
		Short: "[dev] Time skip - simulate aging and attribute decay",
		Long:  "Shift the pet birthday backward and simulate attribute decay over the elapsed time.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if hours == 0 && days == 0 {
				return fmt.Errorf("specify at least --hours or --days")
			}
			if err := requirePet(); err != nil {
				return err
			}

			pet, err := petStore.Load()
			if err != nil {
				return fmt.Errorf("load pet: %w", err)
			}

			totalDuration := time.Duration(hours*float64(time.Hour)) + time.Duration(days*24*float64(time.Hour))

			oldAge := pet.AgeHours()
			oldHunger := pet.Hunger
			oldHappiness := pet.Happiness
			oldHealth := pet.Health
			oldEnergy := pet.Energy

			pet.Birthday = pet.Birthday.Add(-totalDuration)
			pet.SimulateDecay(totalDuration)

			if err := petStore.Save(pet); err != nil {
				return fmt.Errorf("save: %w", err)
			}

			fmt.Println("timeskip done")
			fmt.Printf("  elapsed: %.1f hours\n", totalDuration.Hours())
			fmt.Printf("  age:     %.1fh -> %.1fh\n", oldAge, pet.AgeHours())
			fmt.Printf("  hunger:  %d -> %d\n", oldHunger, pet.Hunger)
			fmt.Printf("  happy:   %d -> %d\n", oldHappiness, pet.Happiness)
			fmt.Printf("  health:  %d -> %d\n", oldHealth, pet.Health)
			fmt.Printf("  energy:  %d -> %d\n", oldEnergy, pet.Energy)
			if !pet.Alive {
				fmt.Println("  WARNING: pet died during timeskip!")
			}

			// Check evolution after timeskip
			candidates := game.CheckEvolution(pet, registry)
			if len(candidates) > 0 {
				best := game.BestCandidate(candidates)
				if best != nil {
					oldID := pet.StageID
					game.DoEvolve(pet, *best)
					_ = petStore.Save(pet)
					fmt.Printf("evolve: %s -> %s (%s)\n", oldID, best.ToStage.ID, best.ToStage.Phase)
				}
			}
			return nil
		},
	}

	cmd.Flags().Float64Var(&hours, "hours", 0, "hours to skip")
	cmd.Flags().Float64Var(&days, "days", 0, "days to skip")

	return cmd
}
