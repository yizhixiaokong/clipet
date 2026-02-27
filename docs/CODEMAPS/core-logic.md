<!-- Generated: 2026-02-27 | Files scanned: 17 | Token estimate: ~950 -->

# Core Game Logic

## Package: internal/game/

**Principle**: Zero UI dependencies (no TUI/CLI imports)

### Key Files

| File | Lines | Purpose |
|------|-------|---------|
| pet.go | ~480 | Pet entity, attributes, actions, decay |
| evolution.go | ~150 | Evolution engine, condition checking |
| adventure.go | ~100 | Adventure system, weighted random |
| lifecycle_manager.go | ~120 | Lifecycle checks and ending triggers (M7) |
| capabilities/types.go | ~95 | Capability and trait definitions (M7) |
| capabilities/registry.go | ~145 | Trait registration and application (M7) |
| attributes/system.go | ~145 | Flexible attribute management (M7) |

## Pet System (pet.go)

### Data Structure

```go
type Pet struct {
    Name, Species, StageID string
    Hunger, Happiness, Health, Energy int  // 0-100

    // Timestamps for cooldowns
    LastFedAt, LastPlayedAt, LastRestedAt time.Time

    // Statistics
    TotalInteractions, FeedCount int
    AccHappiness, AccHealth float64

    // Lifecycle tracking (M7)
    LifecycleWarningShown bool
    CustomAttributes map[string]int  // Plugin-defined attrs

    Alive bool
}
```

### Core Methods

```
Feed() → gain = base * (100 - current) / 100  // Diminishing returns
Play() → same formula
Rest() → energy recovery
Talk() → happiness boost
Heal() → health restoration

MoodScore() → (Hunger + Happiness + Health + Energy) / 4
MoodName() → "开心", "普通", "饥饿", "生病", etc.

GetAttr(name) → unified interface for core + custom attrs (M7)
SetAttr(name, value) → set custom attributes (M7)

ApplyOfflineDecay() → compensate time since last save
SimulateDecay() → hourly decay rates:
  - Hunger: -3/hour
  - Happiness: -2/hour
  - Energy: -1/hour
  - Health: -0.5/hour if Hunger < 20
```

### Cooldown System

| Action | Cooldown | Gain |
|--------|----------|------|
| Feed | 10min | Hunger +25 |
| Play | 5min | Happiness +15 |
| Rest | 15min | Energy +30 |
| Talk | 3min | Happiness +5 |
| Heal | 30min | Health +20 |

## Evolution System (evolution.go)

### Condition Types

```go
type EvolutionCondition struct {
    MinAgeHours float64
    MinAttr map[string]int  // "happiness": 60
    MinInteractions int
    MinFeedRegularity float64
    AttrBias string  // "happiness" or "health"
    NightBias, DayBias bool
}
```

### Evolution Flow

```
CheckEvolution(pet, registry) → []Candidate
  ↓
For each Evolution from current stage:
  ├─ Check age ≥ MinAgeHours
  ├─ Check all MinAttr requirements
  ├─ Check bias conditions (AttrBias, NightBias)
  └─ If all pass → add to candidates

BestCandidate(candidates) → *Candidate
  ↓
Score = ageScore + attrScore + interactionScore
Return highest scored

DoEvolve(pet, candidate) → update StageID
```

### Trigger Points

Evolution check runs automatically after:
- `Feed()`, `Play()`, `Rest()`, `Talk()`, `Heal()`
- Any attribute modification

## Adventure System (adventure.go)

### Structure

```go
type Adventure struct {
    ID, Title, Text string
    Choices []Choice
    Requirements map[string]int
}

type Choice struct {
    Text string
    Effects map[string]int  // "hunger": -20
    Weight int
    RequiredAttr map[string]int
}
```

### Selection Algorithm

```
GetRandomAdventure(pet, registry) → Adventure
  ↓
Filter by Requirements (stage, attributes)
Weighted random selection from available pool

SelectChoice(adventure, pet) → Choice
  ↓
Filter by RequiredAttr
Weighted random pick
```

## Lifecycle System (M7)

### LifecycleManager (lifecycle_manager.go)

**Purpose**: Time-driven lifecycle management and endings

