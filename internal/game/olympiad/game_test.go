package olympiad

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

func testPlayer(t *testing.T, objectID uint32, name string) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(objectID, int64(objectID), 1, name, 40, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer(%s): %v", name, err)
	}
	p.SetLocation(model.NewLocation(0, 0, 0, 0))
	return p
}

func testGame(t *testing.T) (*Game, *Noble, *Noble) {
	t.Helper()
	p1 := testPlayer(t, 1, "Alice")
	p2 := testPlayer(t, 2, "Bob")

	n1 := NewNoble(1, 88)
	n2 := NewNoble(2, 93)

	part1 := NewParticipant(p1, n1)
	part2 := NewParticipant(p2, n2)

	stadiums := NewStadiums()
	g := NewGame(1, stadiums[0], CompClassed, part1, part2)

	return g, n1, n2
}

func TestNewGame(t *testing.T) {
	g, _, _ := testGame(t)

	if g.ID() != 1 {
		t.Errorf("ID() = %d; want 1", g.ID())
	}
	if g.CompType() != CompClassed {
		t.Errorf("CompType() = %v; want CLASSED", g.CompType())
	}
	if g.Stadium() == nil {
		t.Fatal("Stadium() is nil")
	}
	if g.Player1() == nil {
		t.Fatal("Player1() is nil")
	}
	if g.Player2() == nil {
		t.Fatal("Player2() is nil")
	}
	if g.BattleStarted() {
		t.Error("BattleStarted() should be false initially")
	}
	if g.GameIsStarted() {
		t.Error("GameIsStarted() should be false initially")
	}
	if g.IsFinished() {
		t.Error("IsFinished() should be false initially")
	}
}

func TestGame_AtomicBool_BattleStarted(t *testing.T) {
	// ⚠️ Тест fix Java race condition — atomic.Bool для state
	g, _, _ := testGame(t)

	if g.BattleStarted() {
		t.Error("initial BattleStarted should be false")
	}

	g.SetBattleStarted()

	if !g.BattleStarted() {
		t.Error("BattleStarted should be true after SetBattleStarted")
	}
}

func TestGame_AtomicBool_GameIsStarted(t *testing.T) {
	// ⚠️ Тест fix Java race condition — atomic.Bool для state
	g, _, _ := testGame(t)

	if g.GameIsStarted() {
		t.Error("initial GameIsStarted should be false")
	}

	g.SetGameIsStarted()

	if !g.GameIsStarted() {
		t.Error("GameIsStarted should be true after SetGameIsStarted")
	}
}

func TestGame_Start(t *testing.T) {
	g, _, _ := testGame(t)

	g.Start()

	if !g.BattleStarted() {
		t.Error("BattleStarted() should be true after Start()")
	}
	if g.StartTime().IsZero() {
		t.Error("StartTime() should not be zero after Start()")
	}
	remaining := g.RemainingTime()
	if remaining <= 0 || remaining > BattleDuration {
		t.Errorf("RemainingTime() = %v; want (0, %v]", remaining, BattleDuration)
	}
}

func TestGame_RemainingTime_NotStarted(t *testing.T) {
	g, _, _ := testGame(t)

	if g.RemainingTime() != BattleDuration {
		t.Errorf("RemainingTime() = %v; want %v (not started)", g.RemainingTime(), BattleDuration)
	}
}

func TestGame_CheckTimeout(t *testing.T) {
	g, _, _ := testGame(t)

	// Не начат — timeout не может быть
	if g.CheckTimeout() {
		t.Error("CheckTimeout() = true for not started game")
	}

	// Force timeout
	g.mu.Lock()
	g.endTime = time.Now().Add(-1 * time.Second)
	g.mu.Unlock()

	if !g.CheckTimeout() {
		t.Error("CheckTimeout() = false for expired game")
	}
}

func TestGame_CalculateResult_Player1Dead(t *testing.T) {
	g, _, _ := testGame(t)

	g.Player1().Player.SetCurrentHP(0)
	g.Player2().Player.SetCurrentHP(100)

	result := g.CalculateResult()
	if result != ResultPlayer2 {
		t.Errorf("CalculateResult() = %d; want %d (ResultPlayer2)", result, ResultPlayer2)
	}
}

