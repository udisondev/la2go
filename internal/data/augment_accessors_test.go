package data

import "testing"

func TestGetAugmentInfo_NotFound(t *testing.T) {
	t.Parallel()

	// Ensure table is initialized
	if AugmentationTable == nil {
		AugmentationTable = make(map[int32]*augmentationDef)
	}

	info := GetAugmentInfo(-1)
	if info != nil {
		t.Errorf("GetAugmentInfo(-1) = %v, want nil", info)
	}
}

func TestGetAugmentInfo_Found(t *testing.T) {
	t.Parallel()

	// Temporarily inject test data
	oldTable := AugmentationTable
	AugmentationTable = map[int32]*augmentationDef{
		14561: {id: 14561, skillID: 3203, skillLevel: 1, augType: "blue"},
	}
	defer func() { AugmentationTable = oldTable }()

	info := GetAugmentInfo(14561)
	if info == nil {
		t.Fatal("GetAugmentInfo(14561) = nil, want info")
	}
	if info.ID != 14561 {
		t.Errorf("ID = %d, want 14561", info.ID)
	}
	if info.SkillID != 3203 {
		t.Errorf("SkillID = %d, want 3203", info.SkillID)
	}
	if info.SkillLevel != 1 {
		t.Errorf("SkillLevel = %d, want 1", info.SkillLevel)
	}
	if info.AugType != "blue" {
		t.Errorf("AugType = %q, want \"blue\"", info.AugType)
	}
}

func TestAugmentHasSkill(t *testing.T) {
	t.Parallel()

	oldTable := AugmentationTable
	AugmentationTable = map[int32]*augmentationDef{
		100: {id: 100, skillID: 50, skillLevel: 1, augType: "blue"},
		200: {id: 200, skillID: 0, skillLevel: 0, augType: "blue"},
	}
	defer func() { AugmentationTable = oldTable }()

	if !AugmentHasSkill(100) {
		t.Error("AugmentHasSkill(100) = false, want true")
	}
	if AugmentHasSkill(200) {
		t.Error("AugmentHasSkill(200) = true, want false")
	}
	if AugmentHasSkill(999) {
		t.Error("AugmentHasSkill(999) = true, want false")
	}
}

func TestAugmentSkill(t *testing.T) {
	t.Parallel()

	oldTable := AugmentationTable
	AugmentationTable = map[int32]*augmentationDef{
		100: {id: 100, skillID: 50, skillLevel: 3, augType: "red"},
	}
	defer func() { AugmentationTable = oldTable }()

	skillID, skillLevel := AugmentSkill(100)
	if skillID != 50 || skillLevel != 3 {
		t.Errorf("AugmentSkill(100) = (%d, %d), want (50, 3)", skillID, skillLevel)
	}

	skillID, skillLevel = AugmentSkill(999)
	if skillID != 0 || skillLevel != 0 {
		t.Errorf("AugmentSkill(999) = (%d, %d), want (0, 0)", skillID, skillLevel)
	}
}
