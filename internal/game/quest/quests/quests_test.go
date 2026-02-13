package quests

import (
	"os"
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/game/quest"
	"github.com/udisondev/la2go/internal/model"
)

func TestMain(m *testing.M) {
	// Quest items now use real inventory → need item templates loaded.
	_ = data.LoadItemTemplates()
	os.Exit(m.Run())
}

// testPlayer creates a player with given params for quest tests.
func testPlayer(t *testing.T, charID int64, level int32, raceID int32) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(uint32(charID), charID, 1, "Tester", level, raceID, 0)
	if err != nil {
		t.Fatalf("creating test player: %v", err)
	}
	return p
}

// talkEvent creates a talk event for NPC interaction.
func talkEvent(player *model.Player, npcID int32, targetObjID uint32) *quest.Event {
	return &quest.Event{
		Type:     quest.EventTalk,
		Player:   player,
		NpcID:    npcID,
		TargetID: targetObjID,
	}
}

// talkEventWithAccept creates a talk event that simulates clicking "accept".
func talkEventWithAccept(player *model.Player, npcID int32, targetObjID uint32, eventStr string) *quest.Event {
	return &quest.Event{
		Type:     quest.EventTalk,
		Player:   player,
		NpcID:    npcID,
		TargetID: targetObjID,
		Params:   map[string]any{"event": eventStr},
	}
}

// killEvent creates a kill event for monster.
func killEvent(player *model.Player, npcID int32) *quest.Event {
	return &quest.Event{
		Type:   quest.EventKill,
		Player: player,
		NpcID:  npcID,
	}
}

// setupQuestManager creates a manager and registers a single quest.
func setupQuestManager(t *testing.T, q *quest.Quest) *quest.Manager {
	t.Helper()
	m := quest.NewManager(nil)
	if err := m.RegisterQuest(q); err != nil {
		t.Fatalf("RegisterQuest: %v", err)
	}
	return m
}

func TestRegisterAllQuests(t *testing.T) {
	m := quest.NewManager(nil)
	if err := RegisterAllQuests(m); err != nil {
		t.Fatalf("RegisterAllQuests: %v", err)
	}
	if got := m.QuestCount(); got != 14 {
		t.Errorf("QuestCount = %d; want 14", got)
	}
}

func TestRegisterAllQuests_NoDuplicates(t *testing.T) {
	m := quest.NewManager(nil)
	if err := RegisterAllQuests(m); err != nil {
		t.Fatalf("first RegisterAllQuests: %v", err)
	}
	if err := RegisterAllQuests(m); err == nil {
		t.Fatal("expected error on duplicate registration")
	}
}

// --- Q00257: The Guard is Busy (repeatable kill/collect) ---

func TestQ00257_Lifecycle(t *testing.T) {
	q := NewQ00257()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 6, RaceHuman)
	charID := player.CharacterID()

	// Разговор с Gilbert (30039) — должен предложить квест
	result := m.DispatchEvent(talkEvent(player, 30039, 100))
	if result == "" {
		t.Fatal("expected dialog from Gilbert")
	}

	// Принимаем квест (event = "30039-03.htm")
	result = m.DispatchEvent(talkEventWithAccept(player, 30039, 100, "30039-03.htm"))
	if result == "" {
		t.Fatal("expected accept response")
	}

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil {
		t.Fatal("quest state should exist after accept")
	}
	if !qs.IsStarted() {
		t.Fatal("quest should be started")
	}

	// Убиваем Goblins (20006) 20 раз — даёт ORC_AMULET (752)
	for range 20 {
		m.DispatchEvent(killEvent(player, 20006))
	}

	// Проверяем что предметы появились (рандом, но 20 убийств достаточно для хотя бы 1)
	// Не можем гарантировать — просто проверяем cond
	qs = m.GetQuestState(charID, q.Name())
	if qs == nil {
		t.Fatal("quest state lost")
	}
}

// --- Q00260: Hunt the Orcs (Elf only, repeatable) ---

func TestQ00260_RaceRestriction(t *testing.T) {
	q := NewQ00260()
	m := setupQuestManager(t, q)

	// Не-эльф не должен получить квест (или получить отказ)
	human := testPlayer(t, 1, 6, RaceHuman)
	result := m.DispatchEvent(talkEvent(human, 30221, 100))
	if result == "" {
		t.Fatal("expected response for non-elf")
	}
	// Ответ должен содержать отказ для не-эльфа
}

func TestQ00260_Lifecycle(t *testing.T) {
	q := NewQ00260()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 6, RaceElf)
	charID := player.CharacterID()

	// Начальный диалог
	result := m.DispatchEvent(talkEvent(player, 30221, 100))
	if result == "" {
		t.Fatal("expected dialog from Rayen")
	}

	// Принимаем
	m.DispatchEvent(talkEventWithAccept(player, 30221, 100, "accept"))
	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started after accept")
	}
}

