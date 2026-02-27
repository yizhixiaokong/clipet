package main

import (
	"clipet/internal/game"
	"clipet/internal/tui/dev"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

// settableField describes one editable pet attribute.
type settableField = dev.SetField

var settableFields = []settableField{
	// Core attributes (0-100)
	{Key: "hunger", Label: "饱腹", Kind: "int"},
	{Key: "happiness", Label: "快乐", Kind: "int"},
	{Key: "health", Label: "健康", Kind: "int"},
	{Key: "energy", Label: "精力", Kind: "int"},

	// Basic info
	{Key: "name", Label: "名字", Kind: "string"},
	{Key: "species", Label: "物种", Kind: "string"},
	{Key: "stage_id", Label: "阶段ID", Kind: "string"},
	{Key: "age_hours", Label: "年龄(小时)", Kind: "float"},
	{Key: "alive", Label: "存活", Kind: "bool"},

	// Statistics
	{Key: "interactions", Label: "总互动", Kind: "int"},
	{Key: "feeds", Label: "喂食次数", Kind: "int"},
	{Key: "dialogues", Label: "对话次数", Kind: "int"},
	{Key: "adventures", Label: "冒险次数", Kind: "int"},
	{Key: "wins", Label: "游戏胜利", Kind: "int"},

	// Evolution accumulators
	{Key: "acc_happiness", Label: "快乐累积", Kind: "int"},
	{Key: "acc_health", Label: "健康累积", Kind: "int"},
	{Key: "acc_playful", Label: "玩耍累积", Kind: "int"},
	{Key: "night", Label: "夜间互动", Kind: "int"},
	{Key: "day", Label: "日间互动", Kind: "int"},

	// Lifecycle
	{Key: "lifecycle_warning", Label: "生命预警", Kind: "bool"},
}

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [attribute] [value]",
		Short: "[开发] 直接修改宠物属性",
		Long: `直接修改宠物属性。

不带参数进入交互式界面，显示所有属性及当前值，选择后输入新值。
带参数直接执行：set hunger 100

【核心属性】(0-100)
  hunger, happiness, health, energy

【基本信息】
  name, species, stage_id, alive
  age_hours (年龄，以小时为单位)

【统计数据】
  interactions (总互动次数)
  feeds (喂食次数)
  dialogues (对话次数)
  adventures (冒险完成次数)
  wins (游戏胜利次数)

【进化累积】
  acc_happiness, acc_health, acc_playful
  night (夜间互动), day (日间互动)

【生命周期】
  lifecycle_warning (是否显示过生命预警)

示例:
  set hunger 100        # 直接设置饱腹度
  set age_hours 100     # 设置年龄为100小时
  set interactions 500  # 修改总互动次数
  set adventures 30     # 设置冒险次数
  set night 50          # 设置夜间互动次数`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pet, err := loadPet()
			if err != nil {
				return err
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
				// Note: dev commands do not trigger evolution checks
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

// ---------- Interactive set TUI ----------

func runSetTUI(pet *game.Pet) error {
	// Build fields dynamically, including custom attributes
	fields := buildSettableFields(pet)

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

	// Note: dev commands do not trigger evolution checks

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Output changes after TUI exits
	if fm, ok := finalModel.(*dev.SetModel); ok && len(fm.Changes) > 0 {
		for _, change := range fm.Changes {
			fmt.Printf("set %s: %s -> %s\n", change.Field, change.Old, change.New)
		}
	}

	return nil
}

func getCurrentPetValue(pet *game.Pet, key string) string {
	switch key {
	// Core attributes
	case "hunger":
		return strconv.Itoa(pet.Hunger)
	case "happiness":
		return strconv.Itoa(pet.Happiness)
	case "health":
		return strconv.Itoa(pet.Health)
	case "energy":
		return strconv.Itoa(pet.Energy)

	// Basic info
	case "name":
		return pet.Name
	case "species":
		return pet.Species
	case "stage_id":
		return pet.StageID
	case "age_hours":
		return strconv.FormatFloat(pet.AgeHours(), 'f', 1, 64)
	case "alive":
		return strconv.FormatBool(pet.Alive)

	// Statistics
	case "interactions":
		return strconv.Itoa(pet.TotalInteractions)
	case "feeds":
		return strconv.Itoa(pet.FeedCount)
	case "dialogues":
		return strconv.Itoa(pet.DialogueCount)
	case "adventures":
		return strconv.Itoa(pet.AdventuresCompleted)
	case "wins":
		return strconv.Itoa(pet.GamesWon)

	// Evolution accumulators
	case "acc_happiness":
		return strconv.Itoa(pet.AccHappiness)
	case "acc_health":
		return strconv.Itoa(pet.AccHealth)
	case "acc_playful":
		return strconv.Itoa(pet.AccPlayful)
	case "night":
		return strconv.Itoa(pet.NightInteractions)
	case "day":
		return strconv.Itoa(pet.DayInteractions)

	// Lifecycle
	case "lifecycle_warning":
		return strconv.FormatBool(pet.LifecycleWarningShown)

	// Custom attributes (Phase 4)
	default:
		if strings.HasPrefix(key, "custom:") {
			attrName := strings.TrimPrefix(key, "custom:")
			return strconv.Itoa(pet.GetAttr(attrName))
		}
	}
	return ""
}

// buildSettableFields builds the list of settable fields including custom attributes
func buildSettableFields(pet *game.Pet) []dev.SetField {
	fields := make([]dev.SetField, len(settableFields))
	for i, f := range settableFields {
		fields[i] = dev.SetField{
			Key:   f.Key,
			Label: f.Label,
			Kind:  f.Kind,
		}
	}

	// Add custom attributes dynamically (Phase 4)
	if pet.CustomAttributes != nil {
		for name := range pet.CustomAttributes {
			fields = append(fields, dev.SetField{
				Key:   "custom:" + name,
				Label: "自定义:" + name,
				Kind:  "int",
			})
		}
	}

	return fields
}
