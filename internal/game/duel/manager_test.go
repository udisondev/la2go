package duel

import (
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.DuelCount() != 0 {
		t.Errorf("DuelCount() = %d; want 0", m.DuelCount())
	}
}

func TestCanDuel(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*model.Player)
		wantOK bool
	}{
		{
			name:   "healthy player",
			setup:  func(p *model.Player) {},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testPlayer(t, 1, "Test")
			tt.setup(p)
			reason := CanDuel(p)
			if tt.wantOK && reason != "" {
				t.Errorf("CanDuel() = %q; want empty", reason)
			}
			if !tt.wantOK && reason == "" {
				t.Error("CanDuel() = empty; want reason")
			}
		})
	}
}

func TestCreateDuel(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}
	if d == nil {
		t.Fatal("CreateDuel returned nil duel")
	}
	if d.ID() != 1 {
		t.Errorf("duel ID = %d; want 1", d.ID())
	}
	if m.DuelCount() != 1 {
		t.Errorf("DuelCount() = %d; want 1", m.DuelCount())
	}
}

func TestCreateDuel_AlreadyInDuel(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")
	c := testPlayer(t, 3, "Carol")

	if _, err := m.CreateDuel(a, b, false); err != nil {
		t.Fatalf("first CreateDuel: %v", err)
	}

	// Alice already in duel
	_, err := m.CreateDuel(a, c, false)
	if err == nil {
		t.Error("expected error for player already in duel")
	}

	// Bob already in duel
	_, err = m.CreateDuel(c, b, false)
	if err == nil {
		t.Error("expected error for player already in duel")
	}
}

func TestGetDuel(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}

	got := m.GetDuel(d.ID())
	if got != d {
		t.Error("GetDuel returned wrong duel")
	}

	if m.GetDuel(999) != nil {
		t.Error("GetDuel(999) should return nil")
	}
}

func TestGetDuelByPlayer(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}

	gotA := m.GetDuelByPlayer(a.ObjectID())
	if gotA != d {
		t.Error("GetDuelByPlayer(Alice) returned wrong duel")
	}

	gotB := m.GetDuelByPlayer(b.ObjectID())
	if gotB != d {
		t.Error("GetDuelByPlayer(Bob) returned wrong duel")
	}

	if m.GetDuelByPlayer(999) != nil {
		t.Error("GetDuelByPlayer(999) should return nil")
	}
}

func TestIsInDuel(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	if m.IsInDuel(a.ObjectID()) {
		t.Error("should not be in duel initially")
	}

	if _, err := m.CreateDuel(a, b, false); err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}

	if !m.IsInDuel(a.ObjectID()) {
		t.Error("Alice should be in duel after CreateDuel")
	}
	if !m.IsInDuel(b.ObjectID()) {
		t.Error("Bob should be in duel after CreateDuel")
	}
}

func TestRemoveDuel(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}

	m.RemoveDuel(d.ID())

	if m.DuelCount() != 0 {
		t.Errorf("DuelCount after remove = %d; want 0", m.DuelCount())
	}
	if m.IsInDuel(a.ObjectID()) {
		t.Error("Alice should not be in duel after remove")
	}
	if m.IsInDuel(b.ObjectID()) {
		t.Error("Bob should not be in duel after remove")
	}

	// Remove non-existent — should not panic
	m.RemoveDuel(999)
}

func TestOnPlayerDefeat_Manager(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}
	d.InitParticipants()

	m.OnPlayerDefeat(a.ObjectID())

	if d.ParticipantState(a.ObjectID()) != StateDead {
		t.Error("Alice should be StateDead after defeat")
	}
}

func TestOnSurrender_Manager(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}
	d.InitParticipants()

	m.OnSurrender(b.ObjectID())

	if d.surrenderReq.Load() != 2 {
		t.Errorf("surrenderReq = %d; want 2", d.surrenderReq.Load())
	}
}

func TestOnPlayerDefeat_NoDuel(t *testing.T) {
	m := NewManager()
	// Should not panic
	m.OnPlayerDefeat(999)
}

func TestOnSurrender_NoDuel(t *testing.T) {
	m := NewManager()
	// Should not panic
	m.OnSurrender(999)
}

func TestStartDuel_Lifecycle(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}

	// Уменьшаем countdown для быстрого прохождения
	d.countdown.Store(1)
	d.endTime = time.Now().Add(100 * time.Millisecond)

	endCh := make(chan Result, 1)

	m.StartDuel(d,
		func(d *Duel, count int32) {
			// countdown callback
		},
		func(d *Duel) {
			// start callback
		},
		func(d *Duel, result Result) {
			endCh <- result
		},
	)

	select {
	case result := <-endCh:
		if result != ResultTimeout {
			t.Errorf("result = %d; want %d (ResultTimeout)", result, ResultTimeout)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("duel lifecycle timed out")
	}

	// Дождёмся удаления дуэли
	time.Sleep(100 * time.Millisecond)
	if m.DuelCount() != 0 {
		t.Errorf("DuelCount after lifecycle = %d; want 0", m.DuelCount())
	}
}

func TestStartDuel_Cancel(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")

	d, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel: %v", err)
	}

	started := make(chan struct{})
	m.StartDuel(d,
		func(d *Duel, count int32) {},
		func(d *Duel) {
			close(started)
		},
		func(d *Duel, result Result) {},
	)

	// Сразу Finish, чтобы горутина завершилась
	d.Finish()

	// Дождёмся удаления
	time.Sleep(200 * time.Millisecond)
	if m.DuelCount() != 0 {
		t.Errorf("DuelCount after cancel = %d; want 0", m.DuelCount())
	}
}

func TestMultipleDuels(t *testing.T) {
	m := NewManager()
	a := testPlayer(t, 1, "Alice")
	b := testPlayer(t, 2, "Bob")
	c := testPlayer(t, 3, "Carol")
	d := testPlayer(t, 4, "Dave")

	d1, err := m.CreateDuel(a, b, false)
	if err != nil {
		t.Fatalf("CreateDuel 1: %v", err)
	}
	d2, err := m.CreateDuel(c, d, false)
	if err != nil {
		t.Fatalf("CreateDuel 2: %v", err)
	}

	if m.DuelCount() != 2 {
		t.Errorf("DuelCount() = %d; want 2", m.DuelCount())
	}

	if d1.ID() == d2.ID() {
		t.Error("two duels should have different IDs")
	}

	// Remove first, second stays
	m.RemoveDuel(d1.ID())
	if m.DuelCount() != 1 {
		t.Errorf("DuelCount after remove first = %d; want 1", m.DuelCount())
	}
	if m.IsInDuel(a.ObjectID()) {
		t.Error("Alice should not be in duel after remove")
	}
	if !m.IsInDuel(c.ObjectID()) {
		t.Error("Carol should still be in duel")
	}
}
