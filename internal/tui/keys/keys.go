package keys

import (
	"clipet/internal/i18n"

	"charm.land/bubbles/v2/key"
)

// GlobalKeyMap contains keys that are available in all screens.
type GlobalKeyMap struct {
	Quit       key.Binding
	ToggleHelp key.Binding
}

// NewGlobalKeyMap creates a global keymap with i18n support.
func NewGlobalKeyMap(i18n *i18n.Manager) GlobalKeyMap {
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", i18n.T("ui.keys.quit")),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", i18n.T("ui.keys.toggle_help")),
		),
	}
}

// NavigationKeyMap contains navigation keys.
type NavigationKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Enter key.Binding
	Back  key.Binding
}

// NewNavigationKeyMap creates a navigation keymap with i18n support.
func NewNavigationKeyMap(i18n *i18n.Manager) NavigationKeyMap {
	return NavigationKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", i18n.T("ui.keys.up")),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", i18n.T("ui.keys.down")),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", i18n.T("ui.keys.left")),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", i18n.T("ui.keys.right")),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("↵", i18n.T("ui.keys.enter")),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", i18n.T("ui.keys.back")),
		),
	}
}

// HomeActionKeyMap contains action shortcut keys for the home screen.
type HomeActionKeyMap struct {
	Feed  key.Binding
	Play  key.Binding
	Rest  key.Binding
	Heal  key.Binding
	Talk  key.Binding
}

// NewHomeActionKeyMap creates an action keymap with i18n support.
func NewHomeActionKeyMap(i18n *i18n.Manager) HomeActionKeyMap {
	return HomeActionKeyMap{
		Feed: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", i18n.T("ui.keys.feed")),
		),
		Play: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", i18n.T("ui.keys.play")),
		),
		Rest: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", i18n.T("ui.keys.rest")),
		),
		Heal: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", i18n.T("ui.keys.heal")),
		),
		Talk: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", i18n.T("ui.keys.talk")),
		),
	}
}

// HomeKeyMap contains all keys for the home screen.
type HomeKeyMap struct {
	Global     GlobalKeyMap
	Navigation NavigationKeyMap
	Actions    HomeActionKeyMap
}

// NewHomeKeyMap creates a complete home screen keymap.
func NewHomeKeyMap(i18n *i18n.Manager) HomeKeyMap {
	return HomeKeyMap{
		Global:     NewGlobalKeyMap(i18n),
		Navigation: NewNavigationKeyMap(i18n),
		Actions:    NewHomeActionKeyMap(i18n),
	}
}

// ShortHelp returns keybindings for the short help.
func (k HomeKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Navigation.Left,
		k.Navigation.Right,
		k.Navigation.Down,
		k.Navigation.Enter,
		k.Actions.Feed,
		k.Actions.Play,
		k.Actions.Rest,
		k.Actions.Heal,
		k.Actions.Talk,
		k.Global.Quit,
	}
}

// FullHelp returns keybindings for the full help.
func (k HomeKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigation.Left, k.Navigation.Right, k.Navigation.Down, k.Navigation.Enter},
		{k.Actions.Feed, k.Actions.Play, k.Actions.Rest, k.Actions.Heal, k.Actions.Talk},
		{k.Global.Quit, k.Global.ToggleHelp},
	}
}

// TreeKeyMap contains keys for tree navigation (used in evolve, preview, evoinfo).
type TreeKeyMap struct {
	Global     GlobalKeyMap
	Navigation NavigationKeyMap
	Select     key.Binding
	Mark       key.Binding
}

// NewTreeKeyMap creates a tree navigation keymap.
func NewTreeKeyMap(i18n *i18n.Manager) TreeKeyMap {
	return TreeKeyMap{
		Global:     NewGlobalKeyMap(i18n),
		Navigation: NewNavigationKeyMap(i18n),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("↵", i18n.T("ui.keys.enter")),
		),
		Mark: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", i18n.T("ui.keys.mark")),
		),
	}
}

// ShortHelp returns keybindings for the short help.
func (k TreeKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Navigation.Up,
		k.Navigation.Down,
		k.Navigation.Enter,
		k.Navigation.Back,
		k.Global.Quit,
	}
}

// FullHelp returns keybindings for the full help.
func (k TreeKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigation.Up, k.Navigation.Down, k.Navigation.Enter, k.Navigation.Back},
		{k.Mark, k.Global.Quit, k.Global.ToggleHelp},
	}
}

// AdventureKeyMap contains keys for adventure screen.
type AdventureKeyMap struct {
	Global     GlobalKeyMap
	Navigation NavigationKeyMap
}

// NewAdventureKeyMap creates an adventure keymap.
func NewAdventureKeyMap(i18n *i18n.Manager) AdventureKeyMap {
	return AdventureKeyMap{
		Global:     NewGlobalKeyMap(i18n),
		Navigation: NewNavigationKeyMap(i18n),
	}
}

// ShortHelp returns keybindings for the short help.
func (k AdventureKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Navigation.Up,
		k.Navigation.Down,
		k.Navigation.Enter,
		k.Navigation.Back,
		k.Global.Quit,
	}
}

// FullHelp returns keybindings for the full help.
func (k AdventureKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigation.Up, k.Navigation.Down, k.Navigation.Enter, k.Navigation.Back},
		{k.Global.Quit, k.Global.ToggleHelp},
	}
}

