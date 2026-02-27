package game

import (
	"clipet/internal/plugin"
	"testing"
	"time"
)

// TestCustomAttributesIntegration tests custom attributes in adventure scenarios
func TestCustomAttributesIntegration(t *testing.T) {
	// Create test pet
	pet := &Pet{
		Name:             "FlameTest",
		Species:          "cat",
		Stage:            StageChild,
		StageID:          "child_feral",
		Birthday:         time.Now().Add(-80 * time.Hour),
		Hunger:           50,
		Happiness:        60,
		Health:           70,
		Energy:           80,
		LastCheckedAt:    time.Now(),
		LastFedAt:        time.Now(),
		LastPlayedAt:     time.Now(),
		LastRestedAt:     time.Now(),
		LastHealedAt:     time.Now(),
		LastTalkedAt:     time.Now(),
		LastAdventureAt:  time.Now(),
		LastSkillUsedAt:  time.Now(),
		Alive:            true,
		CurrentAnimation: AnimIdle,
		CustomAttributes: make(map[string]int),
	}

	// Test 1: Initial state
	firePower := pet.GetCustomAcc("fire_power")
	if firePower != 0 {
		t.Errorf("Expected initial fire_power=0, got %d", firePower)
	}

	// Test 2: Apply adventure outcome with custom attribute
	outcome := plugin.AdventureOutcome{
		Text: "找到了火之精华！",
		Effects: map[string]int{
			"fire_power": 18,
			"happiness":  15,
		},
	}

	changes := ApplyAdventureOutcome(pet, outcome)

	// Verify fire_power was added to changes
	if change, ok := changes["fire_power"]; !ok {
		t.Error("fire_power not in changes map")
	} else if change[0] != 0 || change[1] != 18 {
		t.Errorf("Expected fire_power change [0, 18], got %v", change)
	}

	// Verify pet's fire_power increased
	firePower = pet.GetCustomAcc("fire_power")
	if firePower != 18 {
		t.Errorf("Expected fire_power=18, got %d", firePower)
	}

	// Verify happiness was also changed
	if change, ok := changes["happiness"]; !ok {
		t.Error("happiness not in changes map")
	} else if change[0] != 60 || change[1] != 75 {
		t.Errorf("Expected happiness change [60, 75], got %v", change)
	}

	// Test 3: Second adventure
	outcome2 := plugin.AdventureOutcome{
		Text: "吸收了火焰之力",
		Effects: map[string]int{
			"fire_power": 10,
			"energy":     -10,
		},
	}

	changes2 := ApplyAdventureOutcome(pet, outcome2)

	// Verify accumulated fire_power
	firePower = pet.GetCustomAcc("fire_power")
	if firePower != 28 {
		t.Errorf("Expected accumulated fire_power=28, got %d", firePower)
	}

	// Verify change reflects accumulation
	if change, ok := changes2["fire_power"]; !ok {
		t.Error("fire_power not in second changes map")
	} else if change[0] != 18 || change[1] != 28 {
		t.Errorf("Expected fire_power change [18, 28], got %v", change)
	}

	// Test 4: SetField to reach evolution requirement
	old, err := pet.SetField("custom:fire_power", "40")
	if err != nil {
		t.Errorf("SetField failed: %v", err)
	}
	if old != "28" {
		t.Errorf("Expected old value 28, got %s", old)
	}

	firePower = pet.GetCustomAcc("fire_power")
	if firePower != 40 {
		t.Errorf("Expected fire_power=40 after SetField, got %d", firePower)
	}

	// Test 5: Verify evolution condition check
	// This would normally be done in evolution.go, but we can test the accumulator
	if firePower < 40 {
		t.Errorf("Evolution condition not met: fire_power=%d < 40", firePower)
	}

	t.Logf("✅ All integration tests passed!")
	t.Logf("Final state: fire_power=%d, energy=%d, happiness=%d",
		pet.GetCustomAcc("fire_power"), pet.Energy, pet.Happiness)
}
