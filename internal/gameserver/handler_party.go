package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleRequestJoinParty processes RequestJoinParty packet (opcode 0x29).
// Player invites another player to party.
//
// Flow:
//  1. Parse packet (target objectID + loot rule)
//  2. Validate: inviter must be in game, target must exist, not in party yet, etc.
//  3. Send AskJoinParty dialog to target
//  4. Store pending invite on target player
//
// Phase 7.3: Party System.
func (h *Handler) handleRequestJoinParty(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestJoinParty(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestJoinParty: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Cannot invite while already in a full party
	if party := player.GetParty(); party != nil {
		if party.MemberCount() >= model.MaxPartyMembers {
			slog.Debug("party full, cannot invite",
				"player", player.Name())
			return 0, true, nil
		}
	}

	// Find target player
	targetClient := h.clientManager.GetClientByObjectID(uint32(pkt.ObjectID))
	if targetClient == nil {
		slog.Debug("party invite target not found",
			"targetObjectID", pkt.ObjectID,
			"player", player.Name())
		return 0, true, nil
	}

	targetPlayer := targetClient.ActivePlayer()
	if targetPlayer == nil {
		return 0, true, nil
	}

	// Target already in a party
	if targetPlayer.IsInParty() {
		slog.Debug("target already in party",
			"target", targetPlayer.Name(),
			"player", player.Name())
		return 0, true, nil
	}

	// Target already has a pending invite
	if targetPlayer.PendingPartyInvite() != nil {
		slog.Debug("target already has pending invite",
			"target", targetPlayer.Name(),
			"player", player.Name())
		return 0, true, nil
	}

	// Store pending invite on target
	targetPlayer.SetPendingPartyInvite(&model.PartyInvite{
		FromObjectID: player.ObjectID(),
		FromName:     player.Name(),
		LootRule:     pkt.ItemDistribution,
	})

	// Send AskJoinParty to target
	askPkt := serverpackets.NewAskJoinParty(player.Name(), pkt.ItemDistribution)
	askData, err := askPkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing AskJoinParty: %w", err)
	}

	if err := h.clientManager.SendToPlayer(targetPlayer.ObjectID(), askData, len(askData)); err != nil {
		slog.Warn("failed to send AskJoinParty",
			"target", targetPlayer.Name(),
			"error", err)
	}

	slog.Debug("party invite sent",
		"from", player.Name(),
		"to", targetPlayer.Name(),
		"lootRule", pkt.ItemDistribution)

	return 0, true, nil
}

// handleAnswerJoinParty processes RequestAnswerJoinParty packet (opcode 0x43).
// Target player accepts or declines party invitation.
//
// Flow (accept):
//  1. Create party if inviter not in one yet
//  2. Add target to party
//  3. Send PartySmallWindowAll to new member
//  4. Send PartySmallWindowAdd to existing members
//  5. Send JoinParty(1) to inviter
//
// Flow (decline):
//  1. Send JoinParty(0) to inviter
//  2. Clear pending invite
//
// Phase 7.3: Party System.
func (h *Handler) handleAnswerJoinParty(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAnswerJoinParty(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestAnswerJoinParty: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	invite := player.PendingPartyInvite()
	if invite == nil {
		return 0, true, nil
	}
	player.ClearPendingPartyInvite()

	// Find inviter
	inviterClient := h.clientManager.GetClientByObjectID(invite.FromObjectID)
	if inviterClient == nil {
		return 0, true, nil
	}
	inviter := inviterClient.ActivePlayer()
	if inviter == nil {
		return 0, true, nil
	}

	if !pkt.IsAccepted() {
		// Declined — notify inviter
		joinPkt := serverpackets.NewJoinParty(0)
		joinData, err := joinPkt.Write()
		if err != nil {
			return 0, true, fmt.Errorf("writing JoinParty(decline): %w", err)
		}
		if err := h.clientManager.SendToPlayer(inviter.ObjectID(), joinData, len(joinData)); err != nil {
			slog.Warn("failed to send JoinParty decline", "error", err)
		}

		slog.Debug("party invite declined",
			"from", inviter.Name(),
			"by", player.Name())
		return 0, true, nil
	}

	// Accepted — create or join party
	partyObj := inviter.GetParty()
	if partyObj == nil {
		// Create new party with inviter as leader
		if h.partyManager == nil {
			return 0, true, fmt.Errorf("party manager not initialized")
		}
		partyObj = h.partyManager.CreateParty(inviter, invite.LootRule)
		inviter.SetParty(partyObj)

		// Send party window to inviter
		inviterWindow := serverpackets.NewPartySmallWindowAll(partyObj, inviter.ObjectID())
		inviterWindowData, err := inviterWindow.Write()
		if err != nil {
			slog.Error("failed to write PartySmallWindowAll for inviter", "error", err)
		} else {
			if err := h.clientManager.SendToPlayer(inviter.ObjectID(), inviterWindowData, len(inviterWindowData)); err != nil {
				slog.Warn("failed to send PartySmallWindowAll to inviter", "error", err)
			}
		}
	}

	// Add new member to party
	if err := partyObj.AddMember(player); err != nil {
		slog.Warn("failed to add member to party",
			"player", player.Name(),
			"error", err)
		return 0, true, nil
	}
	player.SetParty(partyObj)

	// Send JoinParty(accepted) to inviter
	joinPkt := serverpackets.NewJoinParty(1)
	joinData, err := joinPkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing JoinParty(accept): %w", err)
	}
	if err := h.clientManager.SendToPlayer(inviter.ObjectID(), joinData, len(joinData)); err != nil {
		slog.Warn("failed to send JoinParty accept", "error", err)
	}

	// Send PartySmallWindowAll to the new member (shows all existing members)
	windowAll := serverpackets.NewPartySmallWindowAll(partyObj, player.ObjectID())
	windowAllData, err := windowAll.Write()
	if err != nil {
		slog.Error("failed to write PartySmallWindowAll", "error", err)
	} else {
		if err := h.clientManager.SendToPlayer(player.ObjectID(), windowAllData, len(windowAllData)); err != nil {
			slog.Warn("failed to send PartySmallWindowAll to new member", "error", err)
		}
	}

	// Send PartySmallWindowAdd to all existing members (new member info)
	addPkt := serverpackets.NewPartySmallWindowAdd(partyObj, player)
	addData, err := addPkt.Write()
	if err != nil {
		slog.Error("failed to write PartySmallWindowAdd", "error", err)
	} else {
		for _, member := range partyObj.Members() {
			if member.ObjectID() == player.ObjectID() {
				continue
			}
			if err := h.clientManager.SendToPlayer(member.ObjectID(), addData, len(addData)); err != nil {
				slog.Warn("failed to send PartySmallWindowAdd",
					"target", member.Name(),
					"error", err)
			}
		}
	}

	slog.Info("player joined party",
		"player", player.Name(),
		"leader", partyObj.Leader().Name(),
		"members", partyObj.MemberCount())

	return 0, true, nil
}

