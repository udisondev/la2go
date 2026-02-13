package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	skilldata "github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// handleRequestActionUse handles ActionUse packets (opcode 0x45).
// Dispatches pet/summon commands based on actionID.
// Phase 19: Pets/Summons System.
func (h *Handler) handleRequestActionUse(
	_ context.Context,
	client *GameClient,
	data, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Java format: actionId(int32) + ctrlPressed(int32) + shiftPressed(byte) = 9 bytes
	if len(data) < 9 {
		return 0, true, nil
	}

	r := packet.NewReader(data)

	actionID, err := r.ReadInt()
	if err != nil {
		return 0, true, fmt.Errorf("read actionID: %w", err)
	}

	ctrlPressed, err := r.ReadInt()
	if err != nil {
		return 0, true, fmt.Errorf("read ctrlPressed: %w", err)
	}

	shiftPressed, err := r.ReadByte()
	if err != nil {
		return 0, true, fmt.Errorf("read shiftPressed: %w", err)
	}

	_ = ctrlPressed  // reserved for future use (pet skill targeting)
	_ = shiftPressed // reserved for future use (force attack)

	switch actionID {
	case clientpackets.ActionPetFollow, clientpackets.ActionSrvFollow:
		return h.handlePetFollow(player, buf)
	case clientpackets.ActionPetAttack, clientpackets.ActionSrvAttack:
		return h.handlePetAttack(player, buf)
	case clientpackets.ActionPetStop, clientpackets.ActionSrvStop:
		return h.handlePetStop(player, buf)
	case clientpackets.ActionPetUnsummon, clientpackets.ActionSrvUnsummon:
		return h.handlePetUnsummon(player, client, buf)
	default:
		slog.Debug("unknown RequestActionUse",
			"actionID", actionID,
			"player", player.Name())
		return 0, true, nil
	}
}

// handlePetFollow commands pet/summon to follow owner.
func (h *Handler) handlePetFollow(player *model.Player, buf []byte) (int, bool, error) {
	if !player.HasSummon() {
		return 0, true, nil
	}

	slog.Debug("pet follow command",
		"player", player.Name(),
		"summonID", player.Summon().ObjectID())

	player.Summon().SetFollow(true)
	player.Summon().ClearTarget()
	player.Summon().SetIntention(model.IntentionFollow)

	return 0, true, nil
}

// handlePetAttack commands pet/summon to attack player's current target.
func (h *Handler) handlePetAttack(player *model.Player, buf []byte) (int, bool, error) {
	if !player.HasSummon() {
		return 0, true, nil
	}

	target := player.Target()
	if target == nil {
		return 0, true, nil
	}

	slog.Debug("pet attack command",
		"player", player.Name(),
		"summonID", player.Summon().ObjectID(),
		"targetID", target.ObjectID())

	player.Summon().SetTarget(target.ObjectID())
	player.Summon().SetFollow(false)
	player.Summon().SetIntention(model.IntentionAttack)

	return 0, true, nil
}

// handlePetStop commands pet/summon to stop all actions.
func (h *Handler) handlePetStop(player *model.Player, buf []byte) (int, bool, error) {
	if !player.HasSummon() {
		return 0, true, nil
	}

	slog.Debug("pet stop command",
		"player", player.Name(),
		"summonID", player.Summon().ObjectID())

	player.Summon().ClearTarget()
	player.Summon().SetFollow(false)
	player.Summon().SetIntention(model.IntentionIdle)

	return 0, true, nil
}

// handlePetUnsummon removes the pet/summon from the world.
func (h *Handler) handlePetUnsummon(player *model.Player, client *GameClient, buf []byte) (int, bool, error) {
	if !player.HasSummon() {
		return 0, true, nil
	}

	summon := player.Summon()

	slog.Info("unsummoning pet",
		"player", player.Name(),
		"summonID", summon.ObjectID(),
		"type", summon.Type())

	pkt := serverpackets.NewPetDelete(int32(summon.Type()), summon.ObjectID())
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("write PetDelete: %w", err)
	}

	n := copy(buf, pktData)

	player.ClearSummon()

	return n, true, nil
}

// handleRequestChangePetName handles pet name change request.
// Phase 19: Pets/Summons System.
func (h *Handler) handleRequestChangePetName(
	_ context.Context,
	client *GameClient,
	data, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if !player.HasSummon() || !player.Summon().IsPet() {
		return 0, true, nil
	}

	if len(data) < 2 {
		return 0, true, nil
	}

	r := packet.NewReader(data)

	name, err := r.ReadString()
	if err != nil {
		return 0, true, fmt.Errorf("read pet name: %w", err)
	}

	if len(name) < 1 || len(name) > 16 {
		slog.Warn("invalid pet name length",
			"player", player.Name(),
			"nameLen", len(name))
		return 0, true, nil
	}

	player.Summon().SetName(name)

	slog.Info("pet renamed",
		"player", player.Name(),
		"petName", name)

	pkt := serverpackets.NewPetInfo(player.Summon(), player.Name())
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("write PetInfo: %w", err)
	}

	return copy(buf, pktData), true, nil
}

