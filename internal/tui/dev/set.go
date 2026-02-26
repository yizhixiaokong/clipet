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
	"charm.land/lipgloss/v2"
)

// SetField describes one editable pet attribute
type SetField struct {
	Key   string
	Label string
	Kind  string // "int" or "string" or "bool"
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

	// Callbacks for business logic
	GetCurrentValue func(field SetField) string
	SetFieldValue   func(field SetField, value string) (old string, err error)
	OnFieldChanged  func()
}

type setPhase int

const (
	setPhaseSelect setPhase = iota // choosing which field to edit
	setPhaseInput                  // typing new value
)

// NewSetModel creates a new set TUI model
func NewSetModel(pet *game.Pet, fields []SetField) *SetModel {
	return &SetModel{
		Pet:    pet,
		Fields: fields,
		Phase:  setPhaseSelect,
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
			"\n> " + m.Input.View() +
			"\n" + setInfoStyle.Render("Enter确认  Esc取消")
	}

	// Message
	msgArea := ""
	if m.Message != "" {
		msgArea = "\n" + m.Message
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
		Width(m.Width - 2).
		Height(m.Height - 1).
		Render(content)

	v := tea.NewView(panel)
	v.AltScreen = true
	return v
}

func (m *SetModel) updateSelect(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "escape":
		m.Quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Fields)-1 {
			m.Cursor++
		}
	case "enter":
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
	switch msg.String() {
	case "escape":
		m.Phase = setPhaseSelect
		m.Input = nil
		m.Message = ""
	case "enter":
		field := m.Fields[m.Cursor]
		if m.SetFieldValue != nil {
			old, err := m.SetFieldValue(field, m.Input.Value)
			if err != nil {
				m.Message = fmt.Sprintf("❌ %v", err)
			} else {
				m.Message = fmt.Sprintf("✓ %s: %s → %s", field.Label, old, m.Input.Value)
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
