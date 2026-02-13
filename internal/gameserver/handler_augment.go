package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/game/augment"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
)

// handleRequestConfirmTargetItem processes the weapon selection for augmentation (C2S 0xD0:0x29).
// Validates the weapon and sends ExPutItemResultForVariation back.
//
// Phase 28: Augmentation System.
func (h *Handler) handleRequestConfirmTargetItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || h.augmentService == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestConfirmTargetItem(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestConfirmTargetItem: %w", err)
	}

	item := player.Inventory().GetItem(uint32(pkt.ObjectID))
	if item == nil {
		slog.Debug("augment: target item not found", "objectID", pkt.ObjectID, "character", player.Name())
		return 0, true, nil
	}

	if err := h.augmentService.ValidateTarget(item); err != nil {
		slog.Debug("augment: invalid target", "error", err, "character", player.Name())
		return 0, true, nil
	}

	resp := serverpackets.ExPutItemResultForVariation{
		ItemObjectID: pkt.ObjectID,
		ItemID:       item.ItemID(),
	}
	pktData, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ExPutItemResultForVariation: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestConfirmRefinerItem processes life stone selection (C2S 0xD0:0x2A).
// Validates the life stone and sends ExPutCommissionResult with gemstone cost.
//
// Phase 28: Augmentation System.
func (h *Handler) handleRequestConfirmRefinerItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || h.augmentService == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestConfirmRefinerItem(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestConfirmRefinerItem: %w", err)
	}

	refiner := player.Inventory().GetItem(uint32(pkt.RefinerObjectID))
	if refiner == nil {
		slog.Debug("augment: refiner not found", "objectID", pkt.RefinerObjectID, "character", player.Name())
		return 0, true, nil
	}

	if err := h.augmentService.ValidateRefiner(refiner); err != nil {
		slog.Debug("augment: invalid refiner", "error", err, "character", player.Name())
		return 0, true, nil
	}

	// Look up target weapon for crystal grade
	target := player.Inventory().GetItem(uint32(pkt.TargetObjectID))
	if target == nil {
		slog.Debug("augment: target weapon not found", "objectID", pkt.TargetObjectID, "character", player.Name())
		return 0, true, nil
	}

	gemID, gemCount := augment.GemstoneRequirement(target.Template().CrystalType)
	resp := serverpackets.ExPutCommissionResult{
		GemstoneObjectID: gemID,
		GemstoneCount:    gemCount,
	}
	pktData, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ExPutCommissionResult: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestConfirmGemStone processes gemstone confirmation (C2S 0xD0:0x2B).
