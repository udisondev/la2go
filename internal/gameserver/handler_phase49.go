package gameserver

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"

	skilldata "github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/skill"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleRequestExMagicSkillUseGround processes 0xD0:0x2F — ground-targeted AoE skill.
func (h *Handler) handleRequestExMagicSkillUseGround(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestExMagicSkillUseGround(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestExMagicSkillUseGround: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if skill.CastMgr == nil {
		slog.Warn("CastManager not initialized, ignoring ground skill")
		return 0, true, nil
	}

	// Calculate heading from player to target point
	loc := player.Location()
	dx := float64(pkt.X - loc.X)
	dy := float64(pkt.Y - loc.Y)
	heading := uint16(math.Atan2(-dx, -dy) * 32768.0 / math.Pi)

	// Update player heading (Player → Character → WorldObject.SetLocation)
	player.SetLocation(model.NewLocation(loc.X, loc.Y, loc.Z, heading))

	// Use the skill (CastManager handles all validation)
	if err := skill.CastMgr.UseMagic(player, pkt.SkillID, pkt.CtrlPressed, pkt.ShiftPressed); err != nil {
		slog.Debug("ground skill use failed",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"error", err)

		actionFailed := serverpackets.NewActionFailed()
		failedData, err := actionFailed.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing ActionFailed: %w", err)
		}
		n := copy(buf, failedData)
		return n, true, nil
	}

	return 0, true, nil
}

// handleRequestExEnchantSkillInfo processes 0xD0:0x06 — skill enchant info request.
func (h *Handler) handleRequestExEnchantSkillInfo(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestExEnchantSkillInfo(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestExEnchantSkillInfo: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for enchant skill info")
	}

	if pkt.SkillID <= 0 || pkt.SkillLevel <= 0 {
		return 0, true, nil
	}

	// Player must be level 76+
	if player.Level() < 76 {
		slog.Debug("enchant skill info: level too low", "player", player.Name(), "level", player.Level())
		return 0, true, nil
	}

	// Look up enchant data
	enchantRoute := skilldata.GetEnchantSkillRoute(pkt.SkillID, pkt.SkillLevel)
	if enchantRoute == nil {
		slog.Debug("enchant skill info: no route",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"level", pkt.SkillLevel)
		return 0, true, nil
	}

	// Determine if enchant book is required (first enchant: levels 101, 141, 201)
	hasBookReq := pkt.SkillLevel == 101 || pkt.SkillLevel == 141 || pkt.SkillLevel == 201

	resp := &serverpackets.ExEnchantSkillInfo{
		SkillID:    pkt.SkillID,
		SkillLevel: pkt.SkillLevel,
		SpCost:     enchantRoute.SpCost,
		ExpCost:    enchantRoute.ExpCost,
		Rate:       enchantRoute.Rate,
		HasBookReq: hasBookReq,
	}

	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExEnchantSkillInfo: %w", err)
	}
	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestExEnchantSkill processes 0xD0:0x07 — enchant skill execution.
