package main

import (
	"clipet/internal/game"
	"fmt"

	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <attribute> <value>",
		Short: "[dev] Set pet attribute directly",
		Long: `Directly modify a pet attribute in the save file.

Settable attributes:
  hunger, happiness, health, energy  (0-100 integer)
  name, species, stage_id           (string)
  alive                              (true/false)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePet(); err != nil {
				return err
			}

			pet, err := petStore.Load()
			if err != nil {
				return fmt.Errorf("load pet: %w", err)
			}

			field := args[0]
			value := args[1]

			old, err := pet.SetField(field, value)
			if err != nil {
				return fmt.Errorf("set %s: %w", field, err)
			}

			if err := petStore.Save(pet); err != nil {
				return fmt.Errorf("save: %w", err)
			}

			fmt.Printf("set %s: %s -> %s\n", field, old, value)

			// Check evolution after set
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
}