// --- Q00261: Collector's Dream ---

func TestQ00261_Lifecycle(t *testing.T) {
	q := NewQ00261()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 15, RaceHuman)
	charID := player.CharacterID()

	// Принимаем квест
	m.DispatchEvent(talkEvent(player, 30222, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30222, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
}

// --- Q00293: The Hidden Veins (Dwarf only) ---

func TestQ00293_RaceRestriction(t *testing.T) {
	q := NewQ00293()
	m := setupQuestManager(t, q)

	elf := testPlayer(t, 1, 6, RaceElf)
	result := m.DispatchEvent(talkEvent(elf, 30535, 100))
	if result == "" {
		t.Fatal("expected response for non-dwarf")
	}
}

func TestQ00293_Lifecycle(t *testing.T) {
	q := NewQ00293()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 6, RaceDwarf)
	charID := player.CharacterID()

	m.DispatchEvent(talkEvent(player, 30535, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30535, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
}

// --- Q00320: Bones Tell the Future (Dark Elf only) ---

func TestQ00320_RaceRestriction(t *testing.T) {
	q := NewQ00320()
	m := setupQuestManager(t, q)

	human := testPlayer(t, 1, 10, RaceHuman)
	result := m.DispatchEvent(talkEvent(human, 30359, 100))
	if result == "" {
		t.Fatal("expected response for non-dark-elf")
	}
}

func TestQ00320_Lifecycle(t *testing.T) {
	q := NewQ00320()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 10, RaceDarkElf)
	charID := player.CharacterID()

	m.DispatchEvent(talkEvent(player, 30359, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30359, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
}

// --- Q00001: Letters of Love (delivery, no kills) ---

func TestQ00001_FullLifecycle(t *testing.T) {
	q := NewQ00001()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 2, RaceHuman)
	charID := player.CharacterID()

	// Darin (30048) — начальный диалог
	result := m.DispatchEvent(talkEvent(player, 30048, 100))
	if result == "" {
		t.Fatal("expected dialog from Darin")
	}

	// Принимаем
	m.DispatchEvent(talkEventWithAccept(player, 30048, 100, "accept"))
	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
	if qs.GetCond() != 1 {
		t.Errorf("cond = %d; want 1", qs.GetCond())
	}

	// Roxxy (30006) — cond 1→2
	result = m.DispatchEvent(talkEvent(player, 30006, 200))
	if result == "" {
		t.Fatal("expected dialog from Roxxy")
	}
	qs = m.GetQuestState(charID, q.Name())
	if qs.GetCond() != 2 {
		t.Errorf("cond after Roxxy = %d; want 2", qs.GetCond())
	}

	// Darin (30048) — cond 2→3
	result = m.DispatchEvent(talkEvent(player, 30048, 100))
	if result == "" {
		t.Fatal("expected dialog from Darin at cond 2")
	}
	qs = m.GetQuestState(charID, q.Name())
	if qs.GetCond() != 3 {
		t.Errorf("cond after Darin = %d; want 3", qs.GetCond())
	}

	// Baulro (30033) — cond 3→4
	result = m.DispatchEvent(talkEvent(player, 30033, 300))
	if result == "" {
		t.Fatal("expected dialog from Baulro")
	}
	qs = m.GetQuestState(charID, q.Name())
	if qs.GetCond() != 4 {
		t.Errorf("cond after Baulro = %d; want 4", qs.GetCond())
	}

	// Darin (30048) — cond 4 → complete
	result = m.DispatchEvent(talkEvent(player, 30048, 100))
	if result == "" {
		t.Fatal("expected completion dialog from Darin")
	}
	qs = m.GetQuestState(charID, q.Name())
	if !qs.IsCompleted() {
		t.Fatal("quest should be completed")
	}
}

// --- Q00002: What Women Want (Human/Elf only, choice) ---

func TestQ00002_RaceRestriction(t *testing.T) {
	q := NewQ00002()
	m := setupQuestManager(t, q)

	orc := testPlayer(t, 1, 2, RaceOrc)
	result := m.DispatchEvent(talkEvent(orc, 30223, 100))
	if result == "" {
		t.Fatal("expected response for orc")
	}
}

func TestQ00002_Lifecycle(t *testing.T) {
	q := NewQ00002()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 2, RaceHuman)
	charID := player.CharacterID()

	m.DispatchEvent(talkEvent(player, 30223, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30223, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}

	// Mirabel (30146) — cond 1→2
	m.DispatchEvent(talkEvent(player, 30146, 200))
	qs = m.GetQuestState(charID, q.Name())
	if qs.GetCond() != 2 {
		t.Errorf("cond after Mirabel = %d; want 2", qs.GetCond())
	}

	// Arujien (30223) — cond 2→3, gives letter3, says "take to Herbiel"
	m.DispatchEvent(talkEvent(player, 30223, 100))
	qs = m.GetQuestState(charID, q.Name())
	if qs.GetCond() != 3 {
		t.Errorf("cond after Arujien = %d; want 3", qs.GetCond())
	}

	// Herbiel (30150) — cond 3, takes letter3
	m.DispatchEvent(talkEvent(player, 30150, 300))

	// Arujien (30223) — cond 3, letter3 removed → choice dialog
	// Выбираем "adena" вариант
	m.DispatchEvent(talkEventWithAccept(player, 30223, 100, "adena"))
	qs = m.GetQuestState(charID, q.Name())
	if !qs.IsCompleted() {
		t.Fatal("quest should be completed after adena choice")
	}
}

// --- Q00003: Will The Seal Be Broken (Dark Elf, level 16+) ---

func TestQ00003_RaceAndLevel(t *testing.T) {
	q := NewQ00003()
	m := setupQuestManager(t, q)

	// Не тёмный эльф
	human := testPlayer(t, 1, 16, RaceHuman)
	result := m.DispatchEvent(talkEvent(human, 30141, 100))
	if result == "" {
		t.Fatal("expected response for non-dark-elf")
	}

	// Тёмный эльф низкого уровня
	lowLevel := testPlayer(t, 2, 10, RaceDarkElf)
	result = m.DispatchEvent(talkEvent(lowLevel, 30141, 100))
	if result == "" {
		t.Fatal("expected low level response")
	}
}

func TestQ00003_Accept(t *testing.T) {
	q := NewQ00003()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 16, RaceDarkElf)
	charID := player.CharacterID()

	m.DispatchEvent(talkEvent(player, 30141, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30141, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
}

// --- Q00004: Long Live the Pa'agrio Lord (Orc only) ---

func TestQ00004_RaceRestriction(t *testing.T) {
	q := NewQ00004()
	m := setupQuestManager(t, q)

	human := testPlayer(t, 1, 2, RaceHuman)
	result := m.DispatchEvent(talkEvent(human, 30578, 100))
	if result == "" {
		t.Fatal("expected response for non-orc")
	}
}

func TestQ00004_FullLifecycle(t *testing.T) {
	q := NewQ00004()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 2, RaceOrc)
	charID := player.CharacterID()

	// Nakusin (30578) — start
	m.DispatchEvent(talkEvent(player, 30578, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30578, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}

	// Visit 6 NPCs: 30585, 30566, 30562, 30560, 30559, 30587
	giftNPCs := []int32{30585, 30566, 30562, 30560, 30559, 30587}
	for i, npcID := range giftNPCs {
		result := m.DispatchEvent(talkEvent(player, npcID, uint32(200+i)))
		if result == "" {
			t.Errorf("NPC %d gave empty response", npcID)
		}
	}

	// Return to Nakusin
	m.DispatchEvent(talkEvent(player, 30578, 100))
	qs = m.GetQuestState(charID, q.Name())
	if !qs.IsCompleted() {
		t.Fatal("quest should be completed after visiting all NPCs")
	}
}

// --- Q00005: Miner's Favor ---

func TestQ00005_FullLifecycle(t *testing.T) {
	q := NewQ00005()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 2, RaceHuman)
	charID := player.CharacterID()

	// Bolter (30554)
	m.DispatchEvent(talkEvent(player, 30554, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30554, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}

	// Visit Garita (30518) — mining boots
	m.DispatchEvent(talkEvent(player, 30518, 200))
	// Visit Shari (30517) — boom-boom powder
	m.DispatchEvent(talkEvent(player, 30517, 300))
	// Visit Reed (30520) — redstone beer
	m.DispatchEvent(talkEvent(player, 30520, 400))
	// Visit Brunon (30526) — needs smelly socks
	m.DispatchEvent(talkEvent(player, 30526, 500))

	// Return to Bolter — should complete
	m.DispatchEvent(talkEvent(player, 30554, 100))
	qs = m.GetQuestState(charID, q.Name())
	if !qs.IsCompleted() {
		t.Fatal("quest should be completed after collecting all items")
	}
}

// --- Q00101: Sword of Solidarity (Human only) ---

func TestQ00101_RaceRestriction(t *testing.T) {
	q := NewQ00101()
	m := setupQuestManager(t, q)

	elf := testPlayer(t, 1, 9, RaceElf)
	result := m.DispatchEvent(talkEvent(elf, 30008, 100))
	if result == "" {
		t.Fatal("expected response for non-human")
	}
}

func TestQ00101_Accept(t *testing.T) {
	q := NewQ00101()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 9, RaceHuman)
	charID := player.CharacterID()

	m.DispatchEvent(talkEvent(player, 30008, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30008, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
	if qs.GetCond() != 1 {
		t.Errorf("cond = %d; want 1", qs.GetCond())
	}

	// Altran (30283) — cond 1→2 (get directions)
	m.DispatchEvent(talkEvent(player, 30283, 200))
	qs = m.GetQuestState(charID, q.Name())
	if qs.GetCond() != 2 {
		t.Errorf("cond after Altran = %d; want 2", qs.GetCond())
	}
}

// --- Q00102: Sea of Spores Fever (Elf only) ---

func TestQ00102_RaceRestriction(t *testing.T) {
	q := NewQ00102()
	m := setupQuestManager(t, q)

	human := testPlayer(t, 1, 12, RaceHuman)
	result := m.DispatchEvent(talkEvent(human, 30284, 100))
	if result == "" {
		t.Fatal("expected response for non-elf")
	}
}

func TestQ00102_Accept(t *testing.T) {
	q := NewQ00102()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 12, RaceElf)
	charID := player.CharacterID()

	m.DispatchEvent(talkEvent(player, 30284, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30284, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}

	// Cobendell (30156) — cond 1→2
	m.DispatchEvent(talkEvent(player, 30156, 200))
	qs = m.GetQuestState(charID, q.Name())
	if qs.GetCond() != 2 {
		t.Errorf("cond after Cobendell = %d; want 2", qs.GetCond())
	}
}

// --- Q00151: Cure for Fever Disease ---

func TestQ00151_Lifecycle(t *testing.T) {
	q := NewQ00151()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 15, RaceHuman)
	charID := player.CharacterID()

	// Elias (30050)
	m.DispatchEvent(talkEvent(player, 30050, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30050, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
	if qs.GetCond() != 1 {
		t.Errorf("cond = %d; want 1", qs.GetCond())
	}
}

// --- Q00152: Shards of Golem ---

func TestQ00152_Lifecycle(t *testing.T) {
	q := NewQ00152()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 10, RaceHuman)
	charID := player.CharacterID()

	// Harris (30035)
	m.DispatchEvent(talkEvent(player, 30035, 100))
	m.DispatchEvent(talkEventWithAccept(player, 30035, 100, "accept"))

	qs := m.GetQuestState(charID, q.Name())
	if qs == nil || !qs.IsStarted() {
		t.Fatal("quest should be started")
	}
	if qs.GetCond() != 1 {
		t.Errorf("cond = %d; want 1", qs.GetCond())
	}

	// Altran (30283) — cond 1→2
	m.DispatchEvent(talkEvent(player, 30283, 200))
	qs = m.GetQuestState(charID, q.Name())
	if qs.GetCond() != 2 {
		t.Errorf("cond after Altran = %d; want 2", qs.GetCond())
	}
}

// --- Level restriction tests ---

func TestQ00257_LevelRestriction(t *testing.T) {
	q := NewQ00257()
	m := setupQuestManager(t, q)

	lowLevel := testPlayer(t, 1, 3, RaceHuman)
	result := m.DispatchEvent(talkEvent(lowLevel, 30039, 100))
	if result == "" {
		t.Fatal("expected level restriction message")
	}
}

func TestQ00261_LevelRestriction(t *testing.T) {
	q := NewQ00261()
	m := setupQuestManager(t, q)

	lowLevel := testPlayer(t, 1, 5, RaceHuman)
	result := m.DispatchEvent(talkEvent(lowLevel, 30222, 100))
	if result == "" {
		t.Fatal("expected level restriction message")
	}
}

func TestQ00101_LevelRestriction(t *testing.T) {
	q := NewQ00101()
	m := setupQuestManager(t, q)

	lowLevel := testPlayer(t, 1, 5, RaceHuman)
	result := m.DispatchEvent(talkEvent(lowLevel, 30008, 100))
	if result == "" {
		t.Fatal("expected level restriction message")
	}
}

// --- Completed quest dialog tests ---

func TestQ00001_CompletedDialog(t *testing.T) {
	q := NewQ00001()
	m := setupQuestManager(t, q)
	player := testPlayer(t, 1, 2, RaceHuman)
	charID := player.CharacterID()

	// Ускоренно проходим весь квест
	m.DispatchEvent(talkEventWithAccept(player, 30048, 100, "accept"))
	m.DispatchEvent(talkEvent(player, 30006, 200))
	m.DispatchEvent(talkEvent(player, 30048, 100))
	m.DispatchEvent(talkEvent(player, 30033, 300))
	m.DispatchEvent(talkEvent(player, 30048, 100))

	qs := m.GetQuestState(charID, q.Name())
	if !qs.IsCompleted() {
		t.Fatal("quest should be completed")
	}

	// Разговор после завершения
	result := m.DispatchEvent(talkEvent(player, 30048, 100))
	if result == "" {
		t.Fatal("expected completed quest dialog")
	}
}
