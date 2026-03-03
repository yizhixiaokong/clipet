# Clipet Plugin Development Guide

**[中文](plugin-guide.md) | English**

## Overview

Clipet's species system is entirely plugin-pack based. Each species is an independent directory containing TOML configuration files and ASCII animation frame files. Built-in species and external plugins use the exact same format and loading path.

## Plugin Pack Directory Structure

```
my-species-pack/
├── species.toml        # Required — Species definition + evolution tree
├── dialogues.toml      # Optional — Dialogue library
├── adventures.toml     # Optional — Adventure events
├── locales/            # Optional — Multi-language translations (Phase 3+)
│   ├── zh-CN.json      # Chinese translations
│   └── en-US.json      # English translations
└── frames/             # Optional — ASCII animation frames
    ├── egg/              # Stage-organized directories
    │   └── idle.txt
    ├── baby/
    │   └── xxx/
    │       ├── idle.txt
    │       ├── eating.txt
    │       └── ...
    ├── child/
    │   ├── variant_a/
    │   └── variant_b/
    ├── adult/
    │   └── .../
    └── legend/
        └── .../
```

## species.toml

### Species Metadata

```toml
[species]
id = "dragon"                    # Required, unique identifier
name = "Dragon"                  # Required, display name
description = "Ancient Dragon"   # Description text
author = "your-name"             # Author
version = "1.0.0"                # Required, semantic version

[species.base_stats]             # Initial attribute values (0-100)
hunger = 50
happiness = 60
health = 70
energy = 65
```

### Stage Definitions

Each stage is a node on the evolution tree:

```toml
[[stages]]
id = "egg"           # Stage unique ID
name = "Dragon Egg"  # Display name
phase = "egg"        # Life phase: egg | baby | child | adult | legend
```

### Lifecycle Configuration (v2.0+)

Define lifecycle parameters and ending types for pets:

```toml
[lifecycle]
max_age_hours = 240.0       # Maximum lifespan (hours), default 720 hours (30 days)
ending_type = "death"       # Ending type: death | ascend | eternal | loop
warning_threshold = 0.75    # Warning threshold (0.0-1.0), shows gentle reminder when reached
```

**Lifecycle Types**:

| Type | Description |
|------|-------------|
| `death` | Natural passing (default) |
| `ascend` | Ascension/transcendence (warm theme) |
| `eternal` | Eternal existence, never triggers ending |
| `loop` | Rebirth cycle, resets age instead of death when lifespan reached |

### Personality Trait Definitions (v2.0+)

Define species personality traits, divided into three categories: passive traits, active skills, and evolution modifiers.

**Passive Traits** (passive) - Automatically effective features:

```toml
[[traits]]
id = "picky_eater"
name = "Picky Eater"
description = "Picky about food, but happier when fed"
type = "passive"
[traits.passive_effect]
feed_hunger_bonus = -0.2     # Feed hunger -20%
feed_happiness_bonus = 0.1   # But happiness +10%
```

**Active Skills** (active) - Player-triggered skills:

```toml
[[traits]]
id = "purr_heal"
name = "Purr Heal"
description = "Consume energy to heal through purring"
type = "active"
[traits.active_effect]
energy_cost = 10             # Costs 10 energy
health_restore = 15          # Restores 15 health
cooldown = "30m"             # 30 minute cooldown
```

**Evolution Modifiers** (modifier) - Affect evolution point accumulation:

```toml
[[traits]]
id = "night_owl"
name = "Night Owl"
description = "More active at night"
type = "modifier"
[traits.evolution_modifier]
night_interaction_bonus = 1.5  # Night interaction evolution points +50%
```

**Available Fields**:

| Passive Effect Field | Type | Description |
|---------------------|------|-------------|
| `feed_hunger_bonus` | float | Feed hunger gain multiplier (-1.0 to 1.0) |
| `feed_happiness_bonus` | float | Feed happiness gain multiplier |
| `play_happiness_bonus` | float | Play happiness gain multiplier |
| `sleep_energy_bonus` | float | Rest energy gain multiplier |
| `resurrect_chance` | float | Resurrection probability on death (0.0-1.0) |
| `health_restore_percent` | float | Health restore percentage on resurrection |

