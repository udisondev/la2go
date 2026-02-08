package model

import "testing"

func TestNewMonster(t *testing.T) {
	// Non-aggressive NPC (aggroRange = 0)
	templatePassive := NewNpcTemplate(
		2000, "Rabbit", "", 1, 500, 100,
		10, 5, 5, 5, 0, 100, 253, 10, 20, // aggroRange = 0
	)

	monster := NewMonster(1, 2000, templatePassive)

	if monster.ObjectID() != 1 {
		t.Errorf("ObjectID() = %d, want 1", monster.ObjectID())
	}
	if monster.IsAggressive() {
		t.Error("IsAggressive() = true, want false for aggroRange=0")
	}
	if monster.AggroRange() != 0 {
		t.Errorf("AggroRange() = %d, want 0", monster.AggroRange())
	}
}

func TestMonster_Aggressive(t *testing.T) {
	// Aggressive NPC (aggroRange > 0)
	templateAggressive := NewNpcTemplate(
		2001, "Wolf", "", 5, 1500, 800,
		100, 50, 80, 40, 300, 120, 253, 30, 60, // aggroRange = 300
	)

	monster := NewMonster(2, 2001, templateAggressive)

	if !monster.IsAggressive() {
		t.Error("IsAggressive() = false, want true for aggroRange>0")
	}
	if monster.AggroRange() != 300 {
		t.Errorf("AggroRange() = %d, want 300", monster.AggroRange())
	}

	// Change aggressive flag
	monster.SetAggressive(false)
	if monster.IsAggressive() {
		t.Error("after SetAggressive(false) IsAggressive() = true, want false")
	}

	monster.SetAggressive(true)
	if !monster.IsAggressive() {
		t.Error("after SetAggressive(true) IsAggressive() = false, want true")
	}
}

func TestMonster_InheritsNpcFunctionality(t *testing.T) {
	template := NewNpcTemplate(
		2002, "Orc", "Warrior", 10, 2000, 1000,
		150, 75, 100, 50, 400, 100, 273, 60, 120,
	)

	monster := NewMonster(3, 2002, template)

	// Test inherited NPC methods
	if monster.Name() != "Orc" {
		t.Errorf("Name() = %q, want Orc", monster.Name())
	}
	if monster.Title() != "Warrior" {
		t.Errorf("Title() = %q, want Warrior", monster.Title())
	}
	if monster.Level() != 10 {
		t.Errorf("Level() = %d, want 10", monster.Level())
	}
	if monster.PAtk() != 150 {
		t.Errorf("PAtk() = %d, want 150", monster.PAtk())
	}

	// Test intention (inherited from Npc)
	if monster.Intention() != IntentionIdle {
		t.Errorf("initial Intention() = %v, want IDLE", monster.Intention())
	}

	monster.SetIntention(IntentionAttack)
	if monster.Intention() != IntentionAttack {
		t.Errorf("after SetIntention(ATTACK) Intention() = %v, want ATTACK", monster.Intention())
	}
}
