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
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		packs:  make(map[string]*SpeciesPack),
		loader: NewLoader(),
	}
}

// LoadFromFS loads all species packs from a filesystem and registers them.
// Packs with the same ID will be overwritten (external overrides builtin).
func (r *Registry) LoadFromFS(fsys fs.FS, root string, source PluginSource) error {
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

// ListSpecies returns all registered species IDs and names.
func (r *Registry) ListSpecies() []SpeciesInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []SpeciesInfo
	for _, pack := range r.packs {
		list = append(list, SpeciesInfo{
			ID:          pack.Species.ID,
			Name:        pack.Species.Name,
			Description: pack.Species.Description,
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
func (r *Registry) GetStage(speciesID, stageID string) *Stage {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}
	for _, s := range pack.Stages {
		if s.ID == stageID {
			return &s
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
func (r *Registry) GetDialogue(speciesID, stageID, mood string) string {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return ""
	}

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
func (r *Registry) GetAdventures(speciesID, stageID string) []Adventure {
	pack := r.GetSpecies(speciesID)
	if pack == nil {
		return nil
	}

	var result []Adventure
	for _, adv := range pack.Adventures {
		if matchesStage(adv.Stage, stageID) {
			result = append(result, adv)
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
