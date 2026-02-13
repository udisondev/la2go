package olympiad

import (
	"testing"
	"time"
)

func TestNewOlympiad(t *testing.T) {
	oly := NewOlympiad()

	if oly == nil {
		t.Fatal("NewOlympiad returned nil")
	}
	if oly.Manager() == nil {
		t.Fatal("Manager() is nil")
	}
	if oly.HeroTable() == nil {
		t.Fatal("HeroTable() is nil")
	}
	if oly.Nobles() == nil {
		t.Fatal("Nobles() is nil")
	}
	if oly.Period() != PeriodCompetition {
		t.Errorf("Period() = %v; want %v", oly.Period(), PeriodCompetition)
	}
	if oly.CurrentCycle() != 0 {
		t.Errorf("CurrentCycle() = %d; want 0", oly.CurrentCycle())
	}
	if oly.InCompPeriod() {
		t.Error("InCompPeriod() should be false initially")
	}
	if oly.IsOlympiadEnd() {
		t.Error("IsOlympiadEnd() should be false initially")
	}
}

func TestOlympiad_PeriodControl(t *testing.T) {
	oly := NewOlympiad()

	oly.SetPeriod(PeriodValidation)
	if oly.Period() != PeriodValidation {
		t.Errorf("Period() = %v; want %v", oly.Period(), PeriodValidation)
	}
	if !oly.IsOlympiadEnd() {
		t.Error("IsOlympiadEnd() should be true for PeriodValidation")
	}

	oly.SetPeriod(PeriodCompetition)
	if oly.IsOlympiadEnd() {
		t.Error("IsOlympiadEnd() should be false for PeriodCompetition")
	}
}

func TestOlympiad_CompPeriodControl(t *testing.T) {
	oly := NewOlympiad()

	oly.StartCompPeriod()
	if !oly.InCompPeriod() {
		t.Error("InCompPeriod() should be true after StartCompPeriod")
	}

	oly.EndCompPeriod()
	if oly.InCompPeriod() {
		t.Error("InCompPeriod() should be false after EndCompPeriod")
	}
}

func TestOlympiad_CycleControl(t *testing.T) {
	oly := NewOlympiad()

	oly.SetCurrentCycle(5)
	if oly.CurrentCycle() != 5 {
		t.Errorf("CurrentCycle() = %d; want 5", oly.CurrentCycle())
	}
}

func TestOlympiad_SetCompSchedule(t *testing.T) {
	oly := NewOlympiad()

	start := time.Now()
	oly.SetCompSchedule(start)

	if oly.CompStart().IsZero() {
		t.Error("CompStart() should not be zero")
	}

	expectedEnd := start.Add(CompPeriodDuration).UnixMilli()
	if oly.CompEnd() != expectedEnd {
		t.Errorf("CompEnd() = %d; want %d", oly.CompEnd(), expectedEnd)
	}
}

func TestOlympiad_SetNewOlympiadEnd(t *testing.T) {
	oly := NewOlympiad()

	oly.SetNewOlympiadEnd()

	end := oly.OlympiadEnd()
	if end <= time.Now().UnixMilli() {
		t.Error("OlympiadEnd should be in the future")
	}

	// Конец — 1-е число следующего месяца в 12:00.
	// Оставшееся время зависит от текущей даты (1-31 дней).
	remaining := oly.RemainingTimeToEnd()
	if remaining < 1*time.Hour || remaining > 32*24*time.Hour {
		t.Errorf("RemainingTimeToEnd() = %v; want 1h..32d", remaining)
	}
}

func TestOlympiad_EndMonth(t *testing.T) {
	oly := NewOlympiad()

	// Зарегистрировать noble с достаточным количеством матчей
	n := oly.Nobles().Register(1, 88)
	n.LoadStats(NobleStats{ClassID: 88, Points: 50, CompDone: 15, CompWon: 10})

	candidates := oly.EndMonth()

	// Проверить переход в validation
	if oly.Period() != PeriodValidation {
		t.Errorf("Period() = %v; want %v", oly.Period(), PeriodValidation)
	}

	// Должен быть хотя бы 1 кандидат
	if len(candidates) != 1 {
		t.Fatalf("EndMonth() candidates count = %d; want 1", len(candidates))
	}
	if candidates[0].CharID != 1 {
		t.Errorf("candidate CharID = %d; want 1", candidates[0].CharID)
	}

	// Герой должен быть установлен
	if !oly.HeroTable().IsHero(1) {
		t.Error("charID 1 should be a hero after EndMonth")
	}
}

