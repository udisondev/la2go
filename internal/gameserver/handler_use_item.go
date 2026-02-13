package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/game/enchant"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleUseItem processes UseItem packet (opcode 0x19).
// Client sends this when player double-clicks an item in inventory.
// For equippable items (weapon/armor): toggles equip/unequip.
// For consumable items: uses the item (potions, scrolls, soulshots — requires item handler system).
//
// Phase 19: UseItem handler + equipment restrictions.
// Java reference: UseItem.java
func (h *Handler) handleUseItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseUseItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing UseItem: %w", err)
	}

	if client.State() != ClientStateInGame {
		return 0, true, nil
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Dead players cannot use items
	if player.IsDead() {
		return h.sendActionFailed(buf)
	}

	// Players in private store cannot use items
	if player.PrivateStoreType() != model.StoreNone {
		return h.sendActionFailed(buf)
	}

	// Find item in inventory
	inv := player.Inventory()
	item := inv.GetItem(uint32(pkt.ObjectID))
	if item == nil {
		slog.Warn("UseItem: item not found",
			"player", player.Name(),
			"objectID", pkt.ObjectID)
		return h.sendActionFailed(buf)
	}

	tmpl := item.Template()
	if tmpl == nil {
		slog.Error("UseItem: item has nil template",
			"player", player.Name(),
			"objectID", pkt.ObjectID,
			"itemID", item.ItemID())
		return h.sendActionFailed(buf)
	}

	// Quest items cannot be used via UseItem
	if tmpl.Type == model.ItemTypeQuestItem {
		return h.sendActionFailed(buf)
	}

	// Equippable item: toggle equip/unequip
	if tmpl.IsEquippable() {
		return h.handleEquipToggle(client, player, item, tmpl, buf)
	}

	// Enchant scroll: set as active enchant item
	if _, ok := enchant.IsScroll(item.ItemID()); ok {
		player.SetActiveEnchantItemID(int32(item.ObjectID()))
		slog.Debug("UseItem: enchant scroll activated",
			"player", player.Name(),
			"scrollID", item.ItemID(),
			"scrollObjID", item.ObjectID())
		// Клиент откроет окно заточки сам после получения ActionFailed
		return h.sendActionFailed(buf)
	}

	// Dispatch to item handler (Phase 51: Item Handler System)
	return h.handleItemUseDispatch(client, player, item, buf)
}

// handleEquipToggle handles equipping or unequipping an item.
// If the item is currently equipped → unequip it.
// If the item is not equipped → check restrictions and equip it.
func (h *Handler) handleEquipToggle(
	client *GameClient,
	player *model.Player,
	item *model.Item,
	tmpl *model.ItemTemplate,
	buf []byte,
) (int, bool, error) {
	inv := player.Inventory()

	if item.IsEquipped() {
		// Unequip
		return h.handleUnequip(client, player, item, inv, buf)
	}

	// Check grade restriction before equipping
	if err := checkGradeRestriction(player, tmpl); err != nil {
		slog.Debug("UseItem: grade restriction",
			"player", player.Name(),
			"itemName", tmpl.Name,
			"grade", tmpl.CrystalType,
			"error", err)
		// Send SystemMessage "Your grade level does not meet the requirement"
		sysMsg := serverpackets.NewSystemMessage(serverpackets.SysMsgIncompatibleItemGrade)
		sysMsgData, writeErr := sysMsg.Write()
		if writeErr != nil {
			slog.Error("failed to write grade restriction SystemMessage", "error", writeErr)
			return h.sendActionFailed(buf)
		}
		n := copy(buf, sysMsgData)
		return n, true, nil
	}

	// Equip item
	return h.handleEquip(client, player, item, tmpl, inv, buf)
}

