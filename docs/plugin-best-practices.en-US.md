# Clipet Plugin Design Best Practices

**[中文](plugin-best-practices.md) | English**

## Overview

This document provides design principles, best practices, and common pitfalls for plugin development, helping you create balanced, fun, and maintainable species plugins.

## Offline Game Interaction Rhythm Design

### Core Challenge

Clipet is a **mostly offline** pet-raising game. Players typically interact only briefly when opening the interface. This creates unique design challenges:

**Problem Scenario**:
```
1. Player opens game, finds pet hunger at 20
2. Feeds once, recovers to 45, still far from enough
3. But there's a 10-minute cooldown, cannot feed again immediately
4. Player can only wait or leave the game, poor experience
```

### Design Principles

#### 1. **Single Operations Should Be Meaningful**

```toml
# ❌ Wrong: Effect too weak, single operation meaningless
[[actions]]
id = "feed"
cooldown = "10m"
[actions.effects]
hunger = 10   # Too little!

# ✅ Correct: Single operation has noticeable effect
[[actions]]
id = "feed"
cooldown = "10m"
[actions.effects]
hunger = 35   # From 20 → 55, noticeable improvement
happiness = 5
```

**Recommendation**: Single operations should restore 30-40% of attributes, giving players a sense of achievement.

#### 2. **Reasonable Cooldown Times**

```toml
# Cooldown rhythm examples for different species

# Fast-paced species (cat)
[[actions]]
id = "feed"
cooldown = "5m"    # Short cooldown, suitable for frequent interaction
[actions.effects]
hunger = 25

# Slow-paced species (dragon)
[[actions]]
id = "feed"
cooldown = "30m"   # Long cooldown, but strong single effect
[actions.effects]
hunger = 60
```

**Recommendations**:
- Fast-paced species: 5-10 minute cooldown
- Slow-paced species: 15-30 minute cooldown
- Longer cooldowns should have stronger effects

#### 3. **Dynamic Cooldown System (Recommended)**

Clipet uses **dynamic cooldown** to balance offline game experience. Cooldown time automatically adjusts based on attribute urgency:

```toml
[dynamic_cooldown]
# When hunger is very low (< 30), feeding cooldown is extremely short
low_urgency_multiplier = 0.1    # 10% cooldown
low_threshold = 30

# When hunger is medium (30-70), cooldown is moderate
medium_urgency_multiplier = 0.5  # 50% cooldown

# When hunger is high (>= 70), cooldown is normal
high_urgency_multiplier = 1.0    # 100% cooldown
high_threshold = 70
```

**Example** (Base feeding cooldown = 10 minutes):

| Hunger | Cooldown | Design Intent |
|-------|---------|---------|
| 10 | **1 minute** | Emergency, player can quickly feed multiple times |
| 40 | **5 minutes** | Medium urgency, moderate pace |
| 85 | **10 minutes** | Good condition, normal cooldown |

**Design Rationale**:
- **Avoid dead loops**: Low attributes won't trap players in "unable to recover" situations
- **Offline-friendly**: Brief online time can handle emergency states
- **Diminishing returns**: Normal cooldown when attributes are healthy, encourages diverse operations

#### 4. **Diversified Recovery Paths**

```toml
# Don't let players just wait for cooldowns

# Active skill: Consume resources for quick recovery
[[traits]]
id = "emergency_feed"
name = "Emergency Feed"
description = "Consume double energy, ignore cooldown to feed"
type = "active"
[traits.active_effect]
energy_cost = 20      # Higher than normal
hunger_restore = 30
cooldown = "1h"       # Long cooldown, but provides choice
```

### Specific Numerical Recommendations

**Base Cooldown Times** (configured in `[[actions]]`):

| Species Type | Base Cooldown | Single Effect | Characteristics |
|---------|---------|---------|------|
| Fast-paced (cat, rabbit)| 5-10m | 25-35 | Frequent interaction, quick feedback |
| Balanced (dog, fox)| 10-15m | 30-40 | Standard pace |
| Slow-paced (dragon, phoenix)| 20-30m | 50-70 | Fewer but stronger |
| Mythical | 30-60m | 70-90 | Rare but powerful |

**Decay Rate Recommendations** (configured in `[decay]`):

| Species Type | Hunger | Happiness | Energy | Health |
|---------|--------|-----------|--------|--------|
| Low maintenance | 0.5-0.8 | 0.3-0.5 | 0.2-0.3 | 0.1-0.2 |
| Standard | 1.0 | 0.5 | 0.3 | 0.2 |
| High maintenance | 1.5-2.0 | 0.8-1.0 | 0.4-0.5 | 0.3-0.5 |

**Recommended Configuration**: Use unified slow decay (hunger=1.0, happiness=0.5, energy=0.3, health=0.2), suitable for offline game experience.

