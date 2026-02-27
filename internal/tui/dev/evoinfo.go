// Package dev provides TUI models for clipet-dev commands
package dev

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/tui/components"
	"fmt"
)

// PrintEvolutionInfo prints evolution information for a pet
func PrintEvolutionInfo(pet *game.Pet, registry *plugin.Registry, species string) {
	pack := registry.GetSpecies(species)
	if pack == nil {
		fmt.Printf("错误: 找不到物种 %q\n", species)
		return
	}

	stage := registry.GetStage(species, pet.StageID)
	stageName := pet.StageID
	if stage != nil {
		stageName = stage.Name
	}

	fmt.Printf("=== %s (%s) [%s] ===\n", pet.Name, stageName, pet.StageID)
	fmt.Printf("年龄: %.1f 小时\n", pet.AgeHours())
	fmt.Println()

	evos := registry.GetEvolutionsFrom(species, pet.StageID)
	if len(evos) == 0 {
		fmt.Println("当前阶段没有进化路径（已到达终极形态）")
		return
	}

	for i, evo := range evos {
		toStage := registry.GetStage(species, evo.To)
		toName := evo.To
		toPhase := "?"
		if toStage != nil {
			toName = toStage.Name
			toPhase = toStage.Phase
		}

		fmt.Printf("--- 进化路径 %d: %s → %s (%s) ---\n", i+1, pet.StageID, evo.To, toPhase)
		fmt.Printf("    目标: %s\n", toName)

		cond := evo.Condition
		allMet := printConditionChecks(pet, cond)

		if allMet {
			fmt.Printf("    >>> 所有条件已满足！可以进化 <<<\n")
		}
		fmt.Println()
	}

	// Show full evolution tree summary
	fmt.Println("=== 完整进化树 ===")
	PrintEvolutionTree(pack, pet.StageID)
}

// printConditionChecks prints condition checks and returns whether all are met
func printConditionChecks(pet *game.Pet, cond plugin.EvolutionCondition) bool {
	allMet := true

	// min_age_hours
	if cond.MinAgeHours > 0 {
		met := pet.AgeHours() >= cond.MinAgeHours
		fmt.Printf("    %s 年龄 >= %.1f 小时 (当前: %.1f)\n", CheckMark(met), cond.MinAgeHours, pet.AgeHours())
		if !met {
			allMet = false
		}
	}

	// attr_bias
	if cond.AttrBias != "" {
		val := getAccumulator(pet, cond.AttrBias)
		met := val > 0
		fmt.Printf("    %s 属性偏好: %s 累积 > 0 (当前: %d)\n", CheckMark(met), cond.AttrBias, val)
		if !met {
			allMet = false
		}
	}

	// min_dialogues
	if cond.MinDialogues > 0 {
		met := pet.DialogueCount >= cond.MinDialogues
		fmt.Printf("    %s 对话次数 >= %d (当前: %d)\n", CheckMark(met), cond.MinDialogues, pet.DialogueCount)
		if !met {
			allMet = false
		}
	}

	// min_adventures
	if cond.MinAdventures > 0 {
		met := pet.AdventuresCompleted >= cond.MinAdventures
		fmt.Printf("    %s 冒险次数 >= %d (当前: %d)\n", CheckMark(met), cond.MinAdventures, pet.AdventuresCompleted)
		if !met {
			allMet = false
		}
	}

	// min_feed_regularity
	if cond.MinFeedRegularity > 0 {
		pet.UpdateFeedRegularity()
		met := pet.FeedRegularity >= cond.MinFeedRegularity
		fmt.Printf("    %s 喂食规律 >= %.1f (当前: %.2f)\n", CheckMark(met), cond.MinFeedRegularity, pet.FeedRegularity)
		if !met {
			allMet = false
		}
	}

	// night_interactions_bias
	if cond.NightBias {
		met := pet.NightInteractions > pet.DayInteractions
		fmt.Printf("    %s 夜间互动偏好 (夜: %d, 日: %d)\n", CheckMark(met), pet.NightInteractions, pet.DayInteractions)
		if !met {
			allMet = false
		}
	}

	// day_interactions_bias
	if cond.DayBias {
		met := pet.DayInteractions > pet.NightInteractions
		fmt.Printf("    %s 日间互动偏好 (日: %d, 夜: %d)\n", CheckMark(met), pet.DayInteractions, pet.NightInteractions)
		if !met {
			allMet = false
		}
	}

	// min_interactions
	if cond.MinInteractions > 0 {
		met := pet.TotalInteractions >= cond.MinInteractions
		fmt.Printf("    %s 总互动 >= %d (当前: %d)\n", CheckMark(met), cond.MinInteractions, pet.TotalInteractions)
		if !met {
			allMet = false
		}
	}

	// min_attr
	for attr, minVal := range cond.MinAttr {
		val := pet.GetAttr(attr)
		met := val >= minVal
		fmt.Printf("    %s %s >= %d (当前: %d)\n", CheckMark(met), AttrName(attr), minVal, val)
		if !met {
			allMet = false
		}
	}

	return allMet
}

// PrintEvolutionTree prints the evolution tree as plain text
func PrintEvolutionTree(pack *plugin.SpeciesPack, currentStageID string) {
	roots := buildEvoTreeFromPack(pack)

	tree := components.NewTreeList(roots)
	tree.MarkedID = currentStageID
	tree.ExpandAll()

	fmt.Print(tree.RenderPlain())
}

// buildEvoTreeFromPack creates TreeNode tree from SpeciesPack (for CLI output)
func buildEvoTreeFromPack(pack *plugin.SpeciesPack) []*components.TreeNode {
	stageMap := make(map[string]plugin.Stage)
	for _, stage := range pack.Stages {
		stageMap[stage.ID] = stage
	}

	// Build parent lookup from evolutions
	parentMap := make(map[string][]string)
	for _, evo := range pack.Evolutions {
		parentMap[evo.From] = append(parentMap[evo.From], evo.To)
	}

	// Find root stages
	var roots []string
	for stageID := range stageMap {
		hasParent := false
		for _, children := range parentMap {
			for _, child := range children {
				if child == stageID {
					hasParent = true
					break
				}
			}
			if hasParent {
				break
			}
		}
		if !hasParent {
			roots = append(roots, stageID)
		}
	}

	// Build tree
	var treeNodes []*components.TreeNode
	for _, rootID := range roots {
		node := buildStageNodeFromPack(rootID, stageMap, parentMap)
		if node != nil {
			treeNodes = append(treeNodes, node)
		}
	}

	return treeNodes
}

// buildStageNodeFromPack builds a TreeNode for a stage
func buildStageNodeFromPack(stageID string, stageMap map[string]plugin.Stage, parentMap map[string][]string) *components.TreeNode {
	stage, exists := stageMap[stageID]
	if !exists {
		return nil
	}

	label := fmt.Sprintf("%s [%s]", stage.Name, stage.ID)
	node := &components.TreeNode{
		ID:         stageID,
		Label:      label,
		Selectable: true,
		Expanded:   false,
		Data:       stage,
	}

	// Add children
	childIDs := parentMap[stageID]
	for _, childID := range childIDs {
		childNode := buildStageNodeFromPack(childID, stageMap, parentMap)
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}

	return node
}

// CheckMark returns a check or cross mark
func CheckMark(met bool) string {
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

// AttrName returns a human-readable attribute name
func AttrName(attr string) string {
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
