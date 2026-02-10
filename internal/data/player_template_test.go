package data

import (
	"testing"
)

// TestInitStatBonuses verifies stat bonus tables initialization.
func TestInitStatBonuses(t *testing.T) {
	InitStatBonuses()

	// Verify key values (from L2J Mobius statBonus.xml)
	tests := []struct {
		stat     uint8
		expected float64
		delta    float64
	}{
		{35, 1.0, 0.01},     // Base point (stat=35 → bonus=1.0)
		{40, 1.20, 0.01},    // Human Fighter STR (40 → 1.20)
		{50, 1.71, 0.01},    // High stat (50 → 1.71)
		{60, 2.43, 0.01},    // Very high stat (60 → 2.43)
		{70, 3.47, 0.01},    // Exceptional stat (70 → 3.47)
		{80, 4.94, 0.01},    // Maximum realistic stat (80 → 4.94)
	}

	for _, tt := range tests {
		actual := STRBonus[tt.stat]
		if diff := actual - tt.expected; diff < -tt.delta || diff > tt.delta {
			t.Errorf("STRBonus[%d] = %.2f, expected %.2f (±%.2f)",
				tt.stat, actual, tt.expected, tt.delta)
		}
	}
}

// TestLoadPlayerTemplates verifies XML template loading.
func TestLoadPlayerTemplates(t *testing.T) {
	InitStatBonuses()

	err := LoadPlayerTemplates()
	if err != nil {
		t.Fatalf("LoadPlayerTemplates failed: %v", err)
	}

	// Verify templates loaded
	if len(PlayerTemplates) == 0 {
		t.Fatal("No templates loaded")
	}

	t.Logf("Loaded %d player templates", len(PlayerTemplates))

	// Verify key templates exist
	expectedClasses := []uint8{
		0,  // Human Fighter
		10, // Human Mystic
		18, // Elf Fighter
		25, // Elf Mystic
		31, // Dark Fighter
		38, // Orc Fighter
		44, // Dwarf Fighter
	}

	for _, classID := range expectedClasses {
		template := GetTemplate(classID)
		if template == nil {
			t.Errorf("Template classID=%d not loaded", classID)
		}
	}
}

// TestGetTemplate_HumanFighter verifies Human Fighter template data.
func TestGetTemplate_HumanFighter(t *testing.T) {
	InitStatBonuses()
	if err := LoadPlayerTemplates(); err != nil {
		t.Fatalf("LoadPlayerTemplates failed: %v", err)
	}

	template := GetTemplate(0) // Human Fighter
	if template == nil {
		t.Fatal("Human Fighter template not found")
	}

	// Verify base attributes (from Java HumanFighter.xml)
	if template.BaseSTR != 40 {
		t.Errorf("BaseSTR = %d, expected 40", template.BaseSTR)
	}
	if template.BaseCON != 43 {
		t.Errorf("BaseCON = %d, expected 43", template.BaseCON)
	}
	if template.BaseDEX != 30 {
		t.Errorf("BaseDEX = %d, expected 30", template.BaseDEX)
	}
	if template.BaseINT != 21 {
		t.Errorf("BaseINT = %d, expected 21", template.BaseINT)
	}
	if template.BaseWIT != 11 {
		t.Errorf("BaseWIT = %d, expected 11", template.BaseWIT)
	}
	if template.BaseMEN != 25 {
		t.Errorf("BaseMEN = %d, expected 25", template.BaseMEN)
	}

	// Verify combat stats
	if template.BasePAtk != 4 {
		t.Errorf("BasePAtk = %d, expected 4", template.BasePAtk)
	}
	if template.BasePDef != 80 {
		t.Errorf("BasePDef = %d, expected 80", template.BasePDef)
	}
	if template.BasePAtkSpd != 300 {
		t.Errorf("BasePAtkSpd = %d, expected 300", template.BasePAtkSpd)
	}
	if template.BaseCritRate != 4 {
		t.Errorf("BaseCritRate = %d, expected 4", template.BaseCritRate)
	}
	if template.BaseAtkRange != 20 {
		t.Errorf("BaseAtkRange = %d, expected 20", template.BaseAtkRange)
	}

	// Verify HP/MP/CP at level 1
	if template.HPByLevel[0] != 80.0 {
		t.Errorf("HPByLevel[0] = %.1f, expected 80.0", template.HPByLevel[0])
	}
	if template.MPByLevel[0] != 30.0 {
		t.Errorf("MPByLevel[0] = %.1f, expected 30.0", template.MPByLevel[0])
	}
	if template.CPByLevel[0] != 32.0 {
		t.Errorf("CPByLevel[0] = %.1f, expected 32.0", template.CPByLevel[0])
	}

	// Verify slot defense (Physical Defense по слотам)
	if template.SlotDef[SlotChest] != 31 {
		t.Errorf("SlotDef[Chest] = %d, expected 31", template.SlotDef[SlotChest])
	}
	if template.SlotDef[SlotLegs] != 18 {
		t.Errorf("SlotDef[Legs] = %d, expected 18", template.SlotDef[SlotLegs])
	}
}

