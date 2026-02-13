package quest

import (
	"testing"
)

// TestSmokeQuest_CollectArrowheads проверяет полный жизненный цикл квеста.
// Q00303: Minion Lukas просит собрать 10 Orc Arrowheads.
// NPC 30006 = Minion Lukas (start/complete)
// NPC 20006 = Orc Archers (kill target, drops Arrowhead item 963)
func TestSmokeQuest_CollectArrowheads(t *testing.T) {
	const (
		questID       int32 = 303
		questName           = "Q00303_CollectArrowheads"
		npcLukasID    int32 = 30006
		orcArcherID   int32 = 20006
		arrowheadItem int32 = 963
		requiredCount       = 10
	)

	// 1. Создаём квест
	q := NewQuest(questID, questName)
	q.AddQuestItem(arrowheadItem)

	// onTalk: Minion Lukas
	q.AddTalkID(npcLukasID, func(e *Event, qs *QuestState) string {
		switch qs.State() {
		case StateCreated:
			return "<html>Bring me 10 Orc Arrowheads.</html>"
		case StateStarted:
			if qs.GetCond() == 2 {
				// Все стрелы собраны
				return "<html>Excellent work! Here is your reward.</html>"
			}
			return "<html>You still need more arrowheads.</html>"
		case StateCompleted:
			return "<html>Thanks again for your help.</html>"
		}
		return ""
	})

	// onKill: Orc Archer
	q.AddKillID(orcArcherID, func(e *Event, qs *QuestState) string {
		if qs.GetCond() != 1 {
			return ""
		}
		kills := qs.GetCond() // просто для примера используем переменную
		_ = kills

		// Получаем текущее количество убийств
		var killCount int
		if v := qs.Get("kills"); v != "" {
			for _, c := range v {
				killCount = killCount*10 + int(c-'0')
			}
		}
		killCount++
		qs.Set("kills", intToString(killCount))

		if killCount >= requiredCount {
			qs.SetCond(2)
			return "quest_middle" // сигнал для PlaySound
		}
		return ""
	})

	// 2. Регистрируем квест в менеджере
	m := NewManager(nil)
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	// 3. Игрок начинает квест
	player := createTestPlayer(t, 1001)
	charID := player.CharacterID()

	qs, err := m.StartQuest(charID, questID)
	if err != nil {
		t.Fatalf("StartQuest: %v", err)
	}
	if !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
	if qs.GetCond() != 1 {
		t.Errorf("initial cond = %d, want 1", qs.GetCond())
	}

	// 4. Говорим с NPC — должен сказать "you need more"
	talkEvent := &Event{Type: EventTalk, Player: player, NpcID: npcLukasID}
	result := m.DispatchEvent(talkEvent)
	if result != "<html>You still need more arrowheads.</html>" {
		t.Errorf("talk result (in progress) = %q", result)
	}

	// 5. Убиваем 10 Orc Archers
	killEvent := &Event{Type: EventKill, Player: player, NpcID: orcArcherID}
	for range 9 {
		m.DispatchEvent(killEvent)
	}
	// 10-й убийства — квест переходит в cond=2
	killResult := m.DispatchEvent(killEvent)
	if killResult != "quest_middle" {
		t.Errorf("10th kill result = %q, want quest_middle", killResult)
	}

	// Проверяем состояние
	qs = m.GetQuestState(charID, questName)
	if qs.GetCond() != 2 {
		t.Errorf("cond after 10 kills = %d, want 2", qs.GetCond())
	}
	if qs.Get("kills") != "10" {
		t.Errorf("kills var = %q, want 10", qs.Get("kills"))
	}

	// 6. Возвращаемся к Lukas
	result = m.DispatchEvent(talkEvent)
	if result != "<html>Excellent work! Here is your reward.</html>" {
		t.Errorf("talk result (complete) = %q", result)
	}

	// 7. Завершаем квест
	if err := m.ExitQuest(charID, questName, true); err != nil {
		t.Fatalf("ExitQuest: %v", err)
	}

	qs = m.GetQuestState(charID, questName)
	if !qs.IsCompleted() {
		t.Error("quest should be completed")
	}

	// 8. Говорим снова — уже завершённый квест
	result = m.DispatchEvent(talkEvent)
	if result != "<html>Thanks again for your help.</html>" {
		t.Errorf("talk result (after complete) = %q", result)
	}
}

// TestSmokeQuest_AbandonQuest проверяет отмену квеста.
func TestSmokeQuest_AbandonQuest(t *testing.T) {
	m := NewManager(nil)

	q := NewQuest(1, "abandon_test")
	q.AddTalkID(100, func(e *Event, qs *QuestState) string {
		return "hello"
	})
	q.AddQuestItem(999)
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	charID := int64(1001)
	if _, err := m.StartQuest(charID, 1); err != nil {
		t.Fatalf("StartQuest: %v", err)
	}

	// Отменяем
	if err := m.ExitQuest(charID, "abandon_test", false); err != nil {
		t.Fatalf("ExitQuest: %v", err)
	}

	qs := m.GetQuestState(charID, "abandon_test")
	if qs != nil {
		t.Error("quest state should be nil after abandon")
	}

	// Активные квесты должны быть пустыми
	active := m.GetActiveQuests(charID)
	if len(active) != 0 {
		t.Errorf("active quests = %d, want 0", len(active))
	}
}
