package commands

import (
	"fmt"
	"strconv"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// GiveItem handles //give_item <itemID> <count>.
//
// Java reference: AdminCreateItem.java
type GiveItem struct{}

func (c *GiveItem) Names() []string           { return []string{"give_item", "create_item", "itemcreate"} }
func (c *GiveItem) RequiredAccessLevel() int32 { return 2 }

func (c *GiveItem) Handle(player *model.Player, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: //give_item <itemID> <count>")
	}

	itemID, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid itemID %q: %w", args[1], err)
	}

	count, err := strconv.ParseInt(args[2], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid count %q: %w", args[2], err)
	}

	if count < 1 || count > 999999 {
		return fmt.Errorf("count must be between 1 and 999999, got %d", count)
	}

	// Validate item exists
	itemDef := data.GetItemDef(int32(itemID))
	if itemDef == nil {
		return fmt.Errorf("item template %d not found", itemID)
	}

	tmpl := db.ItemDefToTemplate(int32(itemID))
	if tmpl == nil {
		return fmt.Errorf("item template conversion failed for %d", itemID)
	}

	// Determine target player (target or self)
	var target *model.Player
	tgt := player.Target()
	if tgt != nil {
		if p, ok := tgt.Data.(*model.Player); ok {
			target = p
		}
	}
	if target == nil {
		target = player
	}

	// Check if stackable item already exists
	existing := target.Inventory().FindItemByItemID(int32(itemID))
	if existing != nil && !existing.IsEquipped() {
		existing.SetCount(existing.Count() + int32(count))
		player.SetLastAdminMessage(fmt.Sprintf("Added %d %s (ID: %d) to %s (merged stack, total: %d)",
			count, itemDef.Name(), itemID, target.Name(), existing.Count()))
		return nil
	}

	objectID := world.IDGenerator().NextItemID()
	item, err := model.NewItem(objectID, int32(itemID), target.CharacterID(), int32(count), tmpl)
	if err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	if err := target.Inventory().AddItem(item); err != nil {
		return fmt.Errorf("add to inventory: %w", err)
	}

	player.SetLastAdminMessage(fmt.Sprintf("Gave %d %s (ID: %d) to %s",
		count, itemDef.Name(), itemID, target.Name()))
	return nil
}