```
CheckLifecycle(pet) → LifecycleState
  ├─ Calculate agePercent = AgeHours / MaxAgeHours
  ├─ NearEnd = agePercent >= WarningThreshold
  └─ Return {NearEnd, AgePercent}

TriggerEnding(pet) → EndingResult
  ├─ Check custom endings from species pack
  ├─ Fallback to default endings:
  │   ├─ blissful_passing (happiness > 90, interactions > 500)
  │   ├─ adventurous_life (adventures > 30)
  │   └─ peaceful_rest (default)
  └─ Return {Type, Message}
```

### LifecycleHook (hook_lifecycle.go)

**Integration**: Registered with TimeManager at PriorityLow

```
OnTimeAdvance(elapsed, pet):
  ├─ CheckLifecycle(pet)
  ├─ If NearEnd and !LifecycleWarningShown:
  │   └─ Set LifecycleWarningShown = true
  └─ If AgePercent >= 1.0:
      └─ TriggerEnding(pet), Set Alive = false
```

## Capabilities System (M7)

### Types (capabilities/types.go)

```go
type LifecycleConfig struct {
    MaxAgeHours      float64  // Default: 720 (30 days)
    EndingType       string   // death | ascend | eternal
    WarningThreshold float64  // Default: 0.8 (80%)
}

type PersonalityTrait struct {
    ID, Name, Description string
    Type string  // passive | active | modifier

    PassiveEffect *PassiveEffect
    ActiveEffect *ActiveEffect
    EvolutionModifier *EvolutionModifier
}

type PassiveEffect struct {
    FeedHungerBonus     float64  // -0.2 = -20% hunger gain
    FeedHappinessBonus  float64  // 0.1 = +10% happiness gain
    ResurrectChance     float64  // 0.3 = 30% chance
    HealthRestorePercent float64
}

type ActiveEffect struct {
    EnergyCost int
    HealthRestore int
    Cooldown time.Duration
}

type EvolutionModifier struct {
    NightInteractionBonus float64  // 1.5 = +50%
    DayInteractionBonus   float64
}

type Ending struct {
    Type, Name, Message string
    Condition EndingCondition
}
```

### Registry (capabilities/registry.go)

**Thread-safe trait management**

```
RegisterTraits(speciesID, traits)
  └─ Store in map[species_id][trait_id]

ApplyPassiveEffects(speciesID, action, hunger, happiness, health, energy)
  └─ Apply multipliers for all passive traits

GetEvolutionModifier(speciesID) → *EvolutionModifier
  └─ Combine all modifier traits

GetActiveTraits(speciesID) → []PersonalityTrait
  └─ Return all active abilities
```

## Attributes System (M7)

### System (attributes/system.go)

**Purpose**: Manage core 4 + custom plugin-defined attributes

```go
type Definition struct {
    ID, DisplayName string
    Min, Max, Default int
    DecayRate float64  // Per hour
}

type System struct {
    coreAttrs   map[string]Definition  // hunger, happiness, health, energy
    customAttrs map[string]Definition  // Plugin-defined
}
```

### Operations

```
RegisterCustomAttribute(def) → error
  └─ Validate no collision with core attrs

GetDefinition(id) → (Definition, bool)
  └─ Lookup in core + custom

GetDecayRate(id) → float64
  └─ Return hourly decay rate

ValidateValue(id, value) → (int, error)
  └─ Clamp to [Min, Max]

GetAllAttributes() → []string
  └─ Return core + custom IDs
```

### Usage in Pet

```go
// Pet stores custom attributes
CustomAttributes map[string]int

// Unified access
GetAttr(name) → int
  ├─ If core attr (hunger, happiness, health, energy):
  │   └─ Return from struct field
  └─ Else:
      └─ Return from CustomAttributes map

SetAttr(name, value)
  └─ Store in CustomAttributes map
```

## Minigames (games/)

### Interface

```go
type MiniGame interface {
    Start(pet *Pet) tea.Model
    CheckWin() bool
    GetReward() map[string]int
}
```

### Implemented Games

| Game | File | Type | Reward |
|------|------|------|--------|
| Reaction | reaction.go | Speed test | Happiness +10 |
| Guess | guess.go | Number guessing | Happiness +15 |
