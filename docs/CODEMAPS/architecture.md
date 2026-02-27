<!-- Generated: 2026-02-27 | Files scanned: 47 | Token estimate: ~600 -->

# Architecture Overview

## System Type
Single binary CLI application with embedded TUI frontend

## Layered Architecture

```
┌──────────────────────────────────────────┐
│       cmd/clipet/main.go                 │  Entry point
│       cmd/clipet-dev/main.go             │  Dev tools entry
├──────────────────────────────────────────┤
│        internal/cli/ (Cobra)             │  CLI routing
├───────────────┬──────────────────────────┤
│ internal/tui/ │ internal/cli/*.go        │  TUI / Direct CLI
│ (Bubble Tea)  │                          │
├───────────────┴──────────────────────────┤
│        internal/game/ (UI-agnostic)      │  Core logic
├──────────────────────────────────────────┤
│        internal/store/                   │  Persistence
├──────────────────────────────────────────┤
│        internal/plugin/                  │  Plugin system
├──────────────────────────────────────────┤
│        internal/assets/ (go:embed)       │  Builtin species
└──────────────────────────────────────────┘
```

## Design Principles

1. **Game Logic ≠ UI**: `internal/game/` has zero TUI/CLI dependencies
2. **Unified Plugin Pipeline**: Builtin and external use same `fs.FS` → `Loader` → `Registry`
3. **Interface Abstraction**: `Store` interface swappable (JSON/BoltDB/etc.)

## Entry Points

| Binary | Entry File | Purpose |
|--------|-----------|---------|
| `clipet` | cmd/clipet/main.go | Main game (TUI + CLI) |
| `clipet-dev` | cmd/clipet-dev/main.go | Development tools |

## Package Dependency Flow

```
cmd/
 └─→ cli/ (Cobra commands)
       ├─→ tui/ (Bubble Tea screens)
       │     ├─→ components/ (reusable widgets)
       │     └─→ styles/ (Lipgloss theming)
       ├─→ game/ (business logic)
       │     └─→ games/ (minigames)
       ├─→ store/ (JSON persistence)
       └─→ plugin/ (species packs)
             └─→ assets/ (embedded FS)
```

## Data Flow

```
User Input → CLI/TUI → game.Pet methods → store.Save()
                              ↓
                        plugin.Registry → species data
```
