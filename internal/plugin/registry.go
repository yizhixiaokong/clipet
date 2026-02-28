package plugin

import (
	"clipet/internal/game/capabilities"
	"fmt"
	"io/fs"
	"math/rand"
	"strings"
	"sync"
)

// Registry is the central store for all loaded species packs.
// Both builtin and external packs are registered through the same interface.
type Registry struct {
	mu     sync.RWMutex
	packs  map[string]*SpeciesPack // keyed by species ID
	loader *Loader
	lang   string // current language for locale loading
	fallbackLang string // fallback language
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		packs:        make(map[string]*SpeciesPack),
		loader:       NewLoader(),
		lang:         "zh-CN", // default language
		fallbackLang: "en-US",
	}
}

// SetLanguage sets the active language for locale loading.
// Must be called before LoadFromFS to take effect.
func (r *Registry) SetLanguage(lang, fallback string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lang = lang
	r.fallbackLang = fallback
}

// LoadFromFS loads all species packs from a filesystem and registers them.
// Packs with the same ID will be overwritten (external overrides builtin).
// Uses the language set by SetLanguage for locale loading.
func (r *Registry) LoadFromFS(fsys fs.FS, root string, source PluginSource) error {
	r.mu.RLock()
	r.loader.SetLanguage(r.lang, r.fallbackLang)
	r.mu.RUnlock()

	packs, err := r.loader.LoadAll(fsys, root, source)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, pack := range packs {
		r.packs[pack.Species.ID] = pack
	}

	return nil
}

// Register adds or replaces a single species pack in the registry.
func (r *Registry) Register(pack *SpeciesPack) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.packs[pack.Species.ID] = pack
}

// GetSpecies returns a species pack by ID, or nil if not found.
func (r *Registry) GetSpecies(id string) *SpeciesPack {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.packs[id]
}

// GetTraitName returns the localized name for a trait, with fallback to inline TOML name.
func (r *Registry) GetTraitName(speciesID, traitID string) string {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return traitID
	}

	// Find trait definition
	var trait *capabilities.PersonalityTrait
	for i := range pack.Traits {
		if pack.Traits[i].ID == traitID {
			trait = &pack.Traits[i]
			break
		}
	}
	if trait == nil {
		return traitID
	}

	// Try locale first
	if pack.Locale != nil {
		key := fmt.Sprintf("traits.%s.name", traitID)
		if localized := getLocaleValue(pack.Locale.Data, key); localized != "" {
			return localized
		}
	}

	// Fallback to inline TOML name
	return trait.Name
}

// GetTraitDescription returns the localized description for a trait, with fallback to inline TOML description.
func (r *Registry) GetTraitDescription(speciesID, traitID string) string {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return ""
	}

	// Find trait definition
	var trait *capabilities.PersonalityTrait
	for i := range pack.Traits {
		if pack.Traits[i].ID == traitID {
			trait = &pack.Traits[i]
			break
		}
	}
	if trait == nil {
		return ""
	}

	// Try locale first
	if pack.Locale != nil {
		key := fmt.Sprintf("traits.%s.description", traitID)
		if localized := getLocaleValue(pack.Locale.Data, key); localized != "" {
			return localized
		}
	}

	// Fallback to inline TOML description
	return trait.Description
}

// ListSpecies returns all registered species IDs and names.
// Uses locale if available, falls back to inline TOML names.
func (r *Registry) ListSpecies() []SpeciesInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []SpeciesInfo
	for _, pack := range r.packs {
		name := pack.Species.Name
		description := pack.Species.Description

		// Try locale first
		if pack.Locale != nil {
			if localized := getLocaleValue(pack.Locale.Data, "species."+pack.Species.ID+".name"); localized != "" {
				name = localized
			}
			if localized := getLocaleValue(pack.Locale.Data, "species."+pack.Species.ID+".description"); localized != "" {
				description = localized
			}
		}

		list = append(list, SpeciesInfo{
			ID:          pack.Species.ID,
			Name:        name,
			Description: description,
			Author:      pack.Species.Author,
			Version:     pack.Species.Version,
			Source:      pack.Source,
		})
	}
	return list
}

// SpeciesInfo is a summary of a registered species.
type SpeciesInfo struct {
	ID          string
	Name        string
	Description string
	Author      string
	Version     string
	Source      PluginSource
}