func TestOlympiad_EndMonth_NoEligible(t *testing.T) {
	oly := NewOlympiad()

	// Noble без достаточного количества матчей
	n := oly.Nobles().Register(1, 88)
	n.LoadStats(NobleStats{ClassID: 88, Points: 50, CompDone: 5, CompWon: 3})

	candidates := oly.EndMonth()

	if len(candidates) != 0 {
		t.Errorf("EndMonth() candidates count = %d; want 0", len(candidates))
	}
}

func TestOlympiad_EndValidation(t *testing.T) {
	oly := NewOlympiad()
	oly.SetCurrentCycle(3)
	oly.SetPeriod(PeriodValidation)

	oly.EndValidation()

	if oly.Period() != PeriodCompetition {
		t.Errorf("Period() = %v; want %v", oly.Period(), PeriodCompetition)
	}
	if oly.CurrentCycle() != 4 {
		t.Errorf("CurrentCycle() = %d; want 4", oly.CurrentCycle())
	}
}

func TestOlympiad_GrantWeeklyPoints(t *testing.T) {
	oly := NewOlympiad()
	n := oly.Nobles().Register(1, 88)
	n.SetPoints(10) // cap = 0*10+12 = 12

	oly.GrantWeeklyPoints()

	if n.Points() != 12 { // 10+3=13 clamped to 12
		t.Errorf("Points() = %d; want 12", n.Points())
	}
}

func TestOlympiad_GrantWeeklyPoints_ValidationPeriod(t *testing.T) {
	oly := NewOlympiad()
	oly.SetPeriod(PeriodValidation)
	n := oly.Nobles().Register(1, 88)
	n.SetPoints(10)

	oly.GrantWeeklyPoints()

	// Не должны начисляться в validation period
	if n.Points() != 10 {
		t.Errorf("Points() = %d; want 10 (no change in validation)", n.Points())
	}
}

func TestOlympiad_RegisterNoble(t *testing.T) {
	oly := NewOlympiad()
	oly.StartCompPeriod()

	reason := oly.RegisterNoble(1, 88, false)
	if reason != "" {
		t.Errorf("RegisterNoble() = %q; want empty (success)", reason)
	}

	// Проверить что noble зарегистрирован
	if oly.Nobles().Get(1) == nil {
		t.Error("noble should be registered after RegisterNoble")
	}
}

func TestOlympiad_RegisterNoble_NotInCompPeriod(t *testing.T) {
	oly := NewOlympiad()
	// Не запускаем comp period

	reason := oly.RegisterNoble(1, 88, false)
	if reason == "" {
		t.Error("RegisterNoble() should fail when comp period is not active")
	}
}

func TestOlympiad_RegisterNoble_ValidationPeriod(t *testing.T) {
	oly := NewOlympiad()
	oly.StartCompPeriod()
	oly.SetPeriod(PeriodValidation)

	reason := oly.RegisterNoble(1, 88, false)
	if reason == "" {
		t.Error("RegisterNoble() should fail during validation")
	}
}

func TestOlympiad_RegisterNoble_LowPoints_Classed(t *testing.T) {
	oly := NewOlympiad()
	oly.StartCompPeriod()

	// Зарегистрировать noble с 2 очками
	n := oly.Nobles().Register(1, 88)
	n.SetPoints(2) // < 3 для classed

	reason := oly.RegisterNoble(1, 88, true)
	if reason == "" {
		t.Error("RegisterNoble(classed) should fail with < 3 points")
	}
}

func TestOlympiad_RegisterNoble_LowPoints_NonClassed(t *testing.T) {
	oly := NewOlympiad()
	oly.StartCompPeriod()

	n := oly.Nobles().Register(1, 88)
	n.SetPoints(4) // < 5 для non-classed

	reason := oly.RegisterNoble(1, 88, false)
	if reason == "" {
		t.Error("RegisterNoble(non-classed) should fail with < 5 points")
	}
}