func (h *Handler) handleRequestExEnchantSkill(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestExEnchantSkill(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestExEnchantSkill: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for enchant skill")
	}

	if pkt.SkillID <= 0 || pkt.SkillLevel <= 0 {
		return 0, true, nil
	}

	// Player must be level 76+
	if player.Level() < 76 {
		return 0, true, nil
	}

	// Current skill level must be less than requested
	currentLevel := player.GetSkillLevel(pkt.SkillID)
	if currentLevel >= pkt.SkillLevel {
		slog.Warn("enchant skill: already at or above level",
			"player", player.Name(),
			"current", currentLevel,
			"requested", pkt.SkillLevel)
		return 0, true, nil
	}

	// Look up enchant data
	enchantRoute := skilldata.GetEnchantSkillRoute(pkt.SkillID, pkt.SkillLevel)
	if enchantRoute == nil {
		slog.Warn("enchant skill: no route",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"level", pkt.SkillLevel)
		return 0, true, nil
	}

	// SP check
	if player.SP() < int64(enchantRoute.SpCost) {
		slog.Warn("enchant skill: insufficient SP",
			"player", player.Name(),
			"have", player.SP(),
			"need", enchantRoute.SpCost)
		return 0, true, nil
	}

	// EXP check — don't allow if it would delevel
	if player.Experience() < enchantRoute.ExpCost {
		slog.Warn("enchant skill: insufficient EXP",
			"player", player.Name())
		return 0, true, nil
	}

	// Check enchant book requirement (item 6622) for first enchant
	if pkt.SkillLevel == 101 || pkt.SkillLevel == 141 || pkt.SkillLevel == 201 {
		inv := player.Inventory()
		bookCount := int64(0)
		for _, item := range inv.GetItems() {
			if item.ItemID() == 6622 {
				bookCount = int64(item.Count())
				break
			}
		}
		if bookCount < 1 {
			slog.Warn("enchant skill: missing enchant book",
				"player", player.Name())
			return 0, true, nil
		}
		// Consume book
		inv.RemoveItemsByID(6622, 1)
	}

	// Consume SP and EXP
	player.SetSP(player.SP() - int64(enchantRoute.SpCost))
	player.SetExperience(player.Experience() - enchantRoute.ExpCost)

	// Roll success
	success := rand.IntN(100) < int(enchantRoute.Rate)

	var result int32
	if success {
		// Learn enchanted skill level
		isPassive := false
		if tmpl := skilldata.GetSkillTemplate(pkt.SkillID, pkt.SkillLevel); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(pkt.SkillID, pkt.SkillLevel, isPassive)
		result = 1

		slog.Info("player enchanted skill",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"level", pkt.SkillLevel)
	} else {
		// Failure: reset to base level (level % 100, or the non-enchanted max)
		baseLevel := enchantRoute.BaseLevel
		if baseLevel <= 0 {
			baseLevel = pkt.SkillLevel % 100
			if baseLevel <= 0 {
				baseLevel = 1
			}
		}

		isPassive := false
		if tmpl := skilldata.GetSkillTemplate(pkt.SkillID, baseLevel); tmpl != nil {
			isPassive = tmpl.IsPassive()
		}
		player.AddSkill(pkt.SkillID, baseLevel, isPassive)

		slog.Info("player skill enchant failed",
			"player", player.Name(),
			"skillID", pkt.SkillID,
			"requested", pkt.SkillLevel,
			"resetTo", baseLevel)
	}

	// Send enchant result
	resultPkt := &serverpackets.ExEnchantSkillResult{Result: result}
	resultData, err := resultPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExEnchantSkillResult: %w", err)
	}
	n := copy(buf, resultData)

	// Send updated skill list
	skillList := serverpackets.NewSkillList(player.Skills())
	skillListData, err := skillList.Write()
	if err != nil {
		slog.Error("serializing SkillList", "error", err)
	} else {
		n2 := copy(buf[n:], skillListData)
		n += n2
	}

	return n, true, nil
}