// handleEquip equips an item to the appropriate paperdoll slot.
func (h *Handler) handleEquip(
	client *GameClient,
	player *model.Player,
	item *model.Item,
	tmpl *model.ItemTemplate,
	inv *model.Inventory,
	buf []byte,
) (int, bool, error) {
	bodyPart := tmpl.BodyPartStr
	if bodyPart == "" {
		slog.Warn("UseItem: item has no body part",
			"player", player.Name(),
			"itemID", item.ItemID())
		return h.sendActionFailed(buf)
	}

	// Resolve primary slot
	slot := model.BodyPartToPaperdollSlot(bodyPart)
	if slot < 0 {
		slog.Warn("UseItem: unknown body part",
			"player", player.Name(),
			"bodyPart", bodyPart,
			"itemID", item.ItemID())
		return h.sendActionFailed(buf)
	}

	// Handle paired slots (earrings, rings): fill empty slot first
	slot = resolveDoubleSlot(inv, bodyPart, slot)

	// Collect changed items for InventoryUpdate
	var changes []serverpackets.InvUpdateEntry

	// Unequip additional slot if needed (two-handed → remove shield, fullbody → remove legs)
	additionalSlot := model.BodyPartToAdditionalSlot(bodyPart)
	if additionalSlot >= 0 {
		if old := inv.UnequipItem(additionalSlot); old != nil {
			changes = append(changes, serverpackets.InvUpdateEntry{
				ChangeType: serverpackets.InvUpdateModify,
				Item:       old,
			})
		}
	}

	// Equip item (auto-unequips old item in slot)
	oldItem, equipErr := inv.EquipItem(item, slot)
	if equipErr != nil {
		slog.Error("UseItem: equip failed",
			"player", player.Name(),
			"itemID", item.ItemID(),
			"slot", slot,
			"error", equipErr)
		return h.sendActionFailed(buf)
	}

	if oldItem != nil {
		changes = append(changes, serverpackets.InvUpdateEntry{
			ChangeType: serverpackets.InvUpdateModify,
			Item:       oldItem,
		})
	}

	changes = append(changes, serverpackets.InvUpdateEntry{
		ChangeType: serverpackets.InvUpdateModify,
		Item:       item,
	})

	// Send InventoryUpdate as primary response
	invUpdate := serverpackets.NewInventoryUpdate(changes...)
	invUpdateData, err := invUpdate.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing InventoryUpdate: %w", err)
	}
	n := copy(buf, invUpdateData)

	// Send UserInfo for stat refresh (additional packet via Send)
	h.sendUserInfoToClient(client, player)

	// Broadcast CharInfo to visible players (equipment changed)
	h.broadcastCharInfoToVisible(client, player)

	slog.Debug("item equipped",
		"player", player.Name(),
		"item", tmpl.Name,
		"slot", slot,
		"bodyPart", bodyPart)

	return n, true, nil
}

// handleUnequip unequips an item from its paperdoll slot.
func (h *Handler) handleUnequip(
	client *GameClient,
	player *model.Player,
	item *model.Item,
	inv *model.Inventory,
	buf []byte,
) (int, bool, error) {
	slot := item.Slot()
	if slot < 0 || slot >= model.PaperdollTotalSlots {
		return h.sendActionFailed(buf)
	}

	unequipped := inv.UnequipItem(slot)
	if unequipped == nil {
		return h.sendActionFailed(buf)
	}

	changes := []serverpackets.InvUpdateEntry{
		{ChangeType: serverpackets.InvUpdateModify, Item: unequipped},
	}

	// Send InventoryUpdate
	invUpdate := serverpackets.NewInventoryUpdate(changes...)
	invUpdateData, err := invUpdate.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing InventoryUpdate: %w", err)
	}
	n := copy(buf, invUpdateData)

	// Send UserInfo + broadcast CharInfo
	h.sendUserInfoToClient(client, player)
	h.broadcastCharInfoToVisible(client, player)

	slog.Debug("item unequipped",
		"player", player.Name(),
		"item", item.Template().Name,
		"slot", slot)

	return n, true, nil
}

