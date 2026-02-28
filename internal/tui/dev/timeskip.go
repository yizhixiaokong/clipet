// Package dev provides TUI models for clipet-dev commands
package dev

import (
	"clipet/internal/game"
	"clipet/internal/game/capabilities"
	"clipet/internal/plugin"
	"clipet/internal/tui/components"
	"clipet/internal/tui/styles"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// TimeskipInputKeyMap defines keybindings for timeskip command (input phase)
type TimeskipInputKeyMap struct {
	Enter      key.Binding
	Quit       key.Binding
	ToggleHelp key.Binding
}

// TimeskipPreviewKeyMap defines keybindings for timeskip command (preview phase)
type TimeskipPreviewKeyMap struct {
	Yes        key.Binding
	Cancel     key.Binding
	Quit       key.Binding
	ToggleHelp key.Binding
}

// DefaultTimeskipInputKeyMap returns default keybindings for input phase
var DefaultTimeskipInputKeyMap = TimeskipInputKeyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "预览"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q/Ctrl+C/Esc", "退出"),
	),
	ToggleHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "帮助"),
	),
}

// DefaultTimeskipPreviewKeyMap returns default keybindings for preview phase
var DefaultTimeskipPreviewKeyMap = TimeskipPreviewKeyMap{
	Yes: key.NewBinding(
		key.WithKeys("enter", "y"),
		key.WithHelp("Enter/y", "确认"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc", "n"),
		key.WithHelp("Esc/n", "返回"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/Ctrl+C", "退出"),
	),
	ToggleHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "帮助"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k TimeskipInputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Quit, k.ToggleHelp}
}

// FullHelp returns keybindings for the expanded help view
func (k TimeskipInputKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Quit},
		{k.ToggleHelp},
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k TimeskipPreviewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Yes, k.Cancel, k.ToggleHelp}
}

// FullHelp returns keybindings for the expanded help view
func (k TimeskipPreviewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Yes, k.Cancel},
		{k.Quit, k.ToggleHelp},
	}
}

// TimeskipModel is the TUI model for timeskip command
type TimeskipModel struct {
	Pet          *game.Pet
	Registry     *plugin.Registry
	Width        int
	Height       int
	Phase        timeskipPhase
	Input        *components.InputField
	InputErr     string
	InputKeyMap  TimeskipInputKeyMap
	PreviewKeyMap TimeskipPreviewKeyMap
	Help         help.Model

	// Preview data (computed from input)
	PreviewHours float64
	OldAge       float64
	NewAge       float64
	OldStats     [4]int // hunger, happiness, health, energy
	NewStats     [4]int
	WouldDie     bool

	Quitting bool
	Done     bool
}

type timeskipPhase int

const (
	timeskipPhaseInput   timeskipPhase = iota // typing hours
	timeskipPhasePreview                      // showing preview, confirm or cancel
)

// NewTimeskipModel creates a new timeskip TUI model
func NewTimeskipModel(pet *game.Pet, registry *plugin.Registry) *TimeskipModel {
	h := help.New()
	h.ShowAll = false

	return &TimeskipModel{
		Pet: pet,
		Registry: registry,
		Input: components.NewInputField().
			SetValue("24").
			SetFilter(components.NumericFilter(".")),
		Phase:         timeskipPhaseInput,
		InputKeyMap:   DefaultTimeskipInputKeyMap,
		PreviewKeyMap: DefaultTimeskipPreviewKeyMap,
		Help:          h,
	}
}

// Init implements tea.Model
func (m *TimeskipModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *TimeskipModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Help.SetWidth(m.Width)
	case tea.KeyPressMsg:
		if m.Phase == timeskipPhaseInput {
			return m.updateInput(msg)
		}
		return m.updatePreview(msg)
	}
	return m, nil
}

// View implements tea.Model
func (m *TimeskipModel) View() tea.View {
	if m.Quitting || m.Done {
		return tea.NewView("")
	}
	if m.Width == 0 {
		v := tea.NewView("加载中...")
		v.AltScreen = true
		return v
	}

	title := tsHeaderStyle.Render(fmt.Sprintf(" ⏩ 时间跳跃 — %s [%s] ", m.Pet.Name, m.Pet.StageID))

	var content string
	if m.Phase == timeskipPhaseInput {
		content = m.viewInput()
	} else {
		content = m.viewPreview()
	}

	panel := tsPanelStyle.
		Width(m.Width - 2).
		Height(m.Height - 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", content))

	v := tea.NewView(panel)
	v.AltScreen = true
	return v
}

func (m *TimeskipModel) updateInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.InputKeyMap.Quit):
		m.Quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.InputKeyMap.ToggleHelp):
		m.Help.ShowAll = !m.Help.ShowAll
		return m, nil
	case key.Matches(msg, m.InputKeyMap.Enter):
		h, err := strconv.ParseFloat(m.Input.Value(), 64)
		if err != nil || h <= 0 {
			m.InputErr = "请输入正数"
			return m, nil
		}
		m.InputErr = ""
		m.PreviewHours = h
		m.computePreview()
		m.Phase = timeskipPhasePreview
	default:
		// Delegate to InputField component
		m.Input, _ = m.Input.Update(msg)
		m.InputErr = ""
	}
	return m, nil
}

func (m *TimeskipModel) updatePreview(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.PreviewKeyMap.Yes):
		m.Done = true
		return m, tea.Quit
	case key.Matches(msg, m.PreviewKeyMap.Cancel):
		m.Phase = timeskipPhaseInput
	case key.Matches(msg, m.PreviewKeyMap.Quit):
		// q/Ctrl+C quit from preview mode
		m.Quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.PreviewKeyMap.ToggleHelp):
		m.Help.ShowAll = !m.Help.ShowAll
		return m, nil
	}
	return m, nil
}