// GetStage returns the Stage definition for a given species and stage ID.
// Localizes the stage name if locale is available.
func (r *Registry) GetStage(speciesID, stageID string) *Stage {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}
	for _, s := range pack.Stages {
		if s.ID == stageID {
			stage := s // copy

			// Try locale for stage name
			if pack.Locale != nil {
				if localized := getLocaleValue(pack.Locale.Data, "stages."+stageID); localized != "" {
					stage.Name = localized
				}
			}

			return &stage
		}
	}
	return nil
}

// GetEggStage returns the first egg stage for a species.
func (r *Registry) GetEggStage(speciesID string) *Stage {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}
	for _, s := range pack.Stages {
		if s.Phase == PhaseEgg {
			return &s
		}
	}
	return nil
}

// GetEvolutionsFrom returns all possible evolutions from a given stage.
func (r *Registry) GetEvolutionsFrom(speciesID, stageID string) []Evolution {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}
	var evos []Evolution
	for _, e := range pack.Evolutions {
		if e.From == stageID {
			evos = append(evos, e)
		}
	}
	return evos
}

// GetFrames returns the Frame for a specific stage and animation state.
// Falls back to "idle" if the requested animation has no frames.
func (r *Registry) GetFrames(speciesID, stageID, animState string) *Frame {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}
	key := FrameKey(stageID, animState)
	if f, ok := pack.Frames[key]; ok {
		return &f
	}
	// Fallback to idle
	if animState != "idle" {
		key = FrameKey(stageID, "idle")
		if f, ok := pack.Frames[key]; ok {
			return &f
		}
	}
	return nil
}

// GetDialogue returns a random dialogue line matching the stage and mood.
// Uses locale if available, falls back to inline TOML dialogues.
func (r *Registry) GetDialogue(speciesID, stageID, mood string) string {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return ""
	}

	// Try locale first
	if pack.Locale != nil {
		dialogueKey := fmt.Sprintf("dialogues.%s.%s", stageID, mood)
		if lines := getLocaleArray(pack.Locale.Data, dialogueKey); len(lines) > 0 {
			return lines[rand.Intn(len(lines))]
		}
		// Try "normal" as fallback mood
		if mood != "normal" {
			dialogueKey = fmt.Sprintf("dialogues.%s.normal", stageID)
			if lines := getLocaleArray(pack.Locale.Data, dialogueKey); len(lines) > 0 {
				return lines[rand.Intn(len(lines))]
			}
		}
	}

	// Fallback to inline TOML dialogues
	var candidates []string
	for _, dg := range pack.Dialogues {
		if !matchesStage(dg.Stage, stageID) {
			continue
		}
		if !matchesMood(dg.Mood, mood) {
			continue
		}
		candidates = append(candidates, dg.Lines...)
	}

	if len(candidates) == 0 {
		return ""
	}
	return candidates[rand.Intn(len(candidates))]
}

// GetAdventures returns adventures available for the given stage.
// Uses locale if available, falls back to inline TOML adventures.
func (r *Registry) GetAdventures(speciesID, stageID string) []Adventure {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}

	var result []Adventure
	for _, adv := range pack.Adventures {
		if matchesStage(adv.Stage, stageID) {
			// Create a copy for localization
			localizedAdv := adv

			// Try locale for adventure name and description
			if pack.Locale != nil {
				advKey := "adventures." + adv.ID
				if localized := getLocaleValue(pack.Locale.Data, advKey+".name"); localized != "" {
					localizedAdv.Name = localized
				}
				if localized := getLocaleValue(pack.Locale.Data, advKey+".description"); localized != "" {
					localizedAdv.Description = localized
				}

				// Localize choice texts
				for i := range localizedAdv.Choices {
					choiceKey := advKey + ".choices." + fmt.Sprintf("%d", i)
					if localized := getLocaleValue(pack.Locale.Data, choiceKey); localized != "" {
						localizedAdv.Choices[i].Text = localized
					} else {
						// Try by choice ID if available
						choiceIDKey := advKey + ".choices." + localizedAdv.Choices[i].Text
						if locText := getLocaleValue(pack.Locale.Data, choiceIDKey); locText != "" {
							localizedAdv.Choices[i].Text = locText
						}
					}

					// Localize outcome texts
					for j := range localizedAdv.Choices[i].Outcomes {
						outcomeKey := advKey + ".outcomes." + fmt.Sprintf("%d_%d", i, j)
						if localized := getLocaleValue(pack.Locale.Data, outcomeKey); localized != "" {
							localizedAdv.Choices[i].Outcomes[j].Text = localized
						}
					}
				}
			}

			result = append(result, localizedAdv)
		}
	}
	return result
}

