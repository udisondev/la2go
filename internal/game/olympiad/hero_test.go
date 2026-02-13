package olympiad

import "testing"

func TestIsHeroItem(t *testing.T) {
	tests := []struct {
		itemID int32
		want   bool
	}{
		{6842, true},   // Wings of Destiny
		{6611, true},   // Infinity Blade
		{6621, true},   // Infinity Spear
		{6615, true},   // middle weapon
		{6610, false},  // below range
		{6622, false},  // above range
		{100, false},   // random item
		{0, false},
	}

	for _, tt := range tests {
		got := IsHeroItem(tt.itemID)
		if got != tt.want {
			t.Errorf("IsHeroItem(%d) = %v; want %v", tt.itemID, got, tt.want)
		}
	}
}

func TestHeroSkillIDs(t *testing.T) {
	if len(HeroSkillIDs) != 5 {
		t.Fatalf("HeroSkillIDs count = %d; want 5", len(HeroSkillIDs))
	}

	expected := []int32{395, 396, 1374, 1375, 1376}
	for i, id := range expected {
		if HeroSkillIDs[i] != id {
			t.Errorf("HeroSkillIDs[%d] = %d; want %d", i, HeroSkillIDs[i], id)
		}
	}
}

func TestHeroTable_NewHeroTable(t *testing.T) {
	ht := NewHeroTable()

	if ht.IsHero(1) {
		t.Error("IsHero(1) should be false for empty table")
	}
	if ht.GetHero(1) != nil {
		t.Error("GetHero(1) should be nil for empty table")
	}
	if len(ht.AllHeroes()) != 0 {
		t.Errorf("AllHeroes() count = %d; want 0", len(ht.AllHeroes()))
	}
}

func TestHeroTable_ComputeNewHeroes(t *testing.T) {
	ht := NewHeroTable()

	candidates := []*HeroCandidate{
		{CharID: 1, ClassID: 88, Name: "Alice", Points: 50, CompDone: 15, CompWon: 10},
		{CharID: 2, ClassID: 93, Name: "Bob", Points: 45, CompDone: 12, CompWon: 8},
	}

	ht.ComputeNewHeroes(candidates)

	if !ht.IsHero(1) {
		t.Error("Alice should be a hero")
	}
	if !ht.IsHero(2) {
		t.Error("Bob should be a hero")
	}

	h1 := ht.GetHero(1)
	if h1 == nil {
		t.Fatal("GetHero(1) returned nil")
	}
	if h1.Name != "Alice" {
		t.Errorf("h1.Name = %q; want %q", h1.Name, "Alice")
	}
	if h1.Count != 1 {
		t.Errorf("h1.Count = %d; want 1", h1.Count)
	}
	if !h1.Played {
		t.Error("h1.Played should be true")
	}
	if h1.Claimed {
		t.Error("h1.Claimed should be false initially")
	}
}

func TestHeroTable_ComputeNewHeroes_ReturningHero(t *testing.T) {
	ht := NewHeroTable()

	// Первая олимпиада
	ht.ComputeNewHeroes([]*HeroCandidate{
		{CharID: 1, ClassID: 88, Name: "Alice"},
	})

	// Вторая олимпиада — Alice снова герой
	ht.ComputeNewHeroes([]*HeroCandidate{
		{CharID: 1, ClassID: 88, Name: "Alice"},
	})

	h := ht.GetHero(1)
	if h == nil {
		t.Fatal("GetHero(1) returned nil")
	}
	if h.Count != 2 {
		t.Errorf("Count = %d; want 2 (returning hero)", h.Count)
	}
}

func TestHeroTable_ComputeNewHeroes_ReplacesOld(t *testing.T) {
	ht := NewHeroTable()

	// Первая олимпиада — Alice
	ht.ComputeNewHeroes([]*HeroCandidate{
		{CharID: 1, ClassID: 88, Name: "Alice"},
	})

	// Вторая олимпиада — Bob (Alice заменена)
	ht.ComputeNewHeroes([]*HeroCandidate{
		{CharID: 2, ClassID: 93, Name: "Bob"},
	})

	if ht.IsHero(1) {
		t.Error("Alice should no longer be a hero")
	}
	if !ht.IsHero(2) {
		t.Error("Bob should be a hero")
	}
}

func TestHeroTable_ClaimHero(t *testing.T) {
	ht := NewHeroTable()
	ht.ComputeNewHeroes([]*HeroCandidate{
		{CharID: 1, ClassID: 88, Name: "Alice"},
	})

	if !ht.ClaimHero(1) {
		t.Error("ClaimHero(1) should return true")
	}

	h := ht.GetHero(1)
	if !h.Claimed {
		t.Error("Hero should be claimed")
	}

	// Повторный claim — false
	if ht.ClaimHero(1) {
		t.Error("ClaimHero(1) should return false (already claimed)")
	}
}

func TestHeroTable_ClaimHero_NotHero(t *testing.T) {
	ht := NewHeroTable()

	if ht.ClaimHero(999) {
		t.Error("ClaimHero(999) should return false for non-hero")
	}
}

