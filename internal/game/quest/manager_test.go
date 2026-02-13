package quest

import (
	"context"
	"testing"
)

// mockRepo implements QuestRepository for testing.
type mockRepo struct {
	loaded  map[int64][]QuestVar
	saved   map[string]map[string]string // questName → vars
	deleted []string                     // quest names deleted
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		loaded: make(map[int64][]QuestVar),
		saved:  make(map[string]map[string]string),
	}
}

func (r *mockRepo) LoadByCharacterID(_ context.Context, charID int64) ([]QuestVar, error) {
	return r.loaded[charID], nil
}

func (r *mockRepo) SaveQuestState(_ context.Context, charID int64, questName string, vars map[string]string) error {
	r.saved[questName] = vars
	return nil
}

func (r *mockRepo) DeleteQuest(_ context.Context, charID int64, questName string) error {
	r.deleted = append(r.deleted, questName)
	return nil
}

func TestManager_RegisterQuest(t *testing.T) {
	m := NewManager(nil)

	q := NewQuest(303, "Q00303_CollectArrowheads")
	q.AddTalkID(30006, func(e *Event, qs *QuestState) string { return "" })

	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	if m.QuestCount() != 1 {
		t.Errorf("QuestCount() = %d, want 1", m.QuestCount())
	}

	// Проверяем получение по ID и имени
	got := m.GetQuest(303)
	if got == nil || got.ID() != 303 {
		t.Error("GetQuest(303) returned nil or wrong quest")
	}
	gotByName := m.GetQuestByName("Q00303_CollectArrowheads")
	if gotByName == nil || gotByName.ID() != 303 {
		t.Error("GetQuestByName returned nil or wrong quest")
	}
}

func TestManager_RegisterDuplicate(t *testing.T) {
	m := NewManager(nil)

	q1 := NewQuest(1, "quest1")
	q2 := NewQuest(1, "quest1_dup")

	if err := m.RegisterQuest(q1); err != nil {
		t.Fatalf("RegisterQuest q1: %v", err)
	}

	err := m.RegisterQuest(q2)
	if err == nil {
		t.Error("RegisterQuest duplicate ID should return error")
	}
}

func TestManager_StartQuest(t *testing.T) {
	m := NewManager(nil)

	q := NewQuest(1, "test_quest")
	q.AddTalkID(100, func(e *Event, qs *QuestState) string { return "" })
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	qs, err := m.StartQuest(1001, 1)
	if err != nil {
		t.Fatalf("StartQuest: %v", err)
	}

	if !qs.IsStarted() {
		t.Error("quest state should be Started")
	}
	if qs.GetCond() != 1 {
		t.Errorf("initial cond = %d, want 1", qs.GetCond())
	}

	// Повторный старт возвращает тот же state
	qs2, err := m.StartQuest(1001, 1)
	if err != nil {
		t.Fatalf("StartQuest again: %v", err)
	}
	if qs2 != qs {
		t.Error("second StartQuest should return same QuestState")
	}
}

func TestManager_StartQuest_UnregisteredQuest(t *testing.T) {
	m := NewManager(nil)

	_, err := m.StartQuest(1001, 999)
	if err == nil {
		t.Error("StartQuest for unregistered quest should return error")
	}
}

func TestManager_GetActiveQuests(t *testing.T) {
	m := NewManager(nil)

	q1 := NewQuest(1, "quest1")
	q2 := NewQuest(2, "quest2")
	if err := m.RegisterQuest(q1); err != nil {
		t.Fatalf("RegisterQuest q1: %v", err)
	}
	if err := m.RegisterQuest(q2); err != nil {
		t.Fatalf("RegisterQuest q2: %v", err)
	}

	// Нет квестов
	active := m.GetActiveQuests(1001)
	if len(active) != 0 {
		t.Errorf("active quests = %d, want 0", len(active))
	}

	// Начинаем два квеста
	if _, err := m.StartQuest(1001, 1); err != nil {
		t.Fatalf("StartQuest 1: %v", err)
	}
	if _, err := m.StartQuest(1001, 2); err != nil {
		t.Fatalf("StartQuest 2: %v", err)
	}

	active = m.GetActiveQuests(1001)
	if len(active) != 2 {
		t.Errorf("active quests = %d, want 2", len(active))
	}
}

func TestManager_ExitQuest_Abandon(t *testing.T) {
	repo := newMockRepo()
	m := NewManager(repo)

	q := NewQuest(1, "test_quest")
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	if _, err := m.StartQuest(1001, 1); err != nil {
		t.Fatalf("StartQuest: %v", err)
	}

	if err := m.ExitQuest(1001, "test_quest", false); err != nil {
		t.Fatalf("ExitQuest: %v", err)
	}

	// Квест должен быть удалён из памяти
	qs := m.GetQuestState(1001, "test_quest")
	if qs != nil {
		t.Error("quest state should be nil after abandon")
	}

	// Должен быть вызван DeleteQuest
	if len(repo.deleted) != 1 || repo.deleted[0] != "test_quest" {
		t.Errorf("repo.deleted = %v, want [test_quest]", repo.deleted)
	}
}

