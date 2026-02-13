package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
)

// Friend/Block system message IDs (Interlude).
// Java reference: SystemMessageId.java
const (
	sysMsgFriendNotFound     = 3    // "The user who requested to become friends is not found in the game."
	sysMsgAlreadyFriend      = 1375 // "This player is already registered in your friends list."
	sysMsgYouAreBusy         = 153  // "You are already busy."
	sysMsgFriendListHeader   = 314  // "Friends List"
	sysMsgFriendS1Online     = 1081 // "$s1 (Online)"
	sysMsgFriendS1Offline    = 1082 // "$s1 (Offline)"
	sysMsgFriendAddedS1      = 1083 // "$s1 has been added to your friends list."
	sysMsgFriendDeletedS1    = 1084 // "$s1 has been deleted from your friends list."
	sysMsgS1RequestedFriend  = 1085 // "$s1 has requested to become friends."
	sysMsgBlockListHeader    = 1107 // "Blocked list"
	sysMsgS1Blocked          = 1109 // "$s1 has been added to your ignore list."
	sysMsgS1Unblocked        = 1110 // "$s1 has been removed from your ignore list."
	sysMsgMsgRefusalOn       = 1111 // "Message refusal mode."
	sysMsgMsgRefusalOff      = 1112 // "Message acceptance mode."
	sysMsgS1InBlockList        = 1105 // "$s1 was already on your ignore list."
	sysMsgCannotBlockYourself  = 1106 // "You cannot put yourself on your own ignore list."
	sysMsgTargetInRefusalMode  = 176  // "That person is in message refusal mode."
	sysMsgTargetPlayerNotFound = 145  // "That player is not online."
)

// handleRequestFriendInvite sends a friend invite to another player (C2S 0x5E).
func (h *Handler) handleRequestFriendInvite(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestFriendInvite(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestFriendInvite: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Найти целевого игрока по имени
	targetClient := h.clientManager.GetClientByName(pkt.Name)
	if targetClient == nil {
		sm := serverpackets.NewSystemMessage(sysMsgFriendNotFound)
		smData, _ := sm.Write()
		client.Send(smData)
		return 0, true, nil
	}

	target := targetClient.ActivePlayer()
	if target == nil {
		sm := serverpackets.NewSystemMessage(sysMsgFriendNotFound)
		smData, _ := sm.Write()
		client.Send(smData)
		return 0, true, nil
	}

	// Нельзя добавить самого себя
	if target.ObjectID() == player.ObjectID() {
		return 0, true, nil
	}

	// Уже друзья
	if player.IsFriend(int32(target.ObjectID())) {
		sm := serverpackets.NewSystemMessage(sysMsgAlreadyFriend)
		smData, _ := sm.Write()
		client.Send(smData)
		return 0, true, nil
	}

	// Цель заблокировала нас
	if target.IsBlocked(int32(player.ObjectID())) {
		return 0, true, nil
	}

	// Цель занята (уже обрабатывает другой запрос)
	if target.IsProcessingTransaction() {
		sm := serverpackets.NewSystemMessage(sysMsgYouAreBusy)
		smData, _ := sm.Write()
		client.Send(smData)
		return 0, true, nil
	}

	// Устанавливаем транзакцию (10s timeout)
	player.OnTransactionRequest(target)

	// Отправляем FriendAddRequest целевому игроку
	reqPkt := serverpackets.NewFriendAddRequest(player.Name())
	reqData, err := reqPkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing FriendAddRequest: %w", err)
	}
	targetClient.Send(reqData)

	slog.Debug("friend invite sent",
		"from", player.Name(),
		"to", target.Name())

	return 0, true, nil
}