// EvolveKeyMap contains keys for evolve screen.
type EvolveKeyMap struct {
	Global     GlobalKeyMap
	Navigation NavigationKeyMap
}

// NewEvolveKeyMap creates an evolve keymap.
func NewEvolveKeyMap(i18n *i18n.Manager) EvolveKeyMap {
	return EvolveKeyMap{
		Global:     NewGlobalKeyMap(i18n),
		Navigation: NewNavigationKeyMap(i18n),
	}
}

// ShortHelp returns keybindings for the short help.
func (k EvolveKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Navigation.Up,
		k.Navigation.Down,
		k.Navigation.Enter,
		k.Navigation.Back,
		k.Global.Quit,
	}
}

// FullHelp returns keybindings for the full help.
func (k EvolveKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigation.Up, k.Navigation.Down, k.Navigation.Enter, k.Navigation.Back},
		{k.Global.Quit, k.Global.ToggleHelp},
	}
}

// OfflineSettlementKeyMap contains keys for offline settlement screen.
type OfflineSettlementKeyMap struct {
	Global GlobalKeyMap
	Up     key.Binding
	Down   key.Binding
	Top    key.Binding
	Bottom key.Binding
	Quit   key.Binding
}

// NewOfflineSettlementKeyMap creates an offline settlement keymap.
func NewOfflineSettlementKeyMap(i18n *i18n.Manager) OfflineSettlementKeyMap {
	return OfflineSettlementKeyMap{
		Global: NewGlobalKeyMap(i18n),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", i18n.T("ui.keys.up")),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", i18n.T("ui.keys.down")),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", i18n.T("ui.keys.top")),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", i18n.T("ui.keys.bottom")),
		),
		Quit: key.NewBinding(
			key.WithKeys("enter", " ", "q"),
			key.WithHelp("↵/q", i18n.T("ui.keys.quit")),
		),
	}
}

// ShortHelp returns keybindings for the short help.
func (k OfflineSettlementKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Quit,
		k.Top,
		k.Bottom,
	}
}

// FullHelp returns keybindings for the full help.
func (k OfflineSettlementKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Quit},
	}
}

// PreviewKeyMap contains keys for preview command (dev tool).
type PreviewKeyMap struct {
	Global    GlobalKeyMap
	Navigation NavigationKeyMap
	SpeedUp   key.Binding
	SlowDown  key.Binding
}

// NewPreviewKeyMap creates a preview keymap.
func NewPreviewKeyMap(i18n *i18n.Manager) PreviewKeyMap {
	return PreviewKeyMap{
		Global:     NewGlobalKeyMap(i18n),
		Navigation: NewNavigationKeyMap(i18n),
		SpeedUp: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", i18n.T("ui.keys.speed_up")),
		),
		SlowDown: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", i18n.T("ui.keys.slow_down")),
		),
	}
}

// ShortHelp returns keybindings for the short help.
func (k PreviewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Navigation.Up,
		k.Navigation.Down,
		k.SpeedUp,
		k.SlowDown,
		k.Global.Quit,
	}
}

// FullHelp returns keybindings for the full help.
func (k PreviewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigation.Up, k.Navigation.Down, k.Navigation.Left, k.Navigation.Right},
		{k.SpeedUp, k.SlowDown, k.Global.Quit, k.Global.ToggleHelp},
	}
}

// SetKeyMap contains keys for set command (dev tool).
type SetKeyMap struct {
	Global     GlobalKeyMap
	Navigation NavigationKeyMap
}

// NewSetKeyMap creates a set keymap.
func NewSetKeyMap(i18n *i18n.Manager) SetKeyMap {
	return SetKeyMap{
		Global:     NewGlobalKeyMap(i18n),
		Navigation: NewNavigationKeyMap(i18n),
	}
}

// ShortHelp returns keybindings for the short help.
func (k SetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Navigation.Up,
		k.Navigation.Down,
		k.Navigation.Enter,
		k.Navigation.Back,
		k.Global.Quit,
	}
}

// FullHelp returns keybindings for the full help.
func (k SetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigation.Up, k.Navigation.Down, k.Navigation.Enter, k.Navigation.Back},
		{k.Global.Quit, k.Global.ToggleHelp},
	}
}

// TimeskipKeyMap contains keys for timeskip command (dev tool).
type TimeskipKeyMap struct {
	Global     GlobalKeyMap
	Navigation NavigationKeyMap
}

// NewTimeskipKeyMap creates a timeskip keymap.
func NewTimeskipKeyMap(i18n *i18n.Manager) TimeskipKeyMap {
	return TimeskipKeyMap{
		Global:     NewGlobalKeyMap(i18n),
		Navigation: NewNavigationKeyMap(i18n),
	}
}

// ShortHelp returns keybindings for the short help.
func (k TimeskipKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Navigation.Up,
		k.Navigation.Down,
		k.Navigation.Enter,
		k.Navigation.Back,
		k.Global.Quit,
	}
}

// FullHelp returns keybindings for the full help.
func (k TimeskipKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Navigation.Up, k.Navigation.Down, k.Navigation.Enter, k.Navigation.Back},
		{k.Global.Quit, k.Global.ToggleHelp},
	}
}
