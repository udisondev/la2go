package model

import "testing"

func TestNewRaidBoss(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(25001, "Test Raid Boss", "Lvl 50", 50,
		50000, 10000, 500, 300, 200, 150, 1000, 120, 300, 0, 0, 100000, 5000)

	rb := NewRaidBoss(9001, 25001, template)

	if rb == nil {
		t.Fatal("NewRaidBoss returned nil")
	}
	if !rb.IsRaid() {
		t.Error("IsRaid() = false; want true")
	}
	if rb.IsLethalable() {
		t.Error("IsLethalable() = true; want false")
	}
	if rb.Status() != RaidStatusAlive {
		t.Errorf("Status() = %d; want %d (ALIVE)", rb.Status(), RaidStatusAlive)
	}
	if !rb.UseRaidCurse() {
		t.Error("UseRaidCurse() = false; want true")
	}
	if rb.ObjectID() != 9001 {
		t.Errorf("ObjectID() = %d; want 9001", rb.ObjectID())
	}
	if rb.Name() != "Test Raid Boss" {
		t.Errorf("Name() = %q; want %q", rb.Name(), "Test Raid Boss")
	}
}

func TestRaidBoss_StatusTransitions(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(25001, "RB", "", 50,
		50000, 10000, 500, 300, 200, 150, 1000, 120, 300, 0, 0, 100000, 5000)

	rb := NewRaidBoss(9002, 25001, template)

	// Initial: ALIVE
	if rb.Status() != RaidStatusAlive {
		t.Fatalf("initial Status() = %d; want ALIVE(%d)", rb.Status(), RaidStatusAlive)
	}

	// Transition to FIGHTING
	rb.SetStatus(RaidStatusFighting)
	if rb.Status() != RaidStatusFighting {
		t.Errorf("Status() = %d; want FIGHTING(%d)", rb.Status(), RaidStatusFighting)
	}

	// Transition to DEAD
	rb.SetStatus(RaidStatusDead)
	if rb.Status() != RaidStatusDead {
		t.Errorf("Status() = %d; want DEAD(%d)", rb.Status(), RaidStatusDead)
	}

	// Back to ALIVE (respawn)
	rb.SetStatus(RaidStatusAlive)
	if rb.Status() != RaidStatusAlive {
		t.Errorf("Status() = %d; want ALIVE(%d)", rb.Status(), RaidStatusAlive)
	}
}

func TestRaidBoss_UseRaidCurse(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(25001, "RB", "", 50,
		50000, 10000, 500, 300, 200, 150, 1000, 120, 300, 0, 0, 100000, 5000)

	rb := NewRaidBoss(9003, 25001, template)

	// Default: true
	if !rb.UseRaidCurse() {
		t.Error("UseRaidCurse() default = false; want true")
	}

	rb.SetUseRaidCurse(false)
	if rb.UseRaidCurse() {
		t.Error("UseRaidCurse() = true after SetUseRaidCurse(false); want false")
	}

	rb.SetUseRaidCurse(true)
	if !rb.UseRaidCurse() {
		t.Error("UseRaidCurse() = false after SetUseRaidCurse(true); want true")
	}
}

func TestRaidBoss_DataTypeAssertion(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(25001, "RB", "", 50,
		50000, 10000, 500, 300, 200, 150, 1000, 120, 300, 0, 0, 100000, 5000)

	rb := NewRaidBoss(9004, 25001, template)

	// WorldObject.Data should point to *RaidBoss (not *Monster)
	data := rb.Npc.WorldObject.Data
	if data == nil {
		t.Fatal("WorldObject.Data is nil")
	}

	if _, ok := data.(*RaidBoss); !ok {
		t.Errorf("WorldObject.Data type = %T; want *RaidBoss", data)
	}

	// Should NOT be assertable as *Monster directly
	if _, ok := data.(*Monster); ok {
		t.Error("WorldObject.Data asserts as *Monster; want only *RaidBoss")
	}
}

func TestRaidBoss_InheritsMonsterBehavior(t *testing.T) {
	t.Parallel()

	template := NewNpcTemplate(25001, "RB", "", 50,
		50000, 10000, 500, 300, 200, 150, 1000, 120, 300, 0, 0, 100000, 5000)

	rb := NewRaidBoss(9005, 25001, template)

	// Monster methods work through embedding
	if !rb.IsAggressive() {
		t.Error("IsAggressive() = false; want true (aggroRange=1000)")
	}
	if rb.AggroRange() != 1000 {
		t.Errorf("AggroRange() = %d; want 1000", rb.AggroRange())
	}
	if rb.AggroList() == nil {
		t.Error("AggroList() = nil; want non-nil")
	}

	// Target tracking
	rb.SetTarget(42)
	if rb.Target() != 42 {
		t.Errorf("Target() = %d; want 42", rb.Target())
	}
	rb.ClearTarget()
	if rb.Target() != 0 {
		t.Errorf("Target() after ClearTarget = %d; want 0", rb.Target())
	}
}
