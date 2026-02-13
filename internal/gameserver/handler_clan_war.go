package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/gameserver/clan"
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
)

// Clan war system message IDs (Interlude).
// Java reference: SystemMessageId.java
const (
	// SysMsgClanWarDeclared — "$s1 has declared war on your clan." (1561)
	sysMsgClanWarDeclared int32 = 1561
	// SysMsgYouDeclaredClanWar — "You have declared war on $s1." (1562)
	sysMsgYouDeclaredClanWar int32 = 1562
	// SysMsgWarWithClanHasEnded — "War with $s1 has ended." (1567)
	sysMsgWarWithClanHasEnded int32 = 1567
	// SysMsgClanWarEndedAgainstUs — "$s1 has ended the war against your clan." (1568)
	sysMsgClanWarEndedAgainstUs int32 = 1568
	// SysMsgYouSurrendered — "You have surrendered to $s1." (1570)
	sysMsgYouSurrendered int32 = 1570
	// SysMsgEnemySurrendered — "$s1 has surrendered to your clan." (1571)
	sysMsgEnemySurrendered int32 = 1571
	// SysMsgYouCannotDeclareWar — "You cannot declare war because your clan level is below 3." (28)
	sysMsgClanLevelTooLow int32 = 28
	// SysMsgTitleChanged — "Your title has been changed." (2303)
	sysMsgTitleChanged int32 = 2303
)

// Reputation cost for war declaration or stop.
const warReputationCost int32 = 500