## Action System Balance Guide

### Attribute Effect Ranges

```toml
# Recommended effect ranges (single operation)

[actions.effects]
hunger = 20-60      # Hunger
happiness = 10-30   # Happiness
health = 15-40      # Health
energy = 20-50      # Energy
```

**Balance Principles**:
1. **Total positive effects**: Sum of positive effects per operation should not exceed 80
2. **Negative effect limit**: Negative effects should not exceed 30% of positive effects
3. **Cooldown proportional to effect**: Stronger effects need longer cooldowns

### Example: Balanced Action Combination

```toml
# Feed: Primary recovery, small side effect
[[actions]]
id = "feed"
cooldown = "10m"
[actions.effects]
hunger = 35      # Primary effect
happiness = 5    # Small bonus

# Play: Strong happiness, consumes energy
[[actions]]
id = "play"
cooldown = "8m"
energy_cost = 10  # Prerequisite
[actions.effects]
happiness = 25   # Primary effect
energy = -15     # Consumption (greater than cost, includes physical exertion)

# Rest: Strong energy recovery, small negative
[[actions]]
id = "rest"
cooldown = "15m"
[actions.effects]
energy = 40      # Primary effect
health = 5       # Small bonus
happiness = -5   # Small negative (resting is boring)

# Heal: Strong health recovery, high cost
[[actions]]
id = "heal"
cooldown = "20m"
energy_cost = 15
[actions.effects]
health = 35      # Primary effect
```

## Evolution Tree Design

### Basic Principles

1. **Reachability**: Ensure all evolution paths can be completed within pet lifespan
2. **Diversity**: Provide multiple evolution choices to encourage replay
3. **Balance**: Different paths should have relatively balanced difficulty
4. **Thematic**: Each evolution path should have a clear theme

### Evolution Condition Design

**Age Limits**:
```toml
# ❌ Wrong: 10-day lifespan, but evolution requires 30 days
[lifecycle]
max_age_hours = 240.0  # 10 days

[[evolutions]]
from = "adult"
to = "legend"
[evolutions.condition]
min_age_hours = 720.0  # 30 days - Unreachable!

# ✅ Correct: Evolution time within lifespan
[lifecycle]
max_age_hours = 240.0  # 10 days

[[evolutions]]
from = "adult"
to = "legend"
[evolutions.condition]
min_age_hours = 200.0  # 8.3 days - Reachable
```

**Interaction Counts**:
```toml
# ❌ Wrong: Requires 1000 interactions, unrealistic
[evolutions.condition]
min_interactions = 1000

# ✅ Correct: Reasonable interaction count
[evolutions.condition]
min_interactions = 50  # Reachable in ~5-10 days
```

**Attribute Bias**:
```toml
# Use attribute bias to create different evolution paths
[[evolutions]]
from = "child"
to = "adult_arcane"
[evolutions.condition]
attr_bias = "happiness"  # Biased toward high happiness

[[evolutions]]
from = "child"
to = "adult_feral"
[evolutions.condition]
attr_bias = "health"     # Biased toward high health
```

### Custom Attributes System Design

Custom attributes are a powerful v3.0+ feature that allows plugins to create unique evolution paths.

**Design Principles**:

1. **Clear Theme**: Each custom attribute represents a clear direction or theme
2. **Progressive**: Accumulate gradually through multiple adventure events
3. **Balanceable**: Ensure different paths have relatively balanced difficulty

**Example: Three-Path Evolution**

```toml
# Arcane path - accumulated through arcane affinity
[[evolutions]]
from = "child"
to = "adult_arcane_shadow"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { arcane_affinity = 50 }

# Feral path - accumulated through feral affinity
[[evolutions]]
from = "child"
to = "adult_feral_flame"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { feral_affinity = 50 }

# Mech path - accumulated through mech affinity
[[evolutions]]
from = "child"
to = "adult_mech_cyber"
[evolutions.condition]
min_age_hours = 72.0
custom_acc = { mech_affinity = 50 }
```

**Adventure Event Design**:

```toml
# Arcane awakening event
[[adventures]]
id = "arcane_spark"
name = "Magic Spark"
description = "A mysterious purple light flickers before your eyes..."

[[adventures.choices]]
text = "Touch the magic spark"
outcomes = [
  { weight = 50, text = "Magical energy surges through your body!", effects = { arcane_affinity = 10, happiness = 15 } },
  { weight = 30, text = "The spark gently surrounds you.", effects = { arcane_affinity = 6, energy = 10 } },
  { weight = 20, text = "Leaves a faint mark.", effects = { arcane_affinity = 3 } },
]
```

**Balance Considerations**:

- **Accumulation speed**: Ensure players can reach target values in reasonable time
- **Weight distribution**: High-risk high-reward weights should be lower
- **Multi-path support**: Allow players to accumulate the same attribute through different events

