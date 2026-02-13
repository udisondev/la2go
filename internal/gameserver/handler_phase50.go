package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/crest"
	"github.com/udisondev/la2go/internal/game/manor"
	"github.com/udisondev/la2go/internal/gameserver/clan"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleRequestChangePartyLeader processes 0xD0:0x04 — transfer party leadership.
func (h *Handler) handleRequestChangePartyLeader(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestChangePartyLeader(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestChangePartyLeader: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for change party leader")
	}

	party := player.GetParty()
	if party == nil {
		slog.Debug("change party leader: not in party", "player", player.Name())
		return 0, true, nil
	}

	if !party.IsLeader(player.ObjectID()) {
		slog.Debug("change party leader: not leader",
			"player", player.Name())
		return 0, true, nil
	}

	if pkt.Name == "" || pkt.Name == player.Name() {
		return 0, true, nil
	}

	// Find target member by name
	var targetFound bool
	for _, m := range party.Members() {
		if m.Name() == pkt.Name {
			party.SetLeader(m)
			targetFound = true
			break
		}
	}

	if !targetFound {
		slog.Debug("change party leader: target not in party",
			"player", player.Name(),
			"target", pkt.Name)
		return 0, true, nil
	}

	// Broadcast updated PartySmallWindowAll to all members
	for _, m := range party.Members() {
		windowAll := serverpackets.NewPartySmallWindowAll(party, m.ObjectID())
		windowData, err := windowAll.Write()
		if err != nil {
			slog.Error("serializing PartySmallWindowAll for leader change", "error", err)
			continue
		}
		if err := h.clientManager.SendToPlayer(m.ObjectID(), windowData, len(windowData)); err != nil {
			slog.Warn("failed to send PartySmallWindowAll",
				"target", m.Name(), "error", err)
		}
	}

	slog.Info("party leader changed",
		"oldLeader", player.Name(),
		"newLeader", pkt.Name,
		"partyID", party.ID())

	return 0, true, nil
}

// handleRequestAllyCrest processes 0x88 — client requests alliance crest image.
func (h *Handler) handleRequestAllyCrest(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAllyCrest(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestAllyCrest: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	var crestData []byte
	if c := h.crestTbl.Crest(pkt.CrestID); c != nil {
		crestData = c.Data()
	}

	resp := &serverpackets.AllyCrest{
		CrestID: pkt.CrestID,
		Data:    crestData,
	}
	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing AllyCrest: %w", err)
	}
	n := copy(buf, respData)
	return n, true, nil
}

