package main

import (
	"clipet/internal/game"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func newTimeskipCmd() *cobra.Command {
	var hours float64
	var days float64

	cmd := &cobra.Command{
		Use:   "timeskip",
		Short: "[dev] Time skip - simulate aging and attribute decay",
		Long: `时间跳跃：模拟时间流逝对宠物的影响。

带参数直接执行: timeskip --hours 24 或 --days 7
不带参数进入交互式界面，输入小时数后预览属性变化。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePet(); err != nil {
				return err
			}

			pet, err := petStore.Load()
			if err != nil {
				return fmt.Errorf("load pet: %w", err)
			}

			// Direct mode with flags
			if hours != 0 || days != 0 {
				totalDuration := time.Duration(hours*float64(time.Hour)) + time.Duration(days*24*float64(time.Hour))
				return doTimeskip(pet, totalDuration)
			}

			// Interactive mode
			return runTimeskipTUI(pet)
		},
	}

	cmd.Flags().Float64Var(&hours, "hours", 0, "hours to skip")
	cmd.Flags().Float64Var(&days, "days", 0, "days to skip")

	return cmd
}

func doTimeskip(pet *game.Pet, dur time.Duration) error {
	oldAge := pet.AgeHours()
	oldHunger := pet.Hunger
	oldHappiness := pet.Happiness
	oldHealth := pet.Health
	oldEnergy := pet.Energy

	pet.Birthday = pet.Birthday.Add(-dur)
	pet.SimulateDecay(dur)

	if err := petStore.Save(pet); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Println("timeskip done")
	fmt.Printf("  elapsed: %.1f hours\n", dur.Hours())
	fmt.Printf("  age:     %.1fh -> %.1fh\n", oldAge, pet.AgeHours())
	fmt.Printf("  hunger:  %d -> %d\n", oldHunger, pet.Hunger)
	fmt.Printf("  happy:   %d -> %d\n", oldHappiness, pet.Happiness)
	fmt.Printf("  health:  %d -> %d\n", oldHealth, pet.Health)
	fmt.Printf("  energy:  %d -> %d\n", oldEnergy, pet.Energy)
	if !pet.Alive {
		fmt.Println("  WARNING: pet died during timeskip!")
	}

	checkEvoAfterChange(pet)
	return nil
}

// ---------- Interactive timeskip TUI ----------

type tsPhase int

const (
	tsPhaseInput   tsPhase = iota // typing hours
	tsPhasePreview                // showing preview, confirm or cancel
)

type tsModel struct {
	pet    *game.Pet
	width  int
	height int

	phase    tsPhase
	input    string // hours input
	inputErr string

	// Preview data (computed from input)
	previewHours float64
	oldAge       float64
	newAge       float64
	oldStats     [4]int // hunger, happiness, health, energy
	newStats     [4]int
	wouldDie     bool

	quitting bool
	done     bool
}

func runTimeskipTUI(pet *game.Pet) error {
	m := tsModel{
		pet:   pet,
		input: "24",
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	fm := finalModel.(tsModel)
	if fm.done {
		dur := time.Duration(fm.previewHours * float64(time.Hour))
		return doTimeskip(fm.pet, dur)
	}
	return nil
}

func (m tsModel) Init() tea.Cmd { return nil }

func (m tsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if m.phase == tsPhaseInput {
			return m.updateInput(msg)
		}
		return m.updatePreview(msg)
	}
	return m, nil
}

func (m tsModel) updateInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "escape":
		m.quitting = true
		return m, tea.Quit
	case "enter":
		h, err := strconv.ParseFloat(m.input, 64)
		if err != nil || h <= 0 {
			m.inputErr = "请输入正数"
			return m, nil
		}
		m.inputErr = ""
		m.previewHours = h
		m.computePreview()
		m.phase = tsPhasePreview
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		m.inputErr = ""
	default:
		ch := msg.String()
		if len(ch) == 1 && (ch[0] >= '0' && ch[0] <= '9' || ch[0] == '.') {
			m.input += ch
			m.inputErr = ""
		}
	}
	return m, nil
}

func (m tsModel) updatePreview(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y":
		m.done = true
		return m, tea.Quit
	case "escape", "n", "q":
		m.phase = tsPhaseInput
	}
	return m, nil
}

func (m *tsModel) computePreview() {
	m.oldAge = m.pet.AgeHours()
	m.newAge = m.oldAge + m.previewHours
	m.oldStats = [4]int{m.pet.Hunger, m.pet.Happiness, m.pet.Health, m.pet.Energy}

	// Simulate decay on a copy
	hours := m.previewHours
	hunger := clampDev(m.pet.Hunger-int(3*hours), 0, 100)
	happiness := clampDev(m.pet.Happiness-int(2*hours), 0, 100)
	energy := clampDev(m.pet.Energy-int(1*hours), 0, 100)
	health := m.pet.Health
	if hunger < 20 {
		health = clampDev(health-int(0.5*hours), 0, 100)
	}
	m.newStats = [4]int{hunger, happiness, health, energy}
	m.wouldDie = health <= 0
}

func clampDev(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (m tsModel) View() tea.View {
	if m.quitting || m.done {
		return tea.NewView("")
	}
	if m.width == 0 {
		v := tea.NewView("加载中...")
		v.AltScreen = true
		return v
	}

	title := tsHeaderStyle.Render(fmt.Sprintf(" ⏩ 时间跳跃 — %s [%s] ", m.pet.Name, m.pet.StageID))

	var content string
	if m.phase == tsPhaseInput {
		content = m.viewInput()
	} else {
		content = m.viewPreview()
	}

	panel := tsPanelStyle.
		Width(m.width - 2).
		Height(m.height - 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", content))

	v := tea.NewView(panel)
	v.AltScreen = true
	return v
}

func (m tsModel) viewInput() string {
	statNames := []string{"饱腹", "快乐", "健康", "精力"}
	stats := []int{m.pet.Hunger, m.pet.Happiness, m.pet.Health, m.pet.Energy}

	var lines []string
	lines = append(lines, tsInfoStyle.Render(fmt.Sprintf("当前年龄: %.1f 小时", m.pet.AgeHours())))
	lines = append(lines, "")
	for i, name := range statNames {
		bar := renderBar(stats[i], 100, 20)
		lines = append(lines, fmt.Sprintf("  %-6s %3d %s", name, stats[i], bar))
	}
	lines = append(lines, "")
	lines = append(lines, tsInputLabelStyle.Render("跳过小时数:"))
	lines = append(lines, "> "+m.input+"█")

	if m.inputErr != "" {
		lines = append(lines, tsErrStyle.Render(m.inputErr))
	}

	lines = append(lines, "")
	lines = append(lines, tsInfoStyle.Render("Enter预览  q退出"))

	return strings.Join(lines, "\n")
}

func (m tsModel) viewPreview() string {
	statNames := []string{"饱腹", "快乐", "健康", "精力"}

	var lines []string
	lines = append(lines, tsInputLabelStyle.Render(fmt.Sprintf("跳过 %.1f 小时后的变化:", m.previewHours)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  年龄   %.1fh → %.1fh", m.oldAge, m.newAge))
	lines = append(lines, "")

	for i, name := range statNames {
		oldBar := renderBar(m.oldStats[i], 100, 15)
		newBar := renderBar(m.newStats[i], 100, 15)
		delta := m.newStats[i] - m.oldStats[i]
		deltaStr := fmt.Sprintf("%+d", delta)
		if delta < 0 {
			deltaStr = tsErrStyle.Render(deltaStr)
		}
		lines = append(lines, fmt.Sprintf("  %-6s %3d %s → %3d %s  %s",
			name, m.oldStats[i], oldBar, m.newStats[i], newBar, deltaStr))
	}

	if m.wouldDie {
		lines = append(lines, "")
		lines = append(lines, tsErrStyle.Render("  ⚠ 警告: 宠物将会死亡!"))
	}

	lines = append(lines, "")
	lines = append(lines, tsInputLabelStyle.Render("确认执行? (Enter/y 确认, Esc/n 返回)"))

	return strings.Join(lines, "\n")
}

var (
	tsPanelStyle = lipgloss.NewStyle().Padding(0, 1)

	tsHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#E94560")).
			Padding(0, 1)

	tsInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555570"))

	tsInputLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)

	tsErrStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6347")).
			Bold(true)
)