## Dialogue Design

### Stage-Adapted Dialogue

Dialogue complexity should match life stages:

| Life Stage | Suggested Dialogue Content | Examples |
|---------|-------------|------|
| **egg** | Simple sounds only | `"Click..."`, `"Thump..."`, `"Shell rustles lightly"` |
| **baby** | Simple syllables, highly repetitive | `"Meow~"`, `"Meow meow~"`, `"Woo..."` |
| **child** | Simple short sentences, childish tone | `"Let's play together!"`, `"I'm so happy~"` |
| **adult** | Complete sentences, expressing own thoughts | `"I need some rest time."`, `"Thank you for taking care of me."` |
| **legend** | Profound, philosophical, grand scope | `"Protecting all beings is my mission."` |

### Mood State Feedback

**positive (happy)**: Active, energetic, anticipating interaction
```toml
[[dialogues]]
stage = ["adult_dragon"]
mood = ["happy"]
lines = [
  "The sunshine today is beautiful!",
  "I'm so happy to be with you!",
  "Let's go on an adventure!"
]
```

**neutral (normal)**: Calm, daily, gentle expression
```toml
[[dialogues]]
stage = ["child_dragon"]
mood = ["normal"]
lines = [
  "Hmm... today is okay.",
  "Want to eat something...",
  "Should we take a rest?"
]
```

**negative (unhappy/sad/miserable)**: Passive, tired, seeking comfort
```toml
[[dialogues]]
stage = ["baby_dragon"]
mood = ["sad"]
lines = [
  "Woo... woo woo...",
  "I'm so sad...",
  "Please comfort me..."
]
```

## Multi-language Support (i18n)

### When to Create Locale Files

**Recommended**: If you want to support multiple languages, or plan to in the future

**Unnecessary**: Only for single-language communities (e.g., Chinese users only)

### Locale File Best Practices

#### 1. **Complete Coverage of All Text**

```json
{
  "species": {
    "dragon": {
      "name": "Dragon",
      "description": "Ancient Dragon"
    }
  },
  "stages": {
    "egg": "Mysterious Egg",
    "baby": "Baby Dragon",
    "child_fire": "Fire Hatchling",
    "adult_fire": "Fire Dragon"
  },
  "dialogues": {
    "baby": {
      "happy": ["Awoo~", "Roar~"],
      "sad": ["Woo...", "Awo..."]
    }
  },
  "adventures": {
    "fire_shrine": {
      "name": "Fire Shrine",
      "description": "An ancient shrine burning with eternal flame...",
      "choices": {
        "enter": "Enter the shrine",
        "wait": "Wait outside"
      }
    }
  }
}
```

#### 2. **Keep Inline TOML Text as Fallback**

Even when creating locale files, keep default language text in TOML:

```toml
# species.toml
[species]
name = "龙"  # Keep Chinese as fallback
description = "远古巨龙"

[[stages]]
id = "egg"
name = "神秘之蛋"  # Keep Chinese as fallback
```

**Reasons**:
- Backward compatibility with older versions
- Fallback chain: `en-US` → `zh-CN` → TOML inline text
- Avoid displaying blank when translations are missing

#### 3. **Progressive Translation Strategy**

No need to translate everything at once, can add progressively:

**Phase 1**: Translate core content first
- Species names and descriptions
- Stage names

**Phase 2**: Then translate common content
- Common dialogues (happy, normal moods)
- Main adventure events

**Phase 3**: Finally translate detailed content
- All mood dialogues
- All adventure text

#### 4. **Locale File Naming Convention**

```
locales/
├── zh-CN.json     # Chinese (Simplified)
├── en-US.json     # English (US)
├── ja-JP.json     # Japanese
└── ko-KR.json     # Korean
```

Use standard language codes: `{language}-{region}` (e.g., `zh-CN`, `en-US`)

#### 5. **Testing Locale Files**

Validate JSON format with `jq`:
```bash
jq . locales/zh-CN.json
jq . locales/en-US.json
```

Test different languages:
```bash
CLIPET_LANG=en-US ./clipet
CLIPET_LANG=zh-CN ./clipet
```

### Multi-language Plugin Release Recommendations

1. **Include primary language by default**: Include at least one complete language in locale files
2. **Mark supported languages**: Specify supported languages in plugin README or species.toml
3. **Community contributions**: Welcome community contributions for new language translations
4. **Version control**: Update plugin version number when locale files change

## Lifecycle Design

### Lifespan and Evolution Matching

**Common Mistake**:
```toml
[lifecycle]
max_age_hours = 240.0  # 10 days

[[evolutions]]
from = "adult"
to = "legend"
[evolutions.condition]
min_age_hours = 720.0  # 30 days - Unreachable!
```

