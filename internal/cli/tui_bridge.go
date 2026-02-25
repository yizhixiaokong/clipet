package cli

import (
	"clipet/internal/game"
	"clipet/internal/plugin"
	"clipet/internal/store"
	"clipet/internal/tui"

	tea "charm.land/bubbletea/v2"
)

// startTUI launches the Bubble Tea TUI application.
func startTUI(pet *game.Pet, reg *plugin.Registry, st *store.JSONStore) error {
	app := tui.NewApp(pet, reg, st)
	p := tea.NewProgram(app)
	_, err := p.Run()
	return err
}
