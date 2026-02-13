package quest

import (
	"testing"
)

func TestNewQuest(t *testing.T) {
	q := NewQuest(303, "Q00303_CollectArrowheads")

	if q.ID() != 303 {
		t.Errorf("ID() = %d, want 303", q.ID())
	}
	if q.Name() != "Q00303_CollectArrowheads" {
		t.Errorf("Name() = %q, want Q00303_CollectArrowheads", q.Name())
	}
}

func TestQuest_AddHooks(t *testing.T) {
	q := NewQuest(1, "test_quest")

	talkCalled := false
	killCalled := false

	q.AddTalkID(30006, func(e *Event, qs *QuestState) string {
		talkCalled = true
		return "talk_response"
	})
	q.AddKillID(20001, func(e *Event, qs *QuestState) string {
		killCalled = true
		return "kill_response"
	})

	// Проверяем наличие хуков
	if !q.HasHook(EventTalk, 30006) {
		t.Error("expected talk hook for NPC 30006")
	}
	if !q.HasHook(EventKill, 20001) {
		t.Error("expected kill hook for NPC 20001")
	}
	if q.HasHook(EventTalk, 99999) {
		t.Error("unexpected talk hook for NPC 99999")
	}
	if q.HasHook(EventAttack, 30006) {
		t.Error("unexpected attack hook for NPC 30006")
	}

	// Вызываем хуки
	qs := NewQuestState(1, "test_quest", 100, StateStarted)
	event := &Event{Type: EventTalk, NpcID: 30006}

	hook := q.GetHook(EventTalk, 30006)
	if hook == nil {
		t.Fatal("GetHook returned nil for registered talk hook")
	}
	result := hook(event, qs)
	if result != "talk_response" {
		t.Errorf("talk hook result = %q, want talk_response", result)
	}
	if !talkCalled {
		t.Error("talk hook was not called")
	}

	killHook := q.GetHook(EventKill, 20001)
	if killHook == nil {
		t.Fatal("GetHook returned nil for registered kill hook")
	}
	killResult := killHook(&Event{Type: EventKill, NpcID: 20001}, qs)
	if killResult != "kill_response" {
		t.Errorf("kill hook result = %q, want kill_response", killResult)
	}
	if !killCalled {
		t.Error("kill hook was not called")
	}
}

func TestQuest_RegisteredNPCs(t *testing.T) {
	q := NewQuest(1, "test")
	q.AddTalkID(100, func(e *Event, qs *QuestState) string { return "" })
	q.AddTalkID(200, func(e *Event, qs *QuestState) string { return "" })
	q.AddKillID(300, func(e *Event, qs *QuestState) string { return "" })

	talkNPCs := q.RegisteredNPCs(EventTalk)
	if len(talkNPCs) != 2 {
		t.Errorf("RegisteredNPCs(Talk) = %d NPCs, want 2", len(talkNPCs))
	}

	killNPCs := q.RegisteredNPCs(EventKill)
	if len(killNPCs) != 1 {
		t.Errorf("RegisteredNPCs(Kill) = %d NPCs, want 1", len(killNPCs))
	}

	attackNPCs := q.RegisteredNPCs(EventAttack)
	if len(attackNPCs) != 0 {
		t.Errorf("RegisteredNPCs(Attack) = %d NPCs, want 0", len(attackNPCs))
	}
}

func TestQuest_QuestItems(t *testing.T) {
	q := NewQuest(1, "test")
	q.AddQuestItem(1000)
	q.AddQuestItem(2000)

	items := q.QuestItems()
	if len(items) != 2 {
		t.Fatalf("QuestItems() = %d items, want 2", len(items))
	}
	if items[0] != 1000 || items[1] != 2000 {
		t.Errorf("QuestItems() = %v, want [1000, 2000]", items)
	}
}

func TestQuest_AllEventTypes(t *testing.T) {
	q := NewQuest(1, "test")
	nop := func(e *Event, qs *QuestState) string { return "ok" }

	q.AddTalkID(1, nop)
	q.AddFirstTalkID(2, nop)
	q.AddKillID(3, nop)
	q.AddAttackID(4, nop)
	q.AddSpawnID(5, nop)
	q.AddSkillSeeID(6, nop)
	q.AddAggroID(7, nop)
	q.SetOnItemUse(nop)
	q.SetOnEnterZone(nop)
	q.SetOnExitZone(nop)

	tests := []struct {
		eventType EventType
		npcID     int32
		wantHook  bool
	}{
		{EventTalk, 1, true},
		{EventFirstTalk, 2, true},
		{EventKill, 3, true},
		{EventAttack, 4, true},
		{EventSpawn, 5, true},
		{EventSkillSee, 6, true},
		{EventAggro, 7, true},
		{EventItemUse, 0, true},
		{EventEnterZone, 0, true},
		{EventExitZone, 0, true},
	}

	for _, tc := range tests {
		hook := q.GetHook(tc.eventType, tc.npcID)
		if (hook != nil) != tc.wantHook {
			t.Errorf("GetHook(%d, %d) = %v, want hook=%v", tc.eventType, tc.npcID, hook != nil, tc.wantHook)
		}
	}
}
