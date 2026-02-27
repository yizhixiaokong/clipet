package game

// InitTimeSystem initializes the time evolution system.
// Must be called once at program startup.
func InitTimeSystem() {
	// Register core hooks (in priority order, highest first)
	RegisterTimeHook(NewDeathCheckHook(), PriorityCritical) // 100
	RegisterTimeHook(NewAttrDecayHook(), PriorityHigh)      // 80
	RegisterTimeHook(NewCooldownHook(), PriorityNormal)     // 50
}
