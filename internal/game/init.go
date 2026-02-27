package game

import (
	"clipet/internal/plugin"
)

// InitTimeSystem initializes the time evolution system.
// Must be called once at program startup.
func InitTimeSystem(pluginRegistry *plugin.Registry) {
	// Register core hooks (in priority order, highest first)
	RegisterTimeHook(NewDeathCheckHook(), PriorityCritical) // 100
	RegisterTimeHook(NewAttrDecayHook(pluginRegistry), PriorityHigh) // 80
	RegisterTimeHook(NewCooldownHook(), PriorityNormal)     // 50
	RegisterTimeHook(NewLifecycleHook(pluginRegistry), PriorityLow) // 20
}
