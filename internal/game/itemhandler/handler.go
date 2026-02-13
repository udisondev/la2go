// Package itemhandler implements the item use handler system.
// Each handler name (from itemDef.handler) maps to an ItemHandler implementation.
//
// Phase 51: Item Handler System.
// Java reference: IItemHandler.java, ItemHandler.java
package itemhandler

import "github.com/udisondev/la2go/internal/model"

// UseResult describes the outcome of an item use.
type UseResult struct {
	ConsumeCount int64 // number of items to consume (0 = don't consume)
	SkillID      int32 // skill to cast (0 = no skill)
	SkillLevel   int32
	ReuseDelay   int32 // ms, item reuse cooldown (0 = no cooldown)
	Message      int32 // system message ID to send (0 = none)
	Broadcast    bool  // true = broadcast MagicSkillUse to visible players
}

// ItemHandler processes UseItem for a specific handler type.
type ItemHandler interface {
	// UseItem processes item usage and returns the result.
	// Returns nil UseResult if the item cannot be used.
	UseItem(player *model.Player, item *model.Item, skillID, skillLevel int32) *UseResult
}

// registry maps handler name → ItemHandler implementation.
var registry = map[string]ItemHandler{}

// Register adds a handler to the registry.
func Register(name string, h ItemHandler) {
	registry[name] = h
}

// Get returns the handler for the given name, or nil if not registered.
func Get(name string) ItemHandler {
	return registry[name]
}

// Init registers all built-in item handlers.
func Init() {
	Register("ItemSkills", &itemSkillsHandler{})
	Register("Elixir", &elixirHandler{})
	Register("SoulShots", &soulShotsHandler{})
	Register("SpiritShot", &spiritShotHandler{})
	Register("BlessedSpiritShot", &blessedSpiritShotHandler{})
	Register("FishShots", &fishShotsHandler{})
	Register("Book", &bookHandler{})
	Register("Recipes", &recipesHandler{})
	Register("RollingDice", &rollingDiceHandler{})
	Register("CharmOfCourage", &charmOfCourageHandler{})
	Register("PetFood", &petFoodHandler{})
}
