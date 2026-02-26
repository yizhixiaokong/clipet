package main

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newEvoInfoSubCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "查看当前阶段的进化条件及达成状态",
		Long:  "显示宠物当前阶段所有可能的进化路径，以及每个条件的达成情况。",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePet(); err != nil {
				return err
			}

			pet, err := petStore.Load()
			if err != nil {
				return fmt.Errorf("load pet: %w", err)
			}

			pack := registry.GetSpecies(pet.Species)
			if pack == nil {
				return fmt.Errorf("species %q not found", pet.Species)
			}

			stage := registry.GetStage(pet.Species, pet.StageID)
			stageName := pet.StageID
			if stage != nil {
				stageName = stage.Name
			}

			fmt.Printf("=== %s (%s) [%s] ===\n", pet.Name, stageName, pet.StageID)
			fmt.Printf("年龄: %.1f 小时\n", pet.AgeHours())
			fmt.Println()

			evos := registry.GetEvolutionsFrom(pet.Species, pet.StageID)
			if len(evos) == 0 {
				fmt.Println("当前阶段没有进化路径（已到达终极形态）")
				return nil
			}

			for i, evo := range evos {
				toStage := registry.GetStage(pet.Species, evo.To)
				toName := evo.To
				toPhase := "?"
				if toStage != nil {
					toName = toStage.Name
					toPhase = toStage.Phase
				}

				fmt.Printf("--- 进化路径 %d: %s → %s (%s) ---\n", i+1, pet.StageID, evo.To, toPhase)
				fmt.Printf("    目标: %s\n", toName)

				cond := evo.Condition
				allMet := true

				// min_age_hours
				if cond.MinAgeHours > 0 {
					met := pet.AgeHours() >= cond.MinAgeHours
					mark := checkMark(met)
					fmt.Printf("    %s 年龄 >= %.1f 小时 (当前: %.1f)\n", mark, cond.MinAgeHours, pet.AgeHours())
					if !met {
						allMet = false
					}
				}

				// attr_bias
				if cond.AttrBias != "" {
					val := getAccumulator(pet, cond.AttrBias)
					met := val > 0
					mark := checkMark(met)
					fmt.Printf("    %s 属性偏好: %s 累积 > 0 (当前: %d)\n", mark, cond.AttrBias, val)
					if !met {
						allMet = false
					}
				}

				// min_dialogues
				if cond.MinDialogues > 0 {
					met := pet.DialogueCount >= cond.MinDialogues
					mark := checkMark(met)
					fmt.Printf("    %s 对话次数 >= %d (当前: %d)\n", mark, cond.MinDialogues, pet.DialogueCount)
					if !met {
						allMet = false
					}
				}

				// min_adventures
				if cond.MinAdventures > 0 {
					met := pet.AdventuresCompleted >= cond.MinAdventures
					mark := checkMark(met)
					fmt.Printf("    %s 冒险次数 >= %d (当前: %d)\n", mark, cond.MinAdventures, pet.AdventuresCompleted)
					if !met {
						allMet = false
					}
				}

				// min_feed_regularity
				if cond.MinFeedRegularity > 0 {
					pet.UpdateFeedRegularity()
					met := pet.FeedRegularity >= cond.MinFeedRegularity
					mark := checkMark(met)
					fmt.Printf("    %s 喂食规律 >= %.1f (当前: %.2f)\n", mark, cond.MinFeedRegularity, pet.FeedRegularity)
					if !met {
						allMet = false
					}
				}

				// night_interactions_bias
				if cond.NightBias {
					met := pet.NightInteractions > pet.DayInteractions
					mark := checkMark(met)
					fmt.Printf("    %s 夜间互动偏好 (夜: %d, 日: %d)\n", mark, pet.NightInteractions, pet.DayInteractions)
					if !met {
						allMet = false
					}
				}

				// day_interactions_bias
				if cond.DayBias {
					met := pet.DayInteractions > pet.NightInteractions
					mark := checkMark(met)
					fmt.Printf("    %s 日间互动偏好 (日: %d, 夜: %d)\n", mark, pet.DayInteractions, pet.NightInteractions)
					if !met {
						allMet = false
					}
				}

				// min_interactions
				if cond.MinInteractions > 0 {
					met := pet.TotalInteractions >= cond.MinInteractions
					mark := checkMark(met)
					fmt.Printf("    %s 总互动 >= %d (当前: %d)\n", mark, cond.MinInteractions, pet.TotalInteractions)
					if !met {
						allMet = false
					}
				}

				// min_attr
				for attr, minVal := range cond.MinAttr {
					val := pet.GetAttr(attr)
					met := val >= minVal
					mark := checkMark(met)
					fmt.Printf("    %s %s >= %d (当前: %d)\n", mark, attrName(attr), minVal, val)
					if !met {
						allMet = false
					}
				}

				if allMet {
					fmt.Printf("    >>> 所有条件已满足！可以进化 <<<\n")
				}
				fmt.Println()
			}

			// Auto-evolve if any path is ready
			candidates := game.CheckEvolution(pet, registry)
			if len(candidates) > 0 {
				best := game.BestCandidate(candidates)
				if best != nil {
					oldID := pet.StageID
					game.DoEvolve(pet, *best)
					_ = petStore.Save(pet)
					fmt.Printf(">>> 自动进化: %s -> %s (%s) <<<\n\n", oldID, best.ToStage.ID, best.ToStage.Phase)
				}
			}

			// Show full evolution tree summary
			fmt.Println("=== 完整进化树 ===")
			printEvoTree(pack, pet.StageID)

			return nil
		},
	}
}

