package main

import (
	"clipet/internal/game"
	"clipet/internal/tui/dev"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

func newTimeskipCmd() *cobra.Command {
	var hours float64
	var days float64

	cmd := &cobra.Command{
		Use:   "timeskip",
		Short: "[开发] 时间跳跃 - 模拟时间流逝和属性衰减",
		Long: `时间跳跃：模拟时间流逝对宠物的影响。

带参数直接执行: timeskip --hours 24 或 --days 7
不带参数进入交互式界面，输入小时数后预览属性变化。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pet, err := loadPet()
			if err != nil {
				return err
			}

			// Direct mode with flags
			if hours != 0 || days != 0 {
				totalDuration := time.Duration(hours*float64(time.Hour)) + time.Duration(days*24*float64(time.Hour))
				return doTimeskip(pet, totalDuration)
			}

			// Interactive mode
			return runTimeskipTUI(pet)
		},
	}

	cmd.Flags().Float64Var(&hours, "hours", 0, "hours to skip")
	cmd.Flags().Float64Var(&days, "days", 0, "days to skip")

	return cmd
}

func doTimeskip(pet *game.Pet, dur time.Duration) error {
	oldAge := pet.AgeHours()
	oldHunger := pet.Hunger
	oldHappiness := pet.Happiness
	oldHealth := pet.Health
	oldEnergy := pet.Energy

	pet.Birthday = pet.Birthday.Add(-dur)
	// Use dev-only decay that doesn't trigger death/evolution hooks
	pet.DevOnlySimulateDecay(dur)

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Println("timeskip done")
	fmt.Printf("  elapsed: %.1f hours\n", dur.Hours())
	fmt.Printf("  age:     %.1fh -> %.1fh\n", oldAge, pet.AgeHours())
	fmt.Printf("  hunger:  %d -> %d\n", oldHunger, pet.Hunger)
	fmt.Printf("  happy:   %d -> %d\n", oldHappiness, pet.Happiness)
	fmt.Printf("  health:  %d -> %d\n", oldHealth, pet.Health)
	fmt.Printf("  energy:  %d -> %d\n", oldEnergy, pet.Energy)
	// Note: No death check in dev mode

	// Note: dev commands do not trigger evolution checks
	return nil
}

func runTimeskipTUI(pet *game.Pet) error {
	m := dev.NewTimeskipModel(pet, registry)

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	fm := finalModel.(*dev.TimeskipModel)
	if fm.Done {
		dur := time.Duration(fm.PreviewHours * float64(time.Hour))
		return doTimeskip(fm.Pet, dur)
	}
	return nil
}
