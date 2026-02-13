// Package quests implements individual Lineage 2 quests using the quest framework.
package quests

import (
	"log/slog"
	"math/rand/v2"

	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/game/quest"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// Race IDs matching L2 Interlude protocol.
const (
	RaceHuman   int32 = 0
	RaceElf     int32 = 1
	RaceDarkElf int32 = 2
	RaceOrc     int32 = 3
	RaceDwarf   int32 = 4
)

// getItemCount returns the count of a quest item in the player's real inventory.
func getItemCount(player *model.Player, itemID int32) int64 {
	if player == nil {
		return 0
	}
	return player.Inventory().CountItemsByID(itemID)
}

// giveItem adds a real quest item to the player's inventory.
// For stackable items, increases count of existing stack. Otherwise creates a new item.
// Java reference: QuestState.giveItems() → Inventory.addItem()
func giveItem(player *model.Player, itemID int32, count int64) {
	if player == nil || count <= 0 {
		return
	}

	inv := player.Inventory()
	tmpl := db.ItemDefToTemplate(itemID)
	if tmpl == nil {
		slog.Error("quest giveItem: unknown item template", "itemID", itemID)
		return
	}

	// Stackable: merge into existing stack
	if tmpl.Stackable {
		existing := inv.FindItemByItemID(itemID)
		if existing != nil {
			if err := existing.SetCount(existing.Count() + int32(count)); err != nil {
				slog.Error("quest giveItem: failed to update count", "itemID", itemID, "error", err)
			}
			return
		}
	}

	// Create new item
	objectID := world.IDGenerator().NextItemID()
	item, err := model.NewItem(objectID, itemID, player.CharacterID(), int32(count), tmpl)
	if err != nil {
		slog.Error("quest giveItem: failed to create item", "itemID", itemID, "error", err)
		return
	}

	if err := inv.AddItem(item); err != nil {
		slog.Error("quest giveItem: failed to add to inventory", "itemID", itemID, "error", err)
	}
}

// takeItem removes quest items from the player's real inventory.
// count < 0 means remove all. Returns actual removed count.
// Java reference: QuestState.takeItems() → Player.destroyItemByItemId()
func takeItem(player *model.Player, itemID int32, count int64) int64 {
	if player == nil {
		return 0
	}
	inv := player.Inventory()

	if count < 0 {
		// Remove all
		total := inv.CountItemsByID(itemID)
		if total == 0 {
			return 0
		}
		return inv.RemoveItemsByID(itemID, total)
	}

	return inv.RemoveItemsByID(itemID, count)
}

// hasItem returns true if the player has at least one of the specified item in inventory.
func hasItem(player *model.Player, itemID int32) bool {
	return getItemCount(player, itemID) > 0
}

// getRandom returns a random int in [0, max).
func getRandom(max int) int {
	if max <= 0 {
		return 0
	}
	return rand.IntN(max)
}

// getEvent extracts the event string from Event params.
// Returns empty string if no event (initial NPC dialog).
func getEvent(e *quest.Event) string {
	if e.Params == nil {
		return ""
	}
	v, _ := e.Params["event"].(string)
	return v
}

// giveAdena adds adena to the player's inventory.
func giveAdena(e *quest.Event, amount int32) {
	if e.Player != nil {
		e.Player.Inventory().AddAdena(amount)
	}
}

// giveExp adds experience to the player.
func giveExp(e *quest.Event, exp int64) {
	if e.Player != nil {
		e.Player.AddExperience(exp)
	}
}
