package gameserver

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/udisondev/la2go/internal/game/crest"
	"github.com/udisondev/la2go/internal/gameserver/clan"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// --- Phase 51: Alliance System ---

// handleRequestJoinAlly processes 0x82 — alliance leader invites another clan leader.
func (h *Handler) handleRequestJoinAlly(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestJoinAlly(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestJoinAlly: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for join ally")
	}

	if player.ClanID() == 0 {
		slog.Debug("join ally: player has no clan", "player", player.Name())
		return 0, true, nil
	}

	playerClan := h.clanTable.Clan(player.ClanID())
	if playerClan == nil {
		return 0, true, nil
	}

	// Must be alliance leader (clan leader + allyID == clanID)
	if !playerClan.IsAllyLeader() || playerClan.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("join ally: not alliance leader", "player", player.Name())
		return 0, true, nil
	}

	// Check penalty: cannot accept clans for 1 day after dismissing
	now := time.Now().UnixMilli()
	if playerClan.AllyPenaltyExpiryTime() > now && playerClan.AllyPenaltyType() == clan.AllyPenaltyDismissClan {
		slog.Debug("join ally: penalty active (dismiss clan)", "player", player.Name())
		return 0, true, nil
	}

	// Check alliance capacity
	if h.clanTable.ClanAllyCount(playerClan.AllyID()) >= clan.MaxClansInAlly {
		slog.Debug("join ally: alliance full", "player", player.Name())
		return 0, true, nil
	}

	// Find target player by object ID
	targetClient := h.clientManager.GetClientByObjectID(uint32(pkt.ObjectID))
	if targetClient == nil {
		slog.Debug("join ally: target client not found", "objectID", pkt.ObjectID)
		return 0, true, nil
	}

	targetActivePlayer := targetClient.ActivePlayer()
	if targetActivePlayer == nil {
		return 0, true, nil
	}

	if targetActivePlayer.ClanID() == 0 {
		slog.Debug("join ally: target has no clan", "target", targetActivePlayer.Name())
		return 0, true, nil
	}

	targetClan := h.clanTable.Clan(targetActivePlayer.ClanID())
	if targetClan == nil {
		return 0, true, nil
	}

	// Target must be clan leader
	if targetClan.LeaderID() != int64(targetActivePlayer.ObjectID()) {
		slog.Debug("join ally: target is not clan leader",
			"target", targetActivePlayer.Name())
		return 0, true, nil
	}

	// Target must not be in an alliance already
	if targetClan.AllyID() != 0 {
		slog.Debug("join ally: target clan already in alliance",
			"target", targetActivePlayer.Name(),
			"allyID", targetClan.AllyID())
		return 0, true, nil
	}

	// Check target penalty
	if targetClan.AllyPenaltyExpiryTime() > now {
		slog.Debug("join ally: target has penalty",
			"target", targetActivePlayer.Name(),
			"penaltyType", targetClan.AllyPenaltyType())
		return 0, true, nil
	}

	// Check if clans are at war
	if playerClan.IsAtWarWith(targetClan.ID()) {
		slog.Debug("join ally: clans are at war",
			"player", player.Name(),
			"target", targetActivePlayer.Name())
		return 0, true, nil
	}

	// Send AskJoinAlly to target
	askPkt := &serverpackets.AskJoinAlly{
		RequestorObjectID: int32(player.ObjectID()),
		AllyName:          playerClan.AllyName(),
	}
	askData, err := askPkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing AskJoinAlly: %w", err)
	}

	if err := h.clientManager.SendToPlayer(targetActivePlayer.ObjectID(), askData, len(askData)); err != nil {
		slog.Warn("join ally: failed to send AskJoinAlly",
			"target", targetActivePlayer.Name(), "error", err)
	}

	slog.Info("alliance invitation sent",
		"from", player.Name(),
		"to", targetActivePlayer.Name(),
		"allyName", playerClan.AllyName())

	return 0, true, nil
}

// handleRequestAnswerJoinAlly processes 0x83 — response to alliance invitation.
func (h *Handler) handleRequestAnswerJoinAlly(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAnswerJoinAlly(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestAnswerJoinAlly: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for answer join ally")
	}

	if pkt.Response == 0 {
		// Declined
		slog.Debug("alliance invitation declined", "player", player.Name())
		return 0, true, nil
	}

	// Accepted — find requestor's alliance and join
	if player.ClanID() == 0 {
		return 0, true, nil
	}

	playerClan := h.clanTable.Clan(player.ClanID())
	if playerClan == nil {
		return 0, true, nil
	}

	// Target clan must be clan leader
	if playerClan.LeaderID() != int64(player.ObjectID()) {
		return 0, true, nil
	}

	// Must not already be in an alliance
	if playerClan.AllyID() != 0 {
		return 0, true, nil
	}

	// NOTE: In a full implementation, we'd retrieve the requestor from a pending request system.
	// For now, log the acceptance. The requestor's alliance info would be applied here:
	// playerClan.SetAllyID(requestorClan.AllyID())
	// playerClan.SetAllyName(requestorClan.AllyName())
	// playerClan.SetAllyCrestID(requestorClan.AllyCrestID())
	// playerClan.SetAllyPenalty(0, clan.AllyPenaltyNone)

	slog.Info("alliance invitation accepted", "player", player.Name())

	return 0, true, nil
}