// Validates gemstone count and sends ExPutIntensiveResult.
//
// Phase 28: Augmentation System.
func (h *Handler) handleRequestConfirmGemStone(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || h.augmentService == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestConfirmGemStone(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestConfirmGemStone: %w", err)
	}

	target := player.Inventory().GetItem(uint32(pkt.TargetObjectID))
	if target == nil {
		slog.Debug("augment: target weapon not found", "objectID", pkt.TargetObjectID, "character", player.Name())
		return 0, true, nil
	}

	gem := player.Inventory().GetItem(uint32(pkt.GemStoneObjectID))
	if gem == nil {
		slog.Debug("augment: gemstone not found", "objectID", pkt.GemStoneObjectID, "character", player.Name())
		return 0, true, nil
	}

	_, requiredCount := augment.GemstoneRequirement(target.Template().CrystalType)
	if int32(pkt.GemStoneCount) < requiredCount {
		slog.Debug("augment: insufficient gemstones",
			"have", pkt.GemStoneCount,
			"need", requiredCount,
			"character", player.Name())
		return 0, true, nil
	}

	resp := serverpackets.ExPutIntensiveResult{
		RefinerObjectID: pkt.GemStoneObjectID,
		LifeStoneItemID: gem.ItemID(),
		GemstoneItemID:  pkt.GemStoneObjectID,
		GemstoneCount:   requiredCount,
	}
	pktData, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ExPutIntensiveResult: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestRefine processes the augmentation execution (C2S 0xD0:0x2C).
// Consumes life stone and gemstones, generates augmentation, sends ExVariationResult.
//
// Phase 28: Augmentation System.
func (h *Handler) handleRequestRefine(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || h.augmentService == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestRefine(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestRefine: %w", err)
	}

	weapon := player.Inventory().GetItem(uint32(pkt.TargetObjectID))
	refiner := player.Inventory().GetItem(uint32(pkt.RefinerObjectID))
	gem := player.Inventory().GetItem(uint32(pkt.GemStoneObjectID))

	if weapon == nil || refiner == nil || gem == nil {
		return h.sendVariationResult(0, 0, 0, buf)
	}

	// Validate again before execution
	if err := h.augmentService.ValidateTarget(weapon); err != nil {
		return h.sendVariationResult(0, 0, 0, buf)
	}
	if err := h.augmentService.ValidateRefiner(refiner); err != nil {
		return h.sendVariationResult(0, 0, 0, buf)
	}

	// Check gemstone count (based on weapon crystal grade)
	weaponGrade := weapon.Template().CrystalType
	gemItemID, requiredCount := augment.GemstoneRequirement(weaponGrade)
	if gem.Count() < requiredCount {
		return h.sendVariationResult(0, 0, 0, buf)
	}

	// Consume life stone (remove entire item)
	if removed := player.Inventory().RemoveItem(refiner.ObjectID()); removed == nil {
		slog.Error("augment: remove life stone failed", "character", player.Name())
		return h.sendVariationResult(0, 0, 0, buf)
	}

	// Consume gemstones by template ID
	removed := player.Inventory().RemoveItemsByID(gemItemID, int64(requiredCount))
	if removed < int64(requiredCount) {
		slog.Error("augment: insufficient gemstones removed",
			"removed", removed,
			"required", requiredCount,
			"character", player.Name())
		return h.sendVariationResult(0, 0, 0, buf)
	}

	// Perform augmentation
	augID, err := h.augmentService.Augment(weapon, refiner.ItemID())
	if err != nil {
		slog.Error("augment: apply augmentation", "error", err, "character", player.Name())
		return h.sendVariationResult(0, 0, 0, buf)
	}

	slog.Info("weapon augmented",
		"character", player.Name(),
		"weapon", weapon.Name(),
		"augmentationID", augID)

	return h.sendVariationResult(augID, 0, 1, buf) // stat12=augID, stat34=0, result=success
}

// handleRequestConfirmCancelItem processes weapon selection for cancel (C2S 0xD0:0x2D).
// Validates the weapon has augmentation. No response packet needed.
//
// Phase 28: Augmentation System.
func (h *Handler) handleRequestConfirmCancelItem(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || h.augmentService == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestConfirmCancelItem(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestConfirmCancelItem: %w", err)
	}

	item := player.Inventory().GetItem(uint32(pkt.ObjectID))
	if item == nil || item.AugmentationID() == 0 {
		slog.Debug("augment cancel: not augmented or not found", "objectID", pkt.ObjectID, "character", player.Name())
	}

	return 0, true, nil
}

// handleRequestRefineCancel processes augmentation removal (C2S 0xD0:0x2E).
// Removes augmentation and sends ExVariationCancelResult.
//
// Phase 28: Augmentation System.
func (h *Handler) handleRequestRefineCancel(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil || h.augmentService == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestRefineCancel(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestRefineCancel: %w", err)
	}

	weapon := player.Inventory().GetItem(uint32(pkt.ObjectID))
	if weapon == nil {
		return h.sendVariationCancelResult(0, buf)
	}

	oldAugID, err := h.augmentService.RemoveAugmentation(weapon)
	if err != nil {
		slog.Debug("augment cancel: remove augmentation", "error", err, "character", player.Name())
		return h.sendVariationCancelResult(0, buf)
	}

	slog.Info("augmentation removed",
		"character", player.Name(),
		"weapon", weapon.Name(),
		"oldAugmentationID", oldAugID)

	return h.sendVariationCancelResult(1, buf)
}

// handleAugmentBypass processes augmentation NPC bypass commands.
// Returns (n, handled, error). If handled==false, caller tries other handlers.
//
// Bypass commands:
//   - Augment — open augmentation window
//   - AugmentCancel — open cancel window
//
// Phase 28: Augmentation System.
func (h *Handler) handleAugmentBypass(cmdName string, buf []byte) (int, bool, error) {
	if h.augmentService == nil {
		return 0, false, nil
	}

	switch cmdName {
	case "Augment":
		resp := serverpackets.ExShowVariationMakeWindow{}
		pktData, err := resp.Write()
		if err != nil {
			return 0, true, fmt.Errorf("writing ExShowVariationMakeWindow: %w", err)
		}
		n := copy(buf, pktData)
		return n, true, nil

	case "AugmentCancel":
		resp := serverpackets.ExShowVariationCancelWindow{}
		pktData, err := resp.Write()
		if err != nil {
			return 0, true, fmt.Errorf("writing ExShowVariationCancelWindow: %w", err)
		}
		n := copy(buf, pktData)
		return n, true, nil

	default:
		return 0, false, nil
	}
}

// sendVariationResult sends ExVariationResult packet.
// Java writes 3 fields: stat12 (aug pair 1), stat34 (aug pair 2), result.
func (h *Handler) sendVariationResult(stat12, stat34, result int32, buf []byte) (int, bool, error) {
	resp := serverpackets.ExVariationResult{
		Stat12: stat12,
		Stat34: stat34,
		Result: result,
	}
	pktData, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ExVariationResult: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// sendVariationCancelResult sends ExVariationCancelResult packet.
func (h *Handler) sendVariationCancelResult(success int32, buf []byte) (int, bool, error) {
	resp := serverpackets.ExVariationCancelResult{
		Result: success,
	}
	pktData, err := resp.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ExVariationCancelResult: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}