| Active Skill Field | Type | Description |
|-------------------|------|-------------|
| `energy_cost` | int | Active skill energy cost |
| `health_restore` | int | Active skill health restore amount |
| `hunger_restore` | int | Active skill hunger restore amount |
| `happiness_boost` | int | Active skill happiness boost amount |
| `cooldown` | duration | Active skill cooldown (e.g., "30m", "1h") |

| Evolution Modifier Field | Type | Description |
|------------------------|------|-------------|
| `night_interaction_bonus` | float | Night interaction evolution points multiplier |
| `day_interaction_bonus` | float | Day interaction evolution points multiplier |
| `feed_bonus` | float | Feed evolution points multiplier |
| `play_bonus` | float | Play evolution points multiplier |
| `adventure_bonus` | float | Adventure evolution points multiplier |

### Ending Definitions (v2.0+)

Define multiple possible endings for the species based on pet lifecycle quality:

```toml
# Blissful Passing - High happiness + Long life
[[endings]]
type = "blissful_passing"
name = "Blissful Passing"
message = "With a satisfied smile, your cat peacefully passed away..."
[endings.condition]
min_happiness = 80
min_age_hours = 200.0

# Adventurous Life - Completed many adventures
[[endings]]
type = "adventurous_life"
name = "Adventurous Life"
message = "After a life full of adventures, it became a legend..."
[endings.condition]
min_adventures = 30

# Peaceful Rest - Default ending
[[endings]]
type = "peaceful_rest"
name = "Peaceful Rest"
message = "After living a peaceful life, it has departed..."
[endings.condition]
```

**Ending Condition Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `min_happiness` | int | Minimum happiness score (0-100) |
| `min_age_hours` | float | Minimum survival time (hours) |
| `min_adventures` | int | Minimum completed adventures |

### Action Configuration (v3.0+, Phase 7)

Define species action behaviors, including cooldown times and effect values:

```toml
# Feed action
[[actions]]
id = "feed"
cooldown = "10m"              # Cooldown time
[actions.effects]
hunger = 25                   # Hunger +25
happiness = 5                 # Happiness +5

# Play action
[[actions]]
id = "play"
cooldown = "5m"
energy_cost = 10              # Requires 10 energy to execute
[actions.effects]
happiness = 20
energy = -10                  # Energy -10

# Rest action
[[actions]]
id = "rest"
cooldown = "15m"
[actions.effects]
energy = 30                   # Energy +30
health = 5                    # Health +5
happiness = -5                # Happiness -5 (resting might be boring)

# Heal action
[[actions]]
id = "heal"
cooldown = "20m"
energy_cost = 15
[actions.effects]
health = 25

# Talk action
[[actions]]
id = "talk"
cooldown = "2m"
[actions.effects]
happiness = 5
```

**Action Field Descriptions**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Action ID, must be one of feed/play/rest/heal/talk |
| `cooldown` | duration | Yes | Cooldown time (e.g., "10m", "1h30m") |
| `energy_cost` | int | No | Energy required to execute (omit if no cost) |
| `effects.hunger` | int | No | Hunger change (positive = increase) |
| `effects.happiness` | int | No | Happiness change |
| `effects.health` | int | No | Health change |
| `effects.energy` | int | No | Energy change (can be negative) |

### Decay Configuration (v3.0+, Phase 7)

Define species attribute decay rates:

```toml
[decay]
hunger = 1.0        # Hunger decays 1 point per hour
happiness = 0.5     # Happiness decays 0.5 points per hour
energy = 0.3        # Energy decays 0.3 points per hour
health = 0.2        # Health decays 0.2 points per hour when hungry
```

### Dynamic Cooldown Configuration (v3.0+, Phase 7)

Define dynamic cooldown system that automatically adjusts cooldown times based on attribute urgency:

```toml
[dynamic_cooldown]
# Low urgency (attribute < 30): Very short cooldown
low_urgency_multiplier = 0.1    # 10% base cooldown
low_threshold = 30

# Medium urgency (30 <= attribute < 70): Medium cooldown
medium_urgency_multiplier = 0.5  # 50% base cooldown

# High urgency (attribute >= 70): Normal cooldown
high_urgency_multiplier = 1.0    # 100% base cooldown
high_threshold = 70
```

**Dynamic Cooldown Example**:

Assuming base feed cooldown is 10 minutes:

- Hunger = 10 (very low) → Cooldown = 10m × 0.1 = **1 minute**
- Hunger = 50 (medium) → Cooldown = 10m × 0.5 = **5 minutes**
- Hunger = 85 (high) → Cooldown = 10m × 1.0 = **10 minutes**

### Evolution Conditions

Evolution conditions support multiple check types:

```toml
[[evolutions]]
from = "child"
to = "adult_fire"
[evolutions.condition]
min_age_hours = 72.0            # Minimum age
min_interactions = 50           # Minimum interaction count
attr_bias = "happiness"         # Attribute bias
min_attr = { happiness = 70 }   # Minimum attribute requirements

# Custom attribute accumulators (v3.0+)
custom_acc = { fire_power = 30 }  # Requires fire power of 30

# Time preference
night_bias = true                # Night preference
day_bias = false                 # Day preference
```

**Evolution Condition Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `min_age_hours` | float | Minimum age (hours) |
| `min_attr` | map | Minimum attribute requirements (e.g., `{happiness = 70}`) |
| `min_interactions` | int | Minimum interaction count |
| `min_feed_count` | int | Minimum feed count |
| `min_dialogues` | int | Minimum dialogue count |
| `min_adventures` | int | Minimum adventure count |
| `min_feed_regularity` | float | Minimum feed regularity (0-1) |
| `attr_bias` | string | Attribute bias ("happiness", "health", "playful") |
| `night_bias` | bool | Night preference |
| `day_bias` | bool | Day preference |
| `custom_acc` | map | Custom accumulator requirements (v3.0+) |

## dialogues.toml

### Basic Format

Each dialogue group contains stage matching, mood matching, and multiple alternative dialogues:

```toml
[[dialogues]]
stage = ["baby_dragon", "child_dragon"]  # Stage matching (supports wildcards)
mood = ["happy", "normal"]                # Mood matching (supports "*" for all)
lines = [
  "Hello! I'm a little dragon~",
  "Nice weather today~",
  "Let's play together!",
]
```

### Stage Matching Rules

- `"*"` — Matches all stages
- `"child_*"` — Prefix wildcard, matches all stages starting with `child_`
- `"baby_dragon"` — Exact match

### Mood Values

| Mood Name | Mood Score Range |
|-----------|-----------------|
| happy | 81-100 |
| normal | 61-80 |
| unhappy | 41-60 |
| sad | 21-40 |
| miserable | 0-20 |

## adventures.toml

### Basic Format

```toml
[[adventures]]
id = "treasure_cave"
name = "Treasure Cave"
stage = ["child_*", "adult_*"]  # Available stages
description = "You discover a shimmering cave entrance..."

[[adventures.choices]]
text = "Brave the depths"

[[adventures.choices.outcomes]]
weight = 60                     # Weight for weighted random
text = "Found a pile of delicious food!"
[adventures.choices.outcomes.effects]
hunger = 20                     # Attribute change amounts
happiness = 10

# Custom attribute accumulator (v3.0+)
fire_power = 5                  # Fire power +5

[[adventures.choices.outcomes]]
weight = 40
text = "The cave is empty."
[adventures.choices.outcomes.effects]
happiness = -5

[[adventures.choices]]
text = "Take a detour"

[[adventures.choices.outcomes]]
weight = 100
text = "Left safely."
[adventures.choices.outcomes.effects]
energy = 5
```

### Adventure Effect Fields