// handleRequestAnswerFriendInvite handles accept/decline of friend invite (C2S 0x5F).
func (h *Handler) handleRequestAnswerFriendInvite(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestAnswerFriendInvite(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestAnswerFriendInvite: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	requestor := player.ActiveRequester()
	if requestor == nil {
		return 0, true, nil
	}

	if pkt.Response == 1 {
		// ACCEPT: добавить обоих в друзья (в памяти)
		player.AddFriend(int32(requestor.ObjectID()))
		requestor.AddFriend(int32(player.ObjectID()))

		// Уведомление запросившему: L2Friend(Add)
		addToReq := serverpackets.NewL2FriendPacket(
			serverpackets.FriendActionAdd,
			player.Name(),
			true,
			int32(player.ObjectID()),
		)
		addReqData, _ := addToReq.Write()
		reqClient := h.clientManager.GetClientByObjectID(requestor.ObjectID())
		if reqClient != nil {
			reqClient.Send(addReqData)
		}

		// SystemMessage запросившему: "$s1 has been added to your friends list."
		smReq := serverpackets.NewSystemMessage(sysMsgFriendAddedS1).AddString(player.Name())
		smReqData, _ := smReq.Write()
		if reqClient != nil {
			reqClient.Send(smReqData)
		}

		// Уведомление принявшему: L2Friend(Add)
		addToPlayer := serverpackets.NewL2FriendPacket(
			serverpackets.FriendActionAdd,
			requestor.Name(),
			true,
			int32(requestor.ObjectID()),
		)
		addPlayerData, _ := addToPlayer.Write()
		client.Send(addPlayerData)

		// SystemMessage принявшему: "$s1 has been added to your friends list."
		smPlayer := serverpackets.NewSystemMessage(sysMsgFriendAddedS1).AddString(requestor.Name())
		smPlayerData, _ := smPlayer.Write()
		client.Send(smPlayerData)

		slog.Info("friend added",
			"player1", player.Name(),
			"player2", requestor.Name())
	}

	// Очистка состояния транзакции
	player.SetActiveRequester(nil)
	requestor.OnTransactionResponse()

	return 0, true, nil
}

// handleRequestFriendList sends the friend list via FriendList packet (C2S 0x60).
func (h *Handler) handleRequestFriendList(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	friendIDs := player.FriendList()
	infos := make([]serverpackets.FriendInfo, 0, len(friendIDs))

	for _, friendID := range friendIDs {
		friendClient := h.clientManager.GetClientByObjectID(uint32(friendID))
		isOnline := friendClient != nil && friendClient.ActivePlayer() != nil
		name := ""
		if isOnline {
			name = friendClient.ActivePlayer().Name()
		}
		infos = append(infos, serverpackets.FriendInfo{
			ObjectID: friendID,
			Name:     name,
			IsOnline: isOnline,
		})
	}

	pkt := serverpackets.NewFriendListPacket(infos)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing FriendList: %w", err)
	}
	client.Send(pktData)

	return 0, true, nil
}

