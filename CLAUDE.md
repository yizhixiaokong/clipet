# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build both main and dev tools
make build

# Build only dev tool
make dev

# Run all tests
go test ./...

# Run single test
go test -run TestFunctionName ./path/to/package

# Format code
make fmt

# Lint (requires staticcheck)
make lint

# Validate a species pack
./clipet-dev validate internal/assets/builtins/cat-pack
```

## Architecture Overview

Clipet is a terminal-based virtual pet game with a plugin-based species system.

### Core Principle: Game Logic ≠ UI

The `internal/game/` package contains all game logic with **zero dependencies on TUI or CLI packages**. This separation allows:
- Game logic to be tested independently
- Multiple UI frontends (TUI, CLI commands) to share the same core
- Clear boundaries between business rules and presentation

### Layered Architecture

```
cmd/clipet (entry) → internal/cli (Cobra) → internal/tui (Bubble Tea) or direct commands
                                          ↓
                               internal/game (UI-agnostic)
                                          ↓
                          internal/store + internal/plugin
```

### Plugin System

**All species are plugins** (including the built-in cat). Species packs are directories containing:
- `species.toml` - Species metadata, stages, evolution paths, custom attribute definitions
- `dialogues.toml` - Mood/stage-based dialogue lines
- `adventures.toml` - Random event definitions
- `frames/` - ASCII animation files organized by stage/variant/anim-state

Both builtin (embedded via `go:embed`) and external (filesystem) packs use the same `fs.FS` → `Loader` → `Registry` pipeline. External packs can override builtin ones by ID.

**Key Files:**
- `internal/plugin/types.go` - Core data structures (SpeciesPack, Stage, Evolution, Frame, etc.)
- `internal/plugin/parser.go` - TOML parsing + frame file scanning + locale loading
- `internal/plugin/registry.go` - Thread-safe central registry with lookup methods

### Internationalization (i18n)

Clipet supports multiple languages through a lightweight, self-built i18n system.

**Language Priority** (highest to lowest):
1. `CLIPET_LANG` environment variable
2. `LANG` environment variable
3. Configuration file (`~/.config/clipet/config.json`)
4. Default (`zh-CN`)

**Switching Languages**:
```bash
# Temporary switch (environment variable)
CLIPET_LANG=en-US clipet

