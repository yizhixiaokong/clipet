package game

import (
	"fmt"
	"time"

	"clipet/internal/game/capabilities"
	"clipet/internal/plugin"
)

// LifecycleHook integrates lifecycle checks with the time advancement system
type LifecycleHook struct {
	lifecycleMgr *LifecycleManager
}

// NewLifecycleHook creates a new lifecycle hook
func NewLifecycleHook(pluginRegistry *plugin.Registry) *LifecycleHook {
	return &LifecycleHook{
		lifecycleMgr: NewLifecycleManager(pluginRegistry),
	}
}

// Name returns the hook name for logging
func (h *LifecycleHook) Name() string {
	return "lifecycle"
}

// OnTimeAdvance is called when time passes for a pet
func (h *LifecycleHook) OnTimeAdvance(elapsed time.Duration, pet *Pet) {
	if !pet.Alive {
		return
	}

	state := h.lifecycleMgr.CheckLifecycle(pet)

	// Show warning when approaching end of life (only once)
	if state.NearEnd && !pet.LifecycleWarningShown {
		pet.LifecycleWarningShown = true
		// Note: In a full implementation, we would emit a message to the UI
		// For now, we just mark the flag so the UI can check it
	}

	// Eternal pets never trigger endings
	if state.IsEternal {
		return
	}

	// Looping lifecycle resets age instead of ending
	if state.IsLooping && state.AgePercent >= 1.0 {
		pet.LifecycleWarningShown = false
		// Note: In a full implementation, we would emit a rebirth message to the UI
		fmt.Printf("[Lifecycle] Pet %s has completed a life cycle and begins anew\n", pet.Name)
		return
	}

	// Trigger ending when pet reaches end of natural lifespan
	if state.AgePercent >= 1.0 {
		result := h.lifecycleMgr.TriggerEnding(pet)
		h.applyEnding(pet, result)
	}
}

// applyEnding applies the ending result to the pet
func (h *LifecycleHook) applyEnding(pet *Pet, result capabilities.EndingResult) {
	pet.Alive = false
	pet.EndingType = result.Type
	pet.EndingMessage = result.Message // Plugin-provided message (may be empty)

	// The UI can use pet.EndingType for i18n lookup, or pet.EndingMessage if provided
	fmt.Printf("[Lifecycle] Pet %s has reached the end: [%s] %s\n", pet.Name, result.Type, result.Message)
}
