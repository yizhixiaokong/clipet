<!-- Generated: 2026-02-27 | Files scanned: 8 | Token estimate: ~900 -->

# Data Structures & Storage

## Plugin System Data

### SpeciesPack (plugin/types.go)

**Top-level container for all species data**

```go
type SpeciesPack struct {
    Species    SpeciesConfig    `toml:"species"`
    Lifecycle  LifecycleConfig  `toml:"lifecycle"`   // M7: Lifecycle config
    Stages     []Stage          `toml:"stages"`
    Evolutions []Evolution      `toml:"evolutions"`
    Traits     []PersonalityTrait `toml:"traits"`    // M7: Personality traits
    Endings    []Ending         `toml:"endings"`     // M7: Custom endings
    Dialogues  []DialogueGroup  `toml:"-"`  // Loaded from dialogues.toml
    Adventures []Adventure      `toml:"-"`  // Loaded from adventures.toml
    Frames     map[string]Frame `toml:"-"`  // Parsed from files

    Source PluginSource  // builtin/external
}
```

### SpeciesConfig

```go
type SpeciesConfig struct {
    ID, Name, Author string
    Version string
    BaseHunger, BaseHappiness, BaseHealth, BaseEnergy int
    FirstStage string
}
```

### LifecycleConfig (M7)

```go
type LifecycleConfig struct {
    MaxAgeHours      float64 `toml:"max_age_hours"`      // Default: 720 (30 days)
    EndingType       string  `toml:"ending_type"`        // death | ascend | eternal
    WarningThreshold float64 `toml:"warning_threshold"`  // Default: 0.8
}

// Defaults() provides backward compatibility for old plugins
```

### PersonalityTrait (M7)

```go
type PersonalityTrait struct {
    ID, Name, Description string
    Type string  // passive | active | modifier

    PassiveEffect *PassiveEffect
    ActiveEffect *ActiveEffect
    EvolutionModifier *EvolutionModifier
}

type PassiveEffect struct {
    FeedHungerBonus, FeedHappinessBonus float64
    PlayHappinessBonus, SleepEnergyBonus float64
    ResurrectChance, HealthRestorePercent float64
    HealthRegenMultiplier string  // Expression like "magic * 0.01"
}

type ActiveEffect struct {
    EnergyCost, HealthRestore, HungerRestore, HappinessBoost int
    Cooldown time.Duration
}

type EvolutionModifier struct {
    NightInteractionBonus, DayInteractionBonus float64
    FeedBonus, PlayBonus, AdventureBonus float64
}
```

### Ending (M7)

```go
type Ending struct {
    Type, Name, Message string
    Condition EndingCondition
}

type EndingCondition struct {
    MinHappiness int
    MinAgeHours float64
    MinAdventures int
}

// Endings are checked in order, first match wins
```

### Stage

```go
type Stage struct {
    ID, Name, Phase string
    Phases: egg | baby | child | adult | legend
}
```

### Evolution

```go
type Evolution struct {
    From, To string  // Stage IDs
    Conditions []EvolutionCondition
}

type EvolutionCondition struct {
    MinAgeHours float64
    MinAttr map[string]int

    // Accumulator requirements
    AccHappiness, AccHealth float64

    // Interaction requirements
    MinInteractions, MinFeedCount int
    MinDialogues, MinAdventures int

    // Biases
    AttrBias string  // "happiness" or "health"
    NightBias, DayBias bool
    MinFeedRegularity float64

    // Custom accumulators (NEW v3.0)
    CustomAcc map[string]int  // Plugin-defined: "fire_power": 30
}
```

### Frame (ASCII Animation)

```go
type Frame struct {
    StageID, AnimState string
    Width int
    Frames []string  // One string per animation frame
}

// File naming: {stage}.{variant}.{anim-state}.txt
// Example: adult.warrior.idle.txt
```

### DialogueGroup

```go
type DialogueGroup struct {
    Stage string
    Moods []string  // "开心", "饥饿", etc.
    Lines []string
}

// Selection: Random from matching stage + mood
```

### Adventure

```go
type Adventure struct {
    ID, Title, Text string
    Choices []Choice
    Requirements map[string]int  // "stage": "adult"
}

type Choice struct {
    Text string
    Effects map[string]int  // "hunger": -20
    Weight int
    RequiredAttr map[string]int
}
```

## Game Logic Data

### Pet (game/pet.go)

**Persistent entity stored in JSON**

```go
type Pet struct {
    // Identity
    Name string
    Species string
    StageID string
    Alive bool

    // Attributes (0-100)
    Hunger int
    Happiness int
    Health int
    Energy int

    // Custom attributes (M7)
    CustomAttributes map[string]int `json:"custom_attributes,omitempty"`

    // Timestamps (for cooldowns)
    Birthday time.Time
    LastFedAt time.Time
    LastPlayedAt time.Time
    LastRestedAt time.Time
    LastHealedAt time.Time
    LastTalkedAt time.Time
    LastAdventureAt time.Time
    LastCheckedAt time.Time  // For offline decay

    // Statistics
    TotalInteractions int
    GamesWon int
    AdventuresCompleted int
    DialogueCount int
    FeedCount int

    // Evolution accumulators
    AccHappiness float64
    AccHealth float64
    AccPlayful float64
    NightCount int
    DayCount int
    FeedRegularity float64

    // Lifecycle tracking (M7)
    LifecycleWarningShown bool `json:"lifecycle_warning_shown"`

    // Display state
    CurrentAnimation string
}
```