// handleRequestStartPledgeWar processes 0x4D -- declare clan war.
//
// Reads enemy clan name, validates the declaring player is clan leader
// with level >= 3, finds the enemy clan, declares war on both sides,
// and sends PledgeReceiveWarList to both clans.
//
// Phase 53: Clan War System.
func (h *Handler) handleRequestStartPledgeWar(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	r := packet.NewReader(data)
	pledgeName, err := r.ReadString()
	if err != nil {
		return 0, true, fmt.Errorf("reading pledge name: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return h.sendActionFailed(buf)
	}

	if h.clanTable == nil {
		return h.sendActionFailed(buf)
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return h.sendActionFailed(buf)
	}

	ourClan := h.clanTable.Clan(clanID)
	if ourClan == nil {
		return h.sendActionFailed(buf)
	}

	// Только лидер клана может объявить войну
	if ourClan.LeaderID() != int64(player.ObjectID()) {
		slog.Debug("non-leader tried to declare war",
			"player", player.Name(),
			"clan", ourClan.Name())
		return h.sendActionFailed(buf)
	}

	// Клан должен быть уровня 3+
	if ourClan.Level() < 3 {
		sm := serverpackets.NewSystemMessage(sysMsgClanLevelTooLow)
		return h.sendSystemMessageToBuf(sm, buf)
	}

	// Клан не должен быть в процессе роспуска
	if ourClan.IsDissolving() {
		return h.sendActionFailed(buf)
	}

	// Найти вражеский клан
	enemyClan := h.clanTable.ClanByName(pledgeName)
	if enemyClan == nil {
		slog.Debug("war target clan not found",
			"target", pledgeName,
			"player", player.Name())
		return h.sendActionFailed(buf)
	}

	// Нельзя объявить войну своему клану
	if enemyClan.ID() == ourClan.ID() {
		return h.sendActionFailed(buf)
	}

	// Вражеский клан тоже должен быть уровня 3+
	if enemyClan.Level() < 3 {
		return h.sendActionFailed(buf)
	}

	// Нельзя воевать с кланом из того же альянса
	if ourClan.AllyID() != 0 && ourClan.AllyID() == enemyClan.AllyID() {
		slog.Debug("cannot declare war on ally",
			"our_clan", ourClan.Name(),
			"enemy_clan", enemyClan.Name())
		return h.sendActionFailed(buf)
	}

	// Вражеский клан не должен быть в процессе роспуска
	if enemyClan.IsDissolving() {
		return h.sendActionFailed(buf)
	}

	// Объявляем войну
	if err := ourClan.DeclareWar(enemyClan.ID()); err != nil {
		slog.Debug("declare war",
			"our_clan", ourClan.Name(),
			"enemy_clan", enemyClan.Name(),
			"error", err)
		return h.sendActionFailed(buf)
	}

	// В Interlude война односторонняя: объявляющая сторона ставит в atWarWith,
	// цель — в atWarAttackers
	enemyClan.AcceptWar(ourClan.ID())

	// Штраф репутации за объявление войны
	ourClan.AddReputation(-warReputationCost)

	slog.Info("clan war declared",
		"attacker", ourClan.Name(),
		"defender", enemyClan.Name())

	// Отправляем системные сообщения
	// Нашему клану: "Вы объявили войну $s1"
	declaredMsg := serverpackets.NewSystemMessage(sysMsgYouDeclaredClanWar).AddString(enemyClan.Name())
	h.broadcastSystemMessageToClan(ourClan, declaredMsg)

	// Вражескому клану: "$s1 объявил войну вашему клану"
	attackedMsg := serverpackets.NewSystemMessage(sysMsgClanWarDeclared).AddString(ourClan.Name())
	h.broadcastSystemMessageToClan(enemyClan, attackedMsg)

	// Обновляем списки войн обоим кланам
	h.sendWarListsToClan(ourClan)
	h.sendWarListsToClan(enemyClan)

	return 0, true, nil
}

// handleRequestReplyStartPledgeWar processes 0x4E -- reply to war declaration.
// In Interlude war is unilateral (auto-accepted), so this is a no-op.
func (h *Handler) handleRequestReplyStartPledgeWar(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestStopPledgeWar processes 0x4F -- stop clan war.
//
// Only the declaring side can stop the war. Costs 500 reputation.
//
// Phase 53: Clan War System.
func (h *Handler) handleRequestStopPledgeWar(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	r := packet.NewReader(data)
	pledgeName, err := r.ReadString()
	if err != nil {
		return 0, true, fmt.Errorf("reading pledge name: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return h.sendActionFailed(buf)
	}

	if h.clanTable == nil {
		return h.sendActionFailed(buf)
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return h.sendActionFailed(buf)
	}

	ourClan := h.clanTable.Clan(clanID)
	if ourClan == nil {
		return h.sendActionFailed(buf)
	}

	// Только лидер
	if ourClan.LeaderID() != int64(player.ObjectID()) {
		return h.sendActionFailed(buf)
	}

	enemyClan := h.clanTable.ClanByName(pledgeName)
	if enemyClan == nil {
		return h.sendActionFailed(buf)
	}

	// Проверяем, что мы в войне с ними
	if !ourClan.IsAtWarWith(enemyClan.ID()) {
		return h.sendActionFailed(buf)
	}

	// Снимаем войну с обеих сторон
	ourClan.EndWar(enemyClan.ID())
	enemyClan.EndWar(ourClan.ID())

	// Штраф репутации за остановку войны
	ourClan.AddReputation(-warReputationCost)

	slog.Info("clan war stopped",
		"stopper", ourClan.Name(),
		"enemy", enemyClan.Name())

	// Системные сообщения
	endedMsg := serverpackets.NewSystemMessage(sysMsgWarWithClanHasEnded).AddString(enemyClan.Name())
	h.broadcastSystemMessageToClan(ourClan, endedMsg)

	enemyEndedMsg := serverpackets.NewSystemMessage(sysMsgClanWarEndedAgainstUs).AddString(ourClan.Name())
	h.broadcastSystemMessageToClan(enemyClan, enemyEndedMsg)

	// Обновляем списки войн
	h.sendWarListsToClan(ourClan)
	h.sendWarListsToClan(enemyClan)

	return 0, true, nil
}

// handleRequestReplyStopPledgeWar processes 0x50 -- reply to stop war.
// In Interlude this is a no-op.
func (h *Handler) handleRequestReplyStopPledgeWar(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestSurrenderPledgeWar processes 0x51 -- surrender in war.
//
// The declaring side gives up. Costs 500 reputation for surrendering clan,
// enemy clan gains 500 reputation.
//
// Phase 53: Clan War System.
func (h *Handler) handleRequestSurrenderPledgeWar(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	r := packet.NewReader(data)
	pledgeName, err := r.ReadString()
	if err != nil {
		return 0, true, fmt.Errorf("reading pledge name: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return h.sendActionFailed(buf)
	}

	if h.clanTable == nil {
		return h.sendActionFailed(buf)
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return h.sendActionFailed(buf)
	}

	ourClan := h.clanTable.Clan(clanID)
	if ourClan == nil {
		return h.sendActionFailed(buf)
	}

	// Только лидер
	if ourClan.LeaderID() != int64(player.ObjectID()) {
		return h.sendActionFailed(buf)
	}

	enemyClan := h.clanTable.ClanByName(pledgeName)
	if enemyClan == nil {
		return h.sendActionFailed(buf)
	}

	// Проверяем, что мы объявляли войну (atWarWith)
	if !ourClan.IsAtWarWith(enemyClan.ID()) {
		return h.sendActionFailed(buf)
	}

	// Завершаем войну с обеих сторон
	ourClan.EndWar(enemyClan.ID())
	enemyClan.EndWar(ourClan.ID())

	// Штраф за капитуляцию: наш клан теряет 500, враг получает 500
	ourClan.AddReputation(-warReputationCost)
	enemyClan.AddReputation(warReputationCost)

	slog.Info("clan war surrender",
		"surrendered", ourClan.Name(),
		"enemy", enemyClan.Name())

	// Системные сообщения
	surrenderMsg := serverpackets.NewSystemMessage(sysMsgYouSurrendered).AddString(enemyClan.Name())
	h.broadcastSystemMessageToClan(ourClan, surrenderMsg)

	enemySurrenderMsg := serverpackets.NewSystemMessage(sysMsgEnemySurrendered).AddString(ourClan.Name())
	h.broadcastSystemMessageToClan(enemyClan, enemySurrenderMsg)

	// Обновляем списки войн
	h.sendWarListsToClan(ourClan)
	h.sendWarListsToClan(enemyClan)

	return 0, true, nil
}

// handleRequestReplySurrenderPledgeWar processes 0x52 -- reply to surrender.
// In Interlude this is a no-op.
func (h *Handler) handleRequestReplySurrenderPledgeWar(_ context.Context, _ *GameClient, _ []byte, _ []byte) (int, bool, error) {
	return 0, true, nil
}

// handleRequestGiveNickName processes 0x55 -- set clan member title.
//
// Reads target player name and title. The requesting player must have
// CL_MANAGE_TITLES privilege (or be clan leader).
//
// Phase 53: Clan War System (moved from stubs).
func (h *Handler) handleRequestGiveNickName(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	r := packet.NewReader(data)
	targetName, err := r.ReadString()
	if err != nil {
		return 0, true, fmt.Errorf("reading target name: %w", err)
	}

	title, err := r.ReadString()
	if err != nil {
		return 0, true, fmt.Errorf("reading title: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return h.sendActionFailed(buf)
	}

	if h.clanTable == nil {
		return h.sendActionFailed(buf)
	}

	clanID := player.ClanID()
	if clanID == 0 {
		return h.sendActionFailed(buf)
	}

	ourClan := h.clanTable.Clan(clanID)
	if ourClan == nil {
		return h.sendActionFailed(buf)
	}

	// Проверяем привилегию на управление титулами
	selfMember := ourClan.Member(int64(player.ObjectID()))
	if selfMember == nil {
		return h.sendActionFailed(buf)
	}

	isLeader := ourClan.LeaderID() == int64(player.ObjectID())
	if !isLeader && !selfMember.HasPrivilege(clan.PrivCLGiveTitles) {
		slog.Debug("no title privilege",
			"player", player.Name())
		return h.sendActionFailed(buf)
	}

	// Находим целевого мембера по имени
	targetMember := ourClan.MemberByName(targetName)
	if targetMember == nil {
		slog.Debug("give title target not found",
			"target", targetName,
			"player", player.Name())
		return h.sendActionFailed(buf)
	}

	// Устанавливаем титул мемберу
	targetMember.SetTitle(title)

	// Если онлайн, обновляем player model
	targetClient := h.clientManager.GetClientByObjectID(uint32(targetMember.PlayerID()))
	if targetClient != nil {
		if tp := targetClient.ActivePlayer(); tp != nil {
			tp.SetClanTitle(title)

			// Отправляем UserInfo целевому игроку для обновления
			userInfo := serverpackets.NewUserInfo(tp)
			uiData, writeErr := userInfo.Write()
			if writeErr != nil {
				slog.Error("serializing UserInfo for title change", "error", writeErr)
			} else if err := h.clientManager.SendToPlayer(tp.ObjectID(), uiData, len(uiData)); err != nil {
				slog.Debug("send UserInfo after title change",
					"target", targetName,
					"error", err)
			}
		}
	}

	slog.Debug("clan title set",
		"target", targetName,
		"title", title,
		"by", player.Name())

	return 0, true, nil
}

// sendSystemMessageToBuf serializes a SystemMessage and writes it to buf.
func (h *Handler) sendSystemMessageToBuf(sm *serverpackets.SystemMessage, buf []byte) (int, bool, error) {
	smData, err := sm.Write()
	if err != nil {
		return 0, true, fmt.Errorf("serializing system message: %w", err)
	}
	n := copy(buf, smData)
	return n, true, nil
}

// broadcastSystemMessageToClan sends a SystemMessage to all online members of a clan.
func (h *Handler) broadcastSystemMessageToClan(c *clan.Clan, sm *serverpackets.SystemMessage) {
	smData, err := sm.Write()
	if err != nil {
		slog.Error("serializing system message for clan broadcast", "error", err)
		return
	}
	h.broadcastToClan(c, smData)
}

// sendWarListsToClan sends both war list tabs (declared and attackers) to all online clan members.
func (h *Handler) sendWarListsToClan(c *clan.Clan) {
	// Tab 0 — войны, которые мы объявили
	warIDs := c.WarList()
	warEntries := make([]serverpackets.PledgeWarEntry, 0, len(warIDs))
	for _, enemyID := range warIDs {
		enemy := h.clanTable.Clan(enemyID)
		if enemy == nil {
			continue
		}
		warEntries = append(warEntries, serverpackets.PledgeWarEntry{
			ClanName: enemy.Name(),
		})
	}

	warPkt := &serverpackets.PledgeReceiveWarList{
		Tab:     0,
		Entries: warEntries,
	}
	warData, err := warPkt.Write()
	if err != nil {
		slog.Error("serializing PledgeReceiveWarList tab=0", "error", err)
	} else {
		h.broadcastToClan(c, warData)
	}

	// Tab 1 — войны, объявленные нам
	attackerIDs := c.AttackerList()
	attackEntries := make([]serverpackets.PledgeWarEntry, 0, len(attackerIDs))
	for _, enemyID := range attackerIDs {
		enemy := h.clanTable.Clan(enemyID)
		if enemy == nil {
			continue
		}
		attackEntries = append(attackEntries, serverpackets.PledgeWarEntry{
			ClanName: enemy.Name(),
		})
	}

	attackPkt := &serverpackets.PledgeReceiveWarList{
		Tab:     1,
		Entries: attackEntries,
	}
	attackData, err := attackPkt.Write()
	if err != nil {
		slog.Error("serializing PledgeReceiveWarList tab=1", "error", err)
	} else {
		h.broadcastToClan(c, attackData)
	}
}
