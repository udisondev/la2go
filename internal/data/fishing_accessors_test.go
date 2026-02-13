package data

import "testing"

func TestGetFishTemplate_Valid(t *testing.T) {
	t.Parallel()

	tpl := GetFishTemplate(1)
	if tpl == nil {
		t.Fatal("GetFishTemplate(1) = nil; want non-nil")
	}
	if tpl.ID != 1 {
		t.Errorf("ID = %d; want 1", tpl.ID)
	}
	if tpl.ItemID != 6411 {
		t.Errorf("ItemID = %d; want 6411", tpl.ItemID)
	}
	if tpl.FishType != "swift" {
		t.Errorf("FishType = %q; want swift", tpl.FishType)
	}
	if tpl.Level != 1 {
		t.Errorf("Level = %d; want 1", tpl.Level)
	}
	if tpl.HP <= 0 {
		t.Errorf("HP = %d; want > 0", tpl.HP)
	}
}

func TestGetFishTemplate_NotFound(t *testing.T) {
	t.Parallel()

	tpl := GetFishTemplate(99999)
	if tpl != nil {
		t.Error("GetFishTemplate(99999) = non-nil; want nil")
	}
}

func TestGetFishTemplate_HasRegen(t *testing.T) {
	t.Parallel()

	tpl := GetFishTemplate(1)
	if tpl == nil {
		t.Fatal("GetFishTemplate(1) = nil")
	}
	if tpl.HPRegen <= 0 {
		t.Errorf("HPRegen = %.1f; want > 0", tpl.HPRegen)
	}
	if tpl.CombatDuration <= 0 {
		t.Errorf("CombatDuration = %d; want > 0", tpl.CombatDuration)
	}
}

func TestGetFishByLevel(t *testing.T) {
	t.Parallel()

	fishes := GetFishByLevel(1)
	if len(fishes) == 0 {
		t.Fatal("GetFishByLevel(1) returned empty")
	}
	for _, f := range fishes {
		if f.Level != 1 {
			t.Errorf("fish %d: Level = %d; want 1", f.ID, f.Level)
		}
	}
}

func TestGetFishByLevel_Empty(t *testing.T) {
	t.Parallel()

	fishes := GetFishByLevel(999)
	if len(fishes) != 0 {
		t.Errorf("GetFishByLevel(999) returned %d; want 0", len(fishes))
	}
}

func TestGetFishByLevelAndGrade(t *testing.T) {
	t.Parallel()

	fishes := GetFishByLevelAndGrade(1, FishGradeNormal)
	if len(fishes) == 0 {
		t.Fatal("GetFishByLevelAndGrade(1, normal) returned empty")
	}
	for _, f := range fishes {
		if f.Level != 1 {
			t.Errorf("fish %d: Level = %d; want 1", f.ID, f.Level)
		}
		if f.FishGrade != FishGradeNormal {
			t.Errorf("fish %d: FishGrade = %d; want %d", f.ID, f.FishGrade, FishGradeNormal)
		}
	}
}

func TestGetFishByLevelAndType(t *testing.T) {
	t.Parallel()

	fishes := GetFishByLevelAndType(1, "swift")
	if len(fishes) == 0 {
		t.Fatal("GetFishByLevelAndType(1, swift) returned empty")
	}
	for _, f := range fishes {
		if f.FishType != "swift" {
			t.Errorf("fish %d: FishType = %q; want swift", f.ID, f.FishType)
		}
	}
}

func TestGetFishingRod_Valid(t *testing.T) {
	t.Parallel()

	rod := GetFishingRod(6529)
	if rod == nil {
		t.Fatal("GetFishingRod(6529) = nil; want non-nil")
	}
	if rod.ItemID != 6529 {
		t.Errorf("ItemID = %d; want 6529", rod.ItemID)
	}
	if rod.Level != 20 {
		t.Errorf("Level = %d; want 20", rod.Level)
	}
	if rod.Damage != 20.0 {
		t.Errorf("Damage = %.1f; want 20.0", rod.Damage)
	}
	if rod.Name != "Baby Duck Rod" {
		t.Errorf("Name = %q; want Baby Duck Rod", rod.Name)
	}
}

func TestGetFishingRod_NotFound(t *testing.T) {
	t.Parallel()

	rod := GetFishingRod(99999)
	if rod != nil {
		t.Error("GetFishingRod(99999) = non-nil; want nil")
	}
}

func TestGetAllFishingRods(t *testing.T) {
	t.Parallel()

	rods := GetAllFishingRods()
	if len(rods) != 6 {
		t.Errorf("GetAllFishingRods() len = %d; want 6", len(rods))
	}
}

func TestGetFishingMonster_Valid(t *testing.T) {
	t.Parallel()

	m := GetFishingMonster(10)
	if m == nil {
		t.Fatal("GetFishingMonster(10) = nil; want non-nil")
	}
	if m.MonsterID != 18319 {
		t.Errorf("MonsterID = %d; want 18319", m.MonsterID)
	}
	if m.Chance != 5 {
		t.Errorf("Chance = %d; want 5", m.Chance)
	}
}

func TestGetFishingMonster_HighLevel(t *testing.T) {
	t.Parallel()

	m := GetFishingMonster(80)
	if m == nil {
		t.Fatal("GetFishingMonster(80) = nil; want non-nil")
	}
	if m.MonsterID != 18326 {
		t.Errorf("MonsterID = %d; want 18326", m.MonsterID)
	}
}

func TestGetFishingMonster_OutOfRange(t *testing.T) {
	t.Parallel()

	m := GetFishingMonster(1)
	if m != nil {
		t.Error("GetFishingMonster(1) = non-nil; want nil (below min level 6)")
	}

	m = GetFishingMonster(100)
	if m != nil {
		t.Error("GetFishingMonster(100) = non-nil; want nil (above max level 85)")
	}
}

func TestGetAllFishingMonsters(t *testing.T) {
	t.Parallel()

	monsters := GetAllFishingMonsters()
	if len(monsters) != 8 {
		t.Errorf("GetAllFishingMonsters() len = %d; want 8", len(monsters))
	}
}

func TestFishGradeConstants(t *testing.T) {
	t.Parallel()

	if FishGradeEasy != 0 {
		t.Errorf("FishGradeEasy = %d; want 0", FishGradeEasy)
	}
	if FishGradeNormal != 1 {
		t.Errorf("FishGradeNormal = %d; want 1", FishGradeNormal)
	}
	if FishGradeHard != 2 {
		t.Errorf("FishGradeHard = %d; want 2", FishGradeHard)
	}
}