// handleWithdrawalParty processes RequestWithdrawalParty packet (opcode 0x2B).
// Player leaves their current party.
//
// Phase 7.3: Party System.
func (h *Handler) handleWithdrawalParty(_ context.Context, client *GameClient, _ []byte, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	partyObj := player.GetParty()
	if partyObj == nil {
		return 0, true, nil
	}

	h.removeFromParty(player, partyObj)
	return 0, true, nil
}

// handleOustPartyMember processes RequestOustPartyMember packet (opcode 0x2C).
// Party leader kicks a member by name.
//
// Phase 7.3: Party System.
func (h *Handler) handleOustPartyMember(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestOustPartyMember(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestOustPartyMember: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	partyObj := player.GetParty()
	if partyObj == nil {
		return 0, true, nil
	}

	// Only leader can kick
	if !partyObj.IsLeader(player.ObjectID()) {
		slog.Debug("non-leader tried to oust",
			"player", player.Name())
		return 0, true, nil
	}

	// Find target by name
	targetClient := h.clientManager.FindClientByPlayerName(pkt.PlayerName)
	if targetClient == nil {
		return 0, true, nil
	}
	targetPlayer := targetClient.ActivePlayer()
	if targetPlayer == nil {
		return 0, true, nil
	}

	// Cannot kick yourself
	if targetPlayer.ObjectID() == player.ObjectID() {
		return 0, true, nil
	}

	// Target must be in the same party
	if !partyObj.IsMember(targetPlayer.ObjectID()) {
		return 0, true, nil
	}

	h.removeFromParty(targetPlayer, partyObj)
	return 0, true, nil
}

// removeFromParty handles player removal from party with all necessary notifications.
func (h *Handler) removeFromParty(player *model.Player, partyObj *model.Party) {
	shouldDisband := partyObj.RemoveMember(player.ObjectID())
	player.SetParty(nil)

	// Send delete notification to remaining members
	delPkt := serverpackets.NewPartySmallWindowDelete(player.ObjectID(), player.Name())
	delData, err := delPkt.Write()
	if err != nil {
		slog.Error("failed to write PartySmallWindowDelete", "error", err)
	}

	if shouldDisband {
		// Party disbanded — notify remaining member (if any) and clean up
		for _, member := range partyObj.Members() {
			member.SetParty(nil)

			// Send DeleteAll to disbanded members
			delAllPkt := serverpackets.NewPartySmallWindowDeleteAll()
			delAllData, err := delAllPkt.Write()
			if err != nil {
				slog.Error("failed to write PartySmallWindowDeleteAll", "error", err)
				continue
			}
			if err := h.clientManager.SendToPlayer(member.ObjectID(), delAllData, len(delAllData)); err != nil {
				slog.Warn("failed to send PartySmallWindowDeleteAll",
					"target", member.Name(),
					"error", err)
			}
		}

		if h.partyManager != nil {
			h.partyManager.DisbandParty(partyObj.ID())
		}

		slog.Info("party disbanded",
			"partyID", partyObj.ID(),
			"reason", "not enough members")
	} else {
		// Notify remaining members about removal
		for _, member := range partyObj.Members() {
			if delData != nil {
				if err := h.clientManager.SendToPlayer(member.ObjectID(), delData, len(delData)); err != nil {
					slog.Warn("failed to send PartySmallWindowDelete",
						"target", member.Name(),
						"error", err)
				}
			}
		}

		slog.Info("player left party",
			"player", player.Name(),
			"partyID", partyObj.ID(),
			"remaining", partyObj.MemberCount())
	}
}
