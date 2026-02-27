package game

import "time"

// DeathCheckHook handles pet death checking
type DeathCheckHook struct{}

func NewDeathCheckHook() *DeathCheckHook {
	return &DeathCheckHook{}
}

func (h *DeathCheckHook) Name() string {
	return "DeathCheck"
}

func (h *DeathCheckHook) OnTimeAdvance(elapsed time.Duration, pet *Pet) {
	if pet.Health <= 0 {
		pet.Alive = false
	}
}
