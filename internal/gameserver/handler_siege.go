package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/game/siege"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
)

// handleRequestSiegeInfo processes RequestSiegeInfo packet (opcode 0x47).
// Sends SiegeInfo for the requested castle.
//
// Phase 21: Siege System.
func (h *Handler) handleRequestSiegeInfo(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSiegeInfo(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestSiegeInfo: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if h.siegeManager == nil {
		return 0, true, nil
	}

	castle := h.siegeManager.Castle(pkt.CastleID)
	if castle == nil {
		slog.Debug("siege info: castle not found", "castle_id", pkt.CastleID)
		return 0, true, nil
	}

	ownerName, leaderName, allyID, allyName := h.resolveCastleOwnerInfo(castle)

	canManage := false
	if castle.OwnerClanID() > 0 && player.ClanID() == castle.OwnerClanID() {
		c := h.clanTable.Clan(player.ClanID())
		if c != nil && c.LeaderID() == int64(player.ObjectID()) {
			canManage = true
		}
	}

	info := &serverpackets.SiegeInfo{
		CastleID:    castle.ID(),
		CanManage:   canManage,
		OwnerClanID: castle.OwnerClanID(),
		OwnerName:   ownerName,
		LeaderName:  leaderName,
		AllyID:      allyID,
		AllyName:    allyName,
		SiegeDate:   castle.SiegeDate(),
		TimeRegOver: castle.IsTimeRegistrationOver(),
	}

	pktData, err := info.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing SiegeInfo: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestSiegeAttackerList processes RequestSiegeAttackerList (opcode 0xA2).
// Sends the list of attacking clans.
//
// Phase 21: Siege System.
func (h *Handler) handleRequestSiegeAttackerList(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSiegeAttackerList(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestSiegeAttackerList: %w", err)
	}

	if client.ActivePlayer() == nil || h.siegeManager == nil {
		return 0, true, nil
	}

	castle := h.siegeManager.Castle(pkt.CastleID)
	if castle == nil {
		return 0, true, nil
	}

	s := castle.Siege()
	if s == nil {
		return 0, true, nil
	}

	attackers := s.AttackerClans()
	entries := make([]serverpackets.SiegeAttackerEntry, 0, len(attackers))
	for _, sc := range attackers {
		entry := h.buildAttackerEntry(sc)
		entries = append(entries, entry)
	}

	list := &serverpackets.SiegeAttackerList{
		CastleID:  castle.ID(),
		Attackers: entries,
	}

	pktData, err := list.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing SiegeAttackerList: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestSiegeDefenderList processes RequestSiegeDefenderList (opcode 0xA3).
// Sends the list of defending clans.
//
// Phase 21: Siege System.
func (h *Handler) handleRequestSiegeDefenderList(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestSiegeDefenderList(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestSiegeDefenderList: %w", err)
	}

	if client.ActivePlayer() == nil || h.siegeManager == nil {
		return 0, true, nil
	}

	castle := h.siegeManager.Castle(pkt.CastleID)
	if castle == nil {
		return 0, true, nil
	}

	s := castle.Siege()
	if s == nil {
		return 0, true, nil
	}

	defenders := s.DefenderClans()
	pending := s.PendingClans()
	entries := make([]serverpackets.SiegeDefenderEntry, 0, len(defenders)+len(pending))

	for _, sc := range defenders {
		entry := h.buildDefenderEntry(sc)
		entries = append(entries, entry)
	}
	for _, sc := range pending {
		entry := h.buildDefenderEntry(sc)
		entries = append(entries, entry)
	}

	list := &serverpackets.SiegeDefenderList{
		CastleID:  castle.ID(),
		Defenders: entries,
	}

	pktData, err := list.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing SiegeDefenderList: %w", err)
	}
	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestJoinSiege processes RequestJoinSiege (opcode 0xA4).
// Registers or unregisters a clan from a siege.
//
// Phase 21: Siege System.
func (h *Handler) handleRequestJoinSiege(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestJoinSiege(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestJoinSiege: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil || h.siegeManager == nil || h.clanTable == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 {
		slog.Debug("siege join: player not in clan", "player", player.Name())
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	// Только лидер клана может регистрироваться на осаду.
	if c.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("siege join: not clan leader", "player", player.Name())
		return 0, true, nil
	}

	if pkt.IsJoining {
		if pkt.IsAttacker {
			if err := h.siegeManager.RegisterAttacker(pkt.CastleID, clanID, c.Name(), c.Level()); err != nil {
				slog.Debug("siege register attacker failed",
					"castle_id", pkt.CastleID, "clan", c.Name(), "error", err)
				return 0, true, nil
			}
		} else {
			if err := h.siegeManager.RegisterDefender(pkt.CastleID, clanID, c.Name(), c.Level()); err != nil {
				slog.Debug("siege register defender failed",
					"castle_id", pkt.CastleID, "clan", c.Name(), "error", err)
				return 0, true, nil
			}
		}
	} else {
		if err := h.siegeManager.Unregister(pkt.CastleID, clanID); err != nil {
			slog.Debug("siege unregister failed",
				"castle_id", pkt.CastleID, "clan", c.Name(), "error", err)
		}
	}

	// Отправляем обновлённую информацию.
	return h.handleRequestSiegeInfo(context.Background(), client, data[:4], buf)
}

// handleRequestConfirmSiegeWaitingList processes RequestConfirmSiegeWaitingList (opcode 0xA5).
// Castle owner approves/rejects pending defender clans.
//
// Phase 21: Siege System.
func (h *Handler) handleRequestConfirmSiegeWaitingList(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestConfirmSiegeWaitingList(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestConfirmSiegeWaitingList: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil || h.siegeManager == nil || h.clanTable == nil {
		return 0, true, nil
	}

	castle := h.siegeManager.Castle(pkt.CastleID)
	if castle == nil {
		return 0, true, nil
	}

	// Только владелец замка может одобрять защитников.
	if player.ClanID() != castle.OwnerClanID() {
		slog.Debug("siege confirm: not castle owner",
			"player", player.Name(), "castle", castle.Name())
		return 0, true, nil
	}

	c := h.clanTable.Clan(player.ClanID())
	if c == nil || c.LeaderID() != int64(player.ObjectID()) {
		return 0, true, nil
	}

	s := castle.Siege()
	if s == nil {
		return 0, true, nil
	}

	if pkt.IsApproval {
		s.ApprovePendingDefender(pkt.ClanID)
		slog.Info("siege: defender approved",
			"castle", castle.Name(), "clan_id", pkt.ClanID)
	} else {
		s.RemoveClan(pkt.ClanID)
		slog.Info("siege: defender rejected",
			"castle", castle.Name(), "clan_id", pkt.ClanID)
	}

	return 0, true, nil
}

// resolveCastleOwnerInfo returns clan name, leader name, ally ID, and ally name for a castle owner.
func (h *Handler) resolveCastleOwnerInfo(castle *siege.Castle) (ownerName, leaderName string, allyID int32, allyName string) {
	if h.clanTable == nil || castle.OwnerClanID() == 0 {
		return "", "", 0, ""
	}
	c := h.clanTable.Clan(castle.OwnerClanID())
	if c == nil {
		return "", "", 0, ""
	}
	leader := c.Leader()
	if leader != nil {
		leaderName = leader.Name()
	}
	return c.Name(), leaderName, c.AllyID(), c.AllyName()
}

// buildAttackerEntry creates a SiegeAttackerEntry from a SiegeClan.
func (h *Handler) buildAttackerEntry(sc *siege.SiegeClan) serverpackets.SiegeAttackerEntry {
	entry := serverpackets.SiegeAttackerEntry{
		ClanID:   sc.ClanID,
		ClanName: sc.ClanName,
	}
	if h.clanTable != nil {
		c := h.clanTable.Clan(sc.ClanID)
		if c != nil {
			entry.CrestID = c.CrestID()
			entry.AllyID = c.AllyID()
			entry.AllyName = c.AllyName()
			entry.AllyCrestID = c.AllyCrestID()
			if leader := c.Leader(); leader != nil {
				entry.LeaderName = leader.Name()
			}
		}
	}
	return entry
}

// buildDefenderEntry creates a SiegeDefenderEntry from a SiegeClan.
func (h *Handler) buildDefenderEntry(sc *siege.SiegeClan) serverpackets.SiegeDefenderEntry {
	defType := serverpackets.DefenderTypeApproved
	switch sc.Type {
	case siege.ClanTypeOwner:
		defType = serverpackets.DefenderTypeOwner
	case siege.ClanTypeDefenderNotApproved:
		defType = serverpackets.DefenderTypePending
	case siege.ClanTypeDefender:
		defType = serverpackets.DefenderTypeApproved
	}

	entry := serverpackets.SiegeDefenderEntry{
		ClanID:   sc.ClanID,
		ClanName: sc.ClanName,
		Type:     int32(defType),
	}
	if h.clanTable != nil {
		c := h.clanTable.Clan(sc.ClanID)
		if c != nil {
			entry.CrestID = c.CrestID()
			entry.AllyID = c.AllyID()
			entry.AllyName = c.AllyName()
			entry.AllyCrestID = c.AllyCrestID()
			if leader := c.Leader(); leader != nil {
				entry.LeaderName = leader.Name()
			}
		}
	}
	return entry
}