### EvolutionCandidate

```go
type EvolutionCandidate struct {
    Evolution *plugin.Evolution
    ToStage *plugin.Stage
    Score float64  // Higher = better match
}
```

## Storage Layer

### Store Interface (store/store.go)

```go
type Store interface {
    Save(pet *game.Pet) error
    Load() (*game.Pet, error)
    Exists() bool
    Delete() error
}
```

### JSONStore Implementation (store/jsonstore.go)

**Default path**: `~/.local/share/clipet/save.json`

**Write strategy**: Atomic write (tmp → rename)

```go
type JSONStore struct {
    path string  // Empty = use default
}

Save(pet):
  1. Marshal pet to JSON
  2. Write to temp file
  3. Rename temp → final (atomic)

Load():
  1. Read JSON file
  2. Unmarshal to Pet struct
  3. Apply offline decay compensation
```

## Plugin Registry

### Registry (plugin/registry.go)

**Thread-safe central store**

```go
type Registry struct {
    mu sync.RWMutex
    packs map[string]*SpeciesPack  // species_id → pack
}
```

**Key Methods**:

```
LoadFromFS(fsys, root, source)
  ↓
Scan for species.toml files
Parse each with loader.go
Validate with validator.go
Store in packs map

GetSpecies(id) → *SpeciesPack

GetStage(species, stageID) → *Stage

GetFrames(species, stage, anim) → []string
  ↓
Lookup: species.stages.{stage}.{anim}
Fallback: if anim not found → use "idle"

GetDialogue(species, stage, mood) → string
  ↓
Filter by stage + mood
Random selection

GetEvolutionsFrom(species, stage) → []Evolution
  ↓
Return all evolutions where From == stage

GetAdventures(species, stage) → []Adventure
  ↓
Filter by Requirements
```

## Capabilities Registry (M7)

### Registry (capabilities/registry.go)

```go
type Registry struct {
    mu sync.RWMutex
    traits map[string]map[string]PersonalityTrait  // species_id → trait_id → trait
}
```

**Key Methods**:

```
RegisterTraits(speciesID, traits)
  └─ Store in nested map

GetTrait(speciesID, traitID) → (PersonalityTrait, bool)
GetAllTraits(speciesID) → []PersonalityTrait

ApplyPassiveEffects(speciesID, action, hunger, happiness, health, energy)
  └─ Return modified attribute values

GetEvolutionModifier(speciesID) → *EvolutionModifier
  └─ Combine all modifier traits

GetActiveTraits(speciesID) → []PersonalityTrait
  └─ Return all active abilities
```

## Attributes System (M7)

### System (attributes/system.go)

```go
type System struct {
    coreAttrs   map[string]Definition  // hunger, happiness, health, energy
    customAttrs map[string]Definition  // Plugin-defined
}

type Definition struct {
    ID, DisplayName string
    Min, Max, Default int
    DecayRate float64  // Per hour
}
```

## Data Relationships

```
SpeciesPack
  ├─→ Lifecycle (1:1, M7)
  │    └─→ EndingType → ending trigger logic
  ├─→ Stages (1:N)
  │    └─→ Frames (N:M, via stage ID)
  ├─→ Evolutions (1:N)
  │    ├─→ From → Stage.ID
  │    └─→ To → Stage.ID
  ├─→ Traits (1:N, M7)
  │    ├─→ PassiveEffect → capability registry
  │    ├─→ ActiveEffect → capability registry
  │    └─→ EvolutionModifier → capability registry
  ├─→ Endings (1:N, M7)
  │    └─→ Condition → lifecycle trigger
  ├─→ Dialogues (1:N)
  │    └─→ Stage → Stage.ID
  └─→ Adventures (1:N)
       └─→ Requirements["stage"] → Stage.ID

Pet
  ├─→ Species → SpeciesPack.ID
  ├─→ StageID → Stage.ID
  └─→ CustomAttributes → attributes.System (M7)

EvolutionCandidate
  ├─→ Evolution → plugin.Evolution
  └─→ ToStage → plugin.Stage
```

## Validation Rules (plugin/validator.go)

**Pack validation checks**:
- Species.ID not empty
- Stages not empty
- All evolution From/To reference valid stages
- All dialogue stages reference valid stages
- All adventure requirements reference valid stages
- Frame files exist and match stage IDs
- No duplicate stage/evolution IDs
- **M7**: Lifecycle config has valid values (or uses defaults)
- **M7**: Trait IDs not duplicated per species
- **M7**: Ending conditions are valid

**Error types**:
- Missing required fields
- Invalid references
- Missing frame files
- Duplicate IDs
- **M7**: Invalid trait types
- **M7**: Invalid lifecycle thresholds
