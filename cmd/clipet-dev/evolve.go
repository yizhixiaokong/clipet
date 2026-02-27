package main

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/tui/dev"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

// newEvoCmd creates the parent "evo" command with subcommands "to" and "info".
func newEvoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evo",
		Short: "[开发] 进化相关命令",
		Long:  "进化相关的开发工具集，包含强制进化和查看进化信息子命令。",
	}

	cmd.AddCommand(newEvoToCmd())
	cmd.AddCommand(newEvoInfoCmd())

	return cmd
}

// newEvoToCmd creates the "evo to" subcommand for forced evolution.
func newEvoToCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "to [stage-id]",
		Short: "[开发] 强制进化到指定阶段",
		Long: `交互式选择目标阶段并强制进化。

不带参数进入交互式界面，显示所有可选的阶段树。
带参数直接执行：evo to adult.cat.warrior`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pet, err := loadPet()
			if err != nil {
				return err
			}

			// Direct mode
			if len(args) == 1 {
				return doEvolve(pet, args[0])
			}

			// Interactive mode
			return runEvoTUI(pet, registry)
		},
	}
}

func doEvolve(pet *game.Pet, toStageID string) error {
	stage := registry.GetStage(pet.Species, toStageID)
	if stage == nil {
		return fmt.Errorf("stage not found: %s", toStageID)
	}

	oldID := pet.StageID
	oldPhase := string(pet.Stage)

	pet.StageID = toStageID
	pet.Stage = game.PetStage(stage.Phase)

	// Reset accumulators for the new stage
	pet.AccHappiness = 0
	pet.AccHealth = 0
	pet.AccPlayful = 0

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Printf("evolve: %s -> %s (%s -> %s)\n", oldID, stage.ID, oldPhase, stage.Phase)
	return nil
}

func runEvoTUI(pet *game.Pet, registry *plugin.Registry) error {
	m := dev.NewEvolveModel(pet, pet.Species, registry)

	m.OnEvolve = func(toStageID string) error {
		return doEvolve(pet, toStageID)
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Output evolution result after TUI exits
	if fm, ok := finalModel.(*dev.EvolveModel); ok && fm.EvolveResult != "" {
		fmt.Println(fm.EvolveResult)
	}

	return nil
}

// newEvoInfoCmd creates the "evo info" subcommand for viewing evolution info.
func newEvoInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "[开发] 查看进化信息和条件",
		Long:  `打印当前宠物的所有进化路径和条件。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pet, err := loadPet()
			if err != nil {
				return err
			}

			printEvoInfo(pet, registry)
			return nil
		},
	}
}

func printEvoInfo(pet *game.Pet, registry *plugin.Registry) {
	// Print evolution info only (no auto-evolution)
	dev.PrintEvolutionInfo(pet, registry, pet.Species)
}
