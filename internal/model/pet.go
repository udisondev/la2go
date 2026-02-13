package model

import (
	"sync"
	"sync/atomic"
)

// Pet represents a summoned pet with inventory, feed system, and experience.
// Extends Summon with pet-specific functionality.
// Phase 19: Pets/Summons System.
type Pet struct {
	*Summon

	petMu sync.RWMutex

	// Control item
	controlItemID uint32 // objectID of collar item in player's inventory
	npcTemplateID int32  // NPC template ID for pet type lookup

	// Experience & leveling
	experience int64
	maxLevel   int32

	// Feed system
	currentFed int32
	maxFed     int32
	feedRate   float64 // feed consumption per tick (1s)

	// Inventory
	inventory *Inventory

	// Respawn flag
	isRespawned atomic.Bool
}

// NewPet creates a new pet from summon data with feed and exp systems.
func NewPet(
	summon *Summon,
	controlItemID uint32,
	npcTemplateID int32,
	exp int64,
	maxLevel int32,
	maxFed int32,
	feedRate float64,
) *Pet {
	p := &Pet{
		Summon:        summon,
		controlItemID: controlItemID,
		npcTemplateID: npcTemplateID,
		experience:    exp,
		maxLevel:      maxLevel,
		currentFed:    maxFed, // start full
		maxFed:        maxFed,
		feedRate:      feedRate,
		inventory:     NewInventory(int64(summon.ObjectID())),
	}

	// Override WorldObject.Data to point to Pet (for type assertion)
	summon.WorldObject.Data = p

	return p
}

// ControlItemID returns objectID of the control item (collar).
func (p *Pet) ControlItemID() uint32 {
	return p.controlItemID
}

// NpcTemplateID returns NPC template ID for pet type lookup.
func (p *Pet) NpcTemplateID() int32 {
	return p.npcTemplateID
}

// Experience returns current experience.
func (p *Pet) Experience() int64 {
	p.petMu.RLock()
	defer p.petMu.RUnlock()
	return p.experience
}

// SetExperience sets current experience.
func (p *Pet) SetExperience(exp int64) {
	p.petMu.Lock()
	defer p.petMu.Unlock()
	if exp < 0 {
		exp = 0
	}
	p.experience = exp
}

// AddExperience adds experience and returns true if level changed.
func (p *Pet) AddExperience(amount int64) bool {
	p.petMu.Lock()
	defer p.petMu.Unlock()

	p.experience += amount
	if p.experience < 0 {
		p.experience = 0
	}
	return false // level check handled externally via pet data tables
}

// MaxLevel returns maximum level for this pet type.
func (p *Pet) MaxLevel() int32 {
	return p.maxLevel
}

// CurrentFed returns current food level.
func (p *Pet) CurrentFed() int32 {
	p.petMu.RLock()
	defer p.petMu.RUnlock()
	return p.currentFed
}

// MaxFed returns maximum food level.
func (p *Pet) MaxFed() int32 {
	p.petMu.RLock()
	defer p.petMu.RUnlock()
	return p.maxFed
}

// SetCurrentFed sets current food level (clamped to 0..maxFed).
func (p *Pet) SetCurrentFed(fed int32) {
	p.petMu.Lock()
	defer p.petMu.Unlock()

	if fed < 0 {
		fed = 0
	}
	if fed > p.maxFed {
		fed = p.maxFed
	}
	p.currentFed = fed
}

// ConsumeFeed reduces food by feedRate. Returns true if pet is hungry (fed <= 0).
func (p *Pet) ConsumeFeed() bool {
	p.petMu.Lock()
	defer p.petMu.Unlock()

	p.currentFed -= int32(p.feedRate)
	if p.currentFed < 0 {
		p.currentFed = 0
	}
	return p.currentFed <= 0
}

// FeedRate returns feed consumption rate per tick.
func (p *Pet) FeedRate() float64 {
	p.petMu.RLock()
	defer p.petMu.RUnlock()
	return p.feedRate
}

// FedPercentage returns food percentage (0.0 - 1.0).
func (p *Pet) FedPercentage() float64 {
	p.petMu.RLock()
	defer p.petMu.RUnlock()

	if p.maxFed == 0 {
		return 0.0
	}
	return float64(p.currentFed) / float64(p.maxFed)
}

// Inventory returns pet's inventory.
func (p *Pet) Inventory() *Inventory {
	return p.inventory
}

// PetName returns pet's display name.
func (p *Pet) PetName() string {
	return p.Name()
}

// SetPetName sets pet's custom name.
func (p *Pet) SetPetName(name string) {
	p.SetName(name)
}

// UpdateFeedStats updates max feed and feed rate (on level change).
func (p *Pet) UpdateFeedStats(maxFed int32, feedRate float64) {
	p.petMu.Lock()
	defer p.petMu.Unlock()

	p.maxFed = maxFed
	p.feedRate = feedRate
	if p.currentFed > p.maxFed {
		p.currentFed = p.maxFed
	}
}

// IsRespawned returns whether pet was respawned (vs newly summoned).
func (p *Pet) IsRespawned() bool {
	return p.isRespawned.Load()
}

// SetRespawned sets respawned flag.
func (p *Pet) SetRespawned(respawned bool) {
	p.isRespawned.Store(respawned)
}
