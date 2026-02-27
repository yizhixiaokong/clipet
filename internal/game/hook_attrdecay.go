package game

import "time"

// AttrDecayHook handles attribute decay over time
type AttrDecayHook struct{}

func NewAttrDecayHook() *AttrDecayHook {
	return &AttrDecayHook{}
}

func (h *AttrDecayHook) Name() string {
	return "AttrDecay"
}

func (h *AttrDecayHook) OnTimeAdvance(elapsed time.Duration, pet *Pet) {
	hours := elapsed.Hours()

	// Basic attribute decay
	pet.Hunger = clamp(pet.Hunger-int(3*hours), 0, 100)
	pet.Happiness = clamp(pet.Happiness-int(2*hours), 0, 100)
	pet.Energy = clamp(pet.Energy-int(1*hours), 0, 100)

	// Health decay due to hunger
	if pet.Hunger < 20 {
		pet.Health = clamp(pet.Health-int(0.5*hours), 0, 100)
	}
}
