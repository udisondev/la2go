package itemhandler

import (
	"log/slog"

	"github.com/udisondev/la2go/internal/model"
)

// bookHandler processes Book items that open an HTML dialog.
// Does not consume the item.
//
// Java reference: Book.java
type bookHandler struct{}

func (h *bookHandler) UseItem(player *model.Player, item *model.Item, _, _ int32) *UseResult {
	slog.Debug("Book: opened",
		"player", player.Name(),
		"itemID", item.ItemID())
	// Books open HTML help/{itemId}.htm — handled by caller via NPC dialog system.
	// Return empty result (no consumption, no skill).
	return &UseResult{}
}

// recipesHandler processes Recipe items.
// Registers the recipe in the player's known recipes and consumes the item.
// The actual craft logic is in game/craft package (Phase 15).
//
// Java reference: Recipes.java
type recipesHandler struct{}

func (h *recipesHandler) UseItem(player *model.Player, item *model.Item, _, _ int32) *UseResult {
	// Recipe registration is handled by craft system — just consume item.
	// Caller checks if recipe is already known.
	return &UseResult{
		ConsumeCount: 1,
	}
}

// rollingDiceHandler processes Rolling Dice items.
// Broadcasts a random 1-6 result. Does not consume the item.
//
// Java reference: RollingDice.java
type rollingDiceHandler struct{}

func (h *rollingDiceHandler) UseItem(player *model.Player, item *model.Item, _, _ int32) *UseResult {
	// Dice roll is broadcast by caller via CreatureSay.
	return &UseResult{}
}

// charmOfCourageHandler processes Charm of Courage items.
// Prevents XP loss on death. Consumes 1 item.
//
// Java reference: CharmOfCourage.java
type charmOfCourageHandler struct{}

func (h *charmOfCourageHandler) UseItem(player *model.Player, item *model.Item, skillID, skillLevel int32) *UseResult {
	if skillID == 0 {
		return nil
	}
	return &UseResult{
		ConsumeCount: 1,
		SkillID:      skillID,
		SkillLevel:   skillLevel,
		Broadcast:    true,
	}
}

// petFoodHandler processes Pet Food items.
// Stub: pets not yet implemented.
//
// Java reference: PetFood.java
type petFoodHandler struct{}

func (h *petFoodHandler) UseItem(_ *model.Player, _ *model.Item, _, _ int32) *UseResult {
	// Pets not implemented yet
	return nil
}
