// Package dev provides TUI models for clipet-dev commands
package dev

import (
	"clipet/internal/game"
	"clipet/internal/i18n"
	"clipet/internal/plugin"
	"clipet/internal/tui/components"
	"clipet/internal/tui/keys"
	"clipet/internal/tui/styles"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// EvolveModel is the TUI model for evolve command
type EvolveModel struct {
	Pet      *game.Pet
	Species   string
	Registry  *plugin.Registry
	Tree      components.TreeList
	Width     int
	Height    int
	Quitting  bool
	KeyMap    keys.TreeKeyMap
	Help      help.Model
	i18n      *i18n.Manager

	// Callback when user selects a stage to evolve to
	OnEvolve func(toStageID string) error

	// Result stored for output after TUI exits
	EvolveResult string
}

// NewEvolveModel creates a new evolve TUI model
func NewEvolveModel(pet *game.Pet, species string, registry *plugin.Registry, i18nMgr *i18n.Manager) *EvolveModel {
	h := help.New()
	h.ShowAll = false

	pack := registry.GetSpecies(species)
	if pack == nil {
		return &EvolveModel{Pet: pet, Species: species, Registry: registry, KeyMap: keys.NewTreeKeyMap(i18nMgr), Help: h, i18n: i18nMgr}
	}

	roots := buildEvoTreeFromPack(pack)

	tree := components.NewTreeList(roots)
	tree.ShowConnectors = true
	tree.MarkedID = pet.StageID
	tree.ExpandToLevel(2)
	tree.SetCursor(pet.StageID)

	return &EvolveModel{
		Pet:     pet,
		Species: species,
		Registry: registry,
		Tree:    tree,
		KeyMap:  keys.NewTreeKeyMap(i18nMgr),
		Help:    h,
		i18n:    i18nMgr,
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

		// Update help width
		m.Help.SetWidth(m.Width)

		// Update tree size
		treeW := m.Width * 40 / 100
		if treeW < 30 {
			treeW = 30
		}
		m.Tree.SetSize(treeW-4, m.Height-10)

	case tea.KeyPressMsg:
		// Handle global keys
		switch {
		case key.Matches(msg, m.KeyMap.Global.Quit):
			m.Quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.Global.ToggleHelp):
			m.Help.ShowAll = !m.Help.ShowAll
			return m, nil
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

	helpView := m.Help.View(m.KeyMap)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		treeStr,
		"",
		info+isCurrent,
		helpView,
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