// handleRequestBuySeed processes 0xC4 — buy seeds from Manor merchant.
func (h *Handler) handleRequestBuySeed(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestBuySeed(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestBuySeed: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for buy seed")
	}

	if h.manorMgr == nil {
		slog.Warn("manor manager not initialized")
		return 0, true, nil
	}

	if len(pkt.Items) == 0 {
		return 0, true, nil
	}

	// Calculate total cost and validate availability
	var totalCost int64
	for _, item := range pkt.Items {
		sp := h.manorMgr.SeedProduct(pkt.ManorID, item.ItemID, false)
		if sp == nil {
			slog.Warn("buy seed: seed not available",
				"player", player.Name(),
				"manorID", pkt.ManorID,
				"itemID", item.ItemID)
			return 0, true, nil
		}

		if sp.Amount() < item.Count {
			slog.Warn("buy seed: insufficient stock",
				"player", player.Name(),
				"itemID", item.ItemID,
				"have", sp.Amount(),
				"want", item.Count)
			return 0, true, nil
		}

		price := sp.Price()
		if price <= 0 {
			return 0, true, nil
		}

		totalCost += price * int64(item.Count)
		if totalCost < 0 { // overflow
			return 0, true, nil
		}
	}

	// Adena check
	inv := player.Inventory()
	if inv.GetAdena() < totalCost {
		slog.Warn("buy seed: insufficient adena",
			"player", player.Name(),
			"have", inv.GetAdena(),
			"need", totalCost)
		return 0, true, nil
	}

	// Execute purchase
	if err := inv.RemoveAdena(int32(totalCost)); err != nil {
		slog.Error("buy seed: failed to remove adena", "error", err)
		return 0, true, nil
	}

	for _, item := range pkt.Items {
		sp := h.manorMgr.SeedProduct(pkt.ManorID, item.ItemID, false)
		if sp != nil {
			sp.DecreaseAmount(item.Count)
		}
		// Add seeds to inventory (simplified — create Item with itemID)
		slog.Debug("seed purchased",
			"player", player.Name(),
			"seedID", item.ItemID,
			"count", item.Count)
	}

	slog.Info("player bought seeds",
		"player", player.Name(),
		"manorID", pkt.ManorID,
		"items", len(pkt.Items),
		"totalCost", totalCost)

	return 0, true, nil
}

// handleRequestManorList processes 0xD0:0x08 — manor list request.
func (h *Handler) handleRequestManorList(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Build castle names list (Interlude: 9 castles)
	castleNames := []string{
		"Gludio", "Dion", "Giran", "Oren", "Aden",
		"Innadril", "Goddard", "Rune", "Schuttgart",
	}

	pkt := &serverpackets.ExSendManorList{CastleNames: castleNames}
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExSendManorList: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestProcureCropList processes 0xD0:0x09 — sell crops to Manor.
func (h *Handler) handleRequestProcureCropList(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestProcureCropList(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestProcureCropList: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for crop sale")
	}

	if h.manorMgr == nil {
		slog.Warn("manor manager not initialized")
		return 0, true, nil
	}

	if len(pkt.Items) == 0 {
		return 0, true, nil
	}

	inv := player.Inventory()

	for _, entry := range pkt.Items {
		// Validate player has the crop item
		invItem := inv.GetItem(uint32(entry.ObjectID))
		if invItem == nil {
			slog.Warn("procure crop: item not found",
				"player", player.Name(),
				"objectID", entry.ObjectID)
			continue
		}

		if invItem.ItemID() != entry.ItemID || int64(invItem.Count()) < int64(entry.Count) {
			slog.Warn("procure crop: item mismatch or insufficient",
				"player", player.Name(),
				"itemID", entry.ItemID)
			continue
		}

		// Validate crop procure quota
		cp := h.manorMgr.CropProcureEntry(entry.ManorID, entry.ItemID, false)
		if cp == nil {
			slog.Warn("procure crop: not available for procure",
				"player", player.Name(),
				"manorID", entry.ManorID,
				"cropID", entry.ItemID)
			continue
		}

		if cp.Amount() < entry.Count {
			slog.Warn("procure crop: insufficient quota",
				"player", player.Name(),
				"have", cp.Amount(),
				"want", entry.Count)
			continue
		}

		// Get seed template for reward calculation
		seed := skilldata.GetSeedByCropID(entry.ItemID)
		if seed == nil {
			slog.Warn("procure crop: no seed for crop", "cropID", entry.ItemID)
			continue
		}

		// Execute: remove crop, decrease quota
		inv.RemoveItemsByID(entry.ItemID, int64(entry.Count))
		cp.DecreaseAmount(entry.Count)

		// Add reward item (Reward1 from seed template)
		rewardID := seed.Reward1
		if rewardID > 0 {
			slog.Debug("crop procured, reward pending",
				"player", player.Name(),
				"cropID", entry.ItemID,
				"count", entry.Count,
				"rewardID", rewardID,
				"manorID", entry.ManorID)
		}
	}

	return 0, true, nil
}
