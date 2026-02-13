package gameserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/udisondev/la2go/internal/game/duel"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleRequestDuelStart processes RequestDuelStart (C2S 0xD0:0x1B).
// Player challenges another player (or party) to a duel.
//
// Flow:
//  1. Parse packet (target name + partyDuel flag)
//  2. Validate: both players in game, not dead, HP/MP ≥ 50%, not already duelling
//  3. Send ExDuelAskStart to target
//  4. Store pending duel request on target
//
// Phase 20: Duel System.
func (h *Handler) handleRequestDuelStart(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestDuelStart(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestDuelStart: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	// Проверяем, может ли challenger дуэлить
	if reason := duel.CanDuel(player); reason != "" {
		slog.Debug("duel request denied for requestor",
			"player", player.Name(), "reason", reason)
		return 0, true, nil
	}

	if h.duelManager.IsInDuel(player.ObjectID()) {
		slog.Debug("player already in duel",
			"player", player.Name())
		return 0, true, nil
	}

	// Ищем target по имени
	targetClient := h.clientManager.GetClientByName(pkt.Name)
	if targetClient == nil {
		slog.Debug("duel target not found",
			"target", pkt.Name,
			"player", player.Name())
		return 0, true, nil
	}

	targetPlayer := targetClient.ActivePlayer()
	if targetPlayer == nil {
		return 0, true, nil
	}

	// Проверяем target
	if reason := duel.CanDuel(targetPlayer); reason != "" {
		slog.Debug("duel request denied for target",
			"target", targetPlayer.Name(), "reason", reason)
		return 0, true, nil
	}

	if h.duelManager.IsInDuel(targetPlayer.ObjectID()) {
		slog.Debug("target already in duel",
			"target", targetPlayer.Name())
		return 0, true, nil
	}

	if targetPlayer.PendingDuelRequest() != nil {
		slog.Debug("target already has pending duel request",
			"target", targetPlayer.Name())
		return 0, true, nil
	}

	// Party duel: оба должны быть в пати и быть лидерами
	if pkt.PartyDuel {
		playerParty := player.GetParty()
		targetParty := targetPlayer.GetParty()

		if playerParty == nil || targetParty == nil {
			slog.Debug("party duel requires both in party",
				"player", player.Name(), "target", targetPlayer.Name())
			return 0, true, nil
		}

		if playerParty.Leader().ObjectID() != player.ObjectID() ||
			targetParty.Leader().ObjectID() != targetPlayer.ObjectID() {
			slog.Debug("party duel requires both to be party leaders",
				"player", player.Name(), "target", targetPlayer.Name())
			return 0, true, nil
		}
	}

	// Сохраняем pending request
	targetPlayer.SetPendingDuelRequest(&model.DuelRequest{
		RequestorID:   player.ObjectID(),
		RequestorName: player.Name(),
		PartyDuel:     pkt.PartyDuel,
	})

	// Отправляем ExDuelAskStart target'у
	askPkt := serverpackets.ExDuelAskStart{
		RequestorName: player.Name(),
		PartyDuel:     pkt.PartyDuel,
	}
	askData, err := askPkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ExDuelAskStart: %w", err)
	}

	if err := h.clientManager.SendToPlayer(targetPlayer.ObjectID(), askData, len(askData)); err != nil {
		slog.Warn("send ExDuelAskStart",
			"target", targetPlayer.Name(), "error", err)
	}

	slog.Debug("duel request sent",
		"from", player.Name(),
		"to", targetPlayer.Name(),
		"party", pkt.PartyDuel)

	return 0, true, nil
}