| Field | Type | Description |
|-------|------|-------------|
| `hunger` | int | Hunger change |
| `happiness` | int | Happiness change |
| `health` | int | Health change |
| `energy` | int | Energy change |
| `{custom_attr}` | int | Custom attribute change (v3.0+) |

## Animation Frame Files

### Directory Layout

**Layout 1: Multi-level Subdirectories (Recommended)**

```
frames/
  {phase}/
    {variant}/
      {animState}.txt     # Sprite sheet
      {animState}_{index}.txt  # Frame-by-frame (legacy)
```

Directory names at each level are joined with `_` to reconstruct the stageID.
For example, `frames/adult/arcane_shadow/idle.txt` → stageID = `adult_arcane_shadow`.

**Layout 2: Root-level Sprite Sheets**

```
frames/{stageID}_{animState}.txt
```

**Layout 3: Root-level Frame Files (Legacy)**

```
frames/{stageID}_{animState}_{index}.txt
```

> **Priority**: Multi-level subdirectories > Root-level sprite sheets > Root-level frame files. Within the same (stageID, animState), sprite sheets take priority over frame files.

### Sprite Sheet Format

Multiple frames of the same action are placed in one file, separated by `---` lines:

```
 /-/\
(' ' )
 | \/
U-U(_/
---
 \-/\
(' ' )
 | \/
U-U(_/
```

### Supported Animation States

| animState | Trigger Condition |
|-----------|------------------|
| idle | Default state (must provide) |
| eating | When feeding |
| sleeping | Energy below 15 |
| playing | When playing |
| happy | Mood > 80 |
| sad | Mood < 40 |

If a frame file is missing for a state, it automatically falls back to `idle`.

## Custom Attributes System (v3.0+)

Plugins can define custom attribute accumulators through adventure events to create unique evolution paths.

### Defining Custom Attributes

Use custom attribute names in adventure event effects:

```toml
[[adventures]]
id = "fire_shrine"
name = "Fire Shrine"
description = "An ancient shrine burning with eternal flame..."

[[adventures.choices]]
text = "Offer power"
outcomes = [
  { weight = 50, text = "Fire power surges through your body!", effects = { fire_power = 10, happiness = 15 } },
  { weight = 30, text = "Feel the warmth of the flames.", effects = { fire_power = 5, energy = 10 } },
]
```

### Using in Evolution Conditions

```toml
[[evolutions]]
from = "child"
to = "adult_fire"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { fire_power = 30 }  # Requires fire power of 30
```

### How It Works

1. **Accumulation**: Accumulate custom attribute values through adventure events
2. **Checking**: Check if custom attribute requirements are met during evolution
3. **Persistence**: Custom attribute values are saved in the save file

**Notes**:
- Custom attribute values start from 0
- The same custom attribute can be increased through multiple adventure events
- Custom attributes do not affect core four attributes (hunger, happiness, health, energy)

## Installing External Plugins

Place the plugin directory in `~/.local/share/clipet/plugins/`:

```
~/.local/share/clipet/plugins/
└── my-dragon-pack/
    ├── species.toml
    ├── dialogues.toml
    ├── adventures.toml
    └── frames/
        ├── egg/
        │   └── idle.txt
        ├── baby/
        │   └── dragon/
        └── ...
```

The program automatically scans this directory on startup and loads all valid species packs.

## Validation Rules

The following validations are automatically performed on load:

1. **Required Fields**: `species.id`, `species.name`, `species.version`
2. **Stage Completeness**: At least one egg stage
3. **Evolution Path Validity**: Referenced stage IDs in from/to must exist
4. **Evolution Chain Connectivity**: All non-egg stages must be reachable from some egg stage
5. **Dialogue References**: Non-wildcard stage references must point to defined stages
6. **Adventure Structure**: Each adventure has at least one choice, each choice has at least one outcome
7. **Frame Files**: Egg stage must have idle frames

On validation failure, the entire plugin pack will be rejected with detailed error message list.

## Development Tools