// handleRequestFriendDel removes a friend from the list (C2S 0x61).
func (h *Handler) handleRequestFriendDel(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestFriendDel(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestFriendDel: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Поиск цели по имени
	targetClient := h.clientManager.GetClientByName(pkt.Name)
	var targetObjectID int32
	if targetClient != nil && targetClient.ActivePlayer() != nil {
		targetObjectID = int32(targetClient.ActivePlayer().ObjectID())
	}

	if targetObjectID == 0 || !player.IsFriend(targetObjectID) {
		// Игрок не в списке друзей
		return 0, true, nil
	}

	// Удаление из обоих списков друзей (двусторонняя связь)
	player.RemoveFriend(targetObjectID)

	if targetClient != nil && targetClient.ActivePlayer() != nil {
		targetClient.ActivePlayer().RemoveFriend(int32(player.ObjectID()))

		// Уведомляем цель: L2Friend(Remove)
		removePkt := serverpackets.NewL2FriendPacket(
			serverpackets.FriendActionRemove,
			player.Name(),
			true,
			int32(player.ObjectID()),
		)
		removeData, _ := removePkt.Write()
		targetClient.Send(removeData)
	}

	// Уведомляем инициатора: L2Friend(Remove) + SystemMessage
	selfRemovePkt := serverpackets.NewL2FriendPacket(
		serverpackets.FriendActionRemove,
		pkt.Name,
		false,
		targetObjectID,
	)
	selfData, _ := selfRemovePkt.Write()
	client.Send(selfData)

	sm := serverpackets.NewSystemMessage(sysMsgFriendDeletedS1).AddString(pkt.Name)
	smData, _ := sm.Write()
	client.Send(smData)

	slog.Info("friend removed",
		"player", player.Name(),
		"removed", pkt.Name)

	return 0, true, nil
}

// handleRequestBlock handles block/unblock/list/allblock/allunblock (C2S 0xA0).
func (h *Handler) handleRequestBlock(ctx context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestBlock(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestBlock: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	switch pkt.Type {
	case clientpackets.BlockTypeBlock:
		targetClient := h.clientManager.GetClientByName(pkt.Name)
		if targetClient == nil || targetClient.ActivePlayer() == nil {
			sm := serverpackets.NewSystemMessage(sysMsgFriendNotFound)
			smData, _ := sm.Write()
			client.Send(smData)
			return 0, true, nil
		}
		target := targetClient.ActivePlayer()
		targetID := int32(target.ObjectID())

		// Нельзя заблокировать себя
		if target.ObjectID() == player.ObjectID() {
			sm := serverpackets.NewSystemMessage(sysMsgCannotBlockYourself)
			smData, _ := sm.Write()
			client.Send(smData)
			return 0, true, nil
		}

		// Нельзя заблокировать друга
		if player.IsFriend(targetID) {
			return 0, true, nil
		}

		// Уже в блок-листе
		if player.IsBlocked(targetID) {
			sm := serverpackets.NewSystemMessage(sysMsgS1InBlockList).AddString(pkt.Name)
			smData, _ := sm.Write()
			client.Send(smData)
			return 0, true, nil
		}

		player.AddBlock(targetID)

		sm := serverpackets.NewSystemMessage(sysMsgS1Blocked).AddString(pkt.Name)
		smData, _ := sm.Write()
		client.Send(smData)

		slog.Debug("player blocked",
			"player", player.Name(),
			"blocked", pkt.Name)

	case clientpackets.BlockTypeUnblock:
		targetClient := h.clientManager.GetClientByName(pkt.Name)
		if targetClient == nil || targetClient.ActivePlayer() == nil {
			return 0, true, nil
		}
		targetID := int32(targetClient.ActivePlayer().ObjectID())

		if !player.IsBlocked(targetID) {
			return 0, true, nil
		}

		player.RemoveBlock(targetID)

		sm := serverpackets.NewSystemMessage(sysMsgS1Unblocked).AddString(pkt.Name)
		smData, _ := sm.Write()
		client.Send(smData)

		slog.Debug("player unblocked",
			"player", player.Name(),
			"unblocked", pkt.Name)

	case clientpackets.BlockTypeList:
		// Отправляем заголовок списка блокировки
		header := serverpackets.NewSystemMessage(sysMsgBlockListHeader)
		headerData, _ := header.Write()
		client.Send(headerData)

		// Каждый заблокированный как отдельное сообщение
		for _, blockedID := range player.BlockList() {
			blockedClient := h.clientManager.GetClientByObjectID(uint32(blockedID))
			name := ""
			if blockedClient != nil && blockedClient.ActivePlayer() != nil {
				name = blockedClient.ActivePlayer().Name()
			}
			if name == "" {
				continue
			}
			sm := serverpackets.NewSystemMessage(sysMsgS1InBlockList).AddString(name)
			smData, _ := sm.Write()
			client.Send(smData)
		}

	case clientpackets.BlockTypeAllBlock:
		player.SetMessageRefusal(true)
		sm := serverpackets.NewSystemMessage(sysMsgMsgRefusalOn)
		smData, _ := sm.Write()
		client.Send(smData)

		slog.Debug("message refusal enabled",
			"player", player.Name())

	case clientpackets.BlockTypeAllUnblock:
		player.SetMessageRefusal(false)
		sm := serverpackets.NewSystemMessage(sysMsgMsgRefusalOff)
		smData, _ := sm.Write()
		client.Send(smData)

		slog.Debug("message refusal disabled",
			"player", player.Name())
	}

	return 0, true, nil
}
