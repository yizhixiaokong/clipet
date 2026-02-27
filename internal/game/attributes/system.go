package attributes

import (
	"fmt"
)

// Definition defines a custom or core attribute
type Definition struct {
	ID          string  `toml:"id"`
	DisplayName string  `toml:"display_name"`
	Min         int     `toml:"min"`
	Max         int     `toml:"max"`
	Default     int     `toml:"default"`
	DecayRate   float64 `toml:"decay_rate"` // Decay per hour
}

// System manages both core attributes (hunger, happiness, health, energy)
// and custom attributes defined by species plugins
type System struct {
	coreAttrs   map[string]Definition // Core 4 attributes
	customAttrs map[string]Definition // Custom attributes from plugins
}

// NewSystem creates a new attribute system
func NewSystem() *System {
	return &System{
		coreAttrs: map[string]Definition{
			"hunger": {
				ID:          "hunger",
				DisplayName: "饱食度",
				Min:         0,
				Max:         100,
				Default:     50,
				DecayRate:   3.0, // -3 per hour
			},
			"happiness": {
				ID:          "happiness",
				DisplayName: "快乐度",
				Min:         0,
				Max:         100,
				Default:     50,
				DecayRate:   2.0, // -2 per hour
			},
			"health": {
				ID:          "health",
				DisplayName: "健康度",
				Min:         0,
				Max:         100,
				Default:     100,
				DecayRate:   0.0, // No natural decay
			},
			"energy": {
				ID:          "energy",
				DisplayName: "精力",
				Min:         0,
				Max:         100,
				Default:     80,
				DecayRate:   1.0, // -1 per hour
			},
		},
		customAttrs: make(map[string]Definition),
	}
}

// RegisterCustomAttribute registers a custom attribute from a species plugin
func (s *System) RegisterCustomAttribute(def Definition) error {
	if _, exists := s.customAttrs[def.ID]; exists {
		return fmt.Errorf("custom attribute %q already registered", def.ID)
	}
	if _, exists := s.coreAttrs[def.ID]; exists {
		return fmt.Errorf("cannot override core attribute %q", def.ID)
	}

	s.customAttrs[def.ID] = def
	return nil
}

// GetDefinition returns the attribute definition
func (s *System) GetDefinition(id string) (Definition, bool) {
	// Check core attributes first
	if def, ok := s.coreAttrs[id]; ok {
		return def, true
	}
	// Then check custom attributes
	def, ok := s.customAttrs[id]
	return def, ok
}

// GetDecayRate returns the decay rate for an attribute
func (s *System) GetDecayRate(id string) float64 {
	def, ok := s.GetDefinition(id)
	if !ok {
		return 0.0
	}
	return def.DecayRate
}

// IsCoreAttribute checks if an attribute is a core attribute
func (s *System) IsCoreAttribute(id string) bool {
	_, ok := s.coreAttrs[id]
	return ok
}

// ValidateValue ensures a value is within the attribute's valid range
func (s *System) ValidateValue(id string, value int) (int, error) {
	def, ok := s.GetDefinition(id)
	if !ok {
		return 0, fmt.Errorf("unknown attribute %q", id)
	}

	if value < def.Min {
		return def.Min, nil
	}
	if value > def.Max {
		return def.Max, nil
	}
	return value, nil
}

// Clamp ensures a value is within the attribute's valid range
func (s *System) Clamp(id string, value int) int {
	clamped, _ := s.ValidateValue(id, value)
	return clamped
}

// GetAllAttributes returns all attribute IDs (core + custom)
func (s *System) GetAllAttributes() []string {
	attrs := make([]string, 0, len(s.coreAttrs)+len(s.customAttrs))
	for id := range s.coreAttrs {
		attrs = append(attrs, id)
	}
	for id := range s.customAttrs {
		attrs = append(attrs, id)
	}
	return attrs
}

// GetCustomAttributes returns only custom attribute IDs
func (s *System) GetCustomAttributes() []string {
	attrs := make([]string, 0, len(s.customAttrs))
	for id := range s.customAttrs {
		attrs = append(attrs, id)
	}
	return attrs
}
