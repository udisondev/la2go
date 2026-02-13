package duel

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

func TestNewDuel(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	if d.ID() != 1 {
		t.Errorf("ID() = %d; want 1", d.ID())
	}
	if d.PlayerA() != a {
		t.Error("PlayerA() mismatch")
	}
	if d.PlayerB() != b {
		t.Error("PlayerB() mismatch")
	}
	if d.IsPartyDuel() {
		t.Error("IsPartyDuel() = true; want false")
	}
	if d.IsFinished() {
		t.Error("IsFinished() = true; want false")
	}
	if d.Countdown() != CountdownStart {
		t.Errorf("Countdown() = %d; want %d", d.Countdown(), CountdownStart)
	}
	if d.RemainingTime() > PlayerDuelDuration {
		t.Errorf("RemainingTime() = %v; want <= %v", d.RemainingTime(), PlayerDuelDuration)
	}
}

func TestNewDuel_PartyDuel(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, true)
	if !d.IsPartyDuel() {
		t.Error("IsPartyDuel() = false; want true")
	}
	if d.RemainingTime() > PartyDuelDuration {
		t.Errorf("RemainingTime() = %v; want <= %v", d.RemainingTime(), PartyDuelDuration)
	}
}

func TestSaveCondition(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.SaveCondition(a)

	cond := d.GetCondition(a.ObjectID())
	if cond == nil {
		t.Fatal("GetCondition returned nil")
	}
	if cond.ObjectID != a.ObjectID() {
		t.Errorf("condition ObjectID = %d; want %d", cond.ObjectID, a.ObjectID())
	}
	if cond.HP != a.CurrentHP() {
		t.Errorf("condition HP = %d; want %d", cond.HP, a.CurrentHP())
	}

	// Non-existent player
	if got := d.GetCondition(999); got != nil {
		t.Error("GetCondition(999) should be nil")
	}
}

func TestSaveAllConditions(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.SaveAllConditions()

	if d.GetCondition(a.ObjectID()) == nil {
		t.Error("condition for Alice not saved")
	}
	if d.GetCondition(b.ObjectID()) == nil {
		t.Error("condition for Bob not saved")
	}
}

func TestParticipantState(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)

	// Default state
	if got := d.ParticipantState(a.ObjectID()); got != StateNoDuel {
		t.Errorf("initial state = %d; want %d (StateNoDuel)", got, StateNoDuel)
	}

	d.SetParticipantState(a.ObjectID(), StateDuelling)
	if got := d.ParticipantState(a.ObjectID()); got != StateDuelling {
		t.Errorf("state after set = %d; want %d (StateDuelling)", got, StateDuelling)
	}
}

func TestInitParticipants(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()

	if got := d.ParticipantState(a.ObjectID()); got != StateDuelling {
		t.Errorf("Alice state = %d; want %d (StateDuelling)", got, StateDuelling)
	}
	if got := d.ParticipantState(b.ObjectID()); got != StateDuelling {
		t.Errorf("Bob state = %d; want %d (StateDuelling)", got, StateDuelling)
	}
}

func TestOnPlayerDefeat_1v1(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()

	// Alice defeated → Bob wins
	result := d.OnPlayerDefeat(a.ObjectID())
	if result != ResultTeam2Win {
		t.Errorf("result = %d; want %d (ResultTeam2Win)", result, ResultTeam2Win)
	}
	if got := d.ParticipantState(a.ObjectID()); got != StateDead {
		t.Errorf("Alice state = %d; want %d (StateDead)", got, StateDead)
	}
	if got := d.ParticipantState(b.ObjectID()); got != StateWinner {
		t.Errorf("Bob state = %d; want %d (StateWinner)", got, StateWinner)
	}
}

func TestOnPlayerDefeat_1v1_Reverse(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()

	// Bob defeated → Alice wins
	result := d.OnPlayerDefeat(b.ObjectID())
	if result != ResultTeam1Win {
		t.Errorf("result = %d; want %d (ResultTeam1Win)", result, ResultTeam1Win)
	}
}

func TestSurrender(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()

	result := d.Surrender(a.ObjectID())
	if result != ResultTeam1Surrender {
		t.Errorf("result = %d; want %d (ResultTeam1Surrender)", result, ResultTeam1Surrender)
	}

	// Second surrender should be ignored (CompareAndSwap fails)
	result2 := d.Surrender(b.ObjectID())
	if result2 != ResultContinue {
		t.Errorf("double surrender result = %d; want %d (ResultContinue)", result2, ResultContinue)
	}
}

func TestSurrender_Team2(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()

	result := d.Surrender(b.ObjectID())
	if result != ResultTeam2Surrender {
		t.Errorf("result = %d; want %d (ResultTeam2Surrender)", result, ResultTeam2Surrender)
	}
}

func TestCheckEndCondition_Timeout(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()
	d.endTime = time.Now().Add(-1 * time.Second) // Force timeout

	result := d.CheckEndCondition()
	if result != ResultTimeout {
		t.Errorf("result = %d; want %d (ResultTimeout)", result, ResultTimeout)
	}
}

func TestCheckEndCondition_Distance(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()

	// Move Alice far away
	a.SetLocation(model.NewLocation(10000, 10000, 0, 0))

	result := d.CheckEndCondition()
	if result != ResultCanceled {
		t.Errorf("result = %d; want %d (ResultCanceled)", result, ResultCanceled)
	}
}