// handleSendBypassBuildCmd processes 0x5B — GM //command bypass.
func (h *Handler) handleSendBypassBuildCmd(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseSendBypassBuildCmd(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing SendBypassBuildCmd: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if !player.IsGM() {
		slog.Warn("non-GM player attempted admin bypass",
			"player", player.Name(),
			"command", pkt.Command)
		return h.sendActionFailed(buf)
	}

	slog.Info("GM bypass command",
		"player", player.Name(),
		"accessLevel", player.AccessLevel(),
		"command", pkt.Command)

	// Delegate to admin handler if available
	if h.adminHandler != nil {
		h.adminHandler.HandleAdminCommand(player, "admin_"+pkt.Command)
	}

	return 0, true, nil
}

// handleRequestExPledgeCrestLarge processes 0xD0:0x10 — request large clan crest image.
func (h *Handler) handleRequestExPledgeCrestLarge(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	r := packet.NewReader(dataBytes)
	crestID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("reading large crest id: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	var crestData []byte
	if c := h.crestTbl.Crest(crestID); c != nil {
		crestData = c.Data()
	}

	resp := &serverpackets.ExPledgeEmblem{
		CrestID: crestID,
		Data:    crestData,
	}
	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExPledgeEmblem: %w", err)
	}
	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestExSetPledgeCrestLarge processes 0xD0:0x11 — upload large clan crest.
func (h *Handler) handleRequestExSetPledgeCrestLarge(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestExSetPledgeCrestLarge(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestExSetPledgeCrestLarge: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if player.ClanID() == 0 || h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(player.ClanID())
	if c == nil {
		return 0, true, nil
	}

	if c.IsDissolving() {
		return 0, true, nil
	}

	// Проверяем привилегию CHANGE_CREST.
	member := c.Member(int64(player.ObjectID()))
	if member == nil || !member.HasPrivilege(clan.PrivCLRegisterCrest) {
		slog.Debug("set large crest: no privilege", "player", player.Name())
		return 0, true, nil
	}

	// Для большого герба требуется минимум 3 уровень клана.
	if c.Level() < 3 {
		slog.Debug("set large crest: clan level too low",
			"player", player.Name(), "level", c.Level())
		return 0, true, nil
	}

	if pkt.Length <= 0 {
		// Удалить герб.
		oldID := c.LargeCrestID()
		if oldID != 0 {
			h.crestTbl.RemoveCrest(oldID)
		}
		c.SetLargeCrestID(0)

		slog.Info("large clan crest removed",
			"clan", c.Name(), "by", player.Name())
	} else {
		newCrest, err := h.crestTbl.CreateCrest(pkt.Data, crest.PledgeLarge)
		if err != nil {
			slog.Warn("set large crest: create crest",
				"player", player.Name(), "error", err)
			return 0, true, nil
		}

		// Удаляем старый герб если был.
		oldID := c.LargeCrestID()
		if oldID != 0 {
			h.crestTbl.RemoveCrest(oldID)
		}
		c.SetLargeCrestID(newCrest.ID())

		slog.Info("large clan crest set",
			"clan", c.Name(), "crestID", newCrest.ID(),
			"size", len(pkt.Data), "by", player.Name())
	}

	// Broadcast UserInfo ко всем онлайн-членам клана.
	h.broadcastClanUserInfo(c)

	return 0, true, nil
}

// handleRequestOlympiadObserverEnd processes 0xD0:0x12 — leave olympiad observer mode.
// Stub: acknowledged until Olympiad system fully supports observer mode.
func (h *Handler) handleRequestOlympiadObserverEnd(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	// TODO: implement when Olympiad observer mode is added
	return 0, true, nil
}

// handleRequestOlympiadMatchList processes 0xD0:0x13 — request olympiad match list.
// Stub: acknowledged until Olympiad system fully supports match listing.
func (h *Handler) handleRequestOlympiadMatchList(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	// TODO: implement when Olympiad match listing is added
	return 0, true, nil
}

// handleRequestExMPCCShowPartyMembersInfo processes 0xD0:0x26 — request MPCC party member info.
// Stub: acknowledged until MPCC (Command Channel) system is implemented.
func (h *Handler) handleRequestExMPCCShowPartyMembersInfo(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	// TODO: implement when MPCC system is added
	return 0, true, nil
}

// handleRequestGmList processes 0x81 — client requests online GM list.
func (h *Handler) handleRequestGmList(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Send empty GM list (no visible GMs)
	pkt := &serverpackets.GmList{Names: nil}
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing GmList: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestAllyInfo processes 0x8E — client requests alliance info.
func (h *Handler) handleRequestAllyInfo(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 || h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil || c.AllyID() == 0 {
		return 0, true, nil
	}

	// Build alliance info from all clans in the alliance
	allyClanIDs := h.clanTable.ClanAllies(c.AllyID())
	clans := make([]serverpackets.AllianceClanInfo, 0, len(allyClanIDs))
	var totalMembers, onlineMembers int32
	leaderClanName := c.Name()
	leaderName := ""

	for _, allyC := range allyClanIDs {
		memberCount := int32(allyC.MemberCount())
		onlineCount := int32(allyC.OnlineMemberCount())
		totalMembers += memberCount
		onlineMembers += onlineCount

		// Find leader name
		leaderMember := allyC.Member(allyC.LeaderID())
		clanLeaderName := ""
		if leaderMember != nil {
			clanLeaderName = leaderMember.Name()
		}

		// The leader of the alliance is the leader of the leader clan
		if allyC.ID() == c.AllyID() || allyC.LeaderID() == c.LeaderID() {
			leaderClanName = allyC.Name()
			leaderName = clanLeaderName
		}

		clans = append(clans, serverpackets.AllianceClanInfo{
			ClanName:          allyC.Name(),
			ClanLevel:         allyC.Level(),
			ClanLeaderName:    clanLeaderName,
			ClanTotalMembers:  memberCount,
			ClanOnlineMembers: onlineCount,
		})
	}

	resp := &serverpackets.AllianceInfo{
		AllyName:       c.AllyName(),
		TotalMembers:   totalMembers,
		OnlineMembers:  onlineMembers,
		LeaderClanName: leaderClanName,
		LeaderName:     leaderName,
		Clans:          clans,
	}

	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing AllianceInfo: %w", err)
	}
	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestSetPledgeCrest processes 0x53 — upload clan crest.
func (h *Handler) handleRequestSetPledgeCrest(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSetPledgeCrest(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestSetPledgeCrest: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if player.ClanID() == 0 || h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(player.ClanID())
	if c == nil {
		return 0, true, nil
	}

	if c.IsDissolving() {
		return 0, true, nil
	}

	// Проверяем привилегию CHANGE_CREST.
	member := c.Member(int64(player.ObjectID()))
	if member == nil || !member.HasPrivilege(clan.PrivCLRegisterCrest) {
		slog.Debug("set pledge crest: no privilege", "player", player.Name())
		return 0, true, nil
	}

	// Для герба требуется минимум 3 уровень клана.
	if c.Level() < 3 {
		slog.Debug("set pledge crest: clan level too low",
			"player", player.Name(), "level", c.Level())
		return 0, true, nil
	}

	if pkt.Length <= 0 {
		// Удалить герб.
		oldID := c.CrestID()
		if oldID != 0 {
			h.crestTbl.RemoveCrest(oldID)
		}
		c.SetCrestID(0)

		slog.Info("clan crest removed",
			"clan", c.Name(), "by", player.Name())
	} else {
		newCrest, err := h.crestTbl.CreateCrest(pkt.Data, crest.Pledge)
		if err != nil {
			slog.Warn("set pledge crest: create crest",
				"player", player.Name(), "error", err)
			return 0, true, nil
		}

		// Удаляем старый герб если был.
		oldID := c.CrestID()
		if oldID != 0 {
			h.crestTbl.RemoveCrest(oldID)
		}
		c.SetCrestID(newCrest.ID())

		slog.Info("clan crest set",
			"clan", c.Name(), "crestID", newCrest.ID(),
			"size", len(pkt.Data), "by", player.Name())
	}

	// Broadcast UserInfo ко всем онлайн-членам клана.
	h.broadcastClanUserInfo(c)

	return 0, true, nil
}

// handleRequestSetAllyCrest — moved to handler_phase51.go (Phase 51: Alliance System)

// handleRequestSetSeed processes 0xD0:0x0A — manor seed production setup.
// Clan leader sets seed production for the next manor period.
func (h *Handler) handleRequestSetSeed(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	if h.manorMgr == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestSetSeed(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestSetSeed: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if !h.manorMgr.IsModifiable() {
		slog.Debug("set seed: manor not modifiable", "player", player.Name())
		return h.sendActionFailed(buf)
	}

	// Валидация: игрок — лидер клана, владеющего замком.
	if !h.isManorOwner(player, pkt.ManorID) {
		slog.Debug("set seed: not manor owner",
			"player", player.Name(), "manorID", pkt.ManorID)
		return h.sendActionFailed(buf)
	}

	seeds := make([]*manor.SeedProduction, 0, len(pkt.Seeds))
	for _, entry := range pkt.Seeds {
		tmpl := data.GetSeedTemplate(entry.SeedID)
		if tmpl == nil || tmpl.CastleID != pkt.ManorID {
			continue
		}

		if entry.StartAmount < 0 || entry.StartAmount > tmpl.LimitSeeds {
			continue
		}

		minPrice := int32(data.SeedMinPrice(entry.SeedID))
		maxPrice := int32(data.SeedMaxPrice(entry.SeedID))
		if entry.Price < minPrice || entry.Price > maxPrice {
			continue
		}

		seeds = append(seeds, manor.NewSeedProduction(
			entry.SeedID, entry.StartAmount, int64(entry.Price), entry.StartAmount,
		))
	}

	h.manorMgr.SetNextSeedProduction(pkt.ManorID, seeds)

	slog.Info("manor seed production set",
		"player", player.Name(),
		"manorID", pkt.ManorID,
		"seeds", len(seeds))

	return 0, true, nil
}

// handleRequestSetCrop processes 0xD0:0x0B — manor crop procurement setup.
// Clan leader sets crop procurement for the next manor period.
func (h *Handler) handleRequestSetCrop(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	if h.manorMgr == nil {
		return 0, true, nil
	}

	pkt, err := clientpackets.ParseRequestSetCrop(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestSetCrop: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if !h.manorMgr.IsModifiable() {
		slog.Debug("set crop: manor not modifiable", "player", player.Name())
		return h.sendActionFailed(buf)
	}

	if !h.isManorOwner(player, pkt.ManorID) {
		slog.Debug("set crop: not manor owner",
			"player", player.Name(), "manorID", pkt.ManorID)
		return h.sendActionFailed(buf)
	}

	crops := make([]*manor.CropProcure, 0, len(pkt.Crops))
	for _, entry := range pkt.Crops {
		tmpl := data.GetSeedByCropID(entry.CropID)
		if tmpl == nil || tmpl.CastleID != pkt.ManorID {
			continue
		}

		if entry.StartAmount < 0 || entry.StartAmount > tmpl.LimitCrops {
			continue
		}

		minPrice := int32(data.CropMinPrice(entry.CropID))
		maxPrice := int32(data.CropMaxPrice(entry.CropID))
		if entry.Price < minPrice || entry.Price > maxPrice {
			continue
		}

		rewardType := int32(entry.RewardType)
		if rewardType != 1 && rewardType != 2 {
			rewardType = 1
		}

		crops = append(crops, manor.NewCropProcure(
			entry.CropID, entry.StartAmount, rewardType, entry.StartAmount, int64(entry.Price),
		))
	}

	h.manorMgr.SetNextCropProcure(pkt.ManorID, crops)

	slog.Info("manor crop procurement set",
		"player", player.Name(),
		"manorID", pkt.ManorID,
		"crops", len(crops))

	return 0, true, nil
}

// handleRequestExShowManorSeedInfo processes 0xD0:0x0C — display seeds for sale.
func (h *Handler) handleRequestExShowManorSeedInfo(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	if h.manorMgr == nil {
		return 0, true, nil
	}

	r := packet.NewReader(dataBytes)
	manorID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("reading manorID: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	seeds := h.manorMgr.SeedProduction(manorID, false)
	resp := &serverpackets.ExShowSeedInfo{
		ManorID:     manorID,
		HideButtons: false,
		Seeds:       seeds,
	}

	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExShowSeedInfo: %w", err)
	}

	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestExShowCropInfo processes 0xD0:0x0D — display crops for procure.
func (h *Handler) handleRequestExShowCropInfo(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	if h.manorMgr == nil {
		return 0, true, nil
	}

	r := packet.NewReader(dataBytes)
	manorID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("reading manorID: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	crops := h.manorMgr.CropProcureList(manorID, false)
	resp := &serverpackets.ExShowCropInfo{
		ManorID:     manorID,
		HideButtons: false,
		Crops:       crops,
	}

	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExShowCropInfo: %w", err)
	}

	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestExShowSeedSetting processes 0xD0:0x0E — display seed production settings.
func (h *Handler) handleRequestExShowSeedSetting(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	if h.manorMgr == nil {
		return 0, true, nil
	}

	r := packet.NewReader(dataBytes)
	manorID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("reading manorID: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	seeds := data.GetSeedsByCastle(manorID)
	resp := &serverpackets.ExShowSeedSetting{
		ManorID:       manorID,
		Seeds:         seeds,
		CurrentPeriod: h.manorMgr.SeedProduction(manorID, false),
		NextPeriod:    h.manorMgr.SeedProduction(manorID, true),
	}

	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExShowSeedSetting: %w", err)
	}

	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestExShowCropSetting processes 0xD0:0x0F — display crop procurement settings.
func (h *Handler) handleRequestExShowCropSetting(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	if h.manorMgr == nil {
		return 0, true, nil
	}

	r := packet.NewReader(dataBytes)
	manorID, err := r.ReadInt()
	if err != nil {
		return 0, false, fmt.Errorf("reading manorID: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	seeds := data.GetSeedsByCastle(manorID)
	resp := &serverpackets.ExShowCropSetting{
		ManorID:       manorID,
		Seeds:         seeds,
		CurrentPeriod: h.manorMgr.CropProcureList(manorID, false),
		NextPeriod:    h.manorMgr.CropProcureList(manorID, true),
	}

	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing ExShowCropSetting: %w", err)
	}

	n := copy(buf, respData)
	return n, true, nil
}

// isManorOwner checks if a player is the clan leader that owns a castle.
func (h *Handler) isManorOwner(player *model.Player, manorID int32) bool {
	clanID := player.ClanID()
	if clanID == 0 || h.clanTable == nil || h.siegeManager == nil {
		return false
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return false
	}

	// Проверяем, что игрок — лидер клана.
	if c.LeaderID() != int64(player.ObjectID()) {
		return false
	}

	// Проверяем, что клан владеет замком с данным manorID.
	castle := h.siegeManager.Castle(manorID)
	if castle == nil {
		return false
	}

	return castle.OwnerClanID() == clanID
}

// handleRequestOustFromPartyRoom processes 0xD0:0x00 — kick from party matching room.
// Stub: acknowledged until party room system is implemented.
func (h *Handler) handleRequestOustFromPartyRoom(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestDismissPartyRoom processes 0xD0:0x01 — dismiss party matching room.
// Stub: acknowledged until party room system is implemented.
func (h *Handler) handleRequestDismissPartyRoom(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestWithdrawPartyRoom processes 0xD0:0x02 — leave party matching room.
// Stub: acknowledged until party room system is implemented.
func (h *Handler) handleRequestWithdrawPartyRoom(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestListPartyMatchingWaitingRoom processes 0xD0:0x03 — party matching wait list.
// Stub: acknowledged until party room system is implemented.
func (h *Handler) handleRequestListPartyMatchingWaitingRoom(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestAskJoinPartyRoom processes 0xD0:0x14 — invite to party room.
// Stub: acknowledged until party room system is implemented.
func (h *Handler) handleRequestAskJoinPartyRoom(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleConfirmJoinPartyRoom processes 0xD0:0x15 — accept party room invite.
// Stub: acknowledged until party room system is implemented.
func (h *Handler) handleConfirmJoinPartyRoom(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestListPartyMatching processes 0xD0:0x16 — party matching search.
// Stub: acknowledged until party room system is implemented.
func (h *Handler) handleRequestListPartyMatching(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// broadcastClanUserInfo sends a UserInfo refresh to every online clan member.
// Used after crest changes so all members see the updated emblem.
func (h *Handler) broadcastClanUserInfo(c *clan.Clan) {
	if h.clientManager == nil {
		return
	}
	c.ForEachMember(func(m *clan.Member) bool {
		if !m.Online() {
			return true
		}
		objectID := uint32(m.PlayerID())
		cl := h.clientManager.GetClientByObjectID(objectID)
		if cl == nil {
			return true
		}
		p := cl.ActivePlayer()
		if p == nil {
			return true
		}
		userInfo := serverpackets.NewUserInfo(p)
		uiData, err := userInfo.Write()
		if err != nil {
			slog.Error("serializing UserInfo for crest broadcast", "error", err)
			return true
		}
		if err := h.clientManager.SendToPlayer(objectID, uiData, len(uiData)); err != nil {
			slog.Warn("send UserInfo for crest broadcast",
				"objectID", objectID, "error", err)
		}
		return true
	})
}