func TestGame_CalculateResult_Player2Dead(t *testing.T) {
	g, _, _ := testGame(t)

	g.Player1().Player.SetCurrentHP(100)
	g.Player2().Player.SetCurrentHP(0)

	result := g.CalculateResult()
	if result != ResultPlayer1 {
		t.Errorf("CalculateResult() = %d; want %d (ResultPlayer1)", result, ResultPlayer1)
	}
}

func TestGame_CalculateResult_BothDead(t *testing.T) {
	g, _, _ := testGame(t)

	g.Player1().Player.SetCurrentHP(0)
	g.Player2().Player.SetCurrentHP(0)

	result := g.CalculateResult()
	if result != ResultDraw {
		t.Errorf("CalculateResult() = %d; want %d (ResultDraw)", result, ResultDraw)
	}
}

func TestGame_CalculateResult_ByDamage_P1Wins(t *testing.T) {
	g, _, _ := testGame(t)

	g.Player1().Player.SetCurrentHP(100)
	g.Player2().Player.SetCurrentHP(100)
	g.Player1().DamageDealt = 500
	g.Player2().DamageDealt = 300

	result := g.CalculateResult()
	if result != ResultPlayer1 {
		t.Errorf("CalculateResult() = %d; want %d (ResultPlayer1 by damage)", result, ResultPlayer1)
	}
}

func TestGame_CalculateResult_ByDamage_P2Wins(t *testing.T) {
	g, _, _ := testGame(t)

	g.Player1().Player.SetCurrentHP(100)
	g.Player2().Player.SetCurrentHP(100)
	g.Player1().DamageDealt = 200
	g.Player2().DamageDealt = 800

	result := g.CalculateResult()
	if result != ResultPlayer2 {
		t.Errorf("CalculateResult() = %d; want %d (ResultPlayer2 by damage)", result, ResultPlayer2)
	}
}

func TestGame_CalculateResult_Draw_EqualDamage(t *testing.T) {
	g, _, _ := testGame(t)

	g.Player1().Player.SetCurrentHP(100)
	g.Player2().Player.SetCurrentHP(100)
	g.Player1().DamageDealt = 400
	g.Player2().DamageDealt = 400

	result := g.CalculateResult()
	if result != ResultDraw {
		t.Errorf("CalculateResult() = %d; want %d (ResultDraw)", result, ResultDraw)
	}
}

func TestGame_CalculatePointDiff_Classed(t *testing.T) {
	g, n1, n2 := testGame(t)
	n1.SetPoints(30)
	n2.SetPoints(20)

	// min(20,30) / 3 = 6 (div=3 for CLASSED)
	diff := g.CalculatePointDiff()
	if diff != 6 {
		t.Errorf("CalculatePointDiff() = %d; want 6", diff)
	}
}

func TestGame_CalculatePointDiff_Classed_Capped(t *testing.T) {
	g, n1, n2 := testGame(t)
	n1.SetPoints(100)
	n2.SetPoints(80)

	// min(80,100) / 3 = 26 → capped to 10
	diff := g.CalculatePointDiff()
	if diff != MaxPoints {
		t.Errorf("CalculatePointDiff() = %d; want %d (MaxPoints)", diff, MaxPoints)
	}
}

func TestGame_CalculatePointDiff_NonClassed(t *testing.T) {
	p1 := testPlayer(t, 1, "Alice")
	p2 := testPlayer(t, 2, "Bob")
	n1 := NewNoble(1, 88)
	n2 := NewNoble(2, 93)
	n1.SetPoints(40)
	n2.SetPoints(30)

	stadiums := NewStadiums()
	g := NewGame(1, stadiums[0], CompNonClassed, NewParticipant(p1, n1), NewParticipant(p2, n2))

	// min(30,40) / 5 = 6 (div=5 for NON_CLASSED)
	diff := g.CalculatePointDiff()
	if diff != 6 {
		t.Errorf("CalculatePointDiff() = %d; want 6", diff)
	}
}

func TestGame_CalculatePointDiff_Minimum(t *testing.T) {
	g, n1, n2 := testGame(t)
	n1.SetPoints(1)
	n2.SetPoints(1)

	// min(1,1) / 3 = 0 → minimum 1
	diff := g.CalculatePointDiff()
	if diff != 1 {
		t.Errorf("CalculatePointDiff() = %d; want 1 (minimum)", diff)
	}
}

