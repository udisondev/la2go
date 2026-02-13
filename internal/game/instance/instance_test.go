package instance

import (
	"testing"
	"time"
)

func TestNewInstance(t *testing.T) {
	inst := NewInstance(1, 100, 1000, 30*time.Minute)

	if inst.ID() != 1 {
		t.Errorf("ID() = %d; want 1", inst.ID())
	}
	if inst.TemplateID() != 100 {
		t.Errorf("TemplateID() = %d; want 100", inst.TemplateID())
	}
	if inst.OwnerID() != 1000 {
		t.Errorf("OwnerID() = %d; want 1000", inst.OwnerID())
	}
	if inst.Duration() != 30*time.Minute {
		t.Errorf("Duration() = %v; want 30m", inst.Duration())
	}
	if inst.State() != StateCreated {
		t.Errorf("State() = %v; want CREATED", inst.State())
	}
	if inst.PlayerCount() != 0 {
		t.Errorf("PlayerCount() = %d; want 0", inst.PlayerCount())
	}
	if inst.NpcCount() != 0 {
		t.Errorf("NpcCount() = %d; want 0", inst.NpcCount())
	}
}

func TestInstance_AddRemovePlayer(t *testing.T) {
	inst := NewInstance(1, 100, 1000, 0)

	// Добавляем игрока.
	if !inst.AddPlayer(1) {
		t.Error("AddPlayer(1) = false; want true")
	}
	if inst.PlayerCount() != 1 {
		t.Errorf("PlayerCount() = %d; want 1", inst.PlayerCount())
	}
	if !inst.HasPlayer(1) {
		t.Error("HasPlayer(1) = false; want true")
	}

	// Повторное добавление — false.
	if inst.AddPlayer(1) {
		t.Error("AddPlayer(1) second time = true; want false")
	}

	// Добавляем второго.
	if !inst.AddPlayer(2) {
		t.Error("AddPlayer(2) = false; want true")
	}
	if inst.PlayerCount() != 2 {
		t.Errorf("PlayerCount() = %d; want 2", inst.PlayerCount())
	}

	// Удаляем первого — инстанс не пуст.
	removed, empty := inst.RemovePlayer(1)
	if !removed {
		t.Error("RemovePlayer(1) removed = false; want true")
	}
	if empty {
		t.Error("RemovePlayer(1) empty = true; want false")
	}
	if inst.HasPlayer(1) {
		t.Error("HasPlayer(1) after remove = true; want false")
	}

	// Удаляем второго — инстанс пуст.
	removed, empty = inst.RemovePlayer(2)
	if !removed {
		t.Error("RemovePlayer(2) removed = false; want true")
	}
	if !empty {
		t.Error("RemovePlayer(2) empty = false; want true")
	}

	// Удаление несуществующего.
	removed, empty = inst.RemovePlayer(999)
	if removed {
		t.Error("RemovePlayer(999) removed = true; want false")
	}
}

func TestInstance_Players(t *testing.T) {
	inst := NewInstance(1, 100, 1000, 0)
	inst.AddPlayer(10)
	inst.AddPlayer(20)
	inst.AddPlayer(30)

	players := inst.Players()
	if len(players) != 3 {
		t.Fatalf("Players() len = %d; want 3", len(players))
	}

	got := make(map[uint32]bool, 3)
	for _, id := range players {
		got[id] = true
	}
	for _, want := range []uint32{10, 20, 30} {
		if !got[want] {
			t.Errorf("Players() missing %d", want)
		}
	}
}

func TestInstance_AddRemoveNpc(t *testing.T) {
	inst := NewInstance(1, 100, 1000, 0)

	inst.AddNpc(5000)
	if !inst.HasNpc(5000) {
		t.Error("HasNpc(5000) = false; want true")
	}
	if inst.NpcCount() != 1 {
		t.Errorf("NpcCount() = %d; want 1", inst.NpcCount())
	}

	inst.RemoveNpc(5000)
	if inst.HasNpc(5000) {
		t.Error("HasNpc(5000) after remove = true; want false")
	}
	if inst.NpcCount() != 0 {
		t.Errorf("NpcCount() = %d; want 0", inst.NpcCount())
	}
}

func TestInstance_Npcs(t *testing.T) {
	inst := NewInstance(1, 100, 1000, 0)
	inst.AddNpc(100)
	inst.AddNpc(200)

	npcs := inst.Npcs()
	if len(npcs) != 2 {
		t.Fatalf("Npcs() len = %d; want 2", len(npcs))
	}

	got := make(map[uint32]bool, 2)
	for _, id := range npcs {
		got[id] = true
	}
	if !got[100] || !got[200] {
		t.Errorf("Npcs() missing expected IDs, got %v", npcs)
	}
}

func TestInstance_IsExpired(t *testing.T) {
	// Нет ограничения.
	inst := NewInstance(1, 100, 1000, 0)
	if inst.IsExpired() {
		t.Error("IsExpired() with zero duration = true; want false")
	}

	// Ещё не истёк.
	inst2 := NewInstance(2, 100, 1000, 1*time.Hour)
	if inst2.IsExpired() {
		t.Error("IsExpired() with 1h duration = true; want false")
	}

	// Искусственно устанавливаем старую дату.
	inst3 := NewInstance(3, 100, 1000, 1*time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	if !inst3.IsExpired() {
		t.Error("IsExpired() after 2ms with 1ms duration = false; want true")
	}
}

func TestInstance_State(t *testing.T) {
	inst := NewInstance(1, 100, 1000, 0)

	if inst.State() != StateCreated {
		t.Errorf("initial State() = %v; want CREATED", inst.State())
	}

	inst.SetState(StateActive)
	if inst.State() != StateActive {
		t.Errorf("State() after SetState(Active) = %v; want ACTIVE", inst.State())
	}

	inst.SetState(StateDestroyed)
	if inst.State() != StateDestroyed {
		t.Errorf("State() after SetState(Destroyed) = %v; want DESTROYED", inst.State())
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateCreated, "CREATED"},
		{StateActive, "ACTIVE"},
		{StateDestroying, "DESTROYING"},
		{StateDestroyed, "DESTROYED"},
		{State(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q; want %q", tt.state, got, tt.want)
		}
	}
}

func TestInstance_EmptyDelay(t *testing.T) {
	inst := NewInstance(1, 100, 1000, 0)
	if inst.EmptyDelay() != DefaultEmptyDelay {
		t.Errorf("EmptyDelay() = %v; want %v", inst.EmptyDelay(), DefaultEmptyDelay)
	}

	inst.SetEmptyDelay(10 * time.Second)
	if inst.EmptyDelay() != 10*time.Second {
		t.Errorf("EmptyDelay() = %v; want 10s", inst.EmptyDelay())
	}
}
