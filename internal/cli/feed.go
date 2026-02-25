package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "feed",
		Short: "喂食宠物",
		RunE:  runFeed,
	}
}

func runFeed(cmd *cobra.Command, args []string) error {
	if !petStore.Exists() {
		return fmt.Errorf("还没有宠物，请先运行 clipet init")
	}

	pet, err := petStore.Load()
	if err != nil {
		return fmt.Errorf("加载存档失败: %w", err)
	}

	if !pet.Alive {
		return fmt.Errorf("你的宠物已经不在了... 😢")
	}

	// Apply offline decay first
	pet.ApplyOfflineDecay()

	oldHunger := pet.Hunger
	pet.Feed()

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}

	fmt.Printf("🍖 %s 吃得很开心！饱腹度: %d → %d/100\n", pet.Name, oldHunger, pet.Hunger)

	// Check evolution
	checkAndReportEvolution(pet)
	return nil
}
