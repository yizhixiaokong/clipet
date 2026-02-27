// Package plugin provides the species pack loading system.
// Both builtin and external species packs are loaded through
// the same Registry.LoadFromFS interface.
package plugin

import (
	"clipet/internal/game/capabilities"
	"time"
)

// PluginSource indicates where a plugin was loaded from.
type PluginSource string

const (
	SourceBuiltin  PluginSource = "builtin"
	SourceExternal PluginSource = "external"
)

// SpeciesPack represents a complete species plugin package,
// containing species definition, evolution tree, dialogues,
// adventures, and ASCII art frames.
type SpeciesPack struct {
	Species       SpeciesConfig       `toml:"species"`
	Lifecycle     capabilities.LifecycleConfig     `toml:"lifecycle"`     // Phase 2: lifecycle configuration
	Decay         capabilities.DecayConfig         `toml:"decay"`         // Phase 7: attribute decay rates
	DynamicCooldown capabilities.DynamicCooldownConfig `toml:"dynamic_cooldown"` // Phase 7: dynamic cooldown config
	Stages        []Stage            `toml:"stages"`
	Evolutions    []Evolution        `toml:"evolutions"`
	Traits        []capabilities.PersonalityTrait `toml:"traits"` // Phase 1: personality traits
	Endings       []capabilities.Ending `toml:"endings"` // Phase 2: possible endings
	Actions       []ActionConfig     `toml:"actions"` // Phase 7: action configurations
	Dialogues     []DialogueGroup    `toml:"-"` // loaded from dialogues.toml
	Adventures    []Adventure        `toml:"-"` // loaded from adventures.toml
	Frames        map[string]Frame   `toml:"-"` // loaded from frames/ directory
	Scripts       ScriptsConfig      `toml:"scripts"`
	Source        PluginSource       `toml:"-"`
}

// SpeciesConfig holds the species metadata and base stats.
type SpeciesConfig struct {
	ID          string    `toml:"id"`
	Name        string    `toml:"name"`
	Description string    `toml:"description"`
	Author      string    `toml:"author"`
	Version     string    `toml:"version"`
	BaseStats   BaseStats `toml:"base_stats"`
}

// BaseStats defines the starting attribute values for a new pet.
type BaseStats struct {
	Hunger    int `toml:"hunger"`
	Happiness int `toml:"happiness"`
	Health    int `toml:"health"`
	Energy    int `toml:"energy"`
}

// Stage represents one node in the evolution tree.
type Stage struct {
	ID    string `toml:"id"`
	Name  string `toml:"name"`
	Phase string `toml:"phase"` // egg, baby, child, adult, legend
}

// Evolution defines a directed edge in the evolution tree.
type Evolution struct {
	From      string             `toml:"from"`
	To        string             `toml:"to"`
	Condition EvolutionCondition `toml:"condition"`
}

// EvolutionCondition specifies the requirements for an evolution to occur.
type EvolutionCondition struct {
	MinAgeHours       float64        `toml:"min_age_hours"`
	AttrBias          string         `toml:"attr_bias"` // happiness, health, playful
	MinDialogues      int            `toml:"min_dialogues"`
	MinAdventures     int            `toml:"min_adventures"`
	MinFeedRegularity float64        `toml:"min_feed_regularity"`
	NightBias         bool           `toml:"night_interactions_bias"`
	DayBias           bool           `toml:"day_interactions_bias"`
	MinInteractions   int            `toml:"min_interactions"`
	MinAttr           map[string]int `toml:"min_attr"`
}

// ActionConfig defines a pet action (feed, play, rest, etc.) - Phase 7
type ActionConfig struct {
	ID         string        `toml:"id"`
	Cooldown   time.Duration `toml:"cooldown"`
	EnergyCost int           `toml:"energy_cost"` // Energy required to perform action
	Effects    ActionEffects `toml:"effects"`
}

// ActionEffects defines the attribute changes from an action - Phase 7
type ActionEffects struct {
	Hunger    int `toml:"hunger"`    // Change to hunger (positive = gain)
	Happiness int `toml:"happiness"` // Change to happiness
	Health    int `toml:"health"`    // Change to health
	Energy    int `toml:"energy"`    // Change to energy (can be negative)
}

// DialogueGroup is a set of dialogue lines associated with
// specific evolution stages and mood conditions.
type DialogueGroup struct {
	Stage []string `toml:"stage"` // stage IDs or "*" for all
	Mood  []string `toml:"mood"`  // mood names or "*" for all
	Lines []string `toml:"lines"`
}

// DialoguesFile is the top-level structure of dialogues.toml.
type DialoguesFile struct {
	Dialogues []DialogueGroup `toml:"dialogues"`
}

// Adventure represents a random event that can occur during gameplay.
type Adventure struct {
	ID          string            `toml:"id"`
	Name        string            `toml:"name"`
	Stage       []string          `toml:"stage"` // stage IDs, supports wildcards
	Description string            `toml:"description"`
	Choices     []AdventureChoice `toml:"choices"`
}

// AdventuresFile is the top-level structure of adventures.toml.
type AdventuresFile struct {
	Adventures []Adventure `toml:"adventures"`
}

// AdventureChoice represents one option the player can pick.
type AdventureChoice struct {
	Text     string             `toml:"text"`
	Outcomes []AdventureOutcome `toml:"outcomes"`
}

// AdventureOutcome is a weighted result of an adventure choice.
type AdventureOutcome struct {
	Weight  int            `toml:"weight"`
	Text    string         `toml:"text"`
	Effects map[string]int `toml:"effects"` // attribute changes
}

// Frame holds the ASCII art frames for a specific stage+animation combination.
type Frame struct {
	StageID   string   // e.g. "baby"
	AnimState string   // e.g. "idle"
	Frames    []string // each element is one frame of ASCII art
	Width     int      // display width (auto-calculated from content)
}

// FrameKey returns the lookup key for a frame set: "{stageID}_{animState}".
func FrameKey(stageID, animState string) string {
	return stageID + "_" + animState
}

// ScriptsConfig holds paths for future script extension support.
type ScriptsConfig struct {
	OnEvolve    string `toml:"on_evolve"`
	OnAdventure string `toml:"on_adventure"`
	CustomMood  string `toml:"custom_mood"`
}

// StagePhase constants.
const (
	PhaseEgg    = "egg"
	PhaseBaby   = "baby"
	PhaseChild  = "child"
	PhaseAdult  = "adult"
	PhaseLegend = "legend"
)

// ValidPhases is the set of valid phase values.
var ValidPhases = map[string]bool{
	PhaseEgg:    true,
	PhaseBaby:   true,
	PhaseChild:  true,
	PhaseAdult:  true,
	PhaseLegend: true,
}
