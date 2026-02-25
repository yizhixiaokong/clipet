package cli

import (
	"clipet/internal/game"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "创建一只新宠物",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	if petStore.Exists() {
		return fmt.Errorf("已经有一只宠物了！如需重新开始，请删除存档：%s", petStore.Path())
	}

	species := registry.ListSpecies()
	if len(species) == 0 {
		return fmt.Errorf("没有可用的物种包，请安装至少一个物种插件")
	}

	// Sort species by name
	sort.Slice(species, func(i, j int) bool {
		return species[i].Name < species[j].Name
	})

	// Display available species
	fmt.Println("🐾 欢迎来到 Clipet！让我们创建你的宠物。")
	fmt.Println()
	fmt.Println("可选物种：")
	for i, s := range species {
		source := ""
		if s.Source == "external" {
			source = " [外部插件]"
		}
		fmt.Printf("  %d. %s — %s%s\n", i+1, s.Name, s.Description, source)
	}
	fmt.Println()

	// Get species choice
	var choice int
	for {
		fmt.Printf("请选择物种 (1-%d): ", len(species))
		_, err := fmt.Scanln(&choice)
		if err != nil || choice < 1 || choice > len(species) {
			fmt.Println("无效选择，请重试。")
			continue
		}
		break
	}
	selected := species[choice-1]

	// Get pet name
	var name string
	for {
		fmt.Print("给你的宠物取个名字: ")
		_, err := fmt.Scanln(&name)
		if err != nil || name == "" {
			fmt.Println("名字不能为空，请重试。")
			continue
		}
		break
	}

	// Get base stats and egg stage
	baseStats := registry.GetBaseStats(selected.ID)
	eggStage := registry.GetEggStage(selected.ID)
	if baseStats == nil || eggStage == nil {
		return fmt.Errorf("物种 %q 数据不完整", selected.ID)
	}

	// Create pet
	pet := game.NewPet(name, selected.ID, eggStage.ID,
		baseStats.Hunger, baseStats.Happiness, baseStats.Health, baseStats.Energy)

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}

	fmt.Println()
	fmt.Printf("🥚 %s 的 %s 已诞生！\n", name, eggStage.Name)
	fmt.Printf("   物种: %s\n", selected.Name)
	fmt.Printf("   阶段: %s\n", eggStage.Name)
	fmt.Println()
	fmt.Println("运行 clipet 启动交互界面，或使用 clipet status 查看状态。")

	return nil
}