func TestCalculateDrawPenalty(t *testing.T) {
	tests := []struct {
		points int32
		want   int32
	}{
		{50, 10},  // 50/5 = 10
		{100, 10}, // 100/5 = 20 → capped to 10
		{3, 1},    // 3/5 = 0 → minimum 1
		{0, 1},    // minimum 1
	}

	for _, tt := range tests {
		got := CalculateDrawPenalty(tt.points)
		if got != tt.want {
			t.Errorf("CalculateDrawPenalty(%d) = %d; want %d", tt.points, got, tt.want)
		}
	}
}

func TestCalculateDefaultPenalty(t *testing.T) {
	tests := []struct {
		points int32
		want   int32
	}{
		{30, 10}, // 30/3 = 10
		{60, 10}, // 60/3 = 20 → capped to 10
		{6, 2},   // 6/3 = 2
		{1, 1},   // minimum 1
	}

	for _, tt := range tests {
		got := CalculateDefaultPenalty(tt.points)
		if got != tt.want {
			t.Errorf("CalculateDefaultPenalty(%d) = %d; want %d", tt.points, got, tt.want)
		}
	}
}

func TestGame_ApplyResult_Player1Wins(t *testing.T) {
	g, n1, n2 := testGame(t)
	n1.SetPoints(30)
	n2.SetPoints(20)

	g.ApplyResult(ResultPlayer1)

	// diff = min(20,30)/3 = 6
	if n1.Points() != 30+6 {
		t.Errorf("n1.Points() = %d; want %d", n1.Points(), 30+6)
	}
	if n2.Points() != 20-6 {
		t.Errorf("n2.Points() = %d; want %d", n2.Points(), 20-6)
	}
	if n1.CompWon() != 1 {
		t.Errorf("n1.CompWon() = %d; want 1", n1.CompWon())
	}
	if n2.CompLost() != 1 {
		t.Errorf("n2.CompLost() = %d; want 1", n2.CompLost())
	}
}

func TestGame_ApplyResult_Player2Wins(t *testing.T) {
	g, n1, n2 := testGame(t)
	n1.SetPoints(30)
	n2.SetPoints(20)

	g.ApplyResult(ResultPlayer2)

	// diff = min(20,30)/3 = 6
	if n2.Points() != 20+6 {
		t.Errorf("n2.Points() = %d; want %d", n2.Points(), 20+6)
	}
	if n1.Points() != 30-6 {
		t.Errorf("n1.Points() = %d; want %d", n1.Points(), 30-6)
	}
}

func TestGame_ApplyResult_Draw(t *testing.T) {
	g, n1, n2 := testGame(t)
	n1.SetPoints(25) // penalty = 25/5 = 5
	n2.SetPoints(15) // penalty = 15/5 = 3

	g.ApplyResult(ResultDraw)

	if n1.Points() != 25-5 {
		t.Errorf("n1.Points() after draw = %d; want %d", n1.Points(), 25-5)
	}
	if n2.Points() != 15-3 {
		t.Errorf("n2.Points() after draw = %d; want %d", n2.Points(), 15-3)
	}
	if n1.CompDrawn() != 1 {
		t.Errorf("n1.CompDrawn() = %d; want 1", n1.CompDrawn())
	}
}

func TestGame_ApplyResult_Player1DC(t *testing.T) {
	g, n1, n2 := testGame(t)
	n1.SetPoints(30)
	n2.SetPoints(20)

	g.ApplyResult(ResultPlayer1DC)

	// p1 gets default penalty: 30/3 = 10
	if n1.Points() != 30-10 {
		t.Errorf("n1.Points() after DC = %d; want %d", n1.Points(), 30-10)
	}
	// p2 gets win diff: min(20,30)/3 = 6
	if n2.Points() != 20+6 {
		t.Errorf("n2.Points() after DC = %d; want %d", n2.Points(), 20+6)
	}
}

