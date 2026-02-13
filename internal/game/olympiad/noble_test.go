package olympiad

import "testing"

func TestNewNoble(t *testing.T) {
	n := NewNoble(100, 88)

	if n.CharID() != 100 {
		t.Errorf("CharID() = %d; want 100", n.CharID())
	}
	if n.ClassID() != 88 {
		t.Errorf("ClassID() = %d; want 88", n.ClassID())
	}
	if n.Points() != StartPoints {
		t.Errorf("Points() = %d; want %d", n.Points(), StartPoints)
	}
	if n.CompDone() != 0 {
		t.Errorf("CompDone() = %d; want 0", n.CompDone())
	}
	if n.CompWon() != 0 {
		t.Errorf("CompWon() = %d; want 0", n.CompWon())
	}
	if n.CompLost() != 0 {
		t.Errorf("CompLost() = %d; want 0", n.CompLost())
	}
	if n.CompDrawn() != 0 {
		t.Errorf("CompDrawn() = %d; want 0", n.CompDrawn())
	}
}

func TestNoble_RecordWin(t *testing.T) {
	n := NewNoble(1, 88)
	initial := n.Points()

	n.RecordWin(5)

	if n.Points() != initial+5 {
		t.Errorf("Points() after win = %d; want %d", n.Points(), initial+5)
	}
	if n.CompDone() != 1 {
		t.Errorf("CompDone() = %d; want 1", n.CompDone())
	}
	if n.CompWon() != 1 {
		t.Errorf("CompWon() = %d; want 1", n.CompWon())
	}
}

func TestNoble_RecordLoss(t *testing.T) {
	n := NewNoble(1, 88)
	initial := n.Points()

	n.RecordLoss(3)

	if n.Points() != initial-3 {
		t.Errorf("Points() after loss = %d; want %d", n.Points(), initial-3)
	}
	if n.CompLost() != 1 {
		t.Errorf("CompLost() = %d; want 1", n.CompLost())
	}
}

func TestNoble_RecordLoss_NoNegative(t *testing.T) {
	n := NewNoble(1, 88)
	n.SetPoints(2)

	n.RecordLoss(10) // штраф > очков

	if n.Points() != 0 {
		t.Errorf("Points() after big loss = %d; want 0", n.Points())
	}
}

func TestNoble_RecordDraw(t *testing.T) {
	n := NewNoble(1, 88)
	initial := n.Points()

	n.RecordDraw(2)

	if n.Points() != initial-2 {
		t.Errorf("Points() after draw = %d; want %d", n.Points(), initial-2)
	}
	if n.CompDrawn() != 1 {
		t.Errorf("CompDrawn() = %d; want 1", n.CompDrawn())
	}
	if n.CompDone() != 1 {
		t.Errorf("CompDone() = %d; want 1", n.CompDone())
	}
}

func TestNoble_AddPoints(t *testing.T) {
	n := NewNoble(1, 88)

	n.AddPoints(5)
	if n.Points() != StartPoints+5 {
		t.Errorf("Points() after +5 = %d; want %d", n.Points(), StartPoints+5)
	}

	n.AddPoints(-100) // клампинг к 0
	if n.Points() != 0 {
		t.Errorf("Points() after -100 = %d; want 0", n.Points())
	}
}

func TestNoble_GrantWeeklyPoints(t *testing.T) {
	n := NewNoble(1, 88)
	// compDone=0 → cap = 0*10+12 = 12
	n.SetPoints(10)

	n.GrantWeeklyPoints()

	if n.Points() != 12 { // 10+3=13 clamped to 12
		t.Errorf("Points() after weekly = %d; want 12", n.Points())
	}
}

func TestNoble_GrantWeeklyPoints_AtCap(t *testing.T) {
	n := NewNoble(1, 88)
	// compDone=0 → cap = 12
	n.SetPoints(12)

	n.GrantWeeklyPoints()

	// Уже на капе — не должно расти
	if n.Points() != 12 {
		t.Errorf("Points() = %d; want 12 (at cap)", n.Points())
	}
}

func TestNoble_GrantWeeklyPoints_WithMatches(t *testing.T) {
	n := NewNoble(1, 88)
	n.RecordWin(0) // compDone=1 → cap = 1*10+12 = 22
	n.SetPoints(20)

	n.GrantWeeklyPoints()

	if n.Points() != 22 { // 20+3=23 clamped to 22
		t.Errorf("Points() = %d; want 22", n.Points())
	}
}

