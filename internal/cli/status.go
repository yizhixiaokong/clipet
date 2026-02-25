package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "查看宠物状态",
		RunE:  runStatus,
	}
	cmd.Flags().BoolP("json", "j", false, "以 JSON 格式输出")
	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	if !petStore.Exists() {
		return fmt.Errorf("还没有宠物，请先运行 clipet init")
	}

	pet, err := petStore.Load()
	if err != nil {
		return fmt.Errorf("加载存档失败: %w", err)
	}

	// Apply offline decay
	pet.ApplyOfflineDecay()
	_ = petStore.Save(pet)

	jsonFlag, _ := cmd.Flags().GetBool("json")
	if jsonFlag {
		data, err := json.MarshalIndent(pet, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Get stage display name
	stageName := pet.StageID
	if stage := registry.GetStage(pet.Species, pet.StageID); stage != nil {
		stageName = stage.Name
	}

	// Get species display name
	speciesName := pet.Species
	if sp := registry.GetSpecies(pet.Species); sp != nil {
		speciesName = sp.Species.Name
	}

	// Format age
	age := time.Since(pet.Birthday)
	ageStr := formatDuration(age)

	// Status display
	alive := "✅ 存活"
	if !pet.Alive {
		alive = "💀 已死亡"
	}

	mood := moodEmoji(pet.MoodName()) + " " + moodChinese(pet.MoodName())

	fmt.Println("┌──────────────────────────────────┐")
	fmt.Printf("│  🐾 %s\n", pet.Name)
	fmt.Println("├──────────────────────────────────┤")
	fmt.Printf("│  物种: %-26s│\n", speciesName)
	fmt.Printf("│  阶段: %-26s│\n", stageName)
	fmt.Printf("│  年龄: %-26s│\n", ageStr)
	fmt.Printf("│  状态: %-26s│\n", alive)
	fmt.Printf("│  心情: %-26s│\n", mood)
	fmt.Println("├──────────────────────────────────┤")
	fmt.Printf("│  饱腹: %s %3d/100     │\n", bar(pet.Hunger), pet.Hunger)
	fmt.Printf("│  快乐: %s %3d/100     │\n", bar(pet.Happiness), pet.Happiness)
	fmt.Printf("│  健康: %s %3d/100     │\n", bar(pet.Health), pet.Health)
	fmt.Printf("│  精力: %s %3d/100     │\n", bar(pet.Energy), pet.Energy)
	fmt.Println("├──────────────────────────────────┤")
	fmt.Printf("│  总互动: %-24d│\n", pet.TotalInteractions)
	fmt.Printf("│  游戏胜场: %-22d│\n", pet.GamesWon)
	fmt.Printf("│  冒险完成: %-22d│\n", pet.AdventuresCompleted)
	fmt.Println("└──────────────────────────────────┘")

	return nil
}

func bar(val int) string {
	filled := val / 10
	empty := 10 - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%d天 %d小时", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时 %d分钟", hours, int(d.Minutes())%60)
	}
	return fmt.Sprintf("%d分钟", int(d.Minutes()))
}

func moodEmoji(mood string) string {
	switch mood {
	case "happy":
		return "😊"
	case "normal":
		return "😐"
	case "unhappy":
		return "😕"
	case "sad":
		return "😢"
	case "miserable":
		return "😭"
	default:
		return "❓"
	}
}

func moodChinese(mood string) string {
	switch mood {
	case "happy":
		return "开心"
	case "normal":
		return "普通"
	case "unhappy":
		return "不太好"
	case "sad":
		return "难过"
	case "miserable":
		return "很差"
	default:
		return "未知"
	}
}
