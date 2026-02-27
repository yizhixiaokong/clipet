package game

import (
	"testing"
)

// TestPhase1_SetFieldCustomAttributes tests Phase 1: Pet.SetField support for custom attributes
func TestPhase1_SetFieldCustomAttributes(t *testing.T) {
	pet := &Pet{
		Name:             "TestPet",
		CustomAttributes: make(map[string]int),
	}

	// Test setting a custom attribute
	old, err := pet.SetField("custom:fire_points", "25")
	if err != nil {
		t.Errorf("SetField failed: %v", err)
	}
	if old != "0" {
		t.Errorf("Expected old value 0, got %s", old)
	}
	if pet.GetAttr("fire_points") != 25 {
		t.Errorf("Expected fire_points=25, got %d", pet.GetAttr("fire_points"))
	}

	// Test updating a custom attribute
	old, err = pet.SetField("custom:fire_points", "50")
	if err != nil {
		t.Errorf("SetField update failed: %v", err)
	}
	if old != "25" {
		t.Errorf("Expected old value 25, got %s", old)
	}
	if pet.GetAttr("fire_points") != 50 {
		t.Errorf("Expected fire_points=50, got %d", pet.GetAttr("fire_points"))
	}
}

// TestPhase2_CustomAccMethods tests Phase 2: Custom accumulator methods
func TestPhase2_CustomAccMethods(t *testing.T) {
	pet := &Pet{
		Name:             "TestPet",
		CustomAttributes: make(map[string]int),
	}

	// Test GetCustomAcc (initially 0)
	val := pet.GetCustomAcc("fire_points")
	if val != 0 {
		t.Errorf("Expected 0, got %d", val)
	}

	// Test AddCustomAcc
	pet.AddCustomAcc("fire_points", 10)
	if pet.GetCustomAcc("fire_points") != 10 {
		t.Errorf("Expected 10, got %d", pet.GetCustomAcc("fire_points"))
	}

	// Test AddCustomAcc adds to existing value
	pet.AddCustomAcc("fire_points", 5)
	if pet.GetCustomAcc("fire_points") != 15 {
		t.Errorf("Expected 15, got %d", pet.GetCustomAcc("fire_points"))
	}

	// Test negative values
	pet.AddCustomAcc("fire_points", -3)
	if pet.GetCustomAcc("fire_points") != 12 {
		t.Errorf("Expected 12, got %d", pet.GetCustomAcc("fire_points"))
	}
}

// TestPhase3_CustomAttributesMap tests Phase 3: CustomAttributes map storage
func TestPhase3_CustomAttributesMap(t *testing.T) {
	pet := &Pet{
		Name:             "TestPet",
		CustomAttributes: make(map[string]int),
	}

	// Test empty map
	if len(pet.CustomAttributes) != 0 {
		t.Errorf("Expected empty map, got %d items", len(pet.CustomAttributes))
	}

	// Test adding items
	pet.SetAttr("fire_points", 30)
	pet.SetAttr("frost_points", 20)
	if len(pet.CustomAttributes) != 2 {
		t.Errorf("Expected 2 items, got %d", len(pet.CustomAttributes))
	}

	// Test retrieval
	if pet.GetAttr("fire_points") != 30 {
		t.Errorf("Expected fire_points=30, got %d", pet.GetAttr("fire_points"))
	}
	if pet.GetAttr("frost_points") != 20 {
		t.Errorf("Expected frost_points=20, got %d", pet.GetAttr("frost_points"))
	}
}