func TestHeroTable_SetHeroMessage(t *testing.T) {
	ht := NewHeroTable()
	ht.ComputeNewHeroes([]*HeroCandidate{
		{CharID: 1, ClassID: 88, Name: "Alice"},
	})

	if !ht.SetHeroMessage(1, "Hello world") {
		t.Error("SetHeroMessage should return true")
	}

	msg := ht.HeroMessage(1)
	if msg != "Hello world" {
		t.Errorf("HeroMessage() = %q; want %q", msg, "Hello world")
	}
}

func TestHeroTable_SetHeroMessage_NotHero(t *testing.T) {
	ht := NewHeroTable()

	if ht.SetHeroMessage(999, "msg") {
		t.Error("SetHeroMessage should return false for non-hero")
	}

	if ht.HeroMessage(999) != "" {
		t.Error("HeroMessage should be empty for non-hero")
	}
}

func TestHeroTable_AddDiaryEntry(t *testing.T) {
	ht := NewHeroTable()

	ht.AddDiaryEntry(1, DiaryRaidKilled, 12345)
	ht.AddDiaryEntry(1, DiariCastleTaken, 3)

	entries := ht.Diary(1)
	if len(entries) != 2 {
		t.Fatalf("Diary(1) count = %d; want 2", len(entries))
	}
	if entries[0].Action != DiaryRaidKilled {
		t.Errorf("entries[0].Action = %d; want %d", entries[0].Action, DiaryRaidKilled)
	}
	if entries[0].Param != 12345 {
		t.Errorf("entries[0].Param = %d; want 12345", entries[0].Param)
	}
	if entries[1].Action != DiariCastleTaken {
		t.Errorf("entries[1].Action = %d; want %d", entries[1].Action, DiariCastleTaken)
	}
}

func TestHeroTable_Diary_Empty(t *testing.T) {
	ht := NewHeroTable()

	entries := ht.Diary(999)
	if len(entries) != 0 {
		t.Errorf("Diary(999) count = %d; want 0", len(entries))
	}
}

func TestHeroTable_AddFightRecord(t *testing.T) {
	ht := NewHeroTable()

	record := FightRecord{
		OpponentCharID: 2,
		OpponentName:   "Bob",
		Classed:        true,
		Result:         FightVictory,
	}
	ht.AddFightRecord(1, record)

	fights := ht.Fights(1)
	if len(fights) != 1 {
		t.Fatalf("Fights(1) count = %d; want 1", len(fights))
	}
	if fights[0].OpponentName != "Bob" {
		t.Errorf("fight.OpponentName = %q; want %q", fights[0].OpponentName, "Bob")
	}
	if fights[0].Result != FightVictory {
		t.Errorf("fight.Result = %d; want %d", fights[0].Result, FightVictory)
	}
}

func TestHeroTable_LoadHero(t *testing.T) {
	ht := NewHeroTable()

	ht.LoadHero(&HeroData{
		CharID:  1,
		ClassID: 88,
		Name:    "Alice",
		Count:   3,
		Played:  true,
		Claimed: true,
	})

	if !ht.IsHero(1) {
		t.Error("IsHero(1) should be true after LoadHero")
	}

	h := ht.GetHero(1)
	if h.Count != 3 {
		t.Errorf("Count = %d; want 3", h.Count)
	}
}

func TestHeroTable_LoadHero_OldHero(t *testing.T) {
	ht := NewHeroTable()

	ht.LoadHero(&HeroData{
		CharID:  1,
		ClassID: 88,
		Name:    "Alice",
		Played:  false, // старый герой
	})

	if ht.IsHero(1) {
		t.Error("IsHero(1) should be false for played=false")
	}
}

func TestHeroTable_AllHeroes(t *testing.T) {
	ht := NewHeroTable()
	ht.ComputeNewHeroes([]*HeroCandidate{
		{CharID: 1, ClassID: 88, Name: "Alice"},
		{CharID: 2, ClassID: 93, Name: "Bob"},
	})

	all := ht.AllHeroes()
	if len(all) != 2 {
		t.Fatalf("AllHeroes() count = %d; want 2", len(all))
	}
}

// --- SelectHeroes ---

func TestSelectHeroes(t *testing.T) {
	tbl := NewNobleTable()

	// Класс 88: 2 игрока
	n1 := tbl.Register(1, 88)
	n1.LoadStats(NobleStats{ClassID: 88, Points: 50, CompDone: 15, CompWon: 10})

	n2 := tbl.Register(2, 88)
	n2.LoadStats(NobleStats{ClassID: 88, Points: 40, CompDone: 12, CompWon: 8})

	// Класс 93: 1 игрок (достаточно матчей)
	n3 := tbl.Register(3, 93)
	n3.LoadStats(NobleStats{ClassID: 93, Points: 60, CompDone: 20, CompWon: 15})

	candidates := SelectHeroes(tbl)

	if len(candidates) != 2 {
		t.Fatalf("SelectHeroes() count = %d; want 2", len(candidates))
	}

	// Проверить что top-1 по классу 88 = charID 1 (50 points)
	found88 := false
	found93 := false
	for _, c := range candidates {
		if c.ClassID == 88 {
			found88 = true
			if c.CharID != 1 {
				t.Errorf("hero for class 88: charID = %d; want 1", c.CharID)
			}
		}
		if c.ClassID == 93 {
			found93 = true
			if c.CharID != 3 {
				t.Errorf("hero for class 93: charID = %d; want 3", c.CharID)
			}
		}
	}
	if !found88 {
		t.Error("no hero selected for class 88")
	}
	if !found93 {
		t.Error("no hero selected for class 93")
	}
}