# Permanent switch (edit config file)
vim ~/.config/clipet/config.json
# Change "language" field to "en-US" or "zh-CN"
```

**Implementation**:
- `internal/i18n/` - i18n framework (Manager, Bundle, Loader, Plural rules)
- `internal/config/` - Configuration and language detection
- `internal/assets/locales/` - Translation files embedded in binary
- Plugin locale support: `locales/{lang}.json` in species pack directory

**Translation File Format**:
```json
{
  "ui": {
    "home": {
      "feed_success": "Feeding successful! Hunger {{.oldHunger}} → {{.newHunger}}"
    }
  }
}
```

**Adding New Strings**:
1. Add translation keys to `internal/assets/locales/zh-CN/tui.json` and `en-US/tui.json`
2. Use `i18n.T("key", "var1", value1, "var2", value2)` in code

**Plugin Locale Support** (Phase 3):
- Species packs can include `locales/zh-CN.json` and `locales/en-US.json`
- Locale files can translate species names, stage names, dialogues, and adventures
- Fallback chain: requested language → fallback language → inline TOML text

**Key Files:**
- `internal/i18n/i18n.go` - Manager with T() and TN() functions
- `internal/i18n/bundle.go` - Translation bundle with fallback chain
- `internal/assets/locales/` - Translation JSON files (embedded)

### Custom Attributes System (v3.0+)

**Purpose**: Allow plugins to define their own evolution conditions without framework hardcoding.

**Architecture**:
- `internal/game/attributes/system.go` - Manages both core attributes (hunger, happiness, health, energy) and custom plugin-defined attributes
- `Pet.CustomAttributes map[string]int` - Storage for custom attribute values
- `EvolutionCondition.CustomAcc map[string]int` - Custom accumulator requirements for evolution

**Workflow**:
1. **Define**: Species plugins define custom attributes in adventures (e.g., `arcane_affinity`, `fire_power`)
2. **Accumulate**: Players trigger adventure events that add to custom attributes via `effects = { arcane_affinity = 10 }`
3. **Check**: Evolution system checks custom attributes: `custom_acc = { arcane_affinity = 50 }`

**Example**: The built-in cat species uses 10 custom attributes to create 3 distinct evolution paths (arcane, feral, mech), replacing hardcoded conditions like `night_bias` and `day_bias`.

**Implementation Notes**:
- `Pet.SetField("custom:attr_name", value)` - Set custom attributes via unified field API
- `Pet.GetAttr(name)` / `Pet.SetAttr(name, value)` - Unified access for both core and custom attributes
- Adventure system automatically records custom attribute changes in `ApplyAdventureOutcome()`
- TUI displays custom attributes in the status panel when they exist
- Dev tools (`set`, `evo info`) support custom attributes

### TUI Architecture

Built on **Bubble Tea v2** with **Lipgloss v2** for styling.

**Structure:**
- `internal/tui/app.go` - Top-level model with screen routing
- `internal/tui/components/` - Reusable UI components (PetView, TreeList, etc.)
- `internal/tui/screens/` - Individual screens (home, evolve, adventure)
- `internal/tui/dev/` - TUI models for dev commands (evolve, preview, set, timeskip)

**Key Pattern:** Components are Bubble Tea models that emit typed messages (e.g., `TreeSelectMsg{Node}`). Parent models handle these messages and update state.

### Game Mechanics

**Pet System** (`internal/game/pet.go`):
- Four core attributes: Hunger, Happiness, Health, Energy (0-100)
- Unlimited custom attributes defined by plugins
- Time-based decay (hunger decays fastest, energy slowest)
- Diminishing returns: `gain = base * (100 - current) / 100` (higher attributes → smaller gains)
- Cooldown system: Actions have cooldown periods (Feed: 10min, Play: 5min, etc.)
- Offline decay: Time since last save is compensated on load

**Evolution System** (`internal/game/evolution.go`):
- Condition-based triggers (age, attributes, interaction counts, custom accumulators)
- Multiple evolution paths from a single stage
- Automatic evolution check after any attribute change

**Adventure System** (`internal/game/adventure.go`):
- Weighted random outcomes
- Effects can include core attributes OR custom attributes
- Returns change log for UI feedback

### Dev Tools

`clipet-dev` provides development utilities (see `cmd/clipet-dev/`):
- `validate <pack-dir>` - Validate species pack structure
- `preview <pack-dir>` - Interactive ASCII frame viewer with TreeList navigation
- `evo to [stage-id]` - Force evolution (interactive tree selection or direct)
- `evo info` - Show evolution conditions and current progress (includes custom attributes)
- `set` - Directly modify pet attributes (interactive, supports custom attributes)
- `timeskip [--hours N]` - Simulate time passing for testing decay/evolution

All dev TUI logic is in `internal/tui/dev/` with reusable components from `internal/tui/components/`.

### Reusable Components

**TreeList** (`internal/tui/components/treelist.go`):
Generic tree navigation component used by evolve, preview, and evoinfo commands.
- Dual rendering: `View()` for TUI, `RenderPlain()` for CLI output
- Visible list pattern: Maintains flattened list of expanded nodes for O(1) cursor ops
- Features: cursor navigation, expand/collapse, marked nodes, tree connectors, scrolling

**PetView** (`internal/tui/components/petview.go`):
Animated ASCII pet display.
- `NormalizeArt(art, width)` - Pads lines to consistent width to prevent centering misalignment
- `DisplayWidth(s)` - Unicode-aware width calculation using charmbracelet/x/ansi

### Important Patterns

**Width Calculation**: Always use `charmbracelet/x/ansi.StringWidth()` for display width (not `len()` or other runewidth libraries). This ensures consistency with Lipgloss centering.

**TUI Model Updates**: Follow Bubble Tea patterns:
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        // Handle key
    case CustomMsg:
        // Handle custom message
    }
    return m, nil  // or return m, tea.Cmd
}
```

**Plugin Pack IDs**: Use format `{species_id}.{stage_id}.{variant}` for stage IDs (e.g., `cat.adult.warrior`). Phase is one of: egg, baby, child, adult, legend.

**Unified Field API**: `Pet.SetField()` supports:
- Core attributes: `hunger`, `happiness`, `health`, `energy`
- Custom attributes: `custom:attr_name` (prefix required)
- Metadata: `name`, `stage_id`, `species_id`

### Documentation

- `docs/CODEMAPS/` - Architecture diagrams and code structure
- `docs/plugin-guide.md` - Plugin development syntax and API reference
- `docs/plugin-best-practices.md` - Design patterns and optimization tips for plugins
- `docs/architecture.md` - Full architecture diagram and dependency graph
