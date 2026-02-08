package model

import "testing"

func TestNewNpcTemplate(t *testing.T) {
	template := NewNpcTemplate(
		1000,           // templateID
		"Wolf",         // name
		"Wild Beast",   // title
		5,              // level
		1500,           // maxHP
		800,            // maxMP
		100, 50,        // pAtk, pDef
		80, 40,         // mAtk, mDef
		300,            // aggroRange
		120,            // moveSpeed
		253,            // atkSpeed
		30, 60,         // respawnMin, respawnMax
	)

	if template.TemplateID() != 1000 {
		t.Errorf("TemplateID() = %d, want 1000", template.TemplateID())
	}
	if template.Name() != "Wolf" {
		t.Errorf("Name() = %q, want Wolf", template.Name())
	}
	if template.Title() != "Wild Beast" {
		t.Errorf("Title() = %q, want Wild Beast", template.Title())
	}
	if template.Level() != 5 {
		t.Errorf("Level() = %d, want 5", template.Level())
	}
	if template.MaxHP() != 1500 {
		t.Errorf("MaxHP() = %d, want 1500", template.MaxHP())
	}
	if template.MaxMP() != 800 {
		t.Errorf("MaxMP() = %d, want 800", template.MaxMP())
	}
	if template.PAtk() != 100 {
		t.Errorf("PAtk() = %d, want 100", template.PAtk())
	}
	if template.PDef() != 50 {
		t.Errorf("PDef() = %d, want 50", template.PDef())
	}
	if template.MAtk() != 80 {
		t.Errorf("MAtk() = %d, want 80", template.MAtk())
	}
	if template.MDef() != 40 {
		t.Errorf("MDef() = %d, want 40", template.MDef())
	}
	if template.AggroRange() != 300 {
		t.Errorf("AggroRange() = %d, want 300", template.AggroRange())
	}
	if template.MoveSpeed() != 120 {
		t.Errorf("MoveSpeed() = %d, want 120", template.MoveSpeed())
	}
	if template.AtkSpeed() != 253 {
		t.Errorf("AtkSpeed() = %d, want 253", template.AtkSpeed())
	}
	if template.RespawnMin() != 30 {
		t.Errorf("RespawnMin() = %d, want 30", template.RespawnMin())
	}
	if template.RespawnMax() != 60 {
		t.Errorf("RespawnMax() = %d, want 60", template.RespawnMax())
	}
}
