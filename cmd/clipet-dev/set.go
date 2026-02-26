package main

import (
	"clipet/internal/game"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// settableField describes one editable pet attribute.
type settableField struct {
	key   string
	label string
	kind  string // "int" or "string" or "bool"
}

var settableFields = []settableField{
	{"hunger", "饱腹", "int"},
	{"happiness", "快乐", "int"},
	{"health", "健康", "int"},
	{"energy", "精力", "int"},
	{"name", "名字", "string"},
	{"species", "物种", "string"},
	{"stage_id", "阶段ID", "string"},
	{"alive", "存活", "bool"},
}

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [attribute] [value]",
		Short: "[dev] Set pet attribute directly",
		Long: `直接修改宠物属性。

不带参数进入交互式界面，显示所有属性及当前值，选择后输入新值。
带参数直接执行：set hunger 100

可设属性: hunger, happiness, health, energy (0-100)
          name, species, stage_id (字符串)
          alive (true/false)`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePet(); err != nil {
				return err
			}

			pet, err := petStore.Load()
			if err != nil {
				return fmt.Errorf("load pet: %w", err)
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
				checkEvoAfterChange(pet)
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

func checkEvoAfterChange(pet *game.Pet) {
	candidates := game.CheckEvolution(pet, registry)
	if len(candidates) > 0 {
		best := game.BestCandidate(candidates)
		if best != nil {
			oldID := pet.StageID
			game.DoEvolve(pet, *best)
			_ = petStore.Save(pet)
			fmt.Printf("evolve: %s -> %s (%s)\n", oldID, best.ToStage.ID, best.ToStage.Phase)
		}
	}
}

// ---------- Interactive set TUI ----------

type setPhase int

const (
	setPhaseSelect setPhase = iota // choosing which field to edit
	setPhaseInput                  // typing new value
)

type setModel struct {
	pet      *game.Pet
	fields   []settableField
	cursor   int
	phase    setPhase
	input    string // text being typed
	width    int
	height   int
	quitting bool
	message  string // feedback message after edit
}

func runSetTUI(pet *game.Pet) error {
	m := setModel{
		pet:    pet,
		fields: settableFields,
	}

	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func (m setModel) Init() tea.Cmd { return nil }

func (m setModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if m.phase == setPhaseSelect {
			return m.updateSelect(msg)
		}
		return m.updateInput(msg)
	}
	return m, nil
}

func (m setModel) updateSelect(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "escape":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.fields)-1 {
			m.cursor++
		}
	case "enter":
		m.phase = setPhaseInput
		m.input = m.getCurrentValue()
		m.message = ""
	}
	return m, nil
}

func (m setModel) updateInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		m.phase = setPhaseSelect
		m.input = ""
		m.message = ""
	case "enter":
		field := m.fields[m.cursor]
		old, err := m.pet.SetField(field.key, m.input)
		if err != nil {
			m.message = fmt.Sprintf("❌ %v", err)
		} else {
			if err := petStore.Save(m.pet); err != nil {
				m.message = fmt.Sprintf("❌ save: %v", err)
			} else {
				m.message = fmt.Sprintf("✓ %s: %s → %s", field.label, old, m.input)
				checkEvoAfterChange(m.pet)
			}
		}
		m.phase = setPhaseSelect
		m.input = ""
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		ch := msg.String()
		if len(ch) == 1 {
			m.input += ch
		}
	}
	return m, nil
}

func (m setModel) getCurrentValue() string {
	f := m.fields[m.cursor]
	switch f.key {
	case "hunger":
		return strconv.Itoa(m.pet.Hunger)
	case "happiness":
		return strconv.Itoa(m.pet.Happiness)
	case "health":
		return strconv.Itoa(m.pet.Health)
	case "energy":
		return strconv.Itoa(m.pet.Energy)
	case "name":
		return m.pet.Name
	case "species":
		return m.pet.Species
	case "stage_id":
		return m.pet.StageID
	case "alive":
		return strconv.FormatBool(m.pet.Alive)
	}
	return ""
}

func (m setModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	if m.width == 0 {
		v := tea.NewView("加载中...")
		v.AltScreen = true
		return v
	}

	title := setHeaderStyle.Render(fmt.Sprintf(" %s [%s] ", m.pet.Name, m.pet.StageID))

	var lines []string
	for i, f := range m.fields {
		val := m.getFieldDisplay(f)
		prefix := "  "
		style := setItemStyle

		if i == m.cursor {
			prefix = "▸ "
			style = setSelStyle
		}

		line := fmt.Sprintf("%s%-8s %s", prefix, f.label, val)
		lines = append(lines, style.Render(line))
	}

	// Input area
	inputArea := ""
	if m.phase == setPhaseInput {
		f := m.fields[m.cursor]
		inputArea = "\n" + setInputLabelStyle.Render(fmt.Sprintf("编辑 %s:", f.label)) +
			"\n> " + m.input + "█" +
			"\n" + setInfoStyle.Render("Enter确认  Esc取消")
	}

	// Message
	msgArea := ""
	if m.message != "" {
		msgArea = "\n" + m.message
	}

	help := setInfoStyle.Render("↑↓选择  Enter编辑  q退出")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		strings.Join(lines, "\n"),
		inputArea,
		msgArea,
		"", help,
	)

	panel := setPanelStyle.
		Width(m.width - 2).
		Height(m.height - 1).
		Render(content)

	v := tea.NewView(panel)
	v.AltScreen = true
	return v
}

func (m setModel) getFieldDisplay(f settableField) string {
	val := ""
	switch f.key {
	case "hunger":
		val = fmt.Sprintf("%d", m.pet.Hunger)
	case "happiness":
		val = fmt.Sprintf("%d", m.pet.Happiness)
	case "health":
		val = fmt.Sprintf("%d", m.pet.Health)
	case "energy":
		val = fmt.Sprintf("%d", m.pet.Energy)
	case "name":
		val = m.pet.Name
	case "species":
		val = m.pet.Species
	case "stage_id":
		val = m.pet.StageID
	case "alive":
		val = strconv.FormatBool(m.pet.Alive)
	}

	if f.kind == "int" {
		n, _ := strconv.Atoi(val)
		bar := renderBar(n, 100, 20)
		return fmt.Sprintf("%-6s %s", val, bar)
	}
	return val
}

func renderBar(val, maxVal, width int) string {
	filled := val * width / maxVal
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return setBarFilledStyle.Render(strings.Repeat("█", filled)) +
		setBarEmptyStyle.Render(strings.Repeat("░", width-filled))
}

var (
	setPanelStyle = lipgloss.NewStyle().Padding(0, 1)

	setHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	setInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555570"))

	setItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EAEAEA"))

	setSelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)

	setInputLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)

	setBarFilledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575"))

	setBarEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#2A2A4A"))
)