// handleAllyLeave processes 0x84 — clan leaves the alliance.
func (h *Handler) handleAllyLeave(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for ally leave")
	}

	if player.ClanID() == 0 {
		slog.Debug("ally leave: player has no clan", "player", player.Name())
		return 0, true, nil
	}

	playerClan := h.clanTable.Clan(player.ClanID())
	if playerClan == nil {
		return 0, true, nil
	}

	// Must be clan leader
	if playerClan.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("ally leave: not clan leader", "player", player.Name())
		return 0, true, nil
	}

	// Must be in an alliance
	if playerClan.AllyID() == 0 {
		slog.Debug("ally leave: not in alliance", "player", player.Name())
		return 0, true, nil
	}

	// Alliance leader cannot leave (must dissolve instead)
	if playerClan.IsAllyLeader() {
		slog.Debug("ally leave: alliance leader cannot leave", "player", player.Name())
		return 0, true, nil
	}

	allyName := playerClan.AllyName()

	// Clear alliance fields
	playerClan.ClearAlly()

	// Set penalty: cannot join another alliance for 1 day
	now := time.Now().UnixMilli()
	playerClan.SetAllyPenalty(now+86_400_000, clan.AllyPenaltyClanLeaved)

	slog.Info("clan left alliance",
		"clan", playerClan.Name(),
		"alliance", allyName,
		"player", player.Name())

	return 0, true, nil
}

// handleAllyDismiss processes 0x85 — alliance leader removes a clan from the alliance.
func (h *Handler) handleAllyDismiss(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseAllyDismiss(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing AllyDismiss: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for ally dismiss")
	}

	if player.ClanID() == 0 {
		return 0, true, nil
	}

	leaderClan := h.clanTable.Clan(player.ClanID())
	if leaderClan == nil {
		return 0, true, nil
	}

	// Must be alliance leader
	if !leaderClan.IsAllyLeader() || leaderClan.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("ally dismiss: not alliance leader", "player", player.Name())
		return 0, true, nil
	}

	// Find target clan by name
	targetClan := h.clanTable.ClanByName(pkt.ClanName)
	if targetClan == nil {
		slog.Debug("ally dismiss: clan not found", "clanName", pkt.ClanName)
		return 0, true, nil
	}

	// Cannot dismiss self
	if targetClan.ID() == leaderClan.ID() {
		slog.Debug("ally dismiss: cannot dismiss own clan", "player", player.Name())
		return 0, true, nil
	}

	// Target must be in same alliance
	if targetClan.AllyID() != leaderClan.AllyID() {
		slog.Debug("ally dismiss: clan not in same alliance",
			"target", pkt.ClanName,
			"targetAllyID", targetClan.AllyID(),
			"leaderAllyID", leaderClan.AllyID())
		return 0, true, nil
	}

	now := time.Now().UnixMilli()

	// Penalty for alliance leader: cannot accept new clans for 1 day
	leaderClan.SetAllyPenalty(now+86_400_000, clan.AllyPenaltyDismissClan)

	// Clear alliance from dismissed clan + penalty: cannot join for 1 day
	targetClan.ClearAlly()
	targetClan.SetAllyPenalty(now+86_400_000, clan.AllyPenaltyClanDismissed)

	slog.Info("clan dismissed from alliance",
		"dismissedClan", targetClan.Name(),
		"alliance", leaderClan.AllyName(),
		"by", player.Name())

	return 0, true, nil
}

// handleRequestDismissAlly processes 0x86 — dissolve the entire alliance.
func (h *Handler) handleRequestDismissAlly(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for dismiss ally")
	}

	if player.ClanID() == 0 {
		return 0, true, nil
	}

	leaderClan := h.clanTable.Clan(player.ClanID())
	if leaderClan == nil {
		return 0, true, nil
	}

	// Must be alliance leader
	if !leaderClan.IsAllyLeader() || leaderClan.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("dismiss ally: not alliance leader", "player", player.Name())
		return 0, true, nil
	}

	allyID := leaderClan.AllyID()
	allyName := leaderClan.AllyName()

	// Clear alliance from all member clans
	for _, c := range h.clanTable.ClanAllies(allyID) {
		if c.ID() != leaderClan.ID() {
			c.ClearAlly()
			c.SetAllyPenalty(0, clan.AllyPenaltyNone)
		}
	}

	now := time.Now().UnixMilli()

	// Clear alliance from leader + penalty: cannot create new alliance for 1 day
	leaderClan.ClearAlly()
	leaderClan.SetAllyPenalty(now+86_400_000, clan.AllyPenaltyDissolveAlly)

	slog.Info("alliance dissolved",
		"alliance", allyName,
		"allyID", allyID,
		"by", player.Name())

	return 0, true, nil
}

