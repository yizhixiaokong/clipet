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
	pet, err := loadPet()
	if err != nil {
		return err
	}

	if !pet.Alive {
		return fmt.Errorf("你的宠物已经不在了... 😢")
	}

	// Apply offline decay first
	pet.AccumulateOfflineTime()

	res := pet.Feed()
	if !res.OK {
		fmt.Printf("feed: %s\n", res.Message)
		return nil
	}

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}

	ch := res.Changes["hunger"]
	fmt.Printf("feed: hunger %d -> %d\n", ch[0], ch[1])
	checkAndReportEvolution(pet)
	return nil
}