// handleRequestGiveItemToPet handles giving an item from player to pet inventory.
// Phase 19: Pets/Summons System.
func (h *Handler) handleRequestGiveItemToPet(
	_ context.Context,
	client *GameClient,
	data, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || !player.HasSummon() || !player.Summon().IsPet() {
		return 0, true, nil
	}

	if len(data) < 12 {
		return 0, true, nil
	}

	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return 0, true, fmt.Errorf("read objectID: %w", err)
	}

	amount, err := r.ReadInt()
	if err != nil {
		return 0, true, fmt.Errorf("read amount: %w", err)
	}

	slog.Debug("give item to pet",
		"player", player.Name(),
		"objectID", objectID,
		"amount", amount)

	// Transfer item from player inventory to pet inventory.
	item := player.Inventory().GetItem(uint32(objectID))
	if item == nil {
		slog.Warn("give item to pet: item not found in player inventory",
			"player", player.Name(),
			"objectID", objectID)
		return 0, true, nil
	}

	// Prevent transferring equipped items
	if item.IsEquipped() {
		slog.Warn("give item to pet: cannot transfer equipped item",
			"player", player.Name(),
			"itemID", item.ItemID())
		return 0, true, nil
	}

	pet, ok := player.Summon().WorldObject.Data.(*model.Pet)
	if !ok {
		return 0, true, nil
	}

	removed := player.Inventory().RemoveItem(uint32(objectID))
	if removed == nil {
		return 0, true, nil
	}

	if err := pet.Inventory().AddItem(removed); err != nil {
		// Rollback: return item to player
		if rbErr := player.Inventory().AddItem(removed); rbErr != nil {
			slog.Error("failed to rollback item to player inventory",
				"player", player.Name(),
				"itemID", removed.ItemID(),
				"error", rbErr)
		}
		return 0, true, fmt.Errorf("adding item to pet inventory: %w", err)
	}

	slog.Info("item transferred to pet",
		"player", player.Name(),
		"itemID", item.ItemID(),
		"objectID", objectID)

	// Send InventoryUpdate to reflect the removal.
	invUpdate := serverpackets.NewInventoryUpdate(serverpackets.InvUpdateEntry{
		ChangeType: serverpackets.InvUpdateRemove,
		Item:       removed,
	})
	invData, err := invUpdate.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryUpdate for pet transfer",
			"error", err)
		return 0, true, nil
	}

	return copy(buf, invData), true, nil
}

// handleRequestGetItemFromPet handles getting an item from pet to player inventory.
// Phase 19: Pets/Summons System.
func (h *Handler) handleRequestGetItemFromPet(
	_ context.Context,
	client *GameClient,
	data, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || !player.HasSummon() || !player.Summon().IsPet() {
		return 0, true, nil
	}

	if len(data) < 12 {
		return 0, true, nil
	}

	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return 0, true, fmt.Errorf("read objectID: %w", err)
	}

	amount, err := r.ReadInt()
	if err != nil {
		return 0, true, fmt.Errorf("read amount: %w", err)
	}

	slog.Debug("get item from pet",
		"player", player.Name(),
		"objectID", objectID,
		"amount", amount)

	// Transfer item from pet inventory to player inventory.
	pet, ok := player.Summon().WorldObject.Data.(*model.Pet)
	if !ok {
		return 0, true, nil
	}

	item := pet.Inventory().GetItem(uint32(objectID))
	if item == nil {
		slog.Warn("get item from pet: item not found in pet inventory",
			"player", player.Name(),
			"objectID", objectID)
		return 0, true, nil
	}

	removed := pet.Inventory().RemoveItem(uint32(objectID))
	if removed == nil {
		return 0, true, nil
	}

	if err := player.Inventory().AddItem(removed); err != nil {
		// Rollback: return item to pet
		if rbErr := pet.Inventory().AddItem(removed); rbErr != nil {
			slog.Error("failed to rollback item to pet inventory",
				"player", player.Name(),
				"itemID", removed.ItemID(),
				"error", rbErr)
		}
		return 0, true, fmt.Errorf("adding item to player inventory: %w", err)
	}

	slog.Info("item transferred from pet",
		"player", player.Name(),
		"itemID", item.ItemID(),
		"objectID", objectID)

	// Send InventoryUpdate to reflect the addition.
	invUpdate := serverpackets.NewInventoryUpdate(serverpackets.InvUpdateEntry{
		ChangeType: serverpackets.InvUpdateAdd,
		Item:       removed,
	})
	invData, err := invUpdate.Write()
	if err != nil {
		slog.Error("failed to serialize InventoryUpdate for pet transfer",
			"error", err)
		return 0, true, nil
	}

	return copy(buf, invData), true, nil
}

