<!-- Generated: 2026-02-28 | Dependencies: 12 | Token estimate: ~450 -->

# Dependencies

## Core Dependencies

### CLI Framework

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/spf13/cobra | v1.8.1 | CLI command routing |

### TUI Framework

| Package | Version | Purpose |
|---------|---------|---------|
| charm.land/bubbletea/v2 | v2.0.0 | TUI framework |
| charm.land/lipgloss/v2 | v2.0.0 | Styling/layout |
| charm.land/bubbles/v2 | v2.0.0 | UI components |

**Bubbles components used**:
- `textinput` - Input fields (set, timeskip)
- `help` - Keybinding help display
- `key` - Keybinding management

### i18n System (v3.1)

**External dependencies**: None ✅

**Built-in packages used**:
- `encoding/json` - Parse locale files
- `text/template` - String interpolation
- `sync` - Thread-safe access (RWMutex)

### Configuration

| Package | Version | Purpose |
|---------|---------|---------|
| encoding/json | stdlib | Config file parsing |
| os | stdlib | Environment variables |

**Note**: Removed spf13/viper dependency in v3.1 - now uses lightweight custom config

### TOML Parsing

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/BurntSushi/toml | v1.4.0 | Parse species.toml files |

### Utilities

| Package | Version | Purpose |
|---------|---------|---------|
| charm.land/x/ansi | - | String width calculation |

**Specific use**: `ansi.StringWidth()` for Unicode-aware display width

## Dependency Graph

```
cmd/clipet
  └─→ internal/cli
        ├─→ github.com/spf13/cobra
        ├─→ github.com/spf13/viper
        ├─→ internal/tui
        │     ├─→ charm.land/bubbletea/v2
        │     ├─→ charm.land/lipgloss/v2
        │     ├─→ charm.land/bubbles/v2
        │     │     ├─→ textinput
        │     │     ├─→ help
        │     │     └─→ key
        │     └─→ charm.land/x/ansi
        ├─→ internal/game
        ├─→ internal/store
        └─→ internal/plugin
              └─→ github.com/BurntSushi/toml
```

## External Services

**None** - Clipet is fully offline, no network dependencies

## Embedded Assets

### Builtin Species Packs

**Location**: `internal/assets/builtins/`

**Loaded via**: `go:embed` → `embed.FS`

**Current builtin**:
- `cat-pack/` - Default cat species

**Override mechanism**: External packs with same ID override builtin

## File System Dependencies

### Data Directory

**Default**: `~/.local/share/clipet/`

**Contents**:
```
~/.local/share/clipet/
├── save.json          (Pet data)
└── plugins/           (External species packs)
    └── {species-id}/
        ├── species.toml
        ├── dialogues.toml
        ├── adventures.toml
        ├── locales/            (v3.1)
        │   ├── zh-CN.json
        │   └── en-US.json
        └── frames/
```

### Config Directory

**Default**: `~/.config/clipet/`

**Current use** (v3.1):
```
~/.config/clipet/
└── config.json       (Language preferences)
    {
      "language": "en-US",
      "fallback_language": "zh-CN",
      "version": "3.1.0"
    }
```

### Embedded Assets (v3.1)

**Location**: `internal/assets/`

**Structure**:
```
internal/assets/
├── locales/              (TUI translations)
│   ├── zh-CN/
│   │   ├── tui.json      (120+ keys)
│   │   └── game.json
│   └── en-US/
│       ├── tui.json
│       └── game.json
└── builtins/
    └── cat-pack/
        ├── locales/      (Plugin translations)
        │   ├── zh-CN.json  (~750 lines)
        │   └── en-US.json  (~750 lines)
        ├── species.toml
        ├── dialogues.toml
        ├── adventures.toml
        └── frames/
```

**Embed mechanism**: `go:embed` → `embed.FS`

**Override chain**:
1. External plugin locale files (if present)
2. Embedded plugin locale files
3. Inline TOML text (fallback)

## Build Dependencies

### Go Version

**Minimum**: Go 1.21+

**Reason**: Uses modern Go features (slog, embed)

### Build Tools

**Makefile targets**:
- `make build` - Build main binary
- `make dev` - Build dev tools
- `make fmt` - Format code
- `make lint` - Run staticcheck

## Dependency Update Strategy

### Stability Policy

| Category | Update Frequency |
|----------|------------------|
| Charmbracelet packages | Follow v2 releases |
| spf13 packages | Patch updates only |
| BurntSushi/toml | Patch updates only |

### Versioning

- All Charmbracelet packages on v2 (latest stable)
- Cobra and Viper on latest v1.x
- No pre-release or unstable dependencies

## Known Issues

### None currently

All dependencies are stable and well-maintained.
