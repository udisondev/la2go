package quest

import (
	"sync/atomic"
	"testing"
	"time"
)

// mockPlayer implements PlayerRef for testing.
type mockPlayer struct {
	objectID uint32
	name     string
}

func (p *mockPlayer) ObjectID() uint32 { return p.objectID }
func (p *mockPlayer) Name() string     { return p.name }

func TestTimerManager_StartAndFire(t *testing.T) {
	tm := NewTimerManager()
	defer tm.Shutdown()

	var fired atomic.Bool
	player := &mockPlayer{objectID: 1, name: "TestPlayer"}

	tm.StartTimer("test_quest", "timer1", 50*time.Millisecond, player, 0, func(name string, p PlayerRef, npcObjID uint32) {
		fired.Store(true)
	})

	if tm.ActiveCount() != 1 {
		t.Errorf("ActiveCount() = %d, want 1", tm.ActiveCount())
	}

	// Ждём срабатывания
	time.Sleep(150 * time.Millisecond)

	if !fired.Load() {
		t.Error("timer did not fire")
	}

	// Таймер должен быть удалён после срабатывания
	if tm.ActiveCount() != 0 {
		t.Errorf("ActiveCount() after fire = %d, want 0", tm.ActiveCount())
	}
}

func TestTimerManager_Cancel(t *testing.T) {
	tm := NewTimerManager()
	defer tm.Shutdown()

	var fired atomic.Bool
	player := &mockPlayer{objectID: 1, name: "TestPlayer"}

	tm.StartTimer("test_quest", "timer1", 200*time.Millisecond, player, 0, func(name string, p PlayerRef, npcObjID uint32) {
		fired.Store(true)
	})

	// Отменяем до срабатывания
	ok := tm.CancelTimer("test_quest", "timer1", 1)
	if !ok {
		t.Error("CancelTimer returned false, want true")
	}

	time.Sleep(300 * time.Millisecond)
	if fired.Load() {
		t.Error("cancelled timer should not fire")
	}
}

func TestTimerManager_CancelNonExistent(t *testing.T) {
	tm := NewTimerManager()
	defer tm.Shutdown()

	ok := tm.CancelTimer("nonexistent", "timer", 999)
	if ok {
		t.Error("CancelTimer for non-existent timer returned true, want false")
	}
}

func TestTimerManager_ReplaceExisting(t *testing.T) {
	tm := NewTimerManager()
	defer tm.Shutdown()

	var count atomic.Int32
	player := &mockPlayer{objectID: 1, name: "TestPlayer"}

	// Создаём первый таймер
	tm.StartTimer("quest", "timer1", 100*time.Millisecond, player, 0, func(name string, p PlayerRef, npcObjID uint32) {
		count.Add(1)
	})

	// Заменяем его вторым
	tm.StartTimer("quest", "timer1", 100*time.Millisecond, player, 0, func(name string, p PlayerRef, npcObjID uint32) {
		count.Add(10)
	})

	if tm.ActiveCount() != 1 {
		t.Errorf("ActiveCount() after replace = %d, want 1", tm.ActiveCount())
	}

	time.Sleep(200 * time.Millisecond)

	// Должен сработать только второй таймер (count = 10, не 11)
	if got := count.Load(); got != 10 {
		t.Errorf("count = %d, want 10 (only replacement timer should fire)", got)
	}
}

func TestTimerManager_CancelAllForPlayer(t *testing.T) {
	tm := NewTimerManager()
	defer tm.Shutdown()

	player1 := &mockPlayer{objectID: 1, name: "Player1"}
	player2 := &mockPlayer{objectID: 2, name: "Player2"}

	var fired1, fired2 atomic.Bool

	tm.StartTimer("quest", "t1", 200*time.Millisecond, player1, 0, func(name string, p PlayerRef, npcObjID uint32) {
		fired1.Store(true)
	})
	tm.StartTimer("quest", "t2", 200*time.Millisecond, player1, 0, func(name string, p PlayerRef, npcObjID uint32) {
		fired1.Store(true)
	})
	tm.StartTimer("quest", "t1", 200*time.Millisecond, player2, 0, func(name string, p PlayerRef, npcObjID uint32) {
		fired2.Store(true)
	})

	if tm.ActiveCount() != 3 {
		t.Errorf("ActiveCount() = %d, want 3", tm.ActiveCount())
	}

	tm.CancelAllForPlayer(1)

	time.Sleep(300 * time.Millisecond)

	if fired1.Load() {
		t.Error("player1 timers should not fire after CancelAllForPlayer")
	}
	if !fired2.Load() {
		t.Error("player2 timer should still fire")
	}
}

func TestTimerManager_CancelAllForQuest(t *testing.T) {
	tm := NewTimerManager()
	defer tm.Shutdown()

	player := &mockPlayer{objectID: 1, name: "Player"}
	var fired1, fired2 atomic.Bool

	tm.StartTimer("questA", "t1", 200*time.Millisecond, player, 0, func(name string, p PlayerRef, npcObjID uint32) {
		fired1.Store(true)
	})
	tm.StartTimer("questB", "t1", 200*time.Millisecond, player, 0, func(name string, p PlayerRef, npcObjID uint32) {
		fired2.Store(true)
	})

	tm.CancelAllForQuest("questA")

	time.Sleep(300 * time.Millisecond)

	if fired1.Load() {
		t.Error("questA timer should not fire")
	}
	if !fired2.Load() {
		t.Error("questB timer should still fire")
	}
}

func TestTimerManager_Shutdown(t *testing.T) {
	tm := NewTimerManager()

	player := &mockPlayer{objectID: 1, name: "Player"}
	var fired atomic.Bool

	tm.StartTimer("quest", "t1", 200*time.Millisecond, player, 0, func(name string, p PlayerRef, npcObjID uint32) {
		fired.Store(true)
	})

	tm.Shutdown()

	time.Sleep(300 * time.Millisecond)
	if fired.Load() {
		t.Error("timer should not fire after Shutdown")
	}
	if tm.ActiveCount() != 0 {
		t.Errorf("ActiveCount() after Shutdown = %d, want 0", tm.ActiveCount())
	}
}

func TestTimerManager_CallbackParams(t *testing.T) {
	tm := NewTimerManager()
	defer tm.Shutdown()

	player := &mockPlayer{objectID: 42, name: "TestPlayer"}
	var gotName string
	var gotPlayerObjID uint32
	var gotNpcObjID uint32

	tm.StartTimer("quest", "myTimer", 50*time.Millisecond, player, 777, func(name string, p PlayerRef, npcObjID uint32) {
		gotName = name
		gotPlayerObjID = p.ObjectID()
		gotNpcObjID = npcObjID
	})

	time.Sleep(150 * time.Millisecond)

	if gotName != "myTimer" {
		t.Errorf("callback timerName = %q, want myTimer", gotName)
	}
	if gotPlayerObjID != 42 {
		t.Errorf("callback player objectID = %d, want 42", gotPlayerObjID)
	}
	if gotNpcObjID != 777 {
		t.Errorf("callback npcObjectID = %d, want 777", gotNpcObjID)
	}
}