// handleRequestPetGetItem handles pet picking up an item from ground.
// Phase 19: Pets/Summons System.
func (h *Handler) handleRequestPetGetItem(
	_ context.Context,
	client *GameClient,
	data, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || !player.HasSummon() {
		return 0, true, nil
	}

	if len(data) < 4 {
		return 0, true, nil
	}

	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return 0, true, fmt.Errorf("read objectID: %w", err)
	}

	slog.Debug("pet get item from ground",
		"player", player.Name(),
		"objectID", objectID)

	// Pet picks up item from ground (same as player pickup, but into pet inventory).
	worldInst := world.Instance()
	droppedItem, ok := worldInst.GetItem(uint32(objectID))
	if !ok {
		slog.Warn("pet pickup: item not found in world",
			"player", player.Name(),
			"objectID", objectID)
		return 0, true, nil
	}

	// Check pet distance to item (must be within pickup range).
	summon := player.Summon()
	petLoc := summon.Location()
	itemLoc := droppedItem.Location()
	dx := float64(petLoc.X - itemLoc.X)
	dy := float64(petLoc.Y - itemLoc.Y)
	if dx*dx+dy*dy > 2500*2500 { // 2500 game units max pickup range
		return 0, true, nil
	}

	item := droppedItem.Item()

	// If pet is a Pet type, add to pet inventory; otherwise to player inventory.
	pet, isPet := summon.WorldObject.Data.(*model.Pet)
	if isPet {
		if err := pet.Inventory().AddItem(item); err != nil {
			slog.Error("pet pickup: failed to add to pet inventory",
				"player", player.Name(),
				"itemID", item.ItemID(),
				"error", err)
			return 0, true, nil
		}
	} else {
		if err := player.Inventory().AddItem(item); err != nil {
			slog.Error("pet pickup: failed to add to player inventory",
				"player", player.Name(),
				"itemID", item.ItemID(),
				"error", err)
			return 0, true, nil
		}
	}

	// Remove from world.
	worldInst.RemoveObject(uint32(objectID))

	// Broadcast DeleteObject to nearby players.
	deleteObj := serverpackets.NewDeleteObject(objectID)
	deleteData, err := deleteObj.Write()
	if err != nil {
		slog.Error("failed to serialize DeleteObject for pet pickup",
			"objectID", objectID,
			"error", err)
	} else {
		h.clientManager.BroadcastToVisible(player, deleteData, len(deleteData))
	}

	slog.Info("pet picked up item",
		"player", player.Name(),
		"itemID", item.ItemID(),
		"objectID", objectID)

	return 0, true, nil
}

// handleRequestPetUseItem handles using an item from the pet inventory.
// For food items the pet's CurrentFed is increased and the item consumed.
// For other items an ActionFailed is returned (equip not yet implemented).
// Phase 52: Pet/Summon system gaps.
// Java reference: RequestPetUseItem.java
func (h *Handler) handleRequestPetUseItem(
	_ context.Context,
	client *GameClient,
	data, buf []byte,
) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || !player.HasSummon() || !player.Summon().IsPet() {
		return 0, true, nil
	}

	if len(data) < 4 {
		return 0, true, nil
	}

	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return 0, true, fmt.Errorf("read objectID: %w", err)
	}

	pet, ok := player.Summon().WorldObject.Data.(*model.Pet)
	if !ok {
		return 0, true, nil
	}

	item := pet.Inventory().GetItem(uint32(objectID))
	if item == nil {
		return 0, true, nil
	}

	itemDef := skilldata.GetItemDef(item.ItemID())

	// Food item: handler == "PetFood"
	if itemDef != nil && itemDef.Handler() == "PetFood" {
		// Увеличиваем сытость на фиксированное значение (стандарт L2: +100).
		const feedAmount = 100
		newFed := pet.CurrentFed() + feedAmount
		pet.SetCurrentFed(newFed)

		// Потребляем 1 единицу еды из инвентаря пета.
		pet.Inventory().RemoveItemsByID(item.ItemID(), 1)

		slog.Debug("pet consumed food",
			"player", player.Name(),
			"itemID", item.ItemID(),
			"newFed", pet.CurrentFed())

		// Отправляем обновленный список предметов пета.
		pkt := serverpackets.NewPetItemList(pet.Inventory().GetItems())
		pktData, pktErr := pkt.Write()
		if pktErr != nil {
			slog.Error("write PetItemList after pet food",
				"error", pktErr)
			return 0, true, nil
		}

		return copy(buf, pktData), true, nil
	}

	// Для остальных предметов (equip и прочее) — пока stub, ActionFailed.
	slog.Debug("pet use item: unhandled item type",
		"player", player.Name(),
		"itemID", item.ItemID(),
		"handler", func() string {
			if itemDef != nil {
				return itemDef.Handler()
			}
			return "unknown"
		}())

	af := serverpackets.NewActionFailed()
	afData, afErr := af.Write()
	if afErr != nil {
		return 0, true, fmt.Errorf("write ActionFailed: %w", afErr)
	}

	return copy(buf, afData), true, nil
}