Use the `clipet-dev` tool for development and testing:

```bash
# Validate species pack
./clipet-dev validate internal/assets/builtins/cat-pack

# Test evolution conditions
./clipet-dev evo info

# Force evolution testing
./clipet-dev evo to legend_arcane_shadow

# Time skip testing
./clipet-dev timeskip --hours 200

# Modify attributes testing
./clipet-dev set happiness 95
./clipet-dev set age_hours 200
```

## Reference: Built-in Cat Species Pack

See the `internal/assets/builtins/cat-pack/` directory for a complete reference implementation.

Evolution tree overview:

```
egg (Mysterious Egg)
 └── baby (Kitten)
      ├── child_arcane (Arcane Kitten)  ← happiness bias + dialogue count
      │    ├── adult_arcane_shadow (Shadow Charm Cat)  ← night bias
      │    │    └── legend_arcane_shadow (Void Walker)
      │    └── adult_arcane_crystal (Crystal Prophet Cat)  ← day bias
      │         └── legend_arcane_crystal (Star Sage)
      ├── child_feral (Combat Kitten)  ← health bias + adventure count
      │    ├── adult_feral_flame (Flame Lion)  ← hunger bias
      │    │    └── legend_feral_flame (Immortal Flame Emperor)
      │    └── adult_feral_frost (Frost Storm Leopard)  ← energy bias
      │         └── legend_feral_frost (Extreme Frost God)
      └── child_mech (Mecha Kitten)  ← playful bias + feed regularity
           ├── adult_mech_cyber (Cyber Lynx)  ← dialogue count
           │    └── legend_mech_cyber (Quantum Ghost)
           └── adult_mech_chrome (Alloy Cheetah)  ← adventure count
                └── legend_mech_chrome (Interstellar Predator)
```

## Multi-language Support (Locale)

### Locale File Structure

Plugins can provide multi-language translation files to support different languages for species names, stages, dialogues, etc.:

```
my-species-pack/
└── locales/
    ├── zh-CN.json     # Chinese translations
    └── en-US.json     # English translations
```

### locale.json Format

```json
{
  "species": {
    "dragon": {
      "name": "Dragon",
      "description": "Ancient dragon, master of elements"
    }
  },
  "stages": {
    "egg": "Mysterious Egg",
    "baby": "Baby Dragon",
    "adult_fire": "Fire Dragon",
    "adult_ice": "Ice Dragon"
  },
  "dialogues": {
    "baby": {
      "happy": ["Awoo~", "Roar~"],
      "sad": ["Whimper...", "Aww..."]
    },
    "adult_fire": {
      "happy": ["ROAR!!", "Fire burns in my heart!"]
    }
  },
  "adventures": {
    "explore_cave": {
      "name": "Explore Dragon Cave",
      "description": "The baby dragon wants to explore an ancient dragon cave...",
      "choices": {
        "enter": "Enter the cave",
        "wait": "Wait outside"
      }
    }
  }
}
```

### Fallback Mechanism

When the requested language is unavailable, the system falls back in order:

1. **Requested language** (e.g., `en-US`)
2. **Fallback language** (e.g., `zh-CN`)
3. **Inline TOML text** (from species.toml, dialogues.toml)

This means plugins work even without locale files (using TOML text).

### Usage Recommendations

- **Create locale files**: If you want to support multiple languages
- **Keep inline text**: Preserve default language text in TOML files as fallback
- **Progressive translation**: You can translate parts first, untranslated parts fall back to inline text

### Example: cat-pack locale

See `internal/assets/builtins/cat-pack/locales/` for complete examples.

## Further Reading

- **Design Best Practices**: [plugin-best-practices.md](plugin-best-practices.md)
- **Architecture Design**: [CODEMAPS/architecture.md](CODEMAPS/architecture.md)
- **Core Logic**: [CODEMAPS/core-logic.md](CODEMAPS/core-logic.md)
- **Data Structures**: [CODEMAPS/data-structures.md](CODEMAPS/data-structures.md)