func TestManager_ExitQuest_Complete(t *testing.T) {
	m := NewManager(nil)

	q := NewQuest(1, "test_quest")
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	if _, err := m.StartQuest(1001, 1); err != nil {
		t.Fatalf("StartQuest: %v", err)
	}

	if err := m.ExitQuest(1001, "test_quest", true); err != nil {
		t.Fatalf("ExitQuest: %v", err)
	}

	qs := m.GetQuestState(1001, "test_quest")
	if qs == nil {
		t.Fatal("quest state should exist after completion")
	}
	if !qs.IsCompleted() {
		t.Error("quest should be completed")
	}
}

func TestManager_LoadPlayerQuests(t *testing.T) {
	repo := newMockRepo()
	repo.loaded[1001] = []QuestVar{
		{QuestName: "test_quest", Variable: "<state>", Value: "3"},
		{QuestName: "test_quest", Variable: "kills", Value: "5"},
	}

	m := NewManager(repo)

	q := NewQuest(1, "test_quest")
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	if err := m.LoadPlayerQuests(context.Background(), 1001); err != nil {
		t.Fatalf("LoadPlayerQuests: %v", err)
	}

	qs := m.GetQuestState(1001, "test_quest")
	if qs == nil {
		t.Fatal("quest state should be loaded")
	}
	if qs.State() != StateStarted {
		t.Errorf("state = %d, want %d (Started)", qs.State(), StateStarted)
	}
	if qs.Get("kills") != "5" {
		t.Errorf("kills = %q, want 5", qs.Get("kills"))
	}
}

func TestManager_SavePlayerQuests(t *testing.T) {
	repo := newMockRepo()
	m := NewManager(repo)

	q := NewQuest(1, "test_quest")
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	qs, err := m.StartQuest(1001, 1)
	if err != nil {
		t.Fatalf("StartQuest: %v", err)
	}
	qs.Set("kills", "10")

	if err := m.SavePlayerQuests(context.Background(), 1001); err != nil {
		t.Fatalf("SavePlayerQuests: %v", err)
	}

	saved, ok := repo.saved["test_quest"]
	if !ok {
		t.Fatal("quest state was not saved")
	}
	if saved["kills"] != "10" {
		t.Errorf("saved kills = %q, want 10", saved["kills"])
	}

	// changed flag должен быть сброшен
	if qs.IsChanged() {
		t.Error("IsChanged() should be false after save")
	}
}

func TestManager_UnloadPlayer(t *testing.T) {
	m := NewManager(nil)

	q := NewQuest(1, "test_quest")
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	if _, err := m.StartQuest(1001, 1); err != nil {
		t.Fatalf("StartQuest: %v", err)
	}

	m.UnloadPlayer(1001)

	qs := m.GetQuestState(1001, "test_quest")
	if qs != nil {
		t.Error("quest state should be nil after UnloadPlayer")
	}
}

func TestManager_DispatchEvent(t *testing.T) {
	m := NewManager(nil)

	talkCalled := false
	q := NewQuest(1, "test_quest")
	q.AddTalkID(30006, func(e *Event, qs *QuestState) string {
		talkCalled = true
		return "<html>Hello</html>"
	})

	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	// Создаём мок-игрока через startQuest, чтобы state был в памяти
	// (DispatchEvent работает даже без state — создаёт CREATED state)
	player := createTestPlayer(t, 1001)
	event := &Event{
		Type:   EventTalk,
		Player: player,
		NpcID:  30006,
	}

	result := m.DispatchEvent(event)
	if result != "<html>Hello</html>" {
		t.Errorf("DispatchEvent result = %q, want <html>Hello</html>", result)
	}
	if !talkCalled {
		t.Error("talk hook was not called")
	}
}

func TestManager_DispatchEvent_NoHook(t *testing.T) {
	m := NewManager(nil)

	q := NewQuest(1, "test_quest")
	q.AddTalkID(30006, func(e *Event, qs *QuestState) string { return "ok" })
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}

	player := createTestPlayer(t, 1001)
	event := &Event{
		Type:   EventTalk,
		Player: player,
		NpcID:  99999, // Нет хука для этого NPC
	}

	result := m.DispatchEvent(event)
	if result != "" {
		t.Errorf("DispatchEvent for unregistered NPC = %q, want empty", result)
	}
}

func TestManager_GetQuestsForNPC(t *testing.T) {
	m := NewManager(nil)

	q1 := NewQuest(1, "quest1")
	q1.AddTalkID(30006, func(e *Event, qs *QuestState) string { return "" })
	q2 := NewQuest(2, "quest2")
	q2.AddTalkID(30006, func(e *Event, qs *QuestState) string { return "" })

	if err := m.RegisterQuest(q1); err != nil {
		t.Fatalf("RegisterQuest q1: %v", err)
	}
	if err := m.RegisterQuest(q2); err != nil {
		t.Fatalf("RegisterQuest q2: %v", err)
	}

	quests := m.GetQuestsForNPC(30006)
	if len(quests) != 2 {
		t.Errorf("GetQuestsForNPC(30006) = %d quests, want 2", len(quests))
	}

	quests = m.GetQuestsForNPC(99999)
	if len(quests) != 0 {
		t.Errorf("GetQuestsForNPC(99999) = %d quests, want 0", len(quests))
	}
}