func TestSelectHeroes_MinMatches(t *testing.T) {
	tbl := NewNobleTable()

	// Класс 88: 1 игрок, но только 5 матчей (< HeroMinMatches=9)
	n := tbl.Register(1, 88)
	n.LoadStats(NobleStats{ClassID: 88, Points: 100, CompDone: 5, CompWon: 5})

	candidates := SelectHeroes(tbl)

	if len(candidates) != 0 {
		t.Errorf("SelectHeroes() count = %d; want 0 (not enough matches)", len(candidates))
	}
}

func TestSelectHeroes_MinWins(t *testing.T) {
	tbl := NewNobleTable()

	// Класс 88: 9 матчей, но 0 побед
	n := tbl.Register(1, 88)
	n.LoadStats(NobleStats{ClassID: 88, Points: 50, CompDone: 9, CompWon: 0, CompLost: 5, CompDrawn: 4})

	candidates := SelectHeroes(tbl)

	if len(candidates) != 0 {
		t.Errorf("SelectHeroes() count = %d; want 0 (no wins)", len(candidates))
	}
}

func TestSelectHeroes_Soulhound(t *testing.T) {
	tbl := NewNobleTable()

	// Male Soulhound (132)
	n1 := tbl.Register(1, 132)
	n1.LoadStats(NobleStats{ClassID: 132, Points: 40, CompDone: 10, CompWon: 5})

	// Female Soulhound (133)
	n2 := tbl.Register(2, 133)
	n2.LoadStats(NobleStats{ClassID: 133, Points: 50, CompDone: 12, CompWon: 8})

	candidates := SelectHeroes(tbl)

	// Должен быть 1 герой (female с 50 points)
	soulhoundFound := false
	for _, c := range candidates {
		if c.ClassID == 132 || c.ClassID == 133 {
			soulhoundFound = true
			if c.CharID != 2 {
				t.Errorf("soulhound hero: charID = %d; want 2 (higher points)", c.CharID)
			}
		}
	}
	if !soulhoundFound {
		t.Error("no soulhound hero selected")
	}
}

func TestSelectHeroes_EmptyTable(t *testing.T) {
	tbl := NewNobleTable()

	candidates := SelectHeroes(tbl)
	if len(candidates) != 0 {
		t.Errorf("SelectHeroes() count = %d; want 0", len(candidates))
	}
}

// --- CalculateRanks ---

func TestCalculateRanks(t *testing.T) {
	// 100 nobles с ≥9 матчей
	stats := make([]NobleStats, 100)
	for i := range 100 {
		stats[i] = NobleStats{
			CharID:   int64(i + 1),
			Points:   int32(100 - i), // убывающие очки
			CompDone: 10,
			CompWon:  5,
		}
	}

	ranks := CalculateRanks(stats)

	if ranks == nil {
		t.Fatal("CalculateRanks returned nil")
	}

	// Top 1% = 1 noble (rank 1)
	rank1Count := 0
	for _, r := range ranks {
		if r == Rank1 {
			rank1Count++
		}
	}
	if rank1Count == 0 {
		t.Error("no rank 1 nobles")
	}

	// Все ранги должны быть 1-5
	for id, r := range ranks {
		if r < Rank1 || r > Rank5 {
			t.Errorf("noble %d has invalid rank %d", id, r)
		}
	}
}

func TestCalculateRanks_NotEnoughMatches(t *testing.T) {
	stats := []NobleStats{
		{CharID: 1, Points: 50, CompDone: 5, CompWon: 3}, // < 9 matches
	}

	ranks := CalculateRanks(stats)
	if ranks != nil {
		t.Errorf("CalculateRanks should return nil for nobles with < 9 matches")
	}
}

func TestCalculateRanks_Empty(t *testing.T) {
	ranks := CalculateRanks(nil)
	if ranks != nil {
		t.Errorf("CalculateRanks(nil) should return nil")
	}
}

func TestCalculateRanks_SingleNoble(t *testing.T) {
	stats := []NobleStats{
		{CharID: 1, Points: 50, CompDone: 10, CompWon: 5},
	}

	ranks := CalculateRanks(stats)
	if ranks == nil {
		t.Fatal("CalculateRanks returned nil")
	}
	if ranks[1] != Rank1 {
		t.Errorf("single noble rank = %d; want %d (Rank1)", ranks[1], Rank1)
	}
}