// sendUserInfoToClient sends a UserInfo packet to the client for stat refresh.
func (h *Handler) sendUserInfoToClient(client *GameClient, player *model.Player) {
	if h.clientManager == nil {
		return
	}

	userInfo := serverpackets.NewUserInfo(player)
	uiData, err := userInfo.Write()
	if err != nil {
		slog.Error("failed to write UserInfo after equip",
			"player", player.Name(),
			"error", err)
		return
	}

	if h.clientManager.writePool == nil {
		return
	}

	encPkt, err := h.clientManager.writePool.EncryptToPooled(client.Encryption(), uiData, len(uiData))
	if err != nil {
		slog.Error("failed to encrypt UserInfo",
			"player", player.Name(),
			"error", err)
		return
	}

	if err := client.Send(encPkt); err != nil {
		slog.Error("failed to send UserInfo",
			"player", player.Name(),
			"error", err)
	}
}

// broadcastCharInfoToVisible broadcasts CharInfo to all visible players.
func (h *Handler) broadcastCharInfoToVisible(client *GameClient, player *model.Player) {
	if h.clientManager == nil {
		return
	}

	charInfo := serverpackets.NewCharInfo(player)
	ciData, err := charInfo.Write()
	if err != nil {
		slog.Error("failed to write CharInfo for equip broadcast",
			"player", player.Name(),
			"error", err)
		return
	}

	h.clientManager.BroadcastToVisibleNear(player, ciData, len(ciData))
}

// checkGradeRestriction verifies that the player's level meets the item's grade requirement.
// Java reference: ItemTemplate.checkCondition() with PlayerCondition for crystal type.
//
// Grade level requirements (Java Config.EXPERTISE_PENALTY):
//
//	No grade (NONE): level 1+
//	D-grade: level 20+
//	C-grade: level 40+
//	B-grade: level 52+
//	A-grade: level 61+
//	S-grade: level 76+
func checkGradeRestriction(player *model.Player, tmpl *model.ItemTemplate) error {
	minLevel := gradeMinLevel(tmpl.CrystalType)
	if player.Level() < minLevel {
		return fmt.Errorf("player level %d < required %d for %s-grade",
			player.Level(), minLevel, tmpl.CrystalType)
	}
	return nil
}

// gradeMinLevel returns the minimum player level for using items of given grade.
// Java reference: Config.EXPERTISE_PENALTY, ExpertisePenalty in Java.
func gradeMinLevel(ct model.CrystalType) int32 {
	switch ct {
	case model.CrystalD:
		return 20
	case model.CrystalC:
		return 40
	case model.CrystalB:
		return 52
	case model.CrystalA:
		return 61
	case model.CrystalS:
		return 76
	default:
		return 1
	}
}

// resolveDoubleSlot resolves paired slots (earrings, rings).
// If a slot has two positions (left/right), uses the first empty one.
// If both occupied, replaces the right one.
//
// Java reference: Inventory.java equipItem() for SLOT_LR_EAR and SLOT_LR_FINGER.
func resolveDoubleSlot(inv *model.Inventory, bodyPart string, defaultSlot int32) int32 {
	switch bodyPart {
	case "rear", "lear":
		// Earring: try both slots
		rSlot, lSlot := model.EarringSlots()
		if inv.GetPaperdollItem(rSlot) == nil {
			return rSlot
		}
		if inv.GetPaperdollItem(lSlot) == nil {
			return lSlot
		}
		return rSlot // Both occupied → replace right

	case "rfinger", "lfinger":
		// Ring: try both slots
		rSlot, lSlot := model.RingSlots()
		if inv.GetPaperdollItem(rSlot) == nil {
			return rSlot
		}
		if inv.GetPaperdollItem(lSlot) == nil {
			return lSlot
		}
		return rSlot // Both occupied → replace right

	default:
		return defaultSlot
	}
}