func (m *TimeskipModel) computePreview() {
	// Calculate total offline time (existing + new)
	totalHours := m.Pet.AccumulatedOfflineDuration.Hours() + m.PreviewHours

	m.OldAge = m.Pet.AgeHours()
	m.NewAge = m.OldAge + totalHours
	m.OldStats = [4]int{m.Pet.Hunger, m.Pet.Happiness, m.Pet.Health, m.Pet.Energy}

	// Get decay config from plugin (same as DevOnlySimulateDecay)
	var decayConfig capabilities.DecayConfig
	if m.Registry != nil {
		decayConfig = m.Registry.GetDecayConfig(m.Pet.Species)
	} else {
		decayConfig = capabilities.DecayConfig{}.Defaults()
	}

	// Simulate decay using plugin-controlled rates (based on TOTAL offline time)
	hunger := clamp(m.Pet.Hunger-int(decayConfig.Hunger*totalHours), 0, 100)
	happiness := clamp(m.Pet.Happiness-int(decayConfig.Happiness*totalHours), 0, 100)
	energy := clamp(m.Pet.Energy-int(decayConfig.Energy*totalHours), 0, 100)
	health := m.Pet.Health
	if hunger < 20 {
		health = clamp(health-int(decayConfig.Health*totalHours), 0, 100)
	}

	m.NewStats = [4]int{hunger, happiness, health, energy}
	m.WouldDie = health <= 0
}

func (m *TimeskipModel) viewInput() string {
	statNames := []string{"饱腹", "快乐", "健康", "精力"}
	stats := []int{m.Pet.Hunger, m.Pet.Happiness, m.Pet.Health, m.Pet.Energy}

	var lines []string
	lines = append(lines, tsInfoStyle.Render(fmt.Sprintf("当前年龄: %.1f 小时", m.Pet.AgeHours())))

	// Show accumulated offline time if any
	if m.Pet.AccumulatedOfflineDuration > 0 {
		lines = append(lines, tsInfoStyle.Render(fmt.Sprintf("离线时间: %.1f 小时 (将在 TUI 启动时结算)", m.Pet.AccumulatedOfflineDuration.Hours())))
	}

	lines = append(lines, "")
	for i, name := range statNames {
		bar := components.NewProgressBar().
			SetValue(stats[i]).
			SetMax(100).
			SetWidth(20).
			SetFilledStyle(styles.DevCommandStyles.BarFilled).
			SetEmptyStyle(styles.DevCommandStyles.BarEmpty).
			Render()
		lines = append(lines, fmt.Sprintf("  %-6s %3d %s", name, stats[i], bar))
	}
	lines = append(lines, "")
	lines = append(lines, tsInputLabelStyle.Render("跳过小时数:"))
	lines = append(lines, "> "+m.Input.View())

	if m.InputErr != "" {
		lines = append(lines, tsErrStyle.Render(m.InputErr))
	}

	lines = append(lines, "")
	lines = append(lines, m.Help.View(m.InputKeyMap))

	return strings.Join(lines, "\n")
}

func (m *TimeskipModel) viewPreview() string {
	statNames := []string{"饱腹", "快乐", "健康", "精力"}

	var lines []string

	// Show breakdown of offline time
	totalHours := m.Pet.AccumulatedOfflineDuration.Hours() + m.PreviewHours
	lines = append(lines, tsInputLabelStyle.Render(fmt.Sprintf("跳过 %.1f 小时后的变化:", m.PreviewHours)))
	if m.Pet.AccumulatedOfflineDuration > 0 {
		lines = append(lines, tsInfoStyle.Render(fmt.Sprintf("  (已累积: %.1fh + 新增: %.1fh = 总计: %.1fh)",
			m.Pet.AccumulatedOfflineDuration.Hours(), m.PreviewHours, totalHours)))
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  年龄   %.1fh → %.1fh", m.OldAge, m.NewAge))
	lines = append(lines, "")

	for i, name := range statNames {
		oldBar := components.NewProgressBar().
			SetValue(m.OldStats[i]).
			SetMax(100).
			SetWidth(15).
			SetFilledStyle(styles.DevCommandStyles.BarFilled).
			SetEmptyStyle(styles.DevCommandStyles.BarEmpty).
			Render()
		newBar := components.NewProgressBar().
			SetValue(m.NewStats[i]).
			SetMax(100).
			SetWidth(15).
			SetFilledStyle(styles.DevCommandStyles.BarFilled).
			SetEmptyStyle(styles.DevCommandStyles.BarEmpty).
			Render()
		delta := m.NewStats[i] - m.OldStats[i]
		deltaStr := fmt.Sprintf("%+d", delta)
		if delta < 0 {
			deltaStr = tsErrStyle.Render(deltaStr)
		}
		lines = append(lines, fmt.Sprintf("  %-6s %3d %s → %3d %s  %s",
			name, m.OldStats[i], oldBar, m.NewStats[i], newBar, deltaStr))
	}

	if m.WouldDie {
		lines = append(lines, "")
		lines = append(lines, tsErrStyle.Render("  ⚠ 警告: 宠物将会死亡!"))
	}

	lines = append(lines, "")
	lines = append(lines, tsInputLabelStyle.Render("确认执行?"))
	lines = append(lines, m.Help.View(m.PreviewKeyMap))

	return strings.Join(lines, "\n")
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Styles
var (
	tsPanelStyle      = styles.DevCommandStyles.Panel
	tsHeaderStyle     = styles.MakeTitleStyle("#E94560") // Timeskip uses red
	tsInfoStyle       = styles.DevCommandStyles.Info
	tsInputLabelStyle = styles.DevCommandStyles.InputLabel
	tsErrStyle        = styles.DevCommandStyles.Error
)
