package main

import (
	"clipet/internal/game"
	"clipet/internal/tui/dev"
	"fmt"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

// settableField describes one editable pet attribute.
type settableField = dev.SetField

var settableFields = []settableField{
	{Key: "hunger", Label: "饱腹", Kind: "int"},
	{Key: "happiness", Label: "快乐", Kind: "int"},
	{Key: "health", Label: "健康", Kind: "int"},
	{Key: "energy", Label: "精力", Kind: "int"},
	{Key: "name", Label: "名字", Kind: "string"},
	{Key: "species", Label: "物种", Kind: "string"},
	{Key: "stage_id", Label: "阶段ID", Kind: "string"},
	{Key: "alive", Label: "存活", Kind: "bool"},
}

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [attribute] [value]",
		Short: "[dev] Set pet attribute directly",
		Long: `直接修改宠物属性。

不带参数进入交互式界面，显示所有属性及当前值，选择后输入新值。
带参数直接执行：set hunger 100

可设属性: hunger, happiness, health, energy (0-100)
          name, species, stage_id (字符串)
          alive (true/false)`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePet(); err != nil {
				return err
			}

			pet, err := petStore.Load()
			if err != nil {
				return fmt.Errorf("load pet: %w", err)
			}

			// Direct mode
			if len(args) == 2 {
				old, err := pet.SetField(args[0], args[1])
				if err != nil {
					return fmt.Errorf("set %s: %w", args[0], err)
				}
				if err := petStore.Save(pet); err != nil {
					return fmt.Errorf("save: %w", err)
				}
				fmt.Printf("set %s: %s -> %s\n", args[0], old, args[1])
				checkEvoAfterChange(pet)
				return nil
			}

			if len(args) == 1 {
				return fmt.Errorf("需要 0 个或 2 个参数 (交互模式或 <attribute> <value>)")
			}

			// Interactive mode
			return runSetTUI(pet)
		},
	}
}

func checkEvoAfterChange(pet *game.Pet) {
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
}

// ---------- Interactive set TUI ----------

func runSetTUI(pet *game.Pet) error {
	// Convert to dev.SetField slice
	fields := make([]dev.SetField, len(settableFields))
	for i, f := range settableFields {
		fields[i] = dev.SetField{
			Key:   f.Key,
			Label: f.Label,
			Kind:  f.Kind,
		}
	}

	m := dev.NewSetModel(pet, fields)

	// Set callbacks
	m.GetCurrentValue = func(field dev.SetField) string {
		return getCurrentPetValue(pet, field.Key)
	}

	m.SetFieldValue = func(field dev.SetField, value string) (string, error) {
		old, err := pet.SetField(field.Key, value)
		if err != nil {
			return "", err
		}
		if err := petStore.Save(pet); err != nil {
			return "", fmt.Errorf("save: %w", err)
		}
		return old, nil
	}

	m.OnFieldChanged = func() {
		checkEvoAfterChange(pet)
	}

	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func getCurrentPetValue(pet *game.Pet, key string) string {
	switch key {
	case "hunger":
		return strconv.Itoa(pet.Hunger)
	case "happiness":
		return strconv.Itoa(pet.Happiness)
	case "health":
		return strconv.Itoa(pet.Health)
	case "energy":
		return strconv.Itoa(pet.Energy)
	case "name":
		return pet.Name
	case "species":
		return pet.Species
	case "stage_id":
		return pet.StageID
	case "alive":
		return strconv.FormatBool(pet.Alive)
	}
	return ""
}
