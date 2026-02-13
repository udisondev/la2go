package model

import "testing"

func TestNewGrandBoss(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(29001, "Antharas", "Dragon", 79,
		200000, 50000, 2000, 1500, 1800, 1200, 2000, 80, 250, 0, 0, 500000, 50000)

	gb := NewGrandBoss(10001, 29001, template, 29001)

	if gb == nil {
		t.Fatal("NewGrandBoss returned nil")
	}
	if !gb.IsRaid() {
		t.Error("IsRaid() = false; want true")
	}
	if gb.IsLethalable() {
		t.Error("IsLethalable() = true; want false")
	}
	if gb.Status() != GrandBossAlive {
		t.Errorf("Status() = %d; want %d (ALIVE)", gb.Status(), GrandBossAlive)
	}
	if gb.BossID() != 29001 {
		t.Errorf("BossID() = %d; want 29001", gb.BossID())
	}
	if gb.ObjectID() != 10001 {
		t.Errorf("ObjectID() = %d; want 10001", gb.ObjectID())
	}
	if gb.Name() != "Antharas" {
		t.Errorf("Name() = %q; want %q", gb.Name(), "Antharas")
	}
}

func TestGrandBoss_StatusTransitions(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(29001, "Antharas", "", 79,
		200000, 50000, 2000, 1500, 1800, 1200, 2000, 80, 250, 0, 0, 500000, 50000)

	gb := NewGrandBoss(10002, 29001, template, 29001)

	// Initial: ALIVE
	if gb.Status() != GrandBossAlive {
		t.Fatalf("initial Status() = %d; want ALIVE(%d)", gb.Status(), GrandBossAlive)
	}

	// All transitions
	transitions := []struct {
		name   string
		status GrandBossStatus
	}{
		{"FIGHTING", GrandBossFighting},
		{"DEAD", GrandBossDead},
		{"WAITING", GrandBossWaiting},
		{"ALIVE", GrandBossAlive},
	}

	for _, tr := range transitions {
		gb.SetStatus(tr.status)
		if gb.Status() != tr.status {
			t.Errorf("after SetStatus(%s): Status() = %d; want %d",
				tr.name, gb.Status(), tr.status)
		}
	}
}

func TestGrandBoss_DataTypeAssertion(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(29001, "Antharas", "", 79,
		200000, 50000, 2000, 1500, 1800, 1200, 2000, 80, 250, 0, 0, 500000, 50000)

	gb := NewGrandBoss(10003, 29001, template, 29001)

	data := gb.Npc.WorldObject.Data
	if data == nil {
		t.Fatal("WorldObject.Data is nil")
	}

	if _, ok := data.(*GrandBoss); !ok {
		t.Errorf("WorldObject.Data type = %T; want *GrandBoss", data)
	}

	// Should NOT be assertable as *Monster or *RaidBoss directly
	if _, ok := data.(*Monster); ok {
		t.Error("WorldObject.Data asserts as *Monster; want only *GrandBoss")
	}
	if _, ok := data.(*RaidBoss); ok {
		t.Error("WorldObject.Data asserts as *RaidBoss; want only *GrandBoss")
	}
}

func TestGrandBoss_InheritsMonsterBehavior(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(29001, "Antharas", "", 79,
		200000, 50000, 2000, 1500, 1800, 1200, 2000, 80, 250, 0, 0, 500000, 50000)

	gb := NewGrandBoss(10004, 29001, template, 29001)

	// Monster methods
	if !gb.IsAggressive() {
		t.Error("IsAggressive() = false; want true")
	}
	if gb.AggroRange() != 2000 {
		t.Errorf("AggroRange() = %d; want 2000", gb.AggroRange())
	}
	if gb.AggroList() == nil {
		t.Error("AggroList() = nil; want non-nil")
	}

	// NPC methods (template-based)
	if gb.PAtk() != 2000 {
		t.Errorf("PAtk() = %d; want 2000", gb.PAtk())
	}
	if gb.Level() != 79 {
		t.Errorf("Level() = %d; want 79", gb.Level())
	}
}
