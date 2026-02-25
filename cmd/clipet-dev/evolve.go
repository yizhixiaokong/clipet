package main

import (
	"clipet/internal/game"
	"fmt"

	"github.com/spf13/cobra"
)

func newEvolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "evolve <stage-id>",
		Short: "[开发] 强制进化到指定阶段",
		Long:  "跳过所有进化条件，直接将宠物设置为指定的 stage ID。",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePet(); err != nil {
				return err
			}

			pet, err := petStore.Load()
			if err != nil {
				return fmt.Errorf("load pet: %w", err)
			}

			targetID := args[0]
			stage := registry.GetStage(pet.Species, targetID)
			if stage == nil {
				// List available stages for this species
				pack := registry.GetSpecies(pet.Species)
				if pack == nil {
					return fmt.Errorf("species %q not found in registry", pet.Species)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "stage %q not found for species %q\n\n", targetID, pet.Species)
				fmt.Fprintln(cmd.ErrOrStderr(), "可用阶段:")
				for _, s := range pack.Stages {
					fmt.Fprintf(cmd.ErrOrStderr(), "  %-25s (%s)\n", s.ID, s.Phase)
				}
				return fmt.Errorf("invalid stage ID %q", targetID)
			}

			oldStageID := pet.StageID
			oldPhase := string(pet.Stage)

			pet.StageID = targetID
			pet.Stage = game.PetStage(stage.Phase)

			if err := petStore.Save(pet); err != nil {
				return fmt.Errorf("save: %w", err)
			}

			fmt.Printf("🔄 进化完成: %s (%s) → %s (%s)\n", oldStageID, oldPhase, pet.StageID, stage.Phase)
			fmt.Printf("   阶段名称: %s\n", stage.Name)
			return nil
		},
	}
}
