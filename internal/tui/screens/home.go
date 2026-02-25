// Package screens provides the TUI screens.
package screens

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/store"
	"clipet/internal/tui/components"
	"clipet/internal/tui/styles"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// menuItem represents an action in the home screen menu.
type menuItem struct {
	label  string
	action string
}

var homeMenuItems = []menuItem{
	{"🍖 喂食", "feed"},
	{"🎮 玩耍", "play"},
	{"💬 对话", "talk"},
}

// HomeModel is the home screen model.
type HomeModel struct {
	pet      *game.Pet
	registry *plugin.Registry
	store    store.Store
	petView  *components.PetView
	theme    styles.Theme

	menuIndex int
	width     int
	height    int
	message   string // transient feedback message
	dialogue  string // last dialogue line
}

// NewHomeModel creates a new home screen model.
func NewHomeModel(
	pet *game.Pet,
	reg *plugin.Registry,
	st store.Store,
	pv *components.PetView,
	theme styles.Theme,
) HomeModel {
	return HomeModel{
		pet:      pet,
		registry: reg,
		store:    st,
		petView:  pv,
		theme:    theme,
	}
}

// SetSize updates the terminal dimensions.
func (h HomeModel) SetSize(w, ht int) HomeModel {
	h.width = w
	h.height = ht
	return h
}

// UpdatePet refreshes the pet reference.
func (h HomeModel) UpdatePet(pet *game.Pet) HomeModel {
	h.pet = pet
	h.petView.SetPet(pet)
	return h
}

// Update handles input for the home screen.
func (h HomeModel) Update(msg tea.Msg) (HomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if h.menuIndex > 0 {
				h.menuIndex--
			}
		case "down", "j":
			if h.menuIndex < len(homeMenuItems)-1 {
				h.menuIndex++
			}
		case "enter":
			h = h.executeAction(homeMenuItems[h.menuIndex].action)
		case "f":
			h = h.executeAction("feed")
		case "p":
			h = h.executeAction("play")
		case "t":
			h = h.executeAction("talk")
		}
	}
	return h, nil
}

func (h HomeModel) executeAction(action string) HomeModel {
	switch action {
	case "feed":
		old := h.pet.Hunger
		h.pet.Feed()
		_ = h.store.Save(h.pet)
		h.message = fmt.Sprintf("🍖 饱腹度: %d → %d", old, h.pet.Hunger)
		h.dialogue = ""

	case "play":
		old := h.pet.Happiness
		h.pet.Play()
		_ = h.store.Save(h.pet)
		h.message = fmt.Sprintf("🎮 快乐度: %d → %d", old, h.pet.Happiness)
		h.dialogue = ""

	case "talk":
		h.pet.Talk()
		line := h.registry.GetDialogue(h.pet.Species, h.pet.StageID, h.pet.MoodName())
		if line == "" {
			line = "......"
		}
		_ = h.store.Save(h.pet)
		h.dialogue = line
		h.message = ""
	}
	return h
}

// View renders the home screen.
func (h HomeModel) View() string {
	if h.width == 0 {
		return "正在加载..."
	}

	// Title
	title := h.theme.Title.Width(h.width).Render("🐾 Clipet — " + h.pet.Name)

	// Pet ASCII art
	petArt := h.petView.Render()
	petInfo := h.petView.RenderInfo()
	petSection := h.theme.PetBox.Render(petArt + "\n\n" + petInfo)

	// Stats
	stats := h.renderStats()

	// Menu
	menu := h.renderMenu()

	// Dialogue / Message
	feedbackLine := ""
	if h.dialogue != "" {
		feedbackLine = h.theme.Dialogue.Render("💬 " + h.dialogue)
	} else if h.message != "" {
		feedbackLine = h.message
	}

	// Help
	help := h.theme.Help.Render("↑↓/jk:选择  Enter:确认  f:喂食  p:玩耍  t:对话  q:退出")

	// Status bar
	statusBar := h.theme.StatusBar.Width(h.width).Render(
		fmt.Sprintf("互动:%d  年龄:%.1fh  心情:%d/100",
			h.pet.TotalInteractions, h.pet.AgeHours(), h.pet.MoodScore()))

	// Compose
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		petSection,
		"",
		stats,
		"",
		menu,
		"",
		feedbackLine,
		"",
		help,
		statusBar,
	)

	return content
}

func (h HomeModel) renderStats() string {
	bar := func(label string, val int) string {
		filled := val / 10
		empty := 10 - filled
		fStr := h.theme.StatFilled.Render(strings.Repeat("█", filled))
		eStr := h.theme.StatEmpty.Render(strings.Repeat("░", empty))
		return h.theme.StatLabel.Render(label) + " " + fStr + eStr + fmt.Sprintf(" %3d", val)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		bar("饱腹", h.pet.Hunger),
		bar("快乐", h.pet.Happiness),
		bar("健康", h.pet.Health),
		bar("精力", h.pet.Energy),
	)
}

func (h HomeModel) renderMenu() string {
	var items []string
	for i, item := range homeMenuItems {
		if i == h.menuIndex {
			items = append(items, h.theme.MenuItemSelected.Render("▸ "+item.label))
		} else {
			items = append(items, h.theme.MenuItem.Render("  "+item.label))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}
