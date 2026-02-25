// Package tui contains the Bubble Tea v2 TUI application for Clipet.
package tui

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/store"
	"clipet/internal/tui/components"
	"clipet/internal/tui/screens"
	"clipet/internal/tui/styles"
	"time"

	tea "charm.land/bubbletea/v2"
)

// tickMsg is sent on each animation/update tick.
type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// App is the top-level Bubble Tea model.
type App struct {
	pet      *game.Pet
	registry *plugin.Registry
	store    store.Store
	petView  *components.PetView
	home     screens.HomeModel

	width    int
	height   int
	quitting bool
}

// NewApp creates the top-level TUI application model.
func NewApp(pet *game.Pet, reg *plugin.Registry, st store.Store) App {
	pv := components.NewPetView(pet, reg)
	theme := styles.DefaultTheme()
	home := screens.NewHomeModel(pet, reg, st, pv, theme)

	return App{
		pet:      pet,
		registry: reg,
		store:    st,
		petView:  pv,
		home:     home,
	}
}

// Init implements tea.Model.
func (a App) Init() tea.Cmd {
	return doTick()
}

// Update implements tea.Model.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.home = a.home.SetSize(msg.Width, msg.Height)
		return a, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			a.quitting = true
			_ = a.store.Save(a.pet)
			return a, tea.Quit
		}

	case tickMsg:
		a.pet.UpdateAnimation()
		a.petView.Tick()
		a.home = a.home.UpdatePet(a.pet)
		return a, doTick()
	}

	// Delegate to home screen
	var cmd tea.Cmd
	a.home, cmd = a.home.Update(msg)
	return a, cmd
}

// View implements tea.Model.
func (a App) View() tea.View {
	if a.quitting {
		return tea.NewView("再见！🐾\n")
	}

	v := tea.NewView(a.home.View())
	v.AltScreen = true
	return v
}