func TestNoble_Stats(t *testing.T) {
	n := NewNoble(42, 93)
	n.RecordWin(5)
	n.RecordLoss(3)

	stats := n.Stats()

	if stats.CharID != 42 {
		t.Errorf("Stats.CharID = %d; want 42", stats.CharID)
	}
	if stats.ClassID != 93 {
		t.Errorf("Stats.ClassID = %d; want 93", stats.ClassID)
	}
	if stats.CompDone != 2 {
		t.Errorf("Stats.CompDone = %d; want 2", stats.CompDone)
	}
	if stats.CompWon != 1 {
		t.Errorf("Stats.CompWon = %d; want 1", stats.CompWon)
	}
	if stats.CompLost != 1 {
		t.Errorf("Stats.CompLost = %d; want 1", stats.CompLost)
	}
}

func TestNoble_LoadStats(t *testing.T) {
	n := NewNoble(1, 88)
	n.LoadStats(NobleStats{
		ClassID:   99,
		Points:    50,
		CompDone:  20,
		CompWon:   15,
		CompLost:  3,
		CompDrawn: 2,
	})

	if n.ClassID() != 99 {
		t.Errorf("ClassID() = %d; want 99", n.ClassID())
	}
	if n.Points() != 50 {
		t.Errorf("Points() = %d; want 50", n.Points())
	}
	if n.CompDone() != 20 {
		t.Errorf("CompDone() = %d; want 20", n.CompDone())
	}
}

func TestNoble_SetClassID(t *testing.T) {
	n := NewNoble(1, 88)
	n.SetClassID(99)
	if n.ClassID() != 99 {
		t.Errorf("ClassID() = %d; want 99", n.ClassID())
	}
}

// --- NobleTable ---

func TestNobleTable_Register(t *testing.T) {
	tbl := NewNobleTable()

	n := tbl.Register(100, 88)
	if n == nil {
		t.Fatal("Register returned nil")
	}
	if n.CharID() != 100 {
		t.Errorf("CharID() = %d; want 100", n.CharID())
	}

	// Повторная регистрация — тот же объект
	n2 := tbl.Register(100, 88)
	if n2 != n {
		t.Error("Register returned different object for same charID")
	}
}

func TestNobleTable_Get(t *testing.T) {
	tbl := NewNobleTable()

	if tbl.Get(999) != nil {
		t.Error("Get(999) should be nil for empty table")
	}

	tbl.Register(1, 88)
	if tbl.Get(1) == nil {
		t.Error("Get(1) should not be nil after Register")
	}
}

func TestNobleTable_Remove(t *testing.T) {
	tbl := NewNobleTable()
	tbl.Register(1, 88)

	tbl.Remove(1)

	if tbl.Get(1) != nil {
		t.Error("Get(1) should be nil after Remove")
	}
	if tbl.Count() != 0 {
		t.Errorf("Count() = %d; want 0", tbl.Count())
	}
}

func TestNobleTable_All(t *testing.T) {
	tbl := NewNobleTable()
	tbl.Register(1, 88)
	tbl.Register(2, 93)

	all := tbl.All()
	if len(all) != 2 {
		t.Fatalf("All() count = %d; want 2", len(all))
	}
}

func TestNobleTable_ByClassID(t *testing.T) {
	tbl := NewNobleTable()
	tbl.Register(1, 88)
	tbl.Register(2, 88)
	tbl.Register(3, 93)

	class88 := tbl.ByClassID(88)
	if len(class88) != 2 {
		t.Errorf("ByClassID(88) count = %d; want 2", len(class88))
	}

	class93 := tbl.ByClassID(93)
	if len(class93) != 1 {
		t.Errorf("ByClassID(93) count = %d; want 1", len(class93))
	}

	class99 := tbl.ByClassID(99)
	if len(class99) != 0 {
		t.Errorf("ByClassID(99) count = %d; want 0", len(class99))
	}
}

func TestNobleTable_Count(t *testing.T) {
	tbl := NewNobleTable()

	if tbl.Count() != 0 {
		t.Errorf("Count() = %d; want 0", tbl.Count())
	}

	tbl.Register(1, 88)
	if tbl.Count() != 1 {
		t.Errorf("Count() = %d; want 1", tbl.Count())
	}
}

func TestNobleTable_GrantAllWeeklyPoints(t *testing.T) {
	tbl := NewNobleTable()
	n1 := tbl.Register(1, 88)
	n2 := tbl.Register(2, 93)

	n1.SetPoints(10)
	n2.SetPoints(10)

	tbl.GrantAllWeeklyPoints()

	// cap=12 → 10+3=13 clamped to 12
	if n1.Points() != 12 {
		t.Errorf("n1.Points() = %d; want 12", n1.Points())
	}
	if n2.Points() != 12 {
		t.Errorf("n2.Points() = %d; want 12", n2.Points())
	}
}