// handleRequestSetAllyCrest processes 0x87 — upload alliance crest.
func (h *Handler) handleRequestSetAllyCrest(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSetAllyCrest(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestSetAllyCrest: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for set ally crest")
	}

	if player.ClanID() == 0 {
		return 0, true, nil
	}

	playerClan := h.clanTable.Clan(player.ClanID())
	if playerClan == nil {
		return 0, true, nil
	}

	// Must be alliance leader
	if !playerClan.IsAllyLeader() || playerClan.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("set ally crest: not alliance leader", "player", player.Name())
		return 0, true, nil
	}

	// Validate crest size (max 192 bytes = 8x12 BMP)
	if pkt.Length > 192 {
		slog.Warn("set ally crest: crest too large",
			"player", player.Name(),
			"size", pkt.Length)
		return 0, true, nil
	}

	allyID := playerClan.AllyID()

	if pkt.Length == 0 {
		// Удалить герб из crestTable и сбросить у всех кланов альянса.
		oldID := playerClan.AllyCrestID()
		if oldID != 0 {
			h.crestTbl.RemoveCrest(oldID)
		}
		for _, c := range h.clanTable.ClanAllies(allyID) {
			c.SetAllyCrestID(0)
		}
		slog.Info("alliance crest removed",
			"alliance", playerClan.AllyName(),
			"by", player.Name())
	} else {
		newCrest, err := h.crestTbl.CreateCrest(pkt.Data, crest.Ally)
		if err != nil {
			slog.Warn("set ally crest: create crest",
				"player", player.Name(), "error", err)
			return 0, true, nil
		}

		// Удаляем старый герб если был.
		oldID := playerClan.AllyCrestID()
		if oldID != 0 {
			h.crestTbl.RemoveCrest(oldID)
		}

		for _, c := range h.clanTable.ClanAllies(allyID) {
			c.SetAllyCrestID(newCrest.ID())
		}
		slog.Info("alliance crest set",
			"alliance", playerClan.AllyName(),
			"crestID", newCrest.ID(),
			"size", len(pkt.Data),
			"by", player.Name())
	}

	return 0, true, nil
}

// handleNpcCreateAlly handles "create_ally" NPC bypass — creates a new alliance.
// Called from NPC bypass routing when player interacts with VillageMaster.
func (h *Handler) handleNpcCreateAlly(player *model.Player, allyName string, buf []byte) (int, bool, error) {
	if player == nil || player.ClanID() == 0 {
		return 0, true, nil
	}

	playerClan := h.clanTable.Clan(player.ClanID())
	if playerClan == nil {
		return 0, true, nil
	}

	// Must be clan leader
	if playerClan.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("create ally: not clan leader", "player", player.Name())
		return 0, true, nil
	}

	// Must not already be in an alliance
	if playerClan.AllyID() != 0 {
		slog.Debug("create ally: already in alliance", "player", player.Name())
		return 0, true, nil
	}

	// Clan level must be >= 5
	if playerClan.Level() < 5 {
		slog.Debug("create ally: clan level too low",
			"player", player.Name(),
			"level", playerClan.Level())
		return 0, true, nil
	}

	// Check dissolve penalty
	now := time.Now().UnixMilli()
	if playerClan.AllyPenaltyExpiryTime() > now && playerClan.AllyPenaltyType() == clan.AllyPenaltyDissolveAlly {
		slog.Debug("create ally: dissolve penalty active", "player", player.Name())
		return 0, true, nil
	}

	// Validate name
	if len(allyName) < clan.MinAllyNameLen || len(allyName) > clan.MaxAllyNameLen {
		slog.Debug("create ally: invalid name length",
			"name", allyName,
			"len", len(allyName))
		return 0, true, nil
	}

	// Check name uniqueness
	if h.clanTable.AllyExists(allyName) {
		slog.Debug("create ally: name already taken", "name", allyName)
		return 0, true, nil
	}

	// Create alliance: allyID = clanID of the leader
	playerClan.SetAllyID(playerClan.ID())
	playerClan.SetAllyName(allyName)
	playerClan.SetAllyPenalty(0, clan.AllyPenaltyNone)

	slog.Info("alliance created",
		"alliance", allyName,
		"clanID", playerClan.ID(),
		"leader", player.Name())

	return 0, true, nil
}
