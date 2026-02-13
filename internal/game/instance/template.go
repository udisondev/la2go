package instance

import "time"

// Template defines the configuration for a type of instance zone.
// Loaded from data/instances/*.xml in Java; here we use in-code definitions.
//
// Phase 26: Instance Zones.
type Template struct {
	ID         int32         // Unique template ID (matches L2J instance IDs)
	Name       string        // Human-readable name
	Duration   time.Duration // Max lifetime (0 = no limit)
	MaxPlayers int32         // Max concurrent players (0 = unlimited)
	MinLevel   int32         // Minimum player level to enter (0 = no restriction)
	MaxLevel   int32         // Maximum player level (0 = no restriction)
	Cooldown   time.Duration // Reentry cooldown per character
	RemoveBuffs bool         // Remove buffs on entry

	// Spawn location inside the instance.
	SpawnX int32
	SpawnY int32
	SpawnZ int32

	// Exit location (where players go when leaving).
	ExitX int32
	ExitY int32
	ExitZ int32
}

// Validate checks that template fields are sensible.
func (t *Template) Validate() error {
	if t.ID <= 0 {
		return ErrInvalidTemplateID
	}
	if t.Name == "" {
		return ErrEmptyTemplateName
	}
	if t.MaxPlayers < 0 {
		return ErrInvalidMaxPlayers
	}
	if t.MinLevel < 0 || t.MinLevel > 80 {
		return ErrInvalidLevel
	}
	if t.MaxLevel < 0 || t.MaxLevel > 80 {
		return ErrInvalidLevel
	}
	if t.MinLevel > 0 && t.MaxLevel > 0 && t.MinLevel > t.MaxLevel {
		return ErrInvalidLevel
	}
	return nil
}