func TestOlympiad_GetNoblePoints(t *testing.T) {
	oly := NewOlympiad()

	if oly.GetNoblePoints(999) != 0 {
		t.Error("GetNoblePoints(999) should be 0 for unregistered")
	}

	oly.Nobles().Register(1, 88)

	if oly.GetNoblePoints(1) != StartPoints {
		t.Errorf("GetNoblePoints(1) = %d; want %d", oly.GetNoblePoints(1), StartPoints)
	}
}

func TestOlympiad_GetNobleStats(t *testing.T) {
	oly := NewOlympiad()

	_, ok := oly.GetNobleStats(999)
	if ok {
		t.Error("GetNobleStats(999) ok = true; want false")
	}

	oly.Nobles().Register(1, 88)
	stats, ok := oly.GetNobleStats(1)
	if !ok {
		t.Error("GetNobleStats(1) ok = false; want true")
	}
	if stats.CharID != 1 {
		t.Errorf("stats.CharID = %d; want 1", stats.CharID)
	}
}

func TestOlympiad_GetClassLeaderboard(t *testing.T) {
	oly := NewOlympiad()

	n1 := oly.Nobles().Register(1, 88)
	n1.SetPoints(50)
	n2 := oly.Nobles().Register(2, 88)
	n2.SetPoints(30)
	n3 := oly.Nobles().Register(3, 88)
	n3.SetPoints(40)

	board := oly.GetClassLeaderboard(88)

	if len(board) != 3 {
		t.Fatalf("leaderboard count = %d; want 3", len(board))
	}
	// Должно быть отсортировано по points DESC
	if board[0].Points != 50 {
		t.Errorf("board[0].Points = %d; want 50", board[0].Points)
	}
	if board[1].Points != 40 {
		t.Errorf("board[1].Points = %d; want 40", board[1].Points)
	}
	if board[2].Points != 30 {
		t.Errorf("board[2].Points = %d; want 30", board[2].Points)
	}
}

func TestOlympiad_GetClassLeaderboard_Empty(t *testing.T) {
	oly := NewOlympiad()

	board := oly.GetClassLeaderboard(99)
	if board != nil {
		t.Errorf("leaderboard should be nil for empty class")
	}
}

func TestOlympiad_GetRank(t *testing.T) {
	oly := NewOlympiad()

	if oly.GetRank(1) != 0 {
		t.Error("GetRank(1) should be 0 initially")
	}
}

func TestOlympiad_FullCycle(t *testing.T) {
	oly := NewOlympiad()

	// Phase 1: Competition
	oly.SetNewOlympiadEnd()
	oly.StartCompPeriod()

	// Зарегистрировать noble
	n := oly.Nobles().Register(1, 88)
	n.LoadStats(NobleStats{ClassID: 88, Points: 50, CompDone: 15, CompWon: 10})

	// Phase 2: End month → Validation
	candidates := oly.EndMonth()
	if oly.Period() != PeriodValidation {
		t.Fatalf("should be in validation period")
	}
	if len(candidates) != 1 {
		t.Fatalf("should have 1 hero candidate")
	}

	// Phase 3: End validation → new cycle
	oly.EndValidation()
	if oly.Period() != PeriodCompetition {
		t.Fatalf("should be back in competition period")
	}
	if oly.CurrentCycle() != 1 {
		t.Fatalf("cycle should be 1")
	}
}

func TestPeriod_String(t *testing.T) {
	tests := []struct {
		p    Period
		want string
	}{
		{PeriodCompetition, "COMPETITION"},
		{PeriodValidation, "VALIDATION"},
		{Period(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.p.String()
		if got != tt.want {
			t.Errorf("Period(%d).String() = %q; want %q", tt.p, got, tt.want)
		}
	}
}

func TestOlympiad_Timestamps(t *testing.T) {
	oly := NewOlympiad()

	now := time.Now().UnixMilli()
	oly.SetOlympiadEnd(now + 3600000)
	oly.SetValidationEnd(now + 86400000)
	oly.SetNextWeeklyChange(now + 604800000)

	if oly.OlympiadEnd() != now+3600000 {
		t.Errorf("OlympiadEnd() = %d; want %d", oly.OlympiadEnd(), now+3600000)
	}
}
