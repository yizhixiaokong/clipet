package game

import (
	"clipet/internal/game/capabilities"
	"clipet/internal/plugin"
	"testing"
	"time"
)

func TestMultiStageDecay_HungerZeroMidway(t *testing.T) {
	// Create test registry with default configs
	reg := plugin.NewRegistry()

	// Create test species pack with default configs
	pack := &plugin.SpeciesPack{
		Species: plugin.SpeciesConfig{
			ID:        "test",
			Name:      "Test Species",
			BaseStats: plugin.BaseStats{Hunger: 50, Happiness: 50, Health: 50, Energy: 50},
		},
		Decay: capabilities.DecayConfig{
			Hunger:    1.0,
			Happiness: 0.5,
			Energy:    0.3,
			Health:    0.2,
		},
		Interactions: capabilities.AttributeInteractionConfig{
			HungerHealthThreshold:         20,
			HungerHealthRate:              0.2,
			HungerZeroHealthMultiplier:    3.0,
			EnergyLowThreshold:            20,
			EnergyCritThreshold:           10,
			EnergyLowHealthRate:           0.1,
			EnergyCritHealthMultiplier:    2.0,
			EnergyCritHappinessMultiplier: 1.5,
			HappinessLowThreshold:         20,
			HappinessZeroHealthMultiplier: 4.0,
			HealthCritThreshold:           20,
			HealthRecoveryPenalty:         0.5,
		},
	}
	reg.Register(pack)

	// Scenario: 18 hours offline, hunger drops to 0 midway
	pet := &Pet{
		Hunger:    18, // Will drop to 0 after 18 hours
		Happiness: 50,
		Health:    80,
		Energy:    50,
		Alive:     true,
		registry:  reg,
	}

	results := pet.ApplyMultiStageDecay(18 * time.Hour)

	// Verify:
	// - Round 1 (0-6h): Hunger 18→12, no health decay
	// - Round 2 (6-12h): Hunger 12→6, no health decay
	// - Round 3 (12-18h): Hunger 6→0, health starts decaying (hunger=0 bonus)

	if len(results) != 3 {
		t.Errorf("Expected 3 rounds, got %d", len(results))
	}

	// Verify final hunger is 0
	if pet.Hunger != 0 {
		t.Errorf("Expected hunger=0, got %d", pet.Hunger)
	}

	// Verify round 3 has critical state
	if !results[2].CriticalState {
		t.Error("Expected round 3 to have critical state (hunger=0)")
	}

	// Verify health decayed due to hunger=0 bonus
	if pet.Health >= 80 {
		t.Errorf("Expected health to decay below 80, got %d", pet.Health)
	}

	// Verify effects contain hunger death warning
	foundHungerDeath := false
	for _, effect := range results[2].Effects {
		if effect == "⚠️ 饥饿致死：健康大量衰减" {
			foundHungerDeath = true
			break
		}
	}
	if !foundHungerDeath {
		t.Error("Expected 'hunger death' effect in round 3")
	}
}

func TestMultiStageDecay_EnergyCritical(t *testing.T) {
	// Create test registry
	reg := plugin.NewRegistry()

	pack := &plugin.SpeciesPack{
		Species: plugin.SpeciesConfig{
			ID:        "test",
			Name:      "Test Species",
			BaseStats: plugin.BaseStats{Hunger: 50, Happiness: 50, Health: 50, Energy: 50},
		},
		Decay: capabilities.DecayConfig{
			Hunger:    1.0,
			Happiness: 0.5,
			Energy:    0.3,
			Health:    0.2,
		},
		Interactions: capabilities.AttributeInteractionConfig{
			HungerHealthThreshold:         20,
			HungerHealthRate:              0.2,
			HungerZeroHealthMultiplier:    3.0,
			EnergyLowThreshold:            20,
			EnergyCritThreshold:           10,
			EnergyLowHealthRate:           0.1,
			EnergyCritHealthMultiplier:    2.0,
			EnergyCritHappinessMultiplier: 1.5,
			HappinessLowThreshold:         20,
			HappinessZeroHealthMultiplier: 4.0,
			HealthCritThreshold:           20,
			HealthRecoveryPenalty:         0.5,
		},
	}
	reg.Register(pack)

	// Scenario: Energy < 10 triggers critical state
	pet := &Pet{
		Hunger:    50,
		Happiness: 50,
		Health:    80,
		Energy:    8, // Critical state
		Alive:     true,
		registry:  reg,
	}

	results := pet.ApplyMultiStageDecay(6 * time.Hour)

	// Verify energy critical triggers health and happiness decay
	if len(results) != 1 {
		t.Fatalf("Expected 1 round, got %d", len(results))
	}

	if !results[0].CriticalState {
		t.Error("Expected critical state due to energy < 10")
	}

	// Verify effects contain energy exhaustion warning
	foundEnergyExhaustion := false
	for _, effect := range results[0].Effects {
		if effect == "🚨 精力耗尽：健康和快乐加速衰减" {
			foundEnergyExhaustion = true
			break
		}
	}
	if !foundEnergyExhaustion {
		t.Error("Expected 'energy exhaustion' effect")
	}

	// Verify health and happiness decayed
	if pet.Health >= 80 {
		t.Errorf("Expected health to decay below 80, got %d", pet.Health)
	}
	if pet.Happiness >= 50 {
		t.Errorf("Expected happiness to decay below 50, got %d", pet.Happiness)
	}
}

