package cli

import (
	"encoding/json"
	"fmt"
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
	pet, err := loadPet()
	if err != nil {
		return err
	}

	// Apply offline decay
	pet.AccumulateOfflineTime()
	_ = petStore.Save(pet)

	// Check and trigger evolution
	checkAndReportEvolution(pet)

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

	fmt.Printf("name=%s species=%s stage=%s(%s) age=%s alive=%t\n",
		pet.Name, speciesName, stageName, pet.StageID, ageStr, pet.Alive)
	fmt.Printf("hunger=%d happiness=%d health=%d energy=%d mood=%s(%d)\n",
		pet.Hunger, pet.Happiness, pet.Health, pet.Energy, pet.MoodName(), pet.MoodScore())
	fmt.Printf("interactions=%d games_won=%d adventures=%d dialogues=%d\n",
		pet.TotalInteractions, pet.GamesWon, pet.AdventuresCompleted, pet.DialogueCount)

	return nil
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
