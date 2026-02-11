package data

import (
	"testing"
)

// TestLoadSkills_Count tests that all skills from XML are loaded.
func TestLoadSkills_Count(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	if len(SkillTable) < 2600 {
		t.Errorf("SkillTable should have >= 2600 skill IDs, got %d", len(SkillTable))
	}

	var totalEntries int
	for _, levels := range SkillTable {
		totalEntries += len(levels)
	}
	if totalEntries < 20000 {
		t.Errorf("total entries should be >= 20000, got %d", totalEntries)
	}
}

// TestLoadSkills_PowerStrike tests Power Strike (id=3) from XML.
func TestLoadSkills_PowerStrike(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	levels, ok := SkillTable[3]
	if !ok {
		t.Fatal("Power Strike (id=3) not found in SkillTable")
	}

	// Power Strike has 9 base levels in XML
	if len(levels) < 9 {
		t.Errorf("Power Strike levels: got %d, want >= 9", len(levels))
	}

	skill := levels[1]
	if skill == nil {
		t.Fatal("Power Strike level 1 is nil")
	}
	if skill.Name != "Power Strike" {
		t.Errorf("name: got %q, want %q", skill.Name, "Power Strike")
	}
	if skill.OperateType != OperateTypeA1 {
		t.Errorf("operateType: got %d, want A1 (%d)", skill.OperateType, OperateTypeA1)
	}
	if skill.IsMagic {
		t.Error("Power Strike should not be magic")
	}
	if skill.TargetType != TargetOne {
		t.Errorf("targetType: got %d, want ONE (%d)", skill.TargetType, TargetOne)
	}
	if skill.CastRange != 40 {
		t.Errorf("castRange: got %d, want 40", skill.CastRange)
	}
	if skill.ReuseDelay != 13000 {
		t.Errorf("reuseDelay: got %d, want 13000", skill.ReuseDelay)
	}
	if skill.Power != 25 {
		t.Errorf("power L1: got %.0f, want 25", skill.Power)
	}
	if skill.MpConsume != 10 {
		t.Errorf("mpConsume L1: got %d, want 10", skill.MpConsume)
	}
	if !skill.OverHit {
		t.Error("Power Strike should have overHit=true")
	}
	if !skill.NextActionAttack {
		t.Error("Power Strike should have nextActionAttack=true")
	}
}

// TestLoadSkills_TripleSlash tests Triple Slash (id=1) with enchant levels.
func TestLoadSkills_TripleSlash(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	// Triple Slash: 37 base + 30 enchant1 + 30 enchant2 = 97 total levels
	levels, ok := SkillTable[1]
	if !ok {
		t.Fatal("Triple Slash (id=1) not found")
	}
	if len(levels) != 97 {
		t.Errorf("Triple Slash total levels: got %d, want 97", len(levels))
	}

	// Base level 1: power=431, mpConsume=47
	skill1 := levels[1]
	if skill1 == nil {
		t.Fatal("Triple Slash level 1 is nil")
	}
	if skill1.Power != 431 {
		t.Errorf("power L1: got %.0f, want 431", skill1.Power)
	}
	if skill1.MpConsume != 47 {
		t.Errorf("mpConsume L1: got %d, want 47", skill1.MpConsume)
	}

	// Base level 37: power=2131
	skill37 := levels[37]
	if skill37 == nil {
		t.Fatal("Triple Slash level 37 is nil")
	}
	if skill37.Power != 2131 {
		t.Errorf("power L37: got %.0f, want 2131", skill37.Power)
	}

	// Enchant1 level 101: power=2151
	skill101 := levels[101]
	if skill101 == nil {
		t.Fatal("Triple Slash enchant level 101 is nil")
	}
	if skill101.Power != 2151 {
		t.Errorf("enchant1 L101 power: got %.0f, want 2151", skill101.Power)
	}

	// Enchant2 level 141: power=2131 (same as max), mpConsume=96
	skill141 := levels[141]
	if skill141 == nil {
		t.Fatal("Triple Slash enchant level 141 is nil")
	}
	if skill141.Power != 2131 {
		t.Errorf("enchant2 L141 power: got %.0f, want 2131", skill141.Power)
	}
	if skill141.MpConsume != 96 {
		t.Errorf("enchant2 L141 mpConsume: got %d, want 96", skill141.MpConsume)
	}
}