func TestMultiStageDecay_HappinessZero(t *testing.T) {
	// Create test registry
	reg := plugin.NewRegistry()

	pack := &plugin.SpeciesPack{
		Species: plugin.SpeciesConfig{
			ID:        "test",
			Name:      "Test Species",
			BaseStats: plugin.BaseStats{Hunger: 50, Happiness: 50, Health: 50, Energy: 50},
		},
		Decay: capabilities.DecayConfig{
			Hunger:    1.0,
			Happiness: 0.5,
			Energy:    0.3,
			Health:    0.2,
		},
		Interactions: capabilities.AttributeInteractionConfig{
			HungerHealthThreshold:         20,
			HungerHealthRate:              0.2,
			HungerZeroHealthMultiplier:    3.0,
			EnergyLowThreshold:            20,
			EnergyCritThreshold:           10,
			EnergyLowHealthRate:           0.1,
			EnergyCritHealthMultiplier:    2.0,
			EnergyCritHappinessMultiplier: 1.5,
			HappinessLowThreshold:         20,
			HappinessZeroHealthMultiplier: 4.0,
			HealthCritThreshold:           20,
			HealthRecoveryPenalty:         0.5,
		},
	}
	reg.Register(pack)

	// Scenario: Happiness=0 causes massive health decay (with death protection)
	// Use very low health (3) so decay would bring it to <= 0, triggering death protection
	pet := &Pet{
		Hunger:    50,
		Happiness: 0, // Depression
		Health:    3, // Very low health, decay would bring it to <= 0
		Energy:    50,
		Alive:     true,
		registry:  reg,
	}

	results := pet.ApplyMultiStageDecay(6 * time.Hour)

	// Verify health decays massively but death protection keeps it at 1
	if len(results) != 1 {
		t.Fatalf("Expected 1 round, got %d", len(results))
	}

	if !results[0].CriticalState {
		t.Error("Expected critical state due to happiness=0")
	}

	// Verify death protection: health should be 1
	if pet.Health != 1 {
		t.Errorf("Expected health=1 (death protection), got %d", pet.Health)
	}

	// Verify effects contain depression and death protection warnings
	foundDepression := false
	foundDeathProtection := false
	for _, effect := range results[0].Effects {
		if effect == "💔 抑郁症：健康大量衰减" {
			foundDepression = true
		}
		if effect == "🛡️ 濒死保护：健康保持1点" {
			foundDeathProtection = true
		}
	}
	if !foundDepression {
		t.Error("Expected 'depression' effect")
	}
	if !foundDeathProtection {
		t.Error("Expected 'death protection' effect")
	}
}

func TestMultiStageDecay_NoDecayWhenAttributesHigh(t *testing.T) {
	// Create test registry
	reg := plugin.NewRegistry()

	pack := &plugin.SpeciesPack{
		Species: plugin.SpeciesConfig{
			ID:        "test",
			Name:      "Test Species",
			BaseStats: plugin.BaseStats{Hunger: 50, Happiness: 50, Health: 50, Energy: 50},
		},
		Decay: capabilities.DecayConfig{
			Hunger:    1.0,
			Happiness: 0.5,
			Energy:    0.3,
			Health:    0.2,
		},
		Interactions: capabilities.AttributeInteractionConfig{
			HungerHealthThreshold:         20,
			HungerHealthRate:              0.2,
			HungerZeroHealthMultiplier:    3.0,
			EnergyLowThreshold:            20,
			EnergyCritThreshold:           10,
			EnergyLowHealthRate:           0.1,
			EnergyCritHealthMultiplier:    2.0,
			EnergyCritHappinessMultiplier: 1.5,
			HappinessLowThreshold:         20,
			HappinessZeroHealthMultiplier: 4.0,
			HealthCritThreshold:           20,
			HealthRecoveryPenalty:         0.5,
		},
	}
	reg.Register(pack)

	// Scenario: All attributes high, only base decay should apply
	pet := &Pet{
		Hunger:    80,
		Happiness: 80,
		Health:    80,
		Energy:    80,
		Alive:     true,
		registry:  reg,
	}

	results := pet.ApplyMultiStageDecay(6 * time.Hour)

	// Verify no critical state triggered
	if len(results) != 1 {
		t.Fatalf("Expected 1 round, got %d", len(results))
	}

	if results[0].CriticalState {
		t.Error("Expected NO critical state when attributes are high")
	}

	// Verify no special effects (only base decay)
	if len(results[0].Effects) != 0 {
		t.Errorf("Expected no special effects, got %v", results[0].Effects)
	}

	// Verify base decay was applied
	// Base decay: hunger -6, happiness -3, energy -1.8
	if pet.Hunger >= 80 {
		t.Errorf("Expected hunger to decay, got %d", pet.Hunger)
	}
	if pet.Happiness >= 80 {
		t.Errorf("Expected happiness to decay, got %d", pet.Happiness)
	}
	if pet.Energy >= 80 {
		t.Errorf("Expected energy to decay, got %d", pet.Energy)
	}
	// Health should NOT decay when hunger >= 20
	if pet.Health != 80 {
		t.Errorf("Expected health to remain 80 (hunger >= 20), got %d", pet.Health)
	}
}
