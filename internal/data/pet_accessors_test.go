package data

import "testing"

func TestGetPetInfo_Valid(t *testing.T) {
	t.Parallel()

	if len(PetTable) == 0 {
		t.Skip("no pet data loaded")
	}

	// Wolf (npcID 12077) — всегда присутствует в данных
	got := GetPetInfo(12077)
	if got == nil {
		t.Fatal("GetPetInfo(12077) = nil, want non-nil")
	}
	if got.NpcID != 12077 {
		t.Errorf("GetPetInfo(12077).NpcID = %d, want 12077", got.NpcID)
	}
	if got.ItemID != 2375 {
		t.Errorf("GetPetInfo(12077).ItemID = %d, want 2375", got.ItemID)
	}
}

func TestGetPetInfo_Invalid(t *testing.T) {
	t.Parallel()

	got := GetPetInfo(-99999)
	if got != nil {
		t.Errorf("GetPetInfo(-99999) = %+v, want nil", got)
	}
}

func TestGetPetByControlItem_Valid(t *testing.T) {
	t.Parallel()

	// itemID 2375 — Wolf collar
	got := GetPetByControlItem(2375)
	if got == nil {
		t.Fatal("GetPetByControlItem(2375) = nil, want non-nil")
	}
	if got.NpcID != 12077 {
		t.Errorf("GetPetByControlItem(2375).NpcID = %d, want 12077", got.NpcID)
	}
}

func TestGetPetByControlItem_Invalid(t *testing.T) {
	t.Parallel()

	got := GetPetByControlItem(-99999)
	if got != nil {
		t.Errorf("GetPetByControlItem(-99999) = %+v, want nil", got)
	}
}

func TestGetPetLevelInfo_Valid(t *testing.T) {
	t.Parallel()

	got := GetPetLevelInfo(12077, 1)
	if got == nil {
		t.Fatal("GetPetLevelInfo(12077, 1) = nil, want non-nil")
	}
	if got.Level != 1 {
		t.Errorf("Level = %d, want 1", got.Level)
	}
	if got.Exp != 0 {
		t.Errorf("Exp = %d, want 0 (level 1)", got.Exp)
	}
	if got.MaxHP <= 0 {
		t.Errorf("MaxHP = %f, want > 0", got.MaxHP)
	}
	if got.MaxFeed <= 0 {
		t.Errorf("MaxFeed = %d, want > 0", got.MaxFeed)
	}
	if got.FeedRate <= 0 {
		t.Errorf("FeedRate = %f, want > 0", got.FeedRate)
	}
}

func TestGetPetLevelInfo_InvalidNpc(t *testing.T) {
	t.Parallel()

	got := GetPetLevelInfo(-1, 1)
	if got != nil {
		t.Errorf("GetPetLevelInfo(-1, 1) = %+v, want nil", got)
	}
}

func TestGetPetLevelInfo_InvalidLevel(t *testing.T) {
	t.Parallel()

	got := GetPetLevelInfo(12077, 999)
	if got != nil {
		t.Errorf("GetPetLevelInfo(12077, 999) = %+v, want nil", got)
	}
}

func TestGetPetMaxLevel_Valid(t *testing.T) {
	t.Parallel()

	got := GetPetMaxLevel(12077)
	if got <= 0 {
		t.Errorf("GetPetMaxLevel(12077) = %d, want > 0", got)
	}
}

func TestGetPetMaxLevel_Invalid(t *testing.T) {
	t.Parallel()

	got := GetPetMaxLevel(-1)
	if got != 0 {
		t.Errorf("GetPetMaxLevel(-1) = %d, want 0", got)
	}
}

func TestGetPetExpForLevel_Valid(t *testing.T) {
	t.Parallel()

	got := GetPetExpForLevel(12077, 1)
	if got != 0 {
		t.Errorf("GetPetExpForLevel(12077, 1) = %d, want 0", got)
	}

	got2 := GetPetExpForLevel(12077, 2)
	if got2 <= 0 {
		t.Errorf("GetPetExpForLevel(12077, 2) = %d, want > 0", got2)
	}
}

func TestGetPetExpForLevel_Invalid(t *testing.T) {
	t.Parallel()

	got := GetPetExpForLevel(-1, 1)
	if got != -1 {
		t.Errorf("GetPetExpForLevel(-1, 1) = %d, want -1", got)
	}
}

func TestGetPetSkills(t *testing.T) {
	t.Parallel()

	// Просто проверяем что не паникует, не все петы имеют скиллы
	_ = GetPetSkills(12077)
	_ = GetPetSkills(-1)
}

func TestIsPetNpc(t *testing.T) {
	t.Parallel()

	if !IsPetNpc(12077) {
		t.Error("IsPetNpc(12077) = false, want true")
	}
	if IsPetNpc(-1) {
		t.Error("IsPetNpc(-1) = true, want false")
	}
}

func TestIsPetControlItem(t *testing.T) {
	t.Parallel()

	if !IsPetControlItem(2375) {
		t.Error("IsPetControlItem(2375) = false, want true")
	}
	if IsPetControlItem(-99999) {
		t.Error("IsPetControlItem(-99999) = true, want false")
	}
}
