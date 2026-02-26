// Package dev provides TUI models for clipet-dev commands
package dev

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/tui/components"
	"clipet/internal/tui/styles"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// EvolveKeyMap defines keybindings for evolve command
type EvolveKeyMap struct {
	Quit key.Binding
}

// DefaultEvolveKeyMap returns default keybindings for evolve command
var DefaultEvolveKeyMap = EvolveKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q/Esc", "退出"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k EvolveKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k EvolveKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit},
	}
}

// EvolveModel is the TUI model for evolve command
type EvolveModel struct {
	Pet      *game.Pet
	Species   string
	Registry  *plugin.Registry
	Tree      components.TreeList
	Width     int
	Height    int
	Quitting  bool
	KeyMap    EvolveKeyMap

	// Callback when user selects a stage to evolve to
	OnEvolve func(toStageID string) error

	// Result stored for output after TUI exits
	EvolveResult string
}

// NewEvolveModel creates a new evolve TUI model
func NewEvolveModel(pet *game.Pet, species string, registry *plugin.Registry) *EvolveModel {
	pack := registry.GetSpecies(species)
	if pack == nil {
		return &EvolveModel{Pet: pet, Species: species, Registry: registry, KeyMap: DefaultEvolveKeyMap}
	}

	roots := buildEvoTreeFromPack(pack)

	tree := components.NewTreeList(roots)
	tree.ShowConnectors = true
	tree.MarkedID = pet.StageID
	tree.ExpandToLevel(2)
	tree.SetCursor(pet.StageID)

	return &EvolveModel{
		Pet:      pet,
		Species:  species,
		Registry: registry,
		Tree:     tree,
		KeyMap:   DefaultEvolveKeyMap,
	}
}

// Init implements tea.Model
func (m *EvolveModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *EvolveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Update tree size
		treeW := m.Width * 40 / 100
		if treeW < 30 {
			treeW = 30
		}
		m.Tree.SetSize(treeW-4, m.Height-10)

	case tea.KeyPressMsg:
		// Handle global keys
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			m.Quitting = true
			return m, tea.Quit
		}

		// Delegate navigation to tree
		var cmd tea.Cmd
		m.Tree, cmd = m.Tree.Update(msg)
		return m, cmd

	case components.TreeSelectMsg:
		// User pressed Enter on a node
		if m.OnEvolve != nil {
			selected := m.Tree.Selected()
			if selected != nil && selected.ID != m.Pet.StageID {
				// Dev command: force evolution without checking conditions
				oldStageID := m.Pet.StageID
				if err := m.OnEvolve(selected.ID); err == nil {
					// Store result for output after TUI exits
					stage := m.Registry.GetStage(m.Species, selected.ID)
					if stage != nil {
						m.EvolveResult = fmt.Sprintf("evolve: %s -> %s (%s)", oldStageID, selected.ID, stage.Phase)
					} else {
						m.EvolveResult = fmt.Sprintf("evolve: %s -> %s", oldStageID, selected.ID)
					}
					m.Quitting = true
					return m, tea.Quit
				}
			}
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *EvolveModel) View() tea.View {
	if m.Quitting {
		return tea.NewView("")
	}
	if m.Width == 0 {
		v := tea.NewView("加载中...")
		v.AltScreen = true
		return v
	}

	title := evoHeaderStyle.Render(fmt.Sprintf(" 🧬 进化选择 — %s [%s] ", m.Pet.Name, m.Pet.StageID))

	treeStr := m.Tree.View()

	selected := m.Tree.Selected()
	info := evoInfoStyle.Render(fmt.Sprintf("选中: %s", selected.Label))

	isCurrent := ""
	if selected.ID == m.Pet.StageID {
		isCurrent = evoInfoStyle.Render("  (当前阶段)")
	}

	help := evoInfoStyle.Render("↑↓导航  ←→折叠/展开  Enter确认  q退出")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		treeStr,
		"",
		info+isCurrent,
		help,
	)

	panel := evoLeftStyle.
		Width(m.Width - 2).
		Height(m.Height - 1).
		Render(content)

	v := tea.NewView(panel)
	v.AltScreen = true
	return v
}

// Styles
var (
	evoLeftStyle   = styles.DevCommandStyles.Panel
	evoHeaderStyle = styles.DevCommandStyles.Title
	evoInfoStyle   = styles.DevCommandStyles.Info
)