// TestLoadSkills_Dash tests Dash (id=4) — buff skill with stat mod.
func TestLoadSkills_Dash(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	skill := GetSkillTemplate(4, 1)
	if skill == nil {
		t.Fatal("Dash level 1 not found")
	}
	if skill.OperateType != OperateTypeA2 {
		t.Errorf("operateType: got %d, want A2 (%d)", skill.OperateType, OperateTypeA2)
	}
	if skill.TargetType != TargetSelf {
		t.Errorf("targetType: got %d, want SELF (%d)", skill.TargetType, TargetSelf)
	}
	if skill.AbnormalType != "SPEED_UP_SPECIAL" {
		t.Errorf("abnormalType: got %q, want %q", skill.AbnormalType, "SPEED_UP_SPECIAL")
	}
	if skill.AbnormalTime != 15 {
		t.Errorf("abnormalTime: got %d, want 15", skill.AbnormalTime)
	}

	// Check effect stat mod (runSpd +40 for level 1)
	if len(skill.Effects) != 1 {
		t.Fatalf("effects count: got %d, want 1", len(skill.Effects))
	}
	if skill.Effects[0].Name != "Buff" {
		t.Errorf("effect name: got %q, want %q", skill.Effects[0].Name, "Buff")
	}
	if len(skill.Effects[0].StatMods) != 1 {
		t.Fatalf("stat mods count: got %d, want 1", len(skill.Effects[0].StatMods))
	}
	sm := skill.Effects[0].StatMods[0]
	if sm.Op != "add" || sm.Stat != "runSpd" || sm.Val != 40 {
		t.Errorf("stat mod: got {%s, %s, %.0f}, want {add, runSpd, 40}", sm.Op, sm.Stat, sm.Val)
	}

	// Level 2: runSpd +66
	skill2 := GetSkillTemplate(4, 2)
	if skill2 == nil {
		t.Fatal("Dash level 2 not found")
	}
	sm2 := skill2.Effects[0].StatMods[0]
	if sm2.Val != 66 {
		t.Errorf("stat mod L2 val: got %.0f, want 66", sm2.Val)
	}
}

// TestLoadSkills_Confusion tests magic skill with effect params.
func TestLoadSkills_Confusion(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	skill := GetSkillTemplate(2, 1)
	if skill == nil {
		t.Fatal("Confusion level 1 not found")
	}
	if !skill.IsMagic {
		t.Error("Confusion should be magic")
	}
	if skill.HitTime != 1500 {
		t.Errorf("hitTime: got %d, want 1500", skill.HitTime)
	}
	if skill.CastRange != 600 {
		t.Errorf("castRange: got %d, want 600", skill.CastRange)
	}
	if skill.Trait != "DERANGEMENT" {
		t.Errorf("trait: got %q, want %q", skill.Trait, "DERANGEMENT")
	}
	if skill.MpInitialConsume != 2 {
		t.Errorf("mpInitialConsume L1: got %d, want 2", skill.MpInitialConsume)
	}

	// Effect: RandomizeHate with chance=80
	if len(skill.Effects) != 1 {
		t.Fatalf("effects count: got %d, want 1", len(skill.Effects))
	}
	if skill.Effects[0].Name != "RandomizeHate" {
		t.Errorf("effect name: got %q, want %q", skill.Effects[0].Name, "RandomizeHate")
	}
	if skill.Effects[0].Params["chance"] != "80" {
		t.Errorf("effect chance: got %q, want %q", skill.Effects[0].Params["chance"], "80")
	}
}

// TestLoadSkills_PassiveSkills tests passive skill detection.
func TestLoadSkills_PassiveSkills(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	// Fast Spell Casting (id=228): passive
	skill := GetSkillTemplate(228, 1)
	if skill == nil {
		t.Fatal("Fast Spell Casting level 1 not found")
	}
	if !skill.IsPassive() {
		t.Error("Fast Spell Casting should be passive")
	}
	if skill.IsActive() {
		t.Error("Fast Spell Casting should not be active")
	}
}

// TestGetSkill_ByIDAndLevel tests GetSkillTemplate lookup.
func TestGetSkill_ByIDAndLevel(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	skill := GetSkillTemplate(3, 1)
	if skill == nil {
		t.Fatal("GetSkillTemplate(3, 1) returned nil")
	}
	if skill.ID != 3 || skill.Level != 1 {
		t.Errorf("got ID=%d, Level=%d, want ID=3, Level=1", skill.ID, skill.Level)
	}

	missing := GetSkillTemplate(99999, 1)
	if missing != nil {
		t.Error("GetSkillTemplate(99999, 1) should return nil")
	}

	badLevel := GetSkillTemplate(3, 100)
	if badLevel != nil {
		t.Error("GetSkillTemplate(3, 100) should return nil")
	}
}

// TestGetSkillMaxLevel tests maximum level lookup.
func TestGetSkillMaxLevel(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	if got := GetSkillMaxLevel(3); got != 9 {
		t.Errorf("Power Strike max level: got %d, want 9", got)
	}
	if got := GetSkillMaxLevel(99999); got != 0 {
		t.Errorf("unknown skill max level: got %d, want 0", got)
	}
}

// TestLoadSkills_MultiLevel tests that multi-level skills create correct number of entries.
func TestLoadSkills_MultiLevel(t *testing.T) {
	if err := LoadSkills(); err != nil {
		t.Fatalf("LoadSkills() failed: %v", err)
	}

	tests := []struct {
		skillID    int32
		name       string
		minLevels  int // at least this many levels
	}{
		{3, "Power Strike", 9},
		{1, "Triple Slash", 37}, // base levels (+ enchant)
		{4, "Dash", 2},
		{2, "Confusion", 19},
	}

	for _, tt := range tests {
		levels, ok := SkillTable[tt.skillID]
		if !ok {
			t.Errorf("%s (id=%d) not found in SkillTable", tt.name, tt.skillID)
			continue
		}
		if len(levels) < tt.minLevels {
			t.Errorf("%s levels: got %d, want >= %d", tt.name, len(levels), tt.minLevels)
		}
	}
}
