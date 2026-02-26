# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build both main and dev tools
make build

# Build only dev tool
make dev

# Run tests
go test ./...

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
- `species.toml` - Species metadata, stages, evolution paths
- `dialogues.toml` - Mood/stage-based dialogue lines
- `adventures.toml` - Random event definitions
- `frames/` - ASCII animation files organized by stage/variant/anim-state

Both builtin (embedded via `go:embed`) and external (filesystem) packs use the same `fs.FS` → `Loader` → `Registry` pipeline. External packs can override builtin ones by ID.

**Key Files:**
- `internal/plugin/types.go` - Core data structures (SpeciesPack, Stage, Evolution, Frame, etc.)
- `internal/plugin/parser.go` - TOML parsing + frame file scanning
- `internal/plugin/registry.go` - Thread-safe central registry with lookup methods

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
- Four attributes: Hunger, Happiness, Health, Energy (0-100)
- Time-based decay (hunger decays fastest, energy slowest)
- Diminishing returns: `gain = base * (100 - current) / 100` (higher attributes → smaller gains)
- Cooldown system: Actions have cooldown periods (Feed: 10min, Play: 5min, etc.)
- Offline decay: Time since last save is compensated on load

**Evolution System** (`internal/game/evolution.go`):
- Condition-based triggers (age, attributes, interaction counts, time-of-day bias, etc.)
- Multiple evolution paths from a single stage
- Automatic evolution check after any attribute change

### Dev Tools

`clipet-dev` provides development utilities (see `cmd/clipet-dev/`):
- `validate <pack-dir>` - Validate species pack structure
- `preview <pack-dir>` - Interactive ASCII frame viewer with TreeList navigation
- `evo to [stage-id]` - Force evolution (interactive tree selection or direct)
- `evo info` - Show evolution conditions and current progress
- `set [attr] [value]` - Directly modify pet attributes (interactive or CLI)
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

### Documentation

- `docs/architecture.md` - Full architecture diagram and dependency graph
- `docs/plugin-guide.md` - Complete guide to creating species packs (TOML formats, dialogue design, frame organization)
