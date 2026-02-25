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

// screen identifies which TUI screen is active.
type screen int

const (
	screenHome screen = iota
	screenEvolve
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
	theme    styles.Theme

	home   screens.HomeModel
	evolve screens.EvolveModel
	active screen

	width        int
	height       int
	quitting     bool
	decayApplied bool // whether offline decay has been applied
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
		theme:    theme,
		home:     home,
		active:   screenHome,
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
		a.evolve = a.evolve.SetSize(msg.Width, msg.Height)
		return a, nil

	case tea.KeyPressMsg:
		// Global quit: ctrl+c always works, q only on home screen
		switch msg.String() {
		case "ctrl+c":
			a.quitting = true
			_ = a.store.Save(a.pet)
			return a, tea.Quit
		case "q":
			if a.active == screenHome {
				a.quitting = true
				_ = a.store.Save(a.pet)
				return a, tea.Quit
			}
		}

	case tickMsg:
		// Apply offline decay on first tick
		if !a.decayApplied {
			a.decayApplied = true
			a.pet.ApplyOfflineDecay()
			_ = a.store.Save(a.pet)
		}

		a.pet.UpdateAnimation()
		a.petView.Tick()

		if a.active == screenEvolve {
			a.evolve = a.evolve.Tick()
			if a.evolve.IsDone() {
				_ = a.store.Save(a.pet)
				a.active = screenHome
				a.home = a.home.UpdatePet(a.pet)
			}
			return a, doTick()
		}

		// On home screen: update pet, tick game/dialogue
		a.home = a.home.UpdatePet(a.pet)
		a.home = a.home.TickGame()
		a.home = a.home.TickAutoDialogue()
		return a, doTick()
	}

	// Delegate input to active screen
	switch a.active {
	case screenHome:
		var cmd tea.Cmd
		a.home, cmd = a.home.Update(msg)
		// Check evolution after user actions (not during games)
		if !a.home.IsPlayingGame() {
			a.checkEvolution()
		}
		return a, cmd

	case screenEvolve:
		var cmd tea.Cmd
		a.evolve, cmd = a.evolve.Update(msg)
		if a.evolve.IsDone() {
			_ = a.store.Save(a.pet)
			a.active = screenHome
			a.home = a.home.UpdatePet(a.pet)
		}
		return a, cmd
	}

	return a, nil
}

// checkEvolution runs the evolution engine and switches to evolve screen if applicable.
func (a *App) checkEvolution() {
	candidates := game.CheckEvolution(a.pet, a.registry)
	if len(candidates) == 0 {
		return
	}
	a.evolve = screens.NewEvolveModel(a.pet, candidates, a.theme)
	a.evolve = a.evolve.SetSize(a.width, a.height)
	a.active = screenEvolve
}

// View implements tea.Model.
func (a App) View() tea.View {
	if a.quitting {
		return tea.NewView("再见！\n")
	}

	var content string
	switch a.active {
	case screenHome:
		content = a.home.View()
	case screenEvolve:
		content = a.evolve.View()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