func checkMark(met bool) string {
	if met {
		return "[✓]"
	}
	return "[✗]"
}

func getAccumulator(pet *game.Pet, bias string) int {
	switch bias {
	case "happiness":
		return pet.AccHappiness
	case "health":
		return pet.AccHealth
	case "playful":
		return pet.AccPlayful
	default:
		return 0
	}
}

func attrName(attr string) string {
	switch attr {
	case "hunger":
		return "饱腹"
	case "happiness":
		return "快乐"
	case "health":
		return "健康"
	case "energy":
		return "精力"
	default:
		return attr
	}
}

func printEvoTree(pack *plugin.SpeciesPack, currentStageID string) {
	// Build adjacency: from -> []to
	adj := make(map[string][]string)
	for _, e := range pack.Evolutions {
		adj[e.From] = append(adj[e.From], e.To)
	}

	// Find root stages (egg phases)
	var roots []string
	for _, s := range pack.Stages {
		if s.Phase == "egg" {
			roots = append(roots, s.ID)
		}
	}

	// Build name lookup
	names := make(map[string]string)
	phases := make(map[string]string)
	for _, s := range pack.Stages {
		names[s.ID] = s.Name
		phases[s.ID] = s.Phase
	}

	var printNode func(id string, prefix string, isLast bool)
	printNode = func(id string, prefix string, isLast bool) {
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		marker := "  "
		if id == currentStageID {
			marker = "▸ "
		}

		name := names[id]
		if name == "" {
			name = id
		}
		phase := phases[id]

		fmt.Printf("%s%s%s%s [%s] (%s)\n", prefix, connector, marker, name, id, phase)

		children := adj[id]
		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}
		for i, child := range children {
			printNode(child, childPrefix, i == len(children)-1)
		}
	}

	for i, root := range roots {
		name := names[root]
		if name == "" {
			name = root
		}
		phase := phases[root]
		marker := "  "
		if root == currentStageID {
			marker = "▸ "
		}

		if i > 0 {
			fmt.Println()
		}
		// Print roots without connector
		children := adj[root]
		fmt.Printf("%s%s [%s] (%s)\n", marker, name, root, phase)
		for j, child := range children {
			printNode(child, "", j == len(children)-1)
		}
	}
}

// condSummary returns a short summary of an evolution condition.
func condSummary(cond plugin.EvolutionCondition) string {
	var parts []string
	if cond.MinAgeHours > 0 {
		parts = append(parts, fmt.Sprintf("年龄>=%.0fh", cond.MinAgeHours))
	}
	if cond.AttrBias != "" {
		parts = append(parts, fmt.Sprintf("偏好:%s", cond.AttrBias))
	}
	if cond.MinDialogues > 0 {
		parts = append(parts, fmt.Sprintf("对话>=%d", cond.MinDialogues))
	}
	if cond.MinAdventures > 0 {
		parts = append(parts, fmt.Sprintf("冒险>=%d", cond.MinAdventures))
	}
	if cond.MinFeedRegularity > 0 {
		parts = append(parts, fmt.Sprintf("喂食率>=%.1f", cond.MinFeedRegularity))
	}
	if cond.NightBias {
		parts = append(parts, "夜间偏好")
	}
	if cond.DayBias {
		parts = append(parts, "日间偏好")
	}
	if cond.MinInteractions > 0 {
		parts = append(parts, fmt.Sprintf("互动>=%d", cond.MinInteractions))
	}
	for attr, val := range cond.MinAttr {
		parts = append(parts, fmt.Sprintf("%s>=%d", attrName(attr), val))
	}
	if len(parts) == 0 {
		return "无条件"
	}
	return strings.Join(parts, ", ")
}
