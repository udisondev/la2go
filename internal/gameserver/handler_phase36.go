package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// Crystal item IDs by grade.
var crystalItemIDs = map[model.CrystalType]int32{
	model.CrystalD: 1458,
	model.CrystalC: 1459,
	model.CrystalB: 1460,
	model.CrystalA: 1461,
	model.CrystalS: 1462,
}

// handleRequestAutoSoulShot processes 0xD0:0x05 — toggle auto soulshot.
func (h *Handler) handleRequestAutoSoulShot(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAutoSoulShot(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestAutoSoulShot: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for auto soulshot")
	}

	if pkt.Type == 1 {
		// Validate item exists in inventory
		inv := player.Inventory()
		found := false
		for _, item := range inv.GetItems() {
			if item.ItemID() == pkt.ItemID {
				found = true
				break
			}
		}
		if !found {
			slog.Warn("auto soulshot: item not in inventory",
				"player", player.Name(),
				"itemID", pkt.ItemID)
			return 0, true, nil
		}
		player.AddAutoSoulShot(pkt.ItemID)
	} else {
		player.RemoveAutoSoulShot(pkt.ItemID)
	}

	// Send confirmation to client
	resp := &serverpackets.ExAutoSoulShot{
		ItemID: pkt.ItemID,
		Type:   pkt.Type,
	}
	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExAutoSoulShot: %w", err)
	}
	n := copy(buf, respData)

	slog.Debug("auto soulshot toggled",
		"player", player.Name(),
		"itemID", pkt.ItemID,
		"enabled", pkt.Type == 1)

	return n, true, nil
}

// handleRequestMakeMacro processes 0xC1 — create/update macro.
func (h *Handler) handleRequestMakeMacro(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestMakeMacro(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestMakeMacro: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for make macro")
	}

	macro := pkt.Macro

	// Validate command count
	if len(macro.Commands) > model.MaxMacroCommands {
		slog.Warn("macro: too many commands",
			"player", player.Name(),
			"count", len(macro.Commands))
		return 0, true, nil
	}

	// Validate total command string length
	totalLen := 0
	for _, cmd := range macro.Commands {
		totalLen += len(cmd.Command)
	}
	if totalLen > model.MaxMacroCmdLen {
		slog.Warn("macro: commands too long",
			"player", player.Name(),
			"totalLen", totalLen)
		return 0, true, nil
	}

	if err := player.RegisterMacro(macro); err != nil {
		slog.Warn("macro: register failed",
			"player", player.Name(),
			"error", err)
		return 0, true, nil
	}

	// Send macro list with the new macro
	resp := &serverpackets.SendMacroList{
		Revision: player.MacroRevision(),
		Count:    int8(len(player.GetMacros())),
		Macro:    macro,
	}
	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SendMacroList: %w", err)
	}
	n := copy(buf, respData)

	slog.Debug("macro created",
		"player", player.Name(),
		"macroID", macro.ID,
		"name", macro.Name)

	return n, true, nil
}

// handleRequestDeleteMacro processes 0xC2 — delete macro.
func (h *Handler) handleRequestDeleteMacro(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestDeleteMacro(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestDeleteMacro: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for delete macro")
	}

	player.DeleteMacro(pkt.MacroID)

	// Send updated macro list (nil macro = delete notification)
	resp := &serverpackets.SendMacroList{
		Revision: player.MacroRevision(),
		Count:    int8(len(player.GetMacros())),
		Macro:    nil,
	}
	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing SendMacroList: %w", err)
	}
	n := copy(buf, respData)

	slog.Debug("macro deleted",
		"player", player.Name(),
		"macroID", pkt.MacroID)

	return n, true, nil
}

