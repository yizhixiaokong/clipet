<!-- Generated: 2026-02-28 | Files scanned: 62 | Token estimate: ~850 -->

# Architecture Overview

## System Type
Single binary CLI application with embedded TUI frontend and i18n support

## Layered Architecture

```
┌──────────────────────────────────────────────┐
│       cmd/clipet/main.go                     │  Entry point
│       cmd/clipet-dev/main.go                 │  Dev tools entry
├──────────────────────────────────────────────┤
│        internal/cli/ (Cobra)                 │  CLI routing
│           └─→ i18n initialization            │
├───────────────┬──────────────────────────────┤
│ internal/tui/ │ internal/cli/*.go            │  TUI / Direct CLI
│ (Bubble Tea)  │     └─→ i18n pass to screens │
├───────────────┴──────────────────────────────┤
│        internal/game/ (UI-agnostic)          │  Core logic
│          ├─→ capabilities/                   │  Personality traits
│          ├─→ attributes/                     │  Custom attributes (v3)
│          └─→ games/                          │  Minigames
│              └─→ ErrorType system (v3.1)     │  Structured errors
├──────────────────────────────────────────────┤
│        internal/i18n/ (v3.1)                 │  Internationalization
│          ├─→ Manager (T/TN functions)        │
│          ├─→ Bundle (translation storage)    │
│          └─→ Loader (locale files)           │
├──────────────────────────────────────────────┤
│        internal/config/ (v3.1)               │  Configuration
│          └─→ Language detection chain        │
├──────────────────────────────────────────────┤
│        internal/store/                       │  Persistence
├──────────────────────────────────────────────┤
│        internal/plugin/                      │  Plugin system
│          └─→ Locale support (v3.1)           │
├──────────────────────────────────────────────┤
│        internal/assets/ (go:embed)           │  Builtin species
│          ├─→ locales/{lang}/tui.json         │  TUI translations
│          └─→ builtins/*/locales/{lang}.json  │  Plugin translations
└──────────────────────────────────────────────┘
```

## Design Principles

1. **Game Logic ≠ UI**: `internal/game/` has zero TUI/CLI/i18n dependencies
2. **Unified Plugin Pipeline**: Builtin and external use same `fs.FS` → `Loader` → `Registry`
3. **Interface Abstraction**: `Store` interface swappable (JSON/BoltDB/etc.)
4. **Extensible Capabilities**: Core provides interfaces, plugins define behavior
5. **i18n at Boundaries**: Game layer returns ErrorType, TUI layer generates localized messages

## Entry Points

| Binary | Entry File | Purpose |
|--------|-----------|---------|
| `clipet` | cmd/clipet/main.go | Main game (TUI + CLI) with i18n |
| `clipet-dev` | cmd/clipet-dev/main.go | Development tools |

## Package Dependency Flow

```
cmd/
  └─→ cli/ (Cobra commands)
       ├─→ config/ (load config, detect language)
       ├─→ i18n/ (initialize Manager)
       │    └─→ assets/locales/ (TUI translations)
       ├─→ tui/ (if interactive mode)
       │    └─→ i18n/ (pass to screens)
       ├─→ game/ (business logic)
       │    └─→ ErrorType (structured errors)
       ├─→ store/ (save/load)
       └─→ plugin/ (species registry)
            └─→ locales/ (plugin translations)
```

## i18n Architecture (NEW v3.1)

### Language Detection Priority
```
1. CLIPET_LANG env var
2. LANG env var
3. LC_ALL env var
4. ~/.config/clipet/config.json
5. Default: zh-CN
```

### Translation Fallback Chain
```
requested language (e.g., en-US)
  ↓ if not found
fallback language (e.g., zh-CN)
  ↓ if not found
inline TOML text (from species.toml)
```

### Template Interpolation
```go
// Code
i18n.T("ui.home.feed_success", "oldHunger", 50, "newHunger", 75)

// Translation file
"feed_success": "Feeding successful! Hunger {{.oldHunger}} → {{.newHunger}}"

// Output
"Feeding successful! Hunger 50 → 75"
```

## ErrorType System (NEW v3.1)

### Game Layer (UI-agnostic)
```go
// internal/game/pet.go
type ActionResult struct {
    OK        bool
    ErrorType string  // Standardized error type
    Message   string  // Internal log message (Chinese)
    // ...
}

const (
    ErrEnergyLow      = "energy_low"
    ErrCooldown       = "cooldown"
    ErrDead           = "dead"
    // ... 10 error types total
)

// Usage
if p.Energy < minEnergy {
    return failResultWithType(ErrEnergyLow, "精力不足")
}
```

### TUI Layer (i18n-aware)
```go
// internal/tui/screens/home.go
func (h HomeModel) localizeGameError(res game.ActionResult) string {
    switch res.ErrorType {
    case game.ErrEnergyLow:
        return h.i18n.T("game.errors.energy_low")
    case game.ErrDead:
        return h.i18n.T("game.errors.dead")
    // ...
    }
}

// Usage
res := h.pet.Feed()
if !res.OK {
    return h.failMsg(h.localizeGameError(res))
}
```

## Key Metrics

- **Total LoC**: ~12,570 (Go files)
- **Packages**: 8 main packages
- **Translation Keys**: 120+ per language
- **Supported Languages**: 2 (zh-CN, en-US)
- **Error Types**: 10 standardized types
- **Plugin Locale Coverage**: 100% (cat-pack)

## Version History

- **v3.1.0** (2026-02-28): i18n system, ErrorType, complete translation
- **v3.0.0** (2026-02-27): Custom attributes, multi-stage offline settlement
- **v2.0.0**: Plugin system, evolution refactor
