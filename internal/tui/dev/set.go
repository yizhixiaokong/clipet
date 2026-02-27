// Package dev provides TUI models for clipet-dev commands
package dev

import (
	"clipet/internal/game"
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

// SetField describes one editable pet attribute
type SetField struct {
	Key   string
	Label string
	Kind  string // "int" or "string" or "bool"
}

// SetKeyMap defines keybindings for set command
type SetKeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Cancel     key.Binding
	Quit       key.Binding
	ToggleHelp key.Binding
}

// DefaultSetKeyMap returns default keybindings for set command
var DefaultSetKeyMap = SetKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "上移"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "下移"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "编辑"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "取消/退出"),
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
func (k SetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k SetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Cancel, k.Quit},
	}
}

// SetModel is the TUI model for set command
type SetModel struct {
	Pet      *game.Pet
	Fields   []SetField
	Cursor   int
	Phase    setPhase
	Input    *components.InputField
	Width    int
	Height   int
	Quitting bool
	Message  string
	KeyMap   SetKeyMap
	Help     help.Model

	// Changes records all successful modifications (for output after TUI exits)
	Changes []FieldChange

	// Callbacks for business logic
	GetCurrentValue func(field SetField) string
	SetFieldValue   func(field SetField, value string) (old string, err error)
	OnFieldChanged  func()
}

// FieldChange records a single field modification
type FieldChange struct {
	Field string
	Old   string
	New   string
}

type setPhase int

const (
	setPhaseSelect setPhase = iota // choosing which field to edit
	setPhaseInput                  // typing new value
)

// NewSetModel creates a new set TUI model
func NewSetModel(pet *game.Pet, fields []SetField) *SetModel {
	h := help.New()
	h.ShowAll = false // Start with short help

	return &SetModel{
		Pet:    pet,
		Fields: fields,
		Phase:  setPhaseSelect,
		KeyMap: DefaultSetKeyMap,
		Help:   h,
	}
}

// Init implements tea.Model
func (m *SetModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *SetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyPressMsg:
		if m.Phase == setPhaseSelect {
			return m.updateSelect(msg)
		}
		return m.updateInput(msg)
	}
	return m, nil
}

// View implements tea.Model
func (m *SetModel) View() tea.View {
	if m.Quitting {
		return tea.NewView("")
	}
	if m.Width == 0 {
		v := tea.NewView("加载中...")
		v.AltScreen = true
		return v
	}

	title := setHeaderStyle.Render(fmt.Sprintf(" %s [%s] ", m.Pet.Name, m.Pet.StageID))

	var lines []string
	for i, f := range m.Fields {
		val := m.getFieldDisplay(f)
		prefix := "  "
		style := setItemStyle

		if i == m.Cursor {
			prefix = "▸ "
			style = setSelStyle
		}

		line := fmt.Sprintf("%s%-8s %s", prefix, f.Label, val)
		lines = append(lines, style.Render(line))
	}

	// Input area
	inputArea := ""
	if m.Phase == setPhaseInput {
		f := m.Fields[m.Cursor]
		inputArea = "\n" + setInputLabelStyle.Render(fmt.Sprintf("编辑 %s:", f.Label)) +
			"\n> " + m.Input.View()
	}

	// Message
	msgArea := ""
	if m.Message != "" {
		msgArea = "\n" + m.Message
	}

	// Help - show different help based on phase
	var helpView string
	if m.Phase == setPhaseInput {
		// Input phase: show only relevant keys
		helpView = m.Help.ShortHelpView([]key.Binding{m.KeyMap.Enter, m.KeyMap.Cancel})
	} else {
		// Select phase: show navigation help
		helpView = m.Help.View(m.KeyMap)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		strings.Join(lines, "\n"),
		inputArea,
		msgArea,
		"", helpView,
	)

	panel := setPanelStyle.
		Width(m.Width - 2).
		Height(m.Height - 1).
		Render(content)

	v := tea.NewView(panel)
	v.AltScreen = true
	return v
}

func (m *SetModel) updateSelect(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Quit), key.Matches(msg, m.KeyMap.Cancel):
		// Both Quit (q/Ctrl+C) and Cancel (Esc) exit in select mode
		m.Quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.KeyMap.ToggleHelp):
		m.Help.ShowAll = !m.Help.ShowAll
		return m, nil
	case key.Matches(msg, m.KeyMap.Up):
		if m.Cursor > 0 {
			m.Cursor--
		}
	case key.Matches(msg, m.KeyMap.Down):
		if m.Cursor < len(m.Fields)-1 {
			m.Cursor++
		}
	case key.Matches(msg, m.KeyMap.Enter):
		m.Phase = setPhaseInput
		currentValue := ""
		if m.GetCurrentValue != nil {
			currentValue = m.GetCurrentValue(m.Fields[m.Cursor])
		}
		m.Input = components.NewInputField().SetValue(currentValue)
		m.Message = ""
	}
	return m, nil
}

func (m *SetModel) updateInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Quit):
		// Allow q/Ctrl+C/Esc to quit from input mode
		m.Quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.KeyMap.Cancel):
		// Esc cancels input, returns to select
		m.Phase = setPhaseSelect
		m.Input = nil
		m.Message = ""
	case key.Matches(msg, m.KeyMap.ToggleHelp):
		m.Help.ShowAll = !m.Help.ShowAll
		return m, nil
	case key.Matches(msg, m.KeyMap.Enter):
		field := m.Fields[m.Cursor]
		if m.SetFieldValue != nil {
			old, err := m.SetFieldValue(field, m.Input.Value())
			if err != nil {
				m.Message = fmt.Sprintf("❌ %v", err)
			} else {
				m.Message = fmt.Sprintf("✓ %s: %s → %s", field.Label, old, m.Input.Value())
				// Record the change for output after TUI exits
				m.Changes = append(m.Changes, FieldChange{
					Field: field.Label,
					Old:   old,
					New:   m.Input.Value(),
				})
				if m.OnFieldChanged != nil {
					m.OnFieldChanged()
				}
			}
		}
		m.Phase = setPhaseSelect
		m.Input = nil
	default:
		// Delegate to InputField component
		m.Input, _ = m.Input.Update(msg)
	}
	return m, nil
}

func (m *SetModel) getFieldDisplay(f SetField) string {
	val := ""
	if m.GetCurrentValue != nil {
		val = m.GetCurrentValue(f)
	}

	if f.Kind == "int" {
		n, _ := strconv.Atoi(val)
		bar := components.NewProgressBar().
			SetValue(n).
			SetMax(100).
			SetWidth(20).
			SetFilledStyle(styles.DevCommandStyles.BarFilled).
			SetEmptyStyle(styles.DevCommandStyles.BarEmpty).
			Render()
		return fmt.Sprintf("%-6s %s", val, bar)
	}
	return val
}

// Styles
var (
	setPanelStyle      = styles.DevCommandStyles.Panel
	setHeaderStyle     = styles.MakeTitleStyle("#7D56F4")
	setInfoStyle       = styles.DevCommandStyles.Info
	setItemStyle       = styles.DevCommandStyles.Item
	setSelStyle        = styles.DevCommandStyles.SelItem
	setInputLabelStyle = styles.DevCommandStyles.InputLabel
)
