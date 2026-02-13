package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/gameserver/clan"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleRequestJoinPledge processes RequestJoinPledge packet (opcode 0x24).
// Player invites another player to their clan.
//
// Flow:
//  1. Validate inviter is in a clan and has invite privilege
//  2. Validate target exists, not already in a clan, no pending invite
//  3. Send AskJoinPledge dialog to target
//
// Phase 18: Clan System.
func (h *Handler) handleRequestJoinPledge(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestJoinPledge(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestJoinPledge: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if h.clanTable == nil {
		return 0, true, nil
	}

	// Inviter must be in a clan
	clanID := player.ClanID()
	if clanID == 0 {
		slog.Debug("player not in clan, cannot invite",
			"player", player.Name())
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	// Check invite privilege
	member := c.Member(int64(player.ObjectID()))
	if member == nil {
		return 0, true, nil
	}
	if !member.HasPrivilege(clan.PrivCLJoinClan) {
		slog.Debug("no invite privilege",
			"player", player.Name())
		return 0, true, nil
	}

	// Find target player
	targetClient := h.clientManager.GetClientByObjectID(uint32(pkt.ObjectID))
	if targetClient == nil {
		slog.Debug("clan invite target not found",
			"targetObjectID", pkt.ObjectID,
			"player", player.Name())
		return 0, true, nil
	}

	targetPlayer := targetClient.ActivePlayer()
	if targetPlayer == nil {
		return 0, true, nil
	}

	// Target must not be in a clan already
	if targetPlayer.ClanID() != 0 {
		slog.Debug("target already in clan",
			"target", targetPlayer.Name(),
			"player", player.Name())
		return 0, true, nil
	}

	// Target must not have a pending invite
	if targetPlayer.PendingClanInvite() != nil {
		slog.Debug("target already has pending clan invite",
			"target", targetPlayer.Name())
		return 0, true, nil
	}

	// Store pending invite
	targetPlayer.SetPendingClanInvite(&model.ClanInvite{
		ClanID:     c.ID(),
		ClanName:   c.Name(),
		InviterID:  player.ObjectID(),
		PledgeType: pkt.PledgeType,
	})

	// Send AskJoinPledge to target
	askPkt := serverpackets.NewAskJoinPledge(int32(player.ObjectID()), c.Name())
	askData, writeErr := askPkt.Write()
	if writeErr != nil {
		return 0, true, fmt.Errorf("writing AskJoinPledge: %w", writeErr)
	}

	if err := h.clientManager.SendToPlayer(targetPlayer.ObjectID(), askData, len(askData)); err != nil {
		slog.Warn("send AskJoinPledge",
			"target", targetPlayer.Name(),
			"error", err)
	}

	slog.Debug("clan invite sent",
		"from", player.Name(),
		"to", targetPlayer.Name(),
		"clan", c.Name())

	return 0, true, nil
}

// handleRequestAnswerJoinPledge processes RequestAnswerJoinPledge packet (opcode 0x25).
// Target player accepts or denies a clan invitation.
//
// Flow (accept):
//  1. Add player as ClanMember
//  2. Set player's ClanID
//  3. Send JoinPledge to new member
//  4. Send PledgeShowMemberListUpdate to all clan members
//
// Flow (deny):
//  1. Clear pending invite
//
// Phase 18: Clan System.
func (h *Handler) handleRequestAnswerJoinPledge(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAnswerJoinPledge(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestAnswerJoinPledge: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	invite := player.PendingClanInvite()
	if invite == nil {
		return 0, true, nil
	}
	player.ClearPendingClanInvite()

	if pkt.Answer != 1 {
		slog.Debug("clan invite declined",
			"player", player.Name(),
			"clan", invite.ClanName)
		return 0, true, nil
	}

	if h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(invite.ClanID)
	if c == nil {
		return 0, true, nil
	}

	// Create clan member
	newMember := clan.NewMember(
		int64(player.ObjectID()),
		player.Name(),
		player.Level(),
		player.ClassID(),
		invite.PledgeType,
		5, // default power grade for new member
	)
	newMember.SetOnline(true)

	if err := c.AddMember(newMember); err != nil {
		slog.Warn("add member to clan",
			"player", player.Name(),
			"clan", c.Name(),
			"error", err)
		return 0, true, nil
	}

	// Update player's clan
	player.SetClanID(c.ID())

	// Send JoinPledge to new member
	joinPkt := serverpackets.NewJoinPledge(c.ID())
	joinData, writeErr := joinPkt.Write()
	if writeErr != nil {
		return 0, true, fmt.Errorf("writing JoinPledge: %w", writeErr)
	}

	if err := h.clientManager.SendToPlayer(player.ObjectID(), joinData, len(joinData)); err != nil {
		slog.Warn("send JoinPledge", "error", err)
	}

	// Send PledgeShowMemberListUpdate to all clan members
	updatePkt := &serverpackets.PledgeShowMemberListUpdate{
		Name:    player.Name(),
		Level:   player.Level(),
		ClassID: player.ClassID(),
		Online:  1,
	}
	updateData, writeErr := updatePkt.Write()
	if writeErr != nil {
		slog.Error("write PledgeShowMemberListUpdate", "error", writeErr)
	} else {
		h.broadcastToClan(c, updateData)
	}

	slog.Info("player joined clan",
		"player", player.Name(),
		"clan", c.Name(),
		"members", c.MemberCount())

	return 0, true, nil
}

// handleRequestWithdrawalPledge processes RequestWithdrawalPledge packet (opcode 0x26).
// Player leaves their clan voluntarily.
//
// Phase 18: Clan System.
func (h *Handler) handleRequestWithdrawalPledge(_ context.Context, client *GameClient, _ []byte, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return 0, true, nil
	}

	if h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	// Leader cannot leave — must disband or transfer leadership
	if c.LeaderID() == int64(player.ObjectID()) {
		slog.Debug("clan leader cannot leave clan",
			"player", player.Name())
		return 0, true, nil
	}

	removed := c.RemoveMember(int64(player.ObjectID()))
	if removed == nil {
		return 0, true, nil
	}

	player.SetClanID(0)
	player.SetClanTitle("")

	// Notify remaining clan members
	delPkt := &serverpackets.PledgeShowMemberListDelete{Name: player.Name()}
	delData, writeErr := delPkt.Write()
	if writeErr != nil {
		slog.Error("write PledgeShowMemberListDelete", "error", writeErr)
	} else {
		h.broadcastToClan(c, delData)
	}

	slog.Info("player left clan",
		"player", player.Name(),
		"clan", c.Name())

	return 0, true, nil
}

// handleRequestOustPledgeMember processes RequestOustPledgeMember packet (opcode 0x27).
// Clan leader/officer kicks a member from the clan.
//
// Phase 18: Clan System.
func (h *Handler) handleRequestOustPledgeMember(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestOustPledgeMember(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestOustPledgeMember: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return 0, true, nil
	}

	if h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	// Check kick privilege
	inviterMember := c.Member(int64(player.ObjectID()))
	if inviterMember == nil {
		return 0, true, nil
	}
	if !inviterMember.HasPrivilege(clan.PrivCLDismiss) {
		slog.Debug("no dismiss privilege",
			"player", player.Name())
		return 0, true, nil
	}

	// Find target member by name
	targetMember := c.MemberByName(pkt.Name)
	if targetMember == nil {
		slog.Debug("oust target not found",
			"name", pkt.Name,
			"player", player.Name())
		return 0, true, nil
	}

	// Cannot kick the leader
	if targetMember.PlayerID() == c.LeaderID() {
		slog.Debug("cannot kick clan leader",
			"player", player.Name())
		return 0, true, nil
	}

	// Remove from clan
	removed := c.RemoveMember(targetMember.PlayerID())
	if removed == nil {
		return 0, true, nil
	}

	// Update target player if online
	targetClient := h.clientManager.GetClientByObjectID(uint32(targetMember.PlayerID()))
	if targetClient != nil {
		if tp := targetClient.ActivePlayer(); tp != nil {
			tp.SetClanID(0)
			tp.SetClanTitle("")
		}
	}

	// Notify remaining clan members
	delPkt := &serverpackets.PledgeShowMemberListDelete{Name: pkt.Name}
	delData, writeErr := delPkt.Write()
	if writeErr != nil {
		slog.Error("write PledgeShowMemberListDelete", "error", writeErr)
	} else {
		h.broadcastToClan(c, delData)
	}

	slog.Info("player ousted from clan",
		"kicked", pkt.Name,
		"by", player.Name(),
		"clan", c.Name())

	return 0, true, nil
}

// handleRequestPledgeInfo processes RequestPledgeInfo packet (opcode 0x3E).
// Client requests clan information (shown when clicking on clan name).
//
// Phase 18: Clan System.
func (h *Handler) handleRequestPledgeInfo(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPledgeInfo(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestPledgeInfo: %w", err)
	}

	if h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(pkt.ClanID)
	if c == nil {
		return 0, true, nil
	}

	infoPkt := serverpackets.NewPledgeInfo(c.ID(), c.Name(), c.AllyName())
	infoData, writeErr := infoPkt.Write()
	if writeErr != nil {
		return 0, true, fmt.Errorf("writing PledgeInfo: %w", writeErr)
	}

	n := copy(buf, infoData)
	return n, true, nil
}

// handleRequestPledgeMemberList processes RequestPledgeMemberList packet (opcode 0x4D).
// Client requests full clan member list.
//
// Phase 18: Clan System.
func (h *Handler) handleRequestPledgeMemberList(_ context.Context, client *GameClient, _ []byte, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return 0, true, nil
	}

	if h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	leader := c.Leader()
	leaderName := ""
	if leader != nil {
		leaderName = leader.Name()
	}

	members := c.Members()
	entries := make([]serverpackets.PledgeMemberEntry, 0, len(members))
	for _, m := range members {
		online := int32(0)
		if m.Online() {
			online = 1
		}
		entries = append(entries, serverpackets.PledgeMemberEntry{
			Name:       m.Name(),
			Level:      m.Level(),
			ClassID:    m.ClassID(),
			Online:     online,
			PledgeType: m.PledgeType(),
		})
	}

	listPkt := &serverpackets.PledgeShowMemberListAll{
		ClanName:   c.Name(),
		LeaderName: leaderName,
		CrestID:    c.CrestID(),
		ClanLevel:  c.Level(),
		Reputation: c.Reputation(),
		AllyID:     c.AllyID(),
		AllyName:   c.AllyName(),
		AllyCrest:  c.AllyCrestID(),
		Members:    entries,
	}

	listData, writeErr := listPkt.Write()
	if writeErr != nil {
		return 0, true, fmt.Errorf("writing PledgeShowMemberListAll: %w", writeErr)
	}

	n := copy(buf, listData)
	return n, true, nil
}

// handleRequestPledgeCrest processes RequestPledgeCrest packet (opcode 0x68).
// Client requests clan crest image data by crestID.
func (h *Handler) handleRequestPledgeCrest(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPledgeCrest(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestPledgeCrest: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	var crestData []byte
	if c := h.crestTbl.Crest(pkt.CrestID); c != nil {
		crestData = c.Data()
	}

	resp := &serverpackets.PledgeCrest{
		CrestID: pkt.CrestID,
		Data:    crestData,
	}
	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing PledgeCrest: %w", err)
	}
	n := copy(buf, respData)
	return n, true, nil
}

// handleRequestPledgeSetMemberPowerGrade processes opcode 0xCC.
// Clan leader/officer sets a member's rank.
//
// Phase 18: Clan System.
func (h *Handler) handleRequestPledgeSetMemberPowerGrade(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPledgeSetMemberPowerGrade(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestPledgeSetMemberPowerGrade: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 || h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	// Check privilege to manage ranks
	selfMember := c.Member(int64(player.ObjectID()))
	if selfMember == nil || !selfMember.HasPrivilege(clan.PrivCLManageRanks) {
		slog.Debug("no manage rank privilege",
			"player", player.Name())
		return 0, true, nil
	}

	targetMember := c.MemberByName(pkt.MemberName)
	if targetMember == nil {
		return 0, true, nil
	}

	// Cannot change leader's grade
	if targetMember.PlayerID() == c.LeaderID() {
		return 0, true, nil
	}

	// Validate grade (2-9, 1 is leader only)
	if pkt.PowerGrade < 2 || pkt.PowerGrade > 9 {
		return 0, true, nil
	}

	targetMember.SetPowerGrade(pkt.PowerGrade)

	slog.Debug("member power grade changed",
		"target", pkt.MemberName,
		"grade", pkt.PowerGrade,
		"by", player.Name())

	return 0, true, nil
}

// handleRequestPledgeReorganizeMember processes opcode 0xCD.
// Move a clan member to a different sub-pledge.
//
// Phase 18: Clan System.
func (h *Handler) handleRequestPledgeReorganizeMember(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPledgeReorganizeMember(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestPledgeReorganizeMember: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 || h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	// Only leader can reorganize
	if c.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("non-leader tried to reorganize",
			"player", player.Name())
		return 0, true, nil
	}

	targetMember := c.MemberByName(pkt.MemberName)
	if targetMember == nil {
		return 0, true, nil
	}

	// Cannot reorganize the leader
	if targetMember.PlayerID() == c.LeaderID() {
		return 0, true, nil
	}

	// Check sub-pledge capacity
	maxMembers := clan.MaxSubPledgeMembers(pkt.NewPledgeType)
	currentCount := c.SubPledgeMemberCount(pkt.NewPledgeType)
	if currentCount >= maxMembers {
		slog.Debug("sub-pledge full",
			"pledgeType", pkt.NewPledgeType,
			"count", currentCount)
		return 0, true, nil
	}

	targetMember.SetPledgeType(pkt.NewPledgeType)

	slog.Debug("member reorganized",
		"target", pkt.MemberName,
		"newPledge", pkt.NewPledgeType,
		"by", player.Name())

	return 0, true, nil
}

// handleRequestPledgePower processes opcode 0xCE.
// Set rank privileges for a power grade.
//
// Phase 18: Clan System.
func (h *Handler) handleRequestPledgePower(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPledgePower(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestPledgePower: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 || h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	// Only leader can change privileges
	if c.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("non-leader tried to set rank privileges",
			"player", player.Name())
		return 0, true, nil
	}

	// Validate grade (2-9)
	if pkt.PowerGrade < 2 || pkt.PowerGrade > 9 {
		return 0, true, nil
	}

	c.SetRankPrivileges(pkt.PowerGrade, clan.Privilege(pkt.Privileges))

	slog.Debug("rank privileges updated",
		"grade", pkt.PowerGrade,
		"privileges", pkt.Privileges,
		"by", player.Name())

	return 0, true, nil
}

// handleRequestPledgeWarList processes opcode 0xCF.
// Client requests list of clan wars.
//
// Phase 18: Clan System.
func (h *Handler) handleRequestPledgeWarList(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPledgeWarList(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestPledgeWarList: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	clanID := player.ClanID()
	if clanID == 0 || h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	// Page 0 = wars we declared, page 1 = wars declared on us
	var clanIDs []int32
	if pkt.Page == 0 {
		clanIDs = c.WarList()
	} else {
		clanIDs = c.AttackerList()
	}

	entries := make([]serverpackets.PledgeWarEntry, 0, len(clanIDs))
	for _, enemyID := range clanIDs {
		enemy := h.clanTable.Clan(enemyID)
		if enemy == nil {
			continue
		}
		entries = append(entries, serverpackets.PledgeWarEntry{
			ClanName: enemy.Name(),
		})
	}

	warPkt := &serverpackets.PledgeReceiveWarList{
		Tab:     pkt.Tab,
		Entries: entries,
	}
	warData, writeErr := warPkt.Write()
	if writeErr != nil {
		return 0, true, fmt.Errorf("writing PledgeReceiveWarList: %w", writeErr)
	}

	n := copy(buf, warData)
	return n, true, nil
}

// handleRequestPledgeMemberInfo processes extended opcode 0xD0:0x1D.
// Returns detailed info about a specific clan member.
//
// Phase 18: Clan System.
func (h *Handler) handleRequestPledgeMemberInfo(_ context.Context, client *GameClient, dataBytes, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestPledgeMemberInfo(dataBytes)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestPledgeMemberInfo: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for pledge member info")
	}

	clanID := player.ClanID()
	if clanID == 0 || h.clanTable == nil {
		return 0, true, nil
	}

	c := h.clanTable.Clan(clanID)
	if c == nil {
		return 0, true, nil
	}

	member := c.MemberByName(pkt.MemberName)
	if member == nil {
		slog.Debug("pledge member info: member not found",
			"player", player.Name(),
			"member", pkt.MemberName)
		return 0, true, nil
	}

	resp := &serverpackets.PledgeReceiveMemberInfo{
		PledgeType: member.PledgeType(),
		Name:       member.Name(),
		Title:      member.Title(),
		PowerGrade: member.PowerGrade(),
		Level:      member.Level(),
		ClassID:    member.ClassID(),
		Gender:     0, // default male, player gender not tracked in Member
		ObjectID:   int32(member.PlayerID()),
		Online:     member.Online(),
		SponsorID:  int32(member.SponsorID()),
	}

	respData, err := resp.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing PledgeReceiveMemberInfo: %w", err)
	}
	n := copy(buf, respData)
	return n, true, nil
}

// broadcastToClan sends a packet to all online clan members.
func (h *Handler) broadcastToClan(c *clan.Clan, payload []byte) {
	c.ForEachMember(func(m *clan.Member) bool {
		if !m.Online() {
			return true
		}
		if err := h.clientManager.SendToPlayer(uint32(m.PlayerID()), payload, len(payload)); err != nil {
			slog.Debug("broadcast to clan member",
				"member", m.Name(),
				"error", err)
		}
		return true
	})
}
