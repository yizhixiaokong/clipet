package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPlayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "play",
		Short: "和宠物玩耍",
		RunE:  runPlay,
	}
}

func runPlay(cmd *cobra.Command, args []string) error {
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

	res := pet.Play()
	if !res.OK {
		fmt.Printf("play: %s\n", res.Message)
		return nil
	}

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}

	chH := res.Changes["happiness"]
	fmt.Printf("play: happiness %d -> %d, energy %d\n", chH[0], chH[1], pet.Energy)
	checkAndReportEvolution(pet)
	return nil
}