// GetBaseStats returns the base stats for a species.
func (r *Registry) GetBaseStats(speciesID string) *BaseStats {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}
	return &pack.Species.BaseStats
}

// GetAction returns the action configuration for a given species and action ID.
// Returns nil if action is not defined (caller should use defaults).
func (r *Registry) GetAction(speciesID, actionID string) *ActionConfig {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}
	for i := range pack.Actions {
		if pack.Actions[i].ID == actionID {
			return &pack.Actions[i]
		}
	}
	return nil
}

// GetDecayConfig returns the decay configuration for a species.
// Returns defaults if not configured.
func (r *Registry) GetDecayConfig(speciesID string) capabilities.DecayConfig {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return capabilities.DecayConfig{}.Defaults()
	}
	return pack.Decay.Defaults()
}

// GetDynamicCooldownConfig returns the dynamic cooldown configuration for a species.
// Returns defaults if not configured.
func (r *Registry) GetDynamicCooldownConfig(speciesID string) capabilities.DynamicCooldownConfig {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return capabilities.DynamicCooldownConfig{}.Defaults()
	}
	return pack.DynamicCooldown.Defaults()
}

// GetAttributeInteractionConfig returns the attribute interaction configuration for a species.
// Returns defaults if not configured.
func (r *Registry) GetAttributeInteractionConfig(speciesID string) capabilities.AttributeInteractionConfig {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return capabilities.AttributeInteractionConfig{}.Defaults()
	}
	return pack.Interactions.Defaults()
}

// GetEndingMessage returns the localized ending message for a given ending type.
// Uses locale if available, falls back to inline TOML message.
func (r *Registry) GetEndingMessage(speciesID, endingType string) string {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return ""
	}

	// Try locale first
	if pack.Locale != nil {
		endingKey := "endings." + endingType
		if msg := getLocaleValue(pack.Locale.Data, endingKey); msg != "" {
			return msg
		}
	}

	// Fallback to inline TOML endings
	for _, ending := range pack.Endings {
		if ending.Type == endingType {
			return ending.Message
		}
	}

	return ""
}

// Count returns the number of registered species.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.packs)
}

// matchesStage checks if a stageID matches any of the stage patterns.
// Supports wildcards: "*" matches all, "child_*" matches any child stage.
func matchesStage(patterns []string, stageID string) bool {
	for _, p := range patterns {
		if p == "*" {
			return true
		}
		if strings.Contains(p, "*") {
			prefix := strings.TrimSuffix(p, "*")
			if strings.HasPrefix(stageID, prefix) {
				return true
			}
		}
		if p == stageID {
			return true
		}
	}
	return false
}

// matchesMood checks if a mood matches any of the mood patterns.
func matchesMood(patterns []string, mood string) bool {
	for _, p := range patterns {
		if p == "*" {
			return true
		}
		if p == mood {
			return true
		}
	}
	return false
}

// Unregister removes a species pack from the registry.
func (r *Registry) Unregister(speciesID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.packs[speciesID]; !ok {
		return fmt.Errorf("species %q not found", speciesID)
	}
	delete(r.packs, speciesID)
	return nil
}

// InstallFromFS installs a plugin from a filesystem directory into the registry.
func (r *Registry) InstallFromFS(fsys fs.FS, dir string) (*SpeciesPack, error) {
	pack, err := r.loader.LoadOne(fsys, dir, SourceExternal)
	if err != nil {
		return nil, err
	}
	r.Register(pack)
	return pack, nil
}

// getLocaleValue navigates a nested map using dot notation key (e.g., "species.cat.name").
// Returns empty string if key not found.
func getLocaleValue(data map[string]interface{}, key string) string {
	parts := strings.Split(key, ".")
	var current interface{} = data

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return ""
			}
			// If this is the last part, return the string value
			if i == len(parts)-1 {
				if str, ok := current.(string); ok {
					return str
				}
				return ""
			}
		default:
			return ""
		}
	}
	return ""
}

// getLocaleArray navigates a nested map and returns a string array.
// Returns empty array if key not found or not an array.
func getLocaleArray(data map[string]interface{}, key string) []string {
	parts := strings.Split(key, ".")
	var current interface{} = data

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil
			}
			// If this is the last part, return the array value
			if i == len(parts)-1 {
				if arr, ok := current.([]interface{}); ok {
					var result []string
					for _, item := range arr {
						if str, ok := item.(string); ok {
							result = append(result, str)
						}
					}
					return result
				}
				return nil
			}
		default:
			return nil
		}
	}
	return nil
}
