<!-- Generated: 2026-02-27 | Files scanned: 6 | Token estimate: ~750 -->

# Data Structures & Storage

## Plugin System Data

### SpeciesPack (plugin/types.go)

**Top-level container for all species data**

```go
type SpeciesPack struct {
    Species    SpeciesConfig    `toml:"species"`
    Stages     []Stage          `toml:"stages"`
    Evolutions []Evolution      `toml:"evolutions"`
    Dialogues  []DialogueGroup  `toml:"dialogues"`
    Adventures []Adventure      `toml:"adventures"`
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

## Data Relationships

```
SpeciesPack
  ├─→ Stages (1:N)
  │    └─→ Frames (N:M, via stage ID)
  ├─→ Evolutions (1:N)
  │    ├─→ From → Stage.ID
  │    └─→ To → Stage.ID
  ├─→ Dialogues (1:N)
  │    └─→ Stage → Stage.ID
  └─→ Adventures (1:N)
       └─→ Requirements["stage"] → Stage.ID

Pet
  ├─→ Species → SpeciesPack.ID
  └─→ StageID → Stage.ID

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

**Error types**:
- Missing required fields
- Invalid references
- Missing frame files
- Duplicate IDs
