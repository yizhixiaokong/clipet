# Clipet — Terminal Pet Companion Design Document

**[中文](design.md) | [English](design.en-US.md)**

## Project Overview

Clipet is a pet raising game (TUI) running in the terminal, with ASCII pixel art animation as the core visual representation. It supports complete interaction systems including feeding, mood, growth evolution, mini games, dialogue, and data persistence.

### Project Positioning

- **Type**: Terminal Desktop Pet / TUI Raising Game
- **Style**: Sci-Fi + Fantasy Mix (Fantasy Adventure)
- **Run Mode**: Hybrid — Full-screen TUI interaction + CLI shortcut commands

## Tech Stack

| Component          | Selection                        | Version  |
|--------------------|----------------------------------|----------|
| Language           | Go                               | 1.25+    |
| TUI Framework      | Bubble Tea v2                    | 2.0.0    |
| Styling            | Lipgloss v2                      | 2.0.0    |
| TUI Components     | Bubbles v2                       | 2.0.0 (indirect dependency) |
| CLI Framework      | Cobra                            | 1.10.2   |
| Config Format      | TOML (BurntSushi/toml)           | 1.6.0    |
| Internationalization | In-house lightweight i18n framework | —        |
| Animation Easing   | Harmonica                        | 0.2.0 (indirect dependency) |

> **Note**: Bubble Tea v2 uses `charm.land/bubbletea/v2` import path,
> not `github.com/charmbracelet/bubbletea/v2`. Same for Lipgloss v2 and Bubbles v2.

## Core Features

### 1. ASCII Pixel Art Animation

- Multi-frame animation system: idle, eating, sleeping, playing, happy, sad, etc.
- Frame files stored in species pack `frames/` directory, organized by stage hierarchy
- Recommended multi-level directory: `frames/{phase}/{variant}/idle.txt` (path joins to stageID)
- Supports sprite sheet format: `{animState}.txt` (multiple frames separated by `---`)
- Compatible root-level formats: `{stageID}_{animState}.txt` and frame-by-frame `{stageID}_{animState}_{index}.txt`
- Frame switching is timer-driven

### 2. Attribute System

Four core attributes (0-100):

| Attribute  | Description                        | Operation Impact (base value)     |
|------------|-----------------------------------|-----------------------------------|
| Hunger     | Satiety, higher = fuller          | Feed +25                          |
| Happiness  | Happiness level                   | Play +20, Feed +5, Talk +5, Rest -5 |
| Health     | Health value                      | Heal +25, Rest +5                 |
| Energy     | Energy level                      | Rest +30, Play -10, Heal -15      |

#### Diminishing Returns

All gain operations use diminishing returns formula, the closer to max the smaller the gain:

```
gain = base × (100 - current) / 100   (minimum 1)
```

Example: Feed base=25, when current Hunger=80, actual gain = 25×20/100 = 5.

#### Attribute Decay (Core Mechanism)

Attributes naturally decay over **real time**, even when the program isn't running. **Decay rates are controlled by plugin config** (Phase 7+):

**Default Decay Rates** (Cat species):

| Attribute  | Decay Rate    | Description                     |
|------------|--------------|--------------------------------|
| Hunger     | -1.0 / hour  | Unified slow decay              |
| Happiness  | -0.5 / hour  | Unified slow decay              |
| Energy     | -0.3 / hour  | Unified slow decay              |
| Health     | -0.2 / hour  | Only triggers when Hunger < 20  |

**Plugin Configurable** (species.toml `[decay]` section):

```toml
[decay]
hunger = 1.0        # Different species can have different decay rates
happiness = 0.5
energy = 0.3
health = 0.2
```

**Design Philosophy**:
- **Unified Slow Decay**: Suitable for offline games, players don't need to be online for long periods
- **Offline Friendly**: Even offline for days, pet state remains manageable
- **Extensibility**: Different species can have different maintenance difficulties

Offline Compensation: On program startup, calculates `time.Since(LastCheckedAt)` and deducts decay amount in one go.

#### Death Mechanism

- **Trigger Condition**: Health drops to 0
- **Permanence**: In current version, death is irreversible, no resurrection mechanism
- **Behavior**: After death, all operations return "Your pet has passed away..."
- **Typical Death Path**: Long-term no feeding → hunger reaches zero → health continuously drops → health reaches zero → death
- From full attributes to starvation takes approximately **~40 hours** of continuous neglect

#### Action Cooldowns & Dynamic System

**Base Cooldown Times** (controlled by plugin config, Phase 7+):

Each action has a minimum interval to prevent infinite attribute grinding:

| Action   | Base Cooldown | Prerequisite   | Plugin Configurable |
|----------|---------------|----------------|---------------------|
| Feed     | 10 minutes    | Hunger < 95    | ✅ |
| Play     | 5 minutes     | Energy ≥ 10    | ✅ |
| Talk     | 2 minutes     | —              | ✅ |
| Rest     | 15 minutes    | Energy < 90    | ✅ |
| Heal     | 20 minutes    | Energy ≥ 10    | ✅ |
| Adventure| 10 minutes    | Energy ≥ 15    | ✅ |

**Dynamic Cooldown System** (Phase 7+):

Cooldown times dynamically adjust based on attribute urgency, helping players handle emergencies:

```toml
[dynamic_cooldown]
# When attribute is very low (< 30): Very short cooldown (10% base cooldown)
low_urgency_multiplier = 0.1
low_threshold = 30

# When attribute is medium (30-70): Medium cooldown (50% base cooldown)
medium_urgency_multiplier = 0.5

# When attribute is high (>= 70): Normal cooldown (100% base cooldown)
high_urgency_multiplier = 1.0
high_threshold = 70
```

**Example** (Base feed cooldown = 10 minutes):
- Hunger = 10 → Cooldown = 1 minute (emergency help)
- Hunger = 50 → Cooldown = 5 minutes (medium)
- Hunger = 85 → Cooldown = 10 minutes (normal)

**Design Philosophy**:
- **Offline Friendly**: Brief online time can handle emergency states
- **Avoid Dead Loops**: Won't fall into "unable to recover" predicament when attributes are low
- **Diminishing Returns**: Normal cooldown when attributes are healthy, encouraging diverse actions

### 3. Mood System

Mood is calculated by weighted attributes:
```
MoodScore = Hunger × 0.25 + Happiness × 0.35 + Health × 0.25 + Energy × 0.15
```

| Score Range | Mood Name   | Animation State |
|-------------|------------|-----------------|
| 81-100      | happy      | AnimHappy       |
| 61-80       | normal     | AnimIdle        |
| 41-60       | unhappy    | AnimIdle        |
| 21-40       | sad        | AnimSad         |
| 0-20        | miserable  | AnimSad         |

### 4. Growth & Evolution System

Five life stages:

```
Egg → Baby → Child → Adult → Legend
       │       │       │       │
     1 stage 3 branches 6 branches 6 branches
```

Evolution condition combinations:
- `min_age_hours` — Minimum age
- `attr_bias` — Attribute bias (happiness/health/playful)
- `min_dialogues` / `min_adventures` — Interaction counts
- `min_feed_regularity` — Feeding regularity
- `night_interactions_bias` / `day_interactions_bias` — Time period preference
- `min_interactions` — Total interaction count
- `min_attr` — Attribute thresholds

### 5. Dialogue System

- Match dialogue content by stage + mood
- Supports wildcard matching (`*` matches all stages/moods)
- Randomly selects candidate lines each dialogue

#### Auto Chat

Pet will actively speak in TUI interface:
- Every 3 minutes has 30% probability to trigger auto dialogue bubble
- 1 minute retry after failure
- Matches current stage + mood dialogue library
- Won't trigger during games/adventures

### 6. Adventure System

- Choice-based adventure events, 4-stage flow: Intro → Choice → Animation → Result
- Filters available adventures by stage (supports wildcards like `child_*`)
- Each choice has weighted random outcomes (`weight` field controls probability)
- Results affect attribute values (display colored +/- changes)
- Fixed cost of 10 energy, 10 minute cooldown
- Completion count `AdventuresCompleted` counts toward evolution conditions

### 7. Persistence

- JSON format save file
- Atomic write (tmp + rename)
- Default path: `~/.local/share/clipet/save.json`

### 8. Plugin System

- Species/evolution trees loaded via plugin packs
- Built-in species packed via `go:embed`
- External plugins loaded from `~/.local/share/clipet/plugins/`
- Unified `fs.FS` loading interface
- TOML description format

### 9. Configuration System

Config file path: `~/.config/clipet/config.json`

**Config Items**:
- `language`: Current language (e.g., `zh-CN`, `en-US`)
- `fallback_language`: Fallback language
- `version`: Config version

**Language Detection Priority**:
1. `CLIPET_LANG` environment variable
2. `LANG` environment variable
3. `LC_ALL` environment variable
4. Config file
5. Default value (`zh-CN`)

### 10. Internationalization System (i18n)

**Architecture**: Zero external dependency in-house lightweight framework

**Core Components**:
- `internal/i18n/`: Manager, Bundle, Loader, Plural
- `internal/config/`: Config management and language detection
- `internal/assets/locales/`: Translation files (embed)
- Plugin locale support: `locales/{lang}.json`

**API**:
```go
i18n.T("ui.home.feed_success", "oldHunger", 50, "newHunger", 75)
i18n.TN("game.stats.interactions", count, "count", count)
```

**Features**:
- Template variable interpolation (`{{.variable}}`)
- Plural handling
- Translation fallback chain: requested language → fallback language → inline text
- Thread-safe (RWMutex)
- Compile-time embed translation files

**Translation File Structure**:
```
internal/assets/locales/
├── zh-CN/
│   ├── tui.json       # TUI interface
│   ├── game.json      # Game logic
│   └── cli.json       # CLI commands
└── en-US/
    └── ...
```

**Currently Supported Languages**:
- 🇨🇳 Chinese (Simplified) - `zh-CN`
- 🇺🇸 English - `en-US`

## Run Modes

### Full-screen TUI Mode

```bash
clipet          # Launch full-screen TUI interactive interface
```

### CLI Shortcut Commands

```bash
clipet init     # Create new pet
clipet status   # View pet status (-j outputs JSON)
clipet          # Launch TUI main interface
```

## Data Paths

| Path                                  | Purpose         |
|---------------------------------------|-----------------|
| `~/.local/share/clipet/save.json`     | Pet save file   |
| `~/.local/share/clipet/plugins/`      | External plugin directory |
