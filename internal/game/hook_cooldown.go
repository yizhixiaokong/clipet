package game

import "time"

// CooldownHook handles cooldown timestamp adjustments
type CooldownHook struct{}

func NewCooldownHook() *CooldownHook {
	return &CooldownHook{}
}

func (h *CooldownHook) Name() string {
	return "Cooldown"
}

func (h *CooldownHook) OnTimeAdvance(elapsed time.Duration, pet *Pet) {
	// Rewind all cooldown timestamps so time.Since() returns larger values
	pet.LastFedAt = pet.LastFedAt.Add(-elapsed)
	pet.LastPlayedAt = pet.LastPlayedAt.Add(-elapsed)
	pet.LastRestedAt = pet.LastRestedAt.Add(-elapsed)
	pet.LastHealedAt = pet.LastHealedAt.Add(-elapsed)
	pet.LastTalkedAt = pet.LastTalkedAt.Add(-elapsed)
	pet.LastAdventureAt = pet.LastAdventureAt.Add(-elapsed)

	// Update last checked time
	pet.LastCheckedAt = time.Now()
}
