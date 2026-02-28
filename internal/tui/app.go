package tui

import (
	"clipet/internal/game"
	"clipet/internal/i18n"
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
	screenOfflineSettlement screen = iota
	screenHome
	screenEvolve
	screenAdventure
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
	i18n     *i18n.Manager
	petView  *components.PetView
	theme    styles.Theme

	offlineSettlement screens.OfflineSettlementModel
	home              screens.HomeModel
	evolve            screens.EvolveModel
	adventure         screens.AdventureModel
	active            screen

	width        int
	height       int
	quitting     bool
	decayApplied bool // whether offline decay has been applied
}

// NewApp creates the top-level TUI application model.
func NewApp(pet *game.Pet, reg *plugin.Registry, st store.Store, i18nMgr *i18n.Manager, offlineResults []game.DecayRoundResult) App {
	pv := components.NewPetView(pet, reg)
	theme := styles.DefaultTheme()
	home := screens.NewHomeModel(pet, reg, st, pv, theme, i18nMgr)

	// Create offline settlement screen if there are results
	var offlineSettlement screens.OfflineSettlementModel
	activeScreen := screenHome
	if len(offlineResults) > 0 {
		offlineSettlement = screens.NewOfflineSettlementModel(offlineResults, theme, i18nMgr)
		activeScreen = screenOfflineSettlement
	}

	return App{
		pet:               pet,
		registry:          reg,
		store:             st,
		i18n:              i18nMgr,
		petView:           pv,
		theme:             theme,
		offlineSettlement: offlineSettlement,
		home:              home,
		active:            activeScreen,
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
		a.offlineSettlement = a.offlineSettlement.SetSize(msg.Width, msg.Height)
		a.home = a.home.SetSize(msg.Width, msg.Height)
		a.evolve = a.evolve.SetSize(msg.Width, msg.Height)
		a.adventure = a.adventure.SetSize(msg.Width, msg.Height)
		return a, nil

	case tea.KeyPressMsg:
		// Global quit: ctrl+c always works, q only on home screen
		switch msg.String() {
		case "ctrl+c":
			a.quitting = true
			a.pet.MarkAsChecked() // Mark as checked before saving
			_ = a.store.Save(a.pet)
			return a, tea.Quit
		case "q":
			if a.active == screenHome {
				a.quitting = true
				a.pet.MarkAsChecked() // Mark as checked before saving
				_ = a.store.Save(a.pet)
				return a, tea.Quit
			}
		}

	case tickMsg:
		// Offline time accumulation is handled in loadPet
		// Just mark that we've started ticking
		if !a.decayApplied {
			a.decayApplied = true
		}

		a.pet.UpdateAnimation()
		a.petView.Tick()

		if a.active == screenEvolve {
			a.evolve = a.evolve.Tick()
			if a.evolve.IsDone() {
				a.pet.MarkAsChecked() // Mark as checked before saving
				_ = a.store.Save(a.pet)
				a.active = screenHome
				a.home = a.home.UpdatePet(a.pet)
			}
			return a, doTick()
		}

		if a.active == screenAdventure {
			a.adventure = a.adventure.Tick()
			if a.adventure.IsDone() {
				a.pet.MarkAsChecked() // Mark as checked before saving
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
		a.home = a.home.TickSuccessAnimation()
		return a, doTick()
	}

	// Delegate input to active screen
	switch a.active {
	case screenOfflineSettlement:
		var cmd tea.Cmd
		a.offlineSettlement, cmd = a.offlineSettlement.Update(msg)
		if a.offlineSettlement.IsDone() {
			a.active = screenHome
			a.home = a.home.UpdatePet(a.pet)
		}
		return a, cmd

	case screenHome:
		var cmd tea.Cmd
		a.home, cmd = a.home.Update(msg)
		// Check if home wants to start an adventure
		if adv := a.home.PendingAdventure(); adv != nil {
			a.home = a.home.ClearPendingAdventure()
			a.adventure = screens.NewAdventureModel(a.pet, *adv, a.theme, a.petView, a.registry)
			a.adventure = a.adventure.SetSize(a.width, a.height)
			a.active = screenAdventure
			return a, cmd
		}
		// Check evolution after user actions (not during games)
		if !a.home.IsPlayingGame() {
			a.checkEvolution()
		}
		return a, cmd

	case screenEvolve:
		var cmd tea.Cmd
		a.evolve, cmd = a.evolve.Update(msg)
		if a.evolve.IsDone() {
			a.pet.MarkAsChecked() // Mark as checked before saving
			_ = a.store.Save(a.pet)
			a.active = screenHome
			a.home = a.home.UpdatePet(a.pet)
		}
		return a, cmd

	case screenAdventure:
		var cmd tea.Cmd
		a.adventure, cmd = a.adventure.Update(msg)
		if a.adventure.IsDone() {
			a.pet.MarkAsChecked() // Mark as checked before saving
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
		return tea.NewView(a.i18n.T("ui.common.quit") + "\n")
	}

	var content string
	switch a.active {
	case screenOfflineSettlement:
		content = a.offlineSettlement.View()
	case screenHome:
		content = a.home.View()
	case screenEvolve:
		content = a.evolve.View()
	case screenAdventure:
		content = a.adventure.View()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
