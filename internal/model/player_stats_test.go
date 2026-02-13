package model

import (
	"testing"

	"github.com/udisondev/la2go/internal/data"
)

func init() {
	// Initialize templates для тестов
	data.InitStatBonuses()
	_ = data.LoadPlayerTemplates()
}

// TestGetBasePAtk_Level1_HumanFighter verifies pAtk calculation for level 1 Human Fighter.
// Expected: basePAtk=4, STRBonus[40]=1.20, levelMod=0.90
// finalPAtk = 4 × 1.20 × 0.90 = 4.32 ≈ 4
func TestGetBasePAtk_Level1_HumanFighter(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pAtk := player.GetBasePAtk()

	// Expected: 4.32 ≈ 4
	if pAtk < 4 || pAtk > 5 {
		t.Errorf("GetBasePAtk() = %d, expected 4-5 (Human Fighter level 1)", pAtk)
	}
}

// TestGetBasePAtk_Level20_HumanFighter verifies pAtk scaling with level.
// Expected: basePAtk=4, STRBonus[40]=1.20, levelMod=1.09
// finalPAtk = 4 × 1.20 × 1.09 = 5.23 ≈ 5
func TestGetBasePAtk_Level20_HumanFighter(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 20, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pAtk := player.GetBasePAtk()

	// Expected: 5.23 ≈ 5
	if pAtk < 5 || pAtk > 6 {
		t.Errorf("GetBasePAtk() = %d, expected 5-6 (Human Fighter level 20)", pAtk)
	}
}

// TestGetBasePAtk_Level80_HumanFighter verifies pAtk at max level.
// Expected: basePAtk=4, STRBonus[40]=1.20, levelMod=1.69
// finalPAtk = 4 × 1.20 × 1.69 = 8.11 ≈ 8
func TestGetBasePAtk_Level80_HumanFighter(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 80, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pAtk := player.GetBasePAtk()

	// Expected: 8.11 ≈ 8
	if pAtk < 7 || pAtk > 9 {
		t.Errorf("GetBasePAtk() = %d, expected 7-9 (Human Fighter level 80)", pAtk)
	}
}

// TestGetBasePDef_Level1_Nude verifies pDef calculation for nude level 1 character.
// Expected: basePDef=80, levelMod=0.90
// finalPDef = 80 × 0.90 = 72
func TestGetBasePDef_Level1_Nude(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pDef := player.GetBasePDef()

	// Expected: 72
	if pDef < 71 || pDef > 73 {
		t.Errorf("GetBasePDef() = %d, expected 72 (Human Fighter level 1 nude)", pDef)
	}
}

// TestGetBasePDef_Level20_Nude verifies pDef scaling with level.
// Expected: basePDef=80, levelMod=1.09
// finalPDef = 80 × 1.09 = 87.2 ≈ 87
func TestGetBasePDef_Level20_Nude(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 20, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pDef := player.GetBasePDef()

	// Expected: 87
	if pDef < 86 || pDef > 88 {
		t.Errorf("GetBasePDef() = %d, expected 87 (Human Fighter level 20 nude)", pDef)
	}
}

// TestClassDifferences_PAtk verifies that different classes have different pAtk.
// Human Fighter (STR=40, basePAtk=4) should have higher pAtk than Elf Mystic (STR=21, basePAtk=3).
func TestClassDifferences_PAtk(t *testing.T) {
	fighter, err := NewPlayer(1, 100, 200, "Fighter", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer Fighter failed: %v", err)
	}

	mystic, err := NewPlayer(2, 101, 201, "Mystic", 1, 0, 25)
	if err != nil {
		t.Fatalf("NewPlayer Mystic failed: %v", err)
	}

	fighterPAtk := fighter.GetBasePAtk()
	mysticPAtk := mystic.GetBasePAtk()

	t.Logf("Fighter pAtk=%d, Mystic pAtk=%d", fighterPAtk, mysticPAtk)

	// Fighter должен иметь больше pAtk
	if fighterPAtk <= mysticPAtk {
		t.Errorf("Fighter pAtk=%d should be > Mystic pAtk=%d", fighterPAtk, mysticPAtk)
	}
}

