package gameserver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/udisondev/la2go/internal/game/quest"
	"github.com/udisondev/la2go/internal/gameserver/clientpackets"
	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// handleNpcQuestBypass handles the "Quest" bypass command from NPC dialogs.
// Format: "Quest <questName>" or just "Quest" (show all quests for this NPC).
func (h *Handler) handleNpcQuestBypass(
	player *model.Player,
	npc *model.Npc,
	arg string,
	buf []byte,
) (int, bool, error) {
	if h.questManager == nil {
		slog.Debug("quest bypass ignored: quest manager not initialized")
		return 0, true, nil
	}

	npcID := npc.TemplateID()

	if arg == "" {
		// Показать список квестов для этого NPC
		return h.showQuestListForNPC(player, npc, npcID, buf)
	}

	// Разбор аргументов: "Q00257_TheGuardIsBusy" или "Q00257_TheGuardIsBusy 30039-03.htm"
	parts := strings.SplitN(arg, " ", 2)
	var eventStr string
	if len(parts) > 1 {
		eventStr = parts[1]
	}

	params := make(map[string]any, 1)
	if eventStr != "" {
		params["event"] = eventStr
	}

	// Конкретный квест: dispatch talk event
	event := &quest.Event{
		Type:     quest.EventTalk,
		Player:   player,
		NpcID:    npcID,
		TargetID: npc.ObjectID(),
		Params:   params,
	}

	htmlResult := h.questManager.DispatchEvent(event)
	if htmlResult == "" {
		return 0, true, nil
	}

	// Отправляем HTML ответ
	pkt := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()), htmlResult)
	data, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing quest html: %w", err)
	}

	n := copy(buf, data)
	return n, true, nil
}

// showQuestListForNPC shows available quests for an NPC.
func (h *Handler) showQuestListForNPC(
	player *model.Player,
	npc *model.Npc,
	npcID int32,
	buf []byte,
) (int, bool, error) {
	quests := h.questManager.GetQuestsForNPC(npcID)
	if len(quests) == 0 {
		return 0, true, nil
	}

	// Если один квест — сразу dispatch
	if len(quests) == 1 {
		event := &quest.Event{
			Type:     quest.EventTalk,
			Player:   player,
			NpcID:    npcID,
			TargetID: npc.ObjectID(),
		}
		htmlResult := h.questManager.DispatchEvent(event)
		if htmlResult == "" {
			return 0, true, nil
		}

		pkt := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()), htmlResult)
		data, err := pkt.Write()
		if err != nil {
			return 0, false, fmt.Errorf("serializing quest html: %w", err)
		}
		n := copy(buf, data)
		return n, true, nil
	}

	// Несколько квестов — показываем список
	htmlContent := "<html><body>Available Quests:<br>"
	for _, q := range quests {
		htmlContent += fmt.Sprintf(
			"<a action=\"bypass -h npc_%d_Quest %s\">%s</a><br>",
			npc.ObjectID(), q.Name(), q.Name(),
		)
	}
	htmlContent += "</body></html>"

	pkt := serverpackets.NewNpcHtmlMessage(int32(npc.ObjectID()), htmlContent)
	data, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing quest list html: %w", err)
	}
	n := copy(buf, data)
	return n, true, nil
}

// handleRequestQuestList processes C2S 0x63 — client opens quest journal.
// Java reference: RequestQuestList.java — empty packet, server replies with QuestList (0x80).
func (h *Handler) handleRequestQuestList(_ context.Context, client *GameClient, _, buf []byte) (int, bool, error) {
	player := client.ActivePlayer()
	if player == nil {
		return 0, false, fmt.Errorf("no active player for quest list")
	}

	entries := h.buildQuestListEntries(player)
	pkt := serverpackets.NewQuestListWithEntries(entries)
	pktData, err := pkt.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing QuestList: %w", err)
	}

	n := copy(buf, pktData)
	return n, true, nil
}

// handleRequestQuestAbort handles the RequestQuestAbort packet (C2S 0x64).
func (h *Handler) handleRequestQuestAbort(
	ctx context.Context,
	client *GameClient,
	data []byte,
	buf []byte,
) (int, bool, error) {
	pkt, err := clientpackets.ParseRequestQuestAbort(data)
	if err != nil {
		return 0, false, fmt.Errorf("parsing RequestQuestAbort: %w", err)
	}

	player := client.ActivePlayer()
	if player == nil {
		return 0, true, nil
	}

	if h.questManager == nil {
		return 0, true, nil
	}

	q := h.questManager.GetQuest(pkt.QuestID)
	if q == nil {
		slog.Warn("quest abort for unknown quest",
			"questID", pkt.QuestID,
			"character", player.Name())
		return 0, true, nil
	}

	charID := player.CharacterID()

	// Удаляем квестовые предметы
	h.questManager.RemoveQuestItems(player, q.Name())

	// Отменяем квест
	if err := h.questManager.ExitQuest(charID, q.Name(), false); err != nil {
		slog.Error("quest abort failed",
			"questID", pkt.QuestID,
			"questName", q.Name(),
			"character", player.Name(),
			"error", err)
		return 0, true, nil
	}

	// Отправляем обновлённый QuestList
	entries := h.buildQuestListEntries(player)
	ql := serverpackets.NewQuestListWithEntries(entries)
	qlData, err := ql.Write()
	if err != nil {
		return 0, false, fmt.Errorf("serializing QuestList after abort: %w", err)
	}

	totalBytes := copy(buf, qlData)

	// PlaySound: quest give up
	sound := serverpackets.NewPlaySound(serverpackets.SoundQuestGiveUp)
	soundData, err := sound.Write()
	if err != nil {
		return totalBytes, true, nil // не критично
	}
	n := copy(buf[totalBytes:], soundData)
	totalBytes += n

	slog.Info("quest aborted",
		"questID", pkt.QuestID,
		"questName", q.Name(),
		"character", player.Name())

	return totalBytes, true, nil
}

// buildQuestListEntries builds QuestEntry list for the player's active quests.
func (h *Handler) buildQuestListEntries(player *model.Player) []serverpackets.QuestEntry {
	if h.questManager == nil {
		return nil
	}

	activeQuests := h.questManager.GetActiveQuests(player.CharacterID())
	if len(activeQuests) == 0 {
		return nil
	}

	entries := make([]serverpackets.QuestEntry, 0, len(activeQuests))
	for _, qs := range activeQuests {
		entries = append(entries, serverpackets.QuestEntry{
			QuestID: qs.QuestID(),
			State:   int32(qs.State()),
		})
	}
	return entries
}
