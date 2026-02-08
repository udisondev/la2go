package model

import "testing"

func TestNewNpc(t *testing.T) {
	template := NewNpcTemplate(
		1000, "Wolf", "Wild Beast", 5, 1500, 800,
		100, 50, 80, 40, 300, 120, 253, 30, 60,
	)

	npc := NewNpc(12345, 1000, template)

	if npc.ObjectID() != 12345 {
		t.Errorf("ObjectID() = %d, want 12345", npc.ObjectID())
	}
	if npc.TemplateID() != 1000 {
		t.Errorf("TemplateID() = %d, want 1000", npc.TemplateID())
	}
	if npc.Name() != "Wolf" {
		t.Errorf("Name() = %q, want Wolf", npc.Name())
	}
	if npc.Title() != "Wild Beast" {
		t.Errorf("Title() = %q, want Wild Beast", npc.Title())
	}
	if npc.Level() != 5 {
		t.Errorf("Level() = %d, want 5", npc.Level())
	}
	if npc.MaxHP() != 1500 {
		t.Errorf("MaxHP() = %d, want 1500", npc.MaxHP())
	}
	if npc.CurrentHP() != 1500 {
		t.Errorf("CurrentHP() = %d, want 1500", npc.CurrentHP())
	}
	if npc.MaxMP() != 800 {
		t.Errorf("MaxMP() = %d, want 800", npc.MaxMP())
	}
	if npc.CurrentMP() != 800 {
		t.Errorf("CurrentMP() = %d, want 800", npc.CurrentMP())
	}
}

func TestNpc_StatsFromTemplate(t *testing.T) {
	template := NewNpcTemplate(
		1001, "Orc", "", 10, 2000, 1000,
		150, 75, 100, 50, 0, 100, 273, 60, 120,
	)

	npc := NewNpc(999, 1001, template)

	if npc.PAtk() != 150 {
		t.Errorf("PAtk() = %d, want 150", npc.PAtk())
	}
	if npc.PDef() != 75 {
		t.Errorf("PDef() = %d, want 75", npc.PDef())
	}
	if npc.MAtk() != 100 {
		t.Errorf("MAtk() = %d, want 100", npc.MAtk())
	}
	if npc.MDef() != 50 {
		t.Errorf("MDef() = %d, want 50", npc.MDef())
	}
	if npc.MoveSpeed() != 100 {
		t.Errorf("MoveSpeed() = %d, want 100", npc.MoveSpeed())
	}
	if npc.AtkSpeed() != 273 {
		t.Errorf("AtkSpeed() = %d, want 273", npc.AtkSpeed())
	}
}

func TestNpc_Intention(t *testing.T) {
	template := NewNpcTemplate(1000, "Test", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60)
	npc := NewNpc(1, 1000, template)

	// Initial intention should be IDLE
	if npc.Intention() != IntentionIdle {
		t.Errorf("initial Intention() = %v, want IDLE", npc.Intention())
	}

	// Set to ACTIVE
	npc.SetIntention(IntentionActive)
	if npc.Intention() != IntentionActive {
		t.Errorf("after SetIntention(ACTIVE) Intention() = %v, want ACTIVE", npc.Intention())
	}

	// Set to ATTACK
	npc.SetIntention(IntentionAttack)
	if npc.Intention() != IntentionAttack {
		t.Errorf("after SetIntention(ATTACK) Intention() = %v, want ATTACK", npc.Intention())
	}
}

func TestNpc_Decayed(t *testing.T) {
	template := NewNpcTemplate(1000, "Test", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60)
	npc := NewNpc(1, 1000, template)

	// Initial state should be not decayed
	if npc.IsDecayed() {
		t.Error("initial IsDecayed() = true, want false")
	}

	// Set decayed
	npc.SetDecayed(true)
	if !npc.IsDecayed() {
		t.Error("after SetDecayed(true) IsDecayed() = false, want true")
	}

	// Unset decayed
	npc.SetDecayed(false)
	if npc.IsDecayed() {
		t.Error("after SetDecayed(false) IsDecayed() = true, want false")
	}
}

func TestNpc_Spawn(t *testing.T) {
	template := NewNpcTemplate(1000, "Test", "", 1, 1000, 500, 0, 0, 0, 0, 0, 80, 253, 30, 60)
	npc := NewNpc(1, 1000, template)
	spawn := NewSpawn(100, 1000, 17000, 170000, -3500, 0, 1, true)

	// Initial spawn should be nil
	if npc.Spawn() != nil {
		t.Error("initial Spawn() != nil, want nil")
	}

	// Set spawn
	npc.SetSpawn(spawn)
	if npc.Spawn() != spawn {
		t.Error("after SetSpawn() Spawn() != expected spawn")
	}
}
