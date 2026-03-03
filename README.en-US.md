# 🐾 Clipet — Terminal Pet Companion

Clipet is a virtual pet game running in the terminal, built with Go and [Bubble Tea](https://charm.sh/).

**[中文](README.md) | English**

## Features

- **TUI Interactive Interface** — ASCII art pet + status panel + two-level categorized menu
- **Pet Care** — Feed, play, talk, rest, heal; attributes decay over time
- **Evolution System** — Condition-triggered evolution based on attributes and interactions, multi-path evolution trees
- **Mini Games** — Reaction speed test, number guessing; wins/losses affect pet attributes
- **Dialogue System** — Stage/mood-specific dialogues, auto-chat bubbles
- **Plugin-based Species** — Customize species, evolution conditions, dialogues via TOML config
- **Internationalization** — Supports Chinese and English interfaces, configurable language switching
- **Offline Decay** — Attributes automatically decay when closed, compensated on reopening
- **Cooldowns & Diminishing Returns** — Actions have cooldowns, higher attributes yield lower gains

## Quick Start

```bash
# Build
make build

# Initialize pet
./clipet init

# Launch TUI (default: Chinese)
./clipet

# Use English interface
CLIPET_LANG=en-US ./clipet

# CLI commands
./clipet status
```

## Controls

### TUI Shortcut Keys

| Key | Function |
|-----|----------|
| `←` `→` | Switch category / Select action |
| `↓` / `Enter` | Enter category / Confirm action |
| `↑` / `Esc` | Go back |
| `f` `p` `r` `c` `t` | Shortcuts: Feed / Play / Rest / Heal / Talk |
| `?` | Toggle help |
| `q` | Quit |

### In-Game

| Key | Function |
|-----|----------|
| Any key | Reaction test: Press when "GO!" appears |
| `0`-`9` + `Enter` | Number guessing: Input and confirm number |
| `Esc` | Exit game |
| `Enter` | Confirm result and return to main interface |

## Project Structure

```
clipet/
├── cmd/
│   ├── clipet/          # Main program entry
│   └── clipet-dev/      # Developer tools (timeskip, set, evolve, validate, preview)
├── internal/
│   ├── assets/          # Built-in species packs (go:embed)
│   ├── cli/             # Cobra CLI commands
│   ├── game/            # Core game logic (pet, evolution)
│   │   └── games/       # Mini games (types, manager, reaction, guess)
│   ├── plugin/          # Plugin system (types, parser, validator, loader, registry)
│   ├── store/           # Persistence (JSON)
│   └── tui/             # Bubble Tea TUI
│       ├── components/  # UI components (petview, dialoguebubble)
│       ├── screens/     # Screens (home, evolve)
│       └── styles/      # Themes and colors
└── docs/                # Documentation
```

## Development

```bash
# Developer tools
make dev

# Time skip to test decay
./clipet-dev timeskip --hours 2

# Manually set attributes (interactive)
./clipet-dev set

# Force evolution (interactive)
./clipet-dev evo to

# View evolution info
./clipet-dev evo info
```

## Custom Species

Clipet supports a fully customizable plugin system. Each species is an independent plugin pack containing:
- Species definition and evolution tree
- Lifecycle and ending configurations
- Dialogue library and adventure events
- ASCII animation frame files
- Personality traits (passives/actives/modifiers)

### Custom Attributes System (v3.0+)

Plugins can define custom attribute accumulators via adventure events to create unique evolution paths.

**How it works**:

1. **Define custom attributes**: Use any name as a custom attribute in adventure event effects
2. **Accumulate values**: Players accumulate custom attributes by completing adventure events
3. **Check evolution conditions**: The evolution system can check if custom attributes reach thresholds

**Example**: The built-in cat species pack defines 10 custom attributes supporting three evolution paths:

| Custom Attribute | Evolution Path | Acquisition Method |
|-----------------|----------------|-------------------|
| `arcane_affinity` | Arcane path | Complete arcane-themed adventures |
| `feral_affinity` | Feral path | Complete combat-themed adventures |
| `mech_affinity` | Mech path | Complete tech-themed adventures |

**TOML Configuration Example**:

```toml
# Accumulate custom attributes in adventure events
[[adventures]]
id = "mystic_shrine"
name = "Mystic Shrine"
[[adventures.choices]]
text = "Explore the shrine"
outcomes = [
  { weight = 50, text = "Feel arcane energy!", effects = { arcane_affinity = 10 } },
]

# Check custom attributes in evolution conditions
[[evolutions]]
from = "child"
to = "adult_arcane_shadow"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { arcane_affinity = 50 }  # Requires arcane affinity of 50
```

This design allows plugins to create unique evolution paths without modifying the core framework code.

Refer to the [Plugin Development Guide](docs/plugin-guide.en-US.md) and [Plugin Design Best Practices](docs/plugin-best-practices.md) to create custom species packs.

## Tech Stack

- **Go** 1.25+ — Language
- **Bubble Tea v2** — TUI framework (event-driven)
- **Lipgloss v2** — Terminal styling
- **Cobra** — CLI framework
- **TOML** — Species/dialogue configuration
- **JSON** — Save persistence + config files
- **i18n** — Lightweight internationalization framework (in-house, zero external dependencies)

## Internationalization (i18n)

Clipet supports multi-language interfaces, currently supporting Chinese and English.

### Switching Languages

**Temporary switch** (recommended):
```bash
# Use English
CLIPET_LANG=en-US ./clipet

# Use Chinese
CLIPET_LANG=zh-CN ./clipet
```

**Permanent switch**:
Edit the config file `~/.config/clipet/config.json`:
```json
{
  "language": "en-US",
  "fallback_language": "zh-CN"
}
```

### Language Detection Priority

1. `CLIPET_LANG` environment variable
2. `LANG` environment variable
3. `LC_ALL` environment variable
4. Config file `~/.config/clipet/config.json`
5. Default value `zh-CN`

For detailed i18n usage and development guide, see [docs/i18n-guide.en-US.md](docs/i18n-guide.en-US.md).

## License

MIT
