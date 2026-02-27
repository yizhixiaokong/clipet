<!-- Generated: 2026-02-27 | Files scanned: 8 | Token estimate: ~700 -->

# Core Game Logic

## Package: internal/game/

**Principle**: Zero UI dependencies (no TUI/CLI imports)

### Key Files

| File | Lines | Purpose |
|------|-------|---------|
| pet.go | ~280 | Pet entity, attributes, actions, decay |
| evolution.go | ~150 | Evolution engine, condition checking |
| adventure.go | ~100 | Adventure system, weighted random |

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