func TestGame_Finish(t *testing.T) {
	g, _, _ := testGame(t)
	g.Stadium().SetInUse(true)

	g.Finish()

	if !g.IsFinished() {
		t.Error("IsFinished() should be true after Finish()")
	}
	if g.Stadium().InUse() {
		t.Error("Stadium should be free after Finish()")
	}

	// cancelCh should be closed
	select {
	case <-g.CancelCh():
	default:
		t.Error("cancelCh should be closed")
	}
}

func TestGame_Finish_Idempotent(t *testing.T) {
	g, _, _ := testGame(t)

	g.Finish()
	g.Finish() // не должна паниковать (CompareAndSwap)

	if !g.IsFinished() {
		t.Error("IsFinished() should still be true")
	}
}

func TestGame_RecordDamage(t *testing.T) {
	g, _, _ := testGame(t)

	g.RecordDamage(g.Player1().ObjectID, 100)
	g.RecordDamage(g.Player1().ObjectID, 50)
	g.RecordDamage(g.Player2().ObjectID, 200)

	if g.Player1().DamageDealt != 150 {
		t.Errorf("P1 DamageDealt = %d; want 150", g.Player1().DamageDealt)
	}
	if g.Player2().DamageDealt != 200 {
		t.Errorf("P2 DamageDealt = %d; want 200", g.Player2().DamageDealt)
	}
}

func TestGame_RecordDamage_UnknownPlayer(t *testing.T) {
	g, _, _ := testGame(t)

	g.RecordDamage(999, 100) // unknown player — no panic

	if g.Player1().DamageDealt != 0 || g.Player2().DamageDealt != 0 {
		t.Error("damage should not be recorded for unknown player")
	}
}

func TestGame_GetOpponent(t *testing.T) {
	g, _, _ := testGame(t)

	opp := g.GetOpponent(g.Player1().ObjectID)
	if opp != g.Player2() {
		t.Error("GetOpponent(p1) should return p2")
	}

	opp = g.GetOpponent(g.Player2().ObjectID)
	if opp != g.Player1() {
		t.Error("GetOpponent(p2) should return p1")
	}

	opp = g.GetOpponent(999)
	if opp != nil {
		t.Error("GetOpponent(unknown) should return nil")
	}
}

func TestGame_GetParticipant(t *testing.T) {
	g, _, _ := testGame(t)

	if g.GetParticipant(g.Player1().ObjectID) != g.Player1() {
		t.Error("GetParticipant(p1.ObjectID) should return p1")
	}
	if g.GetParticipant(g.Player2().ObjectID) != g.Player2() {
		t.Error("GetParticipant(p2.ObjectID) should return p2")
	}
	if g.GetParticipant(999) != nil {
		t.Error("GetParticipant(999) should return nil")
	}
}

func TestGame_RestoreConditions(t *testing.T) {
	g, _, _ := testGame(t)

	// Сохраняем состояния
	g.Player1().SaveCondition()
	g.Player2().SaveCondition()

	origHP1 := g.Player1().Player.CurrentHP()
	origHP2 := g.Player2().Player.CurrentHP()

	// Изменяем HP
	g.Player1().Player.SetCurrentHP(1)
	g.Player2().Player.SetCurrentHP(1)

	// Восстанавливаем
	g.RestoreConditions()

	if g.Player1().Player.CurrentHP() != origHP1 {
		t.Errorf("P1 HP after restore = %d; want %d", g.Player1().Player.CurrentHP(), origHP1)
	}
	if g.Player2().Player.CurrentHP() != origHP2 {
		t.Errorf("P2 HP after restore = %d; want %d", g.Player2().Player.CurrentHP(), origHP2)
	}
}

func TestNewParticipant(t *testing.T) {
	p := testPlayer(t, 1, "Alice")
	n := NewNoble(1, 88)

	part := NewParticipant(p, n)

	if part.Player != p {
		t.Error("Player mismatch")
	}
	if part.Noble != n {
		t.Error("Noble mismatch")
	}
	if part.Name != "Alice" {
		t.Errorf("Name = %q; want %q", part.Name, "Alice")
	}
	if part.ObjectID != 1 {
		t.Errorf("ObjectID = %d; want 1", part.ObjectID)
	}
	if part.ClassID != 88 {
		t.Errorf("ClassID = %d; want 88", part.ClassID)
	}
}