// handleRequestDuelAnswerStart processes RequestDuelAnswerStart (C2S 0xD0:0x1C).
// Target player accepts or declines the duel.
//
// Flow (accept):
//  1. Create duel via DuelManager
//  2. Send ExDuelReady to both players
//  3. Start duel lifecycle (countdown → fight → end)
//
// Flow (decline):
//  1. Clear pending duel request
//
// Phase 20: Duel System.
func (h *Handler) handleRequestDuelAnswerStart(_ context.Context, client *GameClient, data, buf []byte) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestDuelAnswerStart(data)
	if err != nil {
		return 0, true, fmt.Errorf("parsing RequestDuelAnswerStart: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	pending := player.PendingDuelRequest()
	if pending == nil {
		slog.Debug("no pending duel request",
			"player", player.Name())
		return 0, true, nil
	}
	player.ClearPendingDuelRequest()

	if !pkt.Accepted {
		slog.Debug("duel declined",
			"player", player.Name(),
			"requestor", pending.RequestorName)
		return 0, true, nil
	}

	// Ищем requestor
	requestorClient := h.clientManager.GetClientByObjectID(pending.RequestorID)
	if requestorClient == nil {
		slog.Debug("duel requestor disconnected",
			"requestorID", pending.RequestorID)
		return 0, true, nil
	}

	requestor := requestorClient.ActivePlayer()
	if requestor == nil {
		return 0, true, nil
	}

	// Повторно проверяем обоих
	if reason := duel.CanDuel(requestor); reason != "" {
		return 0, true, nil
	}
	if reason := duel.CanDuel(player); reason != "" {
		return 0, true, nil
	}

	// Создаём дуэль: requestor = playerA (challenger), player = playerB (opponent)
	d, err := h.duelManager.CreateDuel(requestor, player, pending.PartyDuel)
	if err != nil {
		slog.Warn("create duel", "error", err)
		return 0, true, nil
	}

	// Ставим duelID игрокам
	requestor.SetDuelID(d.ID())
	player.SetDuelID(d.ID())

	// Отправляем ExDuelReady обоим
	readyPkt := serverpackets.ExDuelReady{PartyDuel: pending.PartyDuel}
	readyData, err := readyPkt.Write()
	if err != nil {
		return 0, true, fmt.Errorf("writing ExDuelReady: %w", err)
	}
	h.sendToPlayer(requestor.ObjectID(), readyData)
	h.sendToPlayer(player.ObjectID(), readyData)

	// Запускаем lifecycle дуэли с коллбэками
	h.duelManager.StartDuel(d,
		// onCountdown: отправляем обновления HP/MP/CP при countdown
		func(d *duel.Duel, count int32) {
			slog.Debug("duel countdown", "duelID", d.ID(), "count", count)
		},
		// onStart: дуэль началась
		func(d *duel.Duel) {
			startPkt := serverpackets.ExDuelStart{PartyDuel: d.IsPartyDuel()}
			startData, err := startPkt.Write()
			if err != nil {
				slog.Error("writing ExDuelStart", "error", err)
				return
			}
			h.sendToPlayer(d.PlayerA().ObjectID(), startData)
			h.sendToPlayer(d.PlayerB().ObjectID(), startData)

			// Отправляем ExDuelUpdateUserInfo обоим
			h.sendDuelUserInfo(d.PlayerA(), d.PlayerB().ObjectID())
			h.sendDuelUserInfo(d.PlayerB(), d.PlayerA().ObjectID())
		},
		// onEnd: дуэль завершилась
		func(d *duel.Duel, result duel.Result) {
			endPkt := serverpackets.ExDuelEnd{PartyDuel: d.IsPartyDuel()}
			endData, err := endPkt.Write()
			if err != nil {
				slog.Error("writing ExDuelEnd", "error", err)
				return
			}
			h.sendToPlayer(d.PlayerA().ObjectID(), endData)
			h.sendToPlayer(d.PlayerB().ObjectID(), endData)

			// Восстанавливаем HP/MP/CP для нормального завершения
			abnormal := result == duel.ResultCanceled
			d.RestoreConditions(abnormal)

			// Сбрасываем duelID
			d.PlayerA().SetDuelID(0)
			d.PlayerB().SetDuelID(0)

			slog.Info("duel ended",
				"duelID", d.ID(),
				"playerA", d.PlayerA().Name(),
				"playerB", d.PlayerB().Name(),
				"result", result)
		},
	)

	slog.Info("duel accepted",
		"duelID", d.ID(),
		"playerA", requestor.Name(),
		"playerB", player.Name(),
		"party", pending.PartyDuel)

	return 0, true, nil
}

// handleRequestDuelSurrender processes RequestDuelSurrender (C2S 0xD0:0x1D).
// Player surrenders in the current duel.
//
// Phase 20: Duel System.
func (h *Handler) handleRequestDuelSurrender(_ context.Context, client *GameClient, _, _ []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	h.duelManager.OnSurrender(player.ObjectID())

	slog.Debug("duel surrender",
		"player", player.Name())

	return 0, true, nil
}

// sendDuelUserInfo sends ExDuelUpdateUserInfo about sourcePlayer to targetObjectID.
func (h *Handler) sendDuelUserInfo(source *model.Player, targetObjID uint32) {
	pkt := serverpackets.ExDuelUpdateUserInfo{
		ObjectID:  source.ObjectID(),
		Name:      source.Name(),
		CurrentHP: source.CurrentHP(),
		MaxHP:     source.MaxHP(),
		CurrentMP: source.CurrentMP(),
		MaxMP:     source.MaxMP(),
		CurrentCP: source.CurrentCP(),
		MaxCP:     source.MaxCP(),
	}
	pktData, err := pkt.Write()
	if err != nil {
		slog.Error("writing ExDuelUpdateUserInfo", "error", err)
		return
	}
	h.sendToPlayer(targetObjID, pktData)
}

// sendToPlayer sends a packet to a player, logging errors.
func (h *Handler) sendToPlayer(objectID uint32, data []byte) {
	if err := h.clientManager.SendToPlayer(objectID, data, len(data)); err != nil {
		slog.Warn("send to player",
			"objectID", objectID, "error", err)
	}
}