func TestCheckEndCondition_Continue(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()

	result := d.CheckEndCondition()
	if result != ResultContinue {
		t.Errorf("result = %d; want %d (ResultContinue)", result, ResultContinue)
	}
}

func TestCheckEndCondition_SurrenderFlag(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()
	d.surrenderReq.Store(2)

	result := d.CheckEndCondition()
	if result != ResultTeam2Surrender {
		t.Errorf("result = %d; want %d (ResultTeam2Surrender)", result, ResultTeam2Surrender)
	}
}

func TestCheckEndCondition_WinnerFlag(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()
	d.SetParticipantState(b.ObjectID(), StateWinner)

	result := d.CheckEndCondition()
	if result != ResultTeam2Win {
		t.Errorf("result = %d; want %d (ResultTeam2Win)", result, ResultTeam2Win)
	}
}

func TestCheckEndCondition_Interrupted(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()
	d.SetParticipantState(a.ObjectID(), StateInterrupted)

	result := d.CheckEndCondition()
	if result != ResultCanceled {
		t.Errorf("result = %d; want %d (ResultCanceled)", result, ResultCanceled)
	}
}

func TestFinish(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	if d.IsFinished() {
		t.Error("should not be finished initially")
	}

	d.Finish()
	if !d.IsFinished() {
		t.Error("should be finished after Finish()")
	}

	// cancelCh should be closed
	select {
	case <-d.CancelCh():
	default:
		t.Error("cancelCh should be closed after Finish()")
	}
}

func TestCheckEndCondition_Finished(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.finished.Store(true)

	result := d.CheckEndCondition()
	if result != ResultContinue {
		t.Errorf("finished duel should return ResultContinue, got %d", result)
	}
}

func TestParticipants(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()

	parts := d.Participants()
	if len(parts) != 2 {
		t.Fatalf("Participants() count = %d; want 2", len(parts))
	}

	found := make(map[uint32]bool)
	for _, id := range parts {
		found[id] = true
	}
	if !found[a.ObjectID()] || !found[b.ObjectID()] {
		t.Errorf("Participants() = %v; want both %d and %d", parts, a.ObjectID(), b.ObjectID())
	}
}

func TestIsTeam1(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)

	if !d.isTeam1(a.ObjectID()) {
		t.Error("Alice should be team1")
	}
	if d.isTeam1(b.ObjectID()) {
		t.Error("Bob should not be team1")
	}
	if d.isTeam1(999) {
		t.Error("unknown player should not be team1")
	}
}

func TestRestoreConditions_NormalEnd(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.SaveAllConditions()

	// Не паникует, просто проходит
	d.RestoreConditions(false)
}

func TestRestoreConditions_AbnormalEnd(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.SaveAllConditions()

	// abnormalEnd=true: ничего не восстанавливается
	d.RestoreConditions(true)
}

func TestCanDuel_Dead(t *testing.T) {
	p := testPlayer(t, 1, "Dead")
	p.SetCurrentHP(0) // Dead player

	reason := CanDuel(p)
	if reason == "" {
		t.Error("dead player should not be able to duel")
	}
}

func TestCanDuel_LowHP(t *testing.T) {
	p := testPlayer(t, 1, "LowHP")
	// Set HP to less than 50% of max
	p.SetCurrentHP(p.MaxHP()/2 - 1)

	reason := CanDuel(p)
	if reason == "" {
		t.Error("low HP player should not be able to duel")
	}
}

func TestCanDuel_LowMP(t *testing.T) {
	p := testPlayer(t, 1, "LowMP")
	// Set MP to less than 50% of max
	p.SetCurrentMP(p.MaxMP()/2 - 1)

	reason := CanDuel(p)
	if reason == "" {
		t.Error("low MP player should not be able to duel")
	}
}

func TestOnPlayerDefeat_PartyDuel_NoPartyAllDead(t *testing.T) {
	// Party duel without actual party members — both leaders checked
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, true) // partyDuel = true
	d.InitParticipants()

	// Alice defeated — team1 should be dead (no party members)
	result := d.OnPlayerDefeat(a.ObjectID())
	if result != ResultTeam2Win {
		t.Errorf("result = %d; want %d (ResultTeam2Win)", result, ResultTeam2Win)
	}
}

func TestOnPlayerDefeat_PartyDuel_OneAlive(t *testing.T) {
	// Party duel: defeat one but not the leader, duel continues
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, true)
	d.InitParticipants()

	// Bob defeated — Bob's team checked, he's leader, so team is dead
	result := d.OnPlayerDefeat(b.ObjectID())
	if result != ResultTeam1Win {
		t.Errorf("result = %d; want %d (ResultTeam1Win)", result, ResultTeam1Win)
	}
}

func TestCheckEndCondition_Team1Winner(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()
	d.SetParticipantState(a.ObjectID(), StateWinner)

	result := d.CheckEndCondition()
	if result != ResultTeam1Win {
		t.Errorf("result = %d; want %d (ResultTeam1Win)", result, ResultTeam1Win)
	}
}

func TestCheckEndCondition_SurrenderReq1(t *testing.T) {
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d := NewDuel(1, a, b, false)
	d.InitParticipants()
	d.surrenderReq.Store(1)

	result := d.CheckEndCondition()
	if result != ResultTeam1Surrender {
		t.Errorf("result = %d; want %d (ResultTeam1Surrender)", result, ResultTeam1Surrender)
	}
}
