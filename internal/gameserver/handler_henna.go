package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/db"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// Adena item ID (L2 standard).
const adenaItemID = 57

// handleRequestHennaItemList sends list of available hennas for the player's class.
func (h *Handler) handleRequestHennaItemList(_ context.Context, client *GameClient, body, buf []byte) (int, bool, error) {
	if _, err := clientpackets.ParseRequestHennaItemList(body); err != nil {
		return 0, true, fmt.Errorf("parsing RequestHennaItemList: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	pkt := serverpackets.NewHennaEquipList(player)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serializing HennaEquipList: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestHennaItemInfo sends detailed info about a specific henna before equipping.
func (h *Handler) handleRequestHennaItemInfo(_ context.Context, client *GameClient, body, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestHennaItemInfo(body)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestHennaItemInfo: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	hennaInfo := data.GetHennaInfo(pkt.SymbolID)
	if hennaInfo == nil {
		slog.Warn("henna not found", "dyeID", pkt.SymbolID)
		return 0, true, nil
	}

	resp := serverpackets.NewHennaItemDrawInfo(player, hennaInfo)
	respData, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serializing HennaItemDrawInfo: %w", err)
	}

	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestHennaEquip processes henna equip request.
func (h *Handler) handleRequestHennaEquip(ctx context.Context, client *GameClient, body, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestHennaEquip(body)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestHennaEquip: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	hennaInfo := data.GetHennaInfo(pkt.SymbolID)
	if hennaInfo == nil {
		slog.Warn("henna not found", "dyeID", pkt.SymbolID)
		return 0, true, nil
	}

	// Validate class
	if !hennaInfo.IsAllowedClass(player.ClassID()) {
		slog.Warn("henna not allowed for class",
			"dyeID", pkt.SymbolID,
			"classID", player.ClassID(),
			"player", player.Name())
		return 0, true, nil
	}

	// Check free slots
	if player.GetHennaEmptySlots() == 0 {
		slog.Warn("no free henna slots", "player", player.Name())
		return h.sendActionFailed(buf)
	}

	// Check dye count
	inv := player.Inventory()
	dyeCount := inv.CountItemsByID(hennaInfo.DyeItemID)
	if dyeCount < int64(hennaInfo.WearCount) {
		slog.Warn("not enough dye items",
			"dyeItemID", hennaInfo.DyeItemID,
			"have", dyeCount,
			"need", hennaInfo.WearCount,
			"player", player.Name())
		return h.sendActionFailed(buf)
	}

	// Check adena
	if inv.GetAdena() < hennaInfo.WearFee {
		slog.Warn("not enough adena for henna",
			"have", inv.GetAdena(),
			"need", hennaInfo.WearFee,
			"player", player.Name())
		return h.sendActionFailed(buf)
	}

	// Consume dye items
	inv.RemoveItemsByID(hennaInfo.DyeItemID, int64(hennaInfo.WearCount))

	// Consume adena
	if err := inv.RemoveAdena(int32(hennaInfo.WearFee)); err != nil {
		slog.Error("remove adena for henna", "error", err)
		return h.sendActionFailed(buf)
	}

	// Add henna to player
	slot, err := player.AddHenna(pkt.SymbolID)
	if err != nil {
		slog.Error("add henna to player", "error", err)
		return h.sendActionFailed(buf)
	}

	// Persist to DB
	h.persistHennaAdd(ctx, player, int32(slot), pkt.SymbolID)

	slog.Info("henna equipped",
		"player", player.Name(),
		"dyeID", pkt.SymbolID,
		"slot", slot)

	// Send HennaInfo (updated stat bonuses)
	return h.sendHennaInfo(player, buf)
}

// handleRequestHennaRemoveList sends list of currently equipped hennas.
func (h *Handler) handleRequestHennaRemoveList(_ context.Context, client *GameClient, body, buf []byte) (int, bool, error) {
	if _, err := clientpackets.ParseRequestHennaRemoveList(body); err != nil {
		return 0, true, fmt.Errorf("parsing RequestHennaRemoveList: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Send HennaRemoveList — re-use HennaInfo which shows equipped hennas
	return h.sendHennaInfo(player, buf)
}

// handleRequestHennaItemRemoveInfo sends detailed info about removing a specific henna.
func (h *Handler) handleRequestHennaItemRemoveInfo(_ context.Context, client *GameClient, body, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestHennaItemRemoveInfo(body)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestHennaItemRemoveInfo: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	hennaInfo := data.GetHennaInfo(pkt.SymbolID)
	if hennaInfo == nil {
		return 0, true, nil
	}

	resp := serverpackets.NewHennaItemRemoveInfo(player, hennaInfo)
	respData, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serializing HennaItemRemoveInfo: %w", err)
	}

	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestHennaRemove processes henna removal request.
func (h *Handler) handleRequestHennaRemove(ctx context.Context, client *GameClient, body, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestHennaRemove(body)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestHennaRemove: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Find which slot has this henna
	slot := 0
	hennaList := player.GetHennaList()
	for i, hs := range hennaList {
		if hs != nil && hs.DyeID == pkt.SymbolID {
			slot = i + 1
			break
		}
	}

	if slot == 0 {
		slog.Warn("henna not found in slots",
			"dyeID", pkt.SymbolID,
			"player", player.Name())
		return h.sendActionFailed(buf)
	}

	hennaInfo := data.GetHennaInfo(pkt.SymbolID)
	if hennaInfo == nil {
		return h.sendActionFailed(buf)
	}

	// Check adena for cancel fee
	if player.Inventory().GetAdena() < hennaInfo.CancelFee {
		slog.Warn("not enough adena for henna removal",
			"have", player.Inventory().GetAdena(),
			"need", hennaInfo.CancelFee,
			"player", player.Name())
		return h.sendActionFailed(buf)
	}

	// Remove henna
	dyeID, err := player.RemoveHenna(slot)
	if err != nil {
		slog.Error("remove henna", "slot", slot, "error", err)
		return h.sendActionFailed(buf)
	}

	// Consume adena for cancel fee
	if err := player.Inventory().RemoveAdena(int32(hennaInfo.CancelFee)); err != nil {
		slog.Error("remove cancel fee adena", "error", err)
	}

	// Return dye items to player (cancelCount)
	if hennaInfo.CancelCount > 0 {
		h.addDyeItemToInventory(player, hennaInfo.DyeItemID, hennaInfo.CancelCount)
	}

	// Persist removal to DB
	h.persistHennaRemove(ctx, player, int32(slot))

	slog.Info("henna removed",
		"player", player.Name(),
		"dyeID", dyeID,
		"slot", slot)

	// Send HennaInfo (updated stat bonuses)
	return h.sendHennaInfo(player, buf)
}

// sendHennaInfo sends HennaInfo packet to client.
func (h *Handler) sendHennaInfo(player *model.Player, buf []byte) (int, bool, error) {
	pkt := serverpackets.NewHennaInfo(player)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serializing HennaInfo: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// sendActionFailed sends ActionFailed packet to client.
func (h *Handler) sendActionFailed(buf []byte) (int, bool, error) {
	pkt := serverpackets.NewActionFailed()
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serializing ActionFailed: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// persistHennaAdd saves henna addition to DB asynchronously.
func (h *Handler) persistHennaAdd(ctx context.Context, player *model.Player, slot, dyeID int32) {
	if h.persister == nil {
		return
	}
	if err := h.persister.SavePlayer(ctx, player); err != nil {
		slog.Error("persist henna add",
			"characterID", player.CharacterID(),
			"slot", slot,
			"dyeID", dyeID,
			"error", err)
	}
}

// persistHennaRemove saves henna removal to DB asynchronously.
func (h *Handler) persistHennaRemove(ctx context.Context, player *model.Player, slot int32) {
	if h.persister == nil {
		return
	}
	if err := h.persister.SavePlayer(ctx, player); err != nil {
		slog.Error("persist henna remove",
			"characterID", player.CharacterID(),
			"slot", slot,
			"error", err)
	}
}

// addDyeItemToInventory adds dye items back to player inventory (for henna cancel).
// Creates a real Item with ObjectID and adds to inventory.
func (h *Handler) addDyeItemToInventory(player *model.Player, dyeItemID, count int32) {
	objectID := world.IDGenerator().NextItemID()
	tmpl := db.ItemDefToTemplate(dyeItemID)
	item, err := model.NewItem(objectID, dyeItemID, player.CharacterID(), count, tmpl)
	if err != nil {
		slog.Error("failed to create dye item for henna cancel",
			"player", player.Name(),
			"dyeItemID", dyeItemID,
			"error", err)
		return
	}
	if err := player.Inventory().AddItem(item); err != nil {
		slog.Error("failed to add dye item to inventory",
			"player", player.Name(),
			"dyeItemID", dyeItemID,
			"error", err)
		return
	}

	slog.Info("dye items returned to inventory",
		"player", player.Name(),
		"dyeItemID", dyeItemID,
		"count", count)
}
