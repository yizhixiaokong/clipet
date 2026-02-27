<!-- Generated: 2026-02-27 | Files scanned: 59 | Token estimate: ~800 -->

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
│          ├─→ capabilities/               │  Personality traits (M7)
│          ├─→ attributes/                 │  Flexible attributes (M7)
│          └─→ games/                      │  Minigames
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
4. **Extensible Capabilities**: Core provides interfaces, plugins define behavior (M7)

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
       │     ├─→ capabilities/ (traits, lifecycle config)
       │     ├─→ attributes/ (core + custom attrs)
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
                              ↓
                        capabilities.Registry → personality traits (M7)
                              ↓
                        attributes.System → attribute management (M7)
                              ↓
                        LifecycleManager → life checks + endings (M7)
```

## Time Evolution System (M2 + M7)

```
TimeManager (global singleton)
  ├─→ TimeHook: DeathCheckHook (PriorityCritical)
  ├─→ TimeHook: AttrDecayHook (PriorityHigh)
  ├─→ TimeHook: CooldownHook (PriorityNormal)
  └─→ TimeHook: LifecycleHook (PriorityLow) [M7]
```

**Hook execution order**: Critical → High → Normal → Low

## M7 New Components

### Capabilities System (`game/capabilities/`)
- **types.go**: LifecycleConfig, PersonalityTrait, PassiveEffect, ActiveEffect, Ending
- **registry.go**: Thread-safe trait registration and application

### Attributes System (`game/attributes/`)
- **system.go**: Manages core 4 attrs + custom attrs, decay rates, validation

### Lifecycle Management (`game/lifecycle_manager.go`)
- CheckLifecycle(): Age-based warning and ending triggers
- TriggerEnding(): Selects appropriate ending based on pet's life quality
