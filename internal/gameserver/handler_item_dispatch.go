package gameserver

import (
	"log/slog"
	"time"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/itemhandler"
	"github.com/udisondev/la2go/internal/game/skill"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleItemUseDispatch dispatches item use to the appropriate item handler.
// Looks up the handler name from itemDef and delegates to the handler registry.
//
// Phase 51: Item Handler System.
// Java reference: UseItem.java → ItemHandler dispatch
func (h *Handler) handleItemUseDispatch(
	client *GameClient,
	player *model.Player,
	item *model.Item,
	buf []byte,
) (int, bool, error) {
	def := data.GetItemDef(item.ItemID())
	if def == nil {
		slog.Debug("UseItem: no item def",
			"player", player.Name(),
			"itemID", item.ItemID())
		return h.sendActionFailed(buf)
	}

	handlerName := def.Handler()
	if handlerName == "" {
		slog.Debug("UseItem: no handler for item",
			"player", player.Name(),
			"itemID", item.ItemID(),
			"itemName", item.Name())
		return h.sendActionFailed(buf)
	}

	handler := itemhandler.Get(handlerName)
	if handler == nil {
		slog.Debug("UseItem: handler not registered",
			"player", player.Name(),
			"handler", handlerName,
			"itemID", item.ItemID())
		return h.sendActionFailed(buf)
	}

	// Check Olympiad restriction
	if def.IsOlyRestricted() && player.IsInOlympiad() {
		slog.Debug("UseItem: olympiad restricted",
			"player", player.Name(),
			"itemID", item.ItemID())
		return h.sendActionFailed(buf)
	}

	// Check item reuse delay
	if def.ReuseDelay() > 0 {
		if player.IsItemOnCooldown(item.ItemID()) {
			slog.Debug("UseItem: item on cooldown",
				"player", player.Name(),
				"itemID", item.ItemID())
			return h.sendActionFailed(buf)
		}
	}

	// Call handler
	result := handler.UseItem(player, item, def.ItemSkillID(), def.ItemSkillLevel())
	if result == nil {
		return h.sendActionFailed(buf)
	}

	// Consume items
	if result.ConsumeCount > 0 {
		inv := player.Inventory()
		removed := inv.RemoveItemsByID(item.ItemID(), result.ConsumeCount)
		if removed < result.ConsumeCount {
			slog.Warn("UseItem: could not consume enough items",
				"player", player.Name(),
				"itemID", item.ItemID(),
				"needed", result.ConsumeCount,
				"removed", removed)
			return h.sendActionFailed(buf)
		}

		// Send InventoryUpdate for consumed items
		h.sendItemConsumedUpdate(client, item, removed)
	}

	// Cast skill if specified
	if result.SkillID > 0 && result.Broadcast {
		if skill.CastMgr != nil {
			if err := skill.CastMgr.UseItemSkill(player, result.SkillID, result.SkillLevel); err != nil {
				slog.Debug("UseItem: item skill cast failed",
					"player", player.Name(),
					"skillID", result.SkillID,
					"error", err)
			}
		}
	}

	// Set item reuse delay
	reuseDelay := result.ReuseDelay
	if reuseDelay == 0 {
		reuseDelay = def.ReuseDelay()
	}
	if reuseDelay > 0 {
		player.SetItemCooldown(item.ItemID(), time.Duration(reuseDelay)*time.Millisecond)
	}

	slog.Debug("UseItem: handler processed",
		"player", player.Name(),
		"handler", handlerName,
		"itemID", item.ItemID(),
		"consumed", result.ConsumeCount,
		"skillID", result.SkillID)

	return 0, true, nil
}

// sendItemConsumedUpdate sends an InventoryUpdate for consumed (removed/decremented) items.
func (h *Handler) sendItemConsumedUpdate(client *GameClient, item *model.Item, _ int64) {
	if h.clientManager == nil {
		return
	}

	var changeType int16
	if item.Count() <= 0 {
		changeType = serverpackets.InvUpdateRemove
	} else {
		changeType = serverpackets.InvUpdateModify
	}

	invUpdate := serverpackets.NewInventoryUpdate(serverpackets.InvUpdateEntry{
		ChangeType: changeType,
		Item:       item,
	})
	invUpdateData, err := invUpdate.Write()
	if err != nil {
		slog.Error("failed to write InventoryUpdate for item consume", "error", err)
		return
	}

	player := client.ActivePlayer()
	if player == nil {
		return
	}
	if err := h.clientManager.SendToPlayer(player.ObjectID(), invUpdateData, len(invUpdateData)); err != nil {
		slog.Warn("failed to send InventoryUpdate",
			"player", player.Name(),
			"error", err)
	}
}