// TestGetTemplate_ElfMystic verifies Elf Mystic template data.
func TestGetTemplate_ElfMystic(t *testing.T) {
	InitStatBonuses()
	if err := LoadPlayerTemplates(); err != nil {
		t.Fatalf("LoadPlayerTemplates failed: %v", err)
	}

	template := GetTemplate(25) // Elf Mystic
	if template == nil {
		t.Fatal("Elf Mystic template not found")
	}

	// Verify base attributes differ from Human Fighter
	if template.BaseSTR >= 40 {
		t.Errorf("BaseSTR = %d, expected < 40 (Mystic weaker than Fighter)", template.BaseSTR)
	}
	if template.BaseINT <= 21 {
		t.Errorf("BaseINT = %d, expected > 21 (Mystic smarter than Fighter)", template.BaseINT)
	}

	// Verify combat stats differ
	if template.BasePAtk >= 4 {
		t.Errorf("BasePAtk = %d, expected < 4 (Mystic weaker physical attack)", template.BasePAtk)
	}
	if template.BasePDef >= 80 {
		t.Errorf("BasePDef = %d, expected < 80 (Mystic weaker defense)", template.BasePDef)
	}
}

// TestClassDifferences verifies that different classes have different stats.
func TestClassDifferences(t *testing.T) {
	InitStatBonuses()
	if err := LoadPlayerTemplates(); err != nil {
		t.Fatalf("LoadPlayerTemplates failed: %v", err)
	}

	fighter := GetTemplate(0)  // Human Fighter
	mystic := GetTemplate(10)  // Human Mystic

	if fighter == nil || mystic == nil {
		t.Fatal("Templates not found")
	}

	// Fighters should have higher STR than Mystics
	if fighter.BaseSTR <= mystic.BaseSTR {
		t.Errorf("Fighter STR=%d should be > Mystic STR=%d",
			fighter.BaseSTR, mystic.BaseSTR)
	}

	// Mystics should have higher INT than Fighters
	if mystic.BaseINT <= fighter.BaseINT {
		t.Errorf("Mystic INT=%d should be > Fighter INT=%d",
			mystic.BaseINT, fighter.BaseINT)
	}

	// Fighters should have higher pAtk
	if fighter.BasePAtk <= mystic.BasePAtk {
		t.Errorf("Fighter pAtk=%d should be > Mystic pAtk=%d",
			fighter.BasePAtk, mystic.BasePAtk)
	}

	// Fighters should have higher pDef
	if fighter.BasePDef <= mystic.BasePDef {
		t.Errorf("Fighter pDef=%d should be > Mystic pDef=%d",
			fighter.BasePDef, mystic.BasePDef)
	}
}