// TestClassDifferences_PDef verifies that different classes have different pDef.
// Human Fighter (basePDef=80) should have higher pDef than Elf Mystic (basePDef=54).
func TestClassDifferences_PDef(t *testing.T) {
	fighter, err := NewPlayer(1, 100, 200, "Fighter", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer Fighter failed: %v", err)
	}

	mystic, err := NewPlayer(2, 101, 201, "Mystic", 1, 0, 25)
	if err != nil {
		t.Fatalf("NewPlayer Mystic failed: %v", err)
	}

	fighterPDef := fighter.GetBasePDef()
	mysticPDef := mystic.GetBasePDef()

	t.Logf("Fighter pDef=%d, Mystic pDef=%d", fighterPDef, mysticPDef)

	// Fighter должен иметь больше pDef
	if fighterPDef <= mysticPDef {
		t.Errorf("Fighter pDef=%d should be > Mystic pDef=%d", fighterPDef, mysticPDef)
	}
}

// TestGetLevelMod verifies level modifier calculation.
func TestGetLevelMod(t *testing.T) {
	tests := []struct {
		level    int32
		expected float64
		delta    float64
	}{
		{1, 0.90, 0.01},   // Level 1: (1+89)/100 = 0.90
		{20, 1.09, 0.01},  // Level 20: (20+89)/100 = 1.09
		{40, 1.29, 0.01},  // Level 40: (40+89)/100 = 1.29
		{80, 1.69, 0.01},  // Level 80: (80+89)/100 = 1.69
	}

	for _, tt := range tests {
		player, err := NewPlayer(1, 100, 200, "Test", tt.level, 0, 0)
		if err != nil {
			t.Fatalf("NewPlayer failed: %v", err)
		}

		levelMod := player.GetLevelMod()
		if diff := levelMod - tt.expected; diff < -tt.delta || diff > tt.delta {
			t.Errorf("GetLevelMod() level=%d = %.2f, expected %.2f (±%.2f)",
				tt.level, levelMod, tt.expected, tt.delta)
		}
	}
}

// TestGetSTR verifies STR attribute from template.
func TestGetSTR(t *testing.T) {
	fighter, err := NewPlayer(1, 100, 200, "Fighter", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer Fighter failed: %v", err)
	}

	str := fighter.GetSTR()

	// Human Fighter base STR = 40
	if str != 40 {
		t.Errorf("GetSTR() = %d, expected 40 (Human Fighter)", str)
	}
}

// TestGetPAtkSpd verifies attack speed from template with DEX bonus.
func TestGetPAtkSpd(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "Test", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pAtkSpd := player.GetPAtkSpd()

	// Human Fighter base PAtkSpd = 300, with DEX bonus applied.
	// Result depends on template BaseDEX and DEXBonus table.
	if pAtkSpd < 250.0 || pAtkSpd > 400.0 {
		t.Errorf("GetPAtkSpd() = %.1f, expected in range [250, 400]", pAtkSpd)
	}
}

// TestMockStatsReplaced verifies mock stats are replaced with real stats.
// Old mock: pAtk = 100 + level×5 (level 1 → 105)
// New real: pAtk = 4 × 1.20 × 0.90 ≈ 4
func TestMockStatsReplaced(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "Test", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pAtk := player.GetBasePAtk()

	// Mock would return 105, real should return ~4
	if pAtk >= 100 {
		t.Errorf("GetBasePAtk() = %d, still using mock formula (expected ~4, not 105)", pAtk)
	}

	if pAtk < 3 || pAtk > 6 {
		t.Errorf("GetBasePAtk() = %d, expected 3-6 (real stats)", pAtk)
	}
}