// handleRequestSendFriendMsg processes 0xCC — send PM to friend.
func (h *Handler) handleRequestSendFriendMsg(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSendFriendMsg(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestSendFriendMsg: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for friend msg")
	}

	// Validate message length
	if len(pkt.Message) == 0 || len(pkt.Message) > 300 {
		slog.Warn("friend msg: invalid message length",
			"player", player.Name(),
			"len", len(pkt.Message))
		return 0, true, nil
	}

	// Find target client
	targetClient := h.clientManager.GetClientByName(pkt.Receiver)
	if targetClient == nil || targetClient.ActivePlayer() == nil {
		sm := serverpackets.NewSystemMessage(sysMsgTargetPlayerNotFound)
		smData, _ := sm.Write()
		client.Send(smData)
		return 0, true, nil
	}

	target := targetClient.ActivePlayer()

	// Цель заблокировала отправителя
	if target.IsBlocked(int32(player.ObjectID())) {
		sm := serverpackets.NewSystemMessage(sysMsgTargetInRefusalMode)
		smData, _ := sm.Write()
		client.Send(smData)
		return 0, true, nil
	}

	// Цель в режиме отказа от сообщений
	if target.MessageRefusal() {
		sm := serverpackets.NewSystemMessage(sysMsgTargetInRefusalMode)
		smData, _ := sm.Write()
		client.Send(smData)
		return 0, true, nil
	}

	// Check sender is in target's friend list
	if !target.IsFriend(int32(player.ObjectID())) {
		slog.Warn("friend msg: sender not in receiver's friend list",
			"sender", player.Name(),
			"receiver", target.Name())
		return 0, true, nil
	}

	// Send L2FriendSay to target
	friendSay := &serverpackets.L2FriendSay{
		Sender:   player.Name(),
		Receiver: pkt.Receiver,
		Message:  pkt.Message,
	}
	friendSayData, err := friendSay.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing L2FriendSay: %w", err)
	}
	targetClient.Send(friendSayData)

	slog.Debug("friend msg sent",
		"from", player.Name(),
		"to", pkt.Receiver)

	return 0, true, nil
}

// handleRequestCrystallizeItem processes 0x72 — break item into crystals.
func (h *Handler) handleRequestCrystallizeItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestCrystallizeItem(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestCrystallizeItem: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for crystallize")
	}

	inv := player.Inventory()
	if inv == nil {
		return 0, false, fmt.Errorf("no inventory for crystallize")
	}

	// Find item
	item := inv.GetItem(uint32(pkt.ObjectID))
	if item == nil {
		slog.Warn("crystallize: item not found",
			"player", player.Name(),
			"objectID", pkt.ObjectID)
		return 0, true, nil
	}

	// Cannot crystallize equipped items
	if item.IsEquipped() {
		slog.Warn("crystallize: item is equipped",
			"player", player.Name(),
			"objectID", pkt.ObjectID)
		return 0, true, nil
	}

	tmpl := item.Template()
	if tmpl == nil {
		return 0, true, nil
	}

	// Cannot crystallize quest items
	if tmpl.Type == model.ItemTypeQuestItem {
		slog.Warn("crystallize: quest item",
			"player", player.Name(),
			"itemID", item.ItemID())
		return 0, true, nil
	}

	// Must have a crystal grade
	if tmpl.CrystalType == model.CrystalNone {
		slog.Warn("crystallize: no crystal type",
			"player", player.Name(),
			"itemID", item.ItemID())
		return 0, true, nil
	}

	crystalItemID, ok := crystalItemIDs[tmpl.CrystalType]
	if !ok {
		slog.Warn("crystallize: unknown crystal type",
			"player", player.Name(),
			"crystalType", tmpl.CrystalType)
		return 0, true, nil
	}

	// Calculate crystal count: base formula
	// Weapons get 3× enchant bonus, armor gets 1×
	var crystalCount int32 = 1 // minimum 1 crystal
	enchant := item.Enchant()
	if tmpl.IsWeapon() {
		crystalCount += 3 * enchant
	} else {
		crystalCount += enchant
	}

	// Remove item from inventory
	inv.RemoveItem(uint32(pkt.ObjectID))

	// Add crystals to inventory
	crystalTmpl := &model.ItemTemplate{
		ItemID:    crystalItemID,
		Name:      fmt.Sprintf("Crystal (%s)", tmpl.CrystalType),
		Type:      model.ItemTypeConsumable,
		Stackable: true,
		Tradeable: true,
	}
	crystalItem, err := model.NewItem(0, crystalItemID, player.CharacterID(), crystalCount, crystalTmpl)
	if err != nil {
		slog.Error("crystallize: create crystal item failed",
			"error", err)
		return 0, true, nil
	}
	if err := inv.AddItem(crystalItem); err != nil {
		slog.Error("crystallize: add crystal to inventory failed",
			"error", err)
		return 0, true, nil
	}

	// Send InventoryUpdate
	invPkt := serverpackets.NewInventoryItemList(inv.GetItems())
	invData, err := invPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing InventoryItemList: %w", err)
	}
	n := copy(buf, invData)

	slog.Debug("item crystallized",
		"player", player.Name(),
		"itemID", item.ItemID(),
		"crystals", crystalCount,
		"crystalType", tmpl.CrystalType)

	return n, true, nil
}