**Correct Approach**:
```toml
[lifecycle]
max_age_hours = 240.0  # 10 days

# Evolution timeline
# 0-1h: egg → baby
# 1-24h: baby → child
# 24-72h: child → adult
# 72-200h: adult → legend  ✓ Within lifespan
[[evolutions]]
from = "adult"
to = "legend"
[evolutions.condition]
min_age_hours = 200.0  # 8.3 days - Reachable!
```

**Principle**: Ensure players can experience the complete evolution chain within pet lifespan.

### Lifecycle Style Choices

```toml
# Short-term challenge (7 days)
[lifecycle]
max_age_hours = 168.0
warning_threshold = 0.6

# Standard experience (10-30 days)
max_age_hours = 240.0-720.0
warning_threshold = 0.75-0.85

# Long-term companionship (30-90 days)
max_age_hours = 720.0-2160.0
warning_threshold = 0.9

# Eternal pet
ending_type = "eternal"
```

## Safety Constraint System

To prevent plugin abuse (excessively short lifespans, extreme values, too many crises), the system implements lightweight safety boundaries.

### Default Constraints

| Constraint Type | Default Value | Description |
|---------|--------|------|
| Minimum lifespan | 24 hours | Prevent instant death (except eternal) |
| Maximum lifespan | 10 years | Prevent never dying (except eternal) |
| Attribute modifiers | 10% - 300% | Prevent extreme buffs/penalties |
| Adventure count limit | 20 | Prevent overwhelming players |
| Dialogue group limit | 100 | Prevent content overload |

### Constraint Override Mechanism

To override default constraints, must provide reason description (at least 50 characters):

```toml
# Species-level constraint override
[species]
id = "challenge_beetle"
name = "Challenge Beetle"

[constraints]
min_lifespan_hours = 1.0      # Override: 1-hour lifespan
reason = "Roguelike challenge mode: Players must complete objectives within 1 hour, experiencing intense and exciting extreme raising"
```

## Testing & Debugging

### Using clipet-dev Tools

```bash
# Validate species pack
./clipet-dev validate internal/assets/builtins/cat-pack

# Test evolution conditions
./clipet-dev evo info

# Force evolution test
./clipet-dev evo to legend_arcane_shadow

# Time skip test
./clipet-dev timeskip --hours 200

# Modify attributes test
./clipet-dev set happiness 95
./clipet-dev set age_hours 200
```

### Testing Checklist

When creating new species packs, ensure testing:

- [ ] **Basic functionality**: All stages have idle frames
- [ ] **Evolution paths**: At least one path can be completed within lifespan
- [ ] **Action balance**: Single operations have noticeable effects
- [ ] **Dialogue coverage**: Each stage has at least 3-5 dialogues
- [ ] **Adventure availability**: At least 3-5 adventures
- [ ] **Reasonable cooldowns**: Won't trap players in dead loops
- [ ] **Trait characteristics**: Trait effects are reasonable, not too strong or weak
- [ ] **Custom attributes**: Custom attributes can be accumulated normally through adventure events

## Common Pitfalls

### 1. Over-balancing

```toml
# ❌ Wrong: All effects are tiny, operations are meaningless
[actions.effects]
hunger = 5
happiness = 2
```

**Solution**: Make operations have noticeable effects, maintain game fun.

### 2. Excessive Cooldowns

```toml
# ❌ Wrong: 1-hour cooldown, player already closed the game
cooldown = "1h"
```

**Solution**: Cooldowns no longer than 30 minutes, unless effects are very strong.

### 3. Overly Strict Evolution Conditions

```toml
# ❌ Wrong: Requires 1000 interactions
[evolutions.condition]
min_interactions = 1000
```

**Solution**: Ensure evolution conditions can be achieved in reasonable time (< 500 interactions).

### 4. Ignoring Diminishing Returns

```toml
# ❌ Wrong: High attributes still restore a lot when feeding
# Should implement diminishing returns in code, not hardcode in config
```

**Solution**: Implement `diminish()` function in code, effects automatically decrease at high attributes.

### 5. Dialogue Mismatch with Stage

```toml
# ❌ Wrong: Baby stage speaking complex sentences
[[dialogues]]
stage = ["baby"]
mood = ["happy"]
lines = [
  "I think this philosophical viewpoint is worth exploring in depth."  # Too complex!
]
```

**Solution**: Ensure dialogue complexity matches life stage.

## Further Reading

- **Plugin Development Guide**: [plugin-guide.md](plugin-guide.md)
- **Architecture Design**: [CODEMAPS/architecture.md](CODEMAPS/architecture.md)
- **Core Logic**: [CODEMAPS/core-logic.md](CODEMAPS/core-logic.md)
- **Data Structures**: [CODEMAPS/data-structures.md](CODEMAPS/data-structures.md)
