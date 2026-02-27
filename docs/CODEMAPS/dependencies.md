<!-- Generated: 2026-02-27 | Dependencies: 12 | Token estimate: ~400 -->

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

### Configuration

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/spf13/viper | v1.19.0 | Config management |

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
        └── frames/
```

### Config Directory

**Default**: `~/.config/clipet/`

**Planned use**: Custom config files (not yet implemented)

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
