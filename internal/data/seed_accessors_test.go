package data

import "testing"

func TestSeedDefToTemplate_Defaults(t *testing.T) {
	t.Parallel()

	def := &seedDef{
		castleID: 1,
		cropID:   5073,
		seedID:   5016,
		matureID: 5103,
		reward1:  1864,
		reward2:  1878,
		level:    10,
		// limitSeeds и limitCrops = 0 → используются дефолты.
	}

	tpl := seedDefToTemplate(def)
	if tpl.SeedID != 5016 {
		t.Errorf("SeedID = %d; want 5016", tpl.SeedID)
	}
	if tpl.CropID != 5073 {
		t.Errorf("CropID = %d; want 5073", tpl.CropID)
	}
	if tpl.CastleID != 1 {
		t.Errorf("CastleID = %d; want 1", tpl.CastleID)
	}
	if tpl.LimitSeeds != defaultLimitSeeds {
		t.Errorf("LimitSeeds = %d; want %d (default)", tpl.LimitSeeds, defaultLimitSeeds)
	}
	if tpl.LimitCrops != defaultLimitCrops {
		t.Errorf("LimitCrops = %d; want %d (default)", tpl.LimitCrops, defaultLimitCrops)
	}
	if tpl.IsAlternative {
		t.Error("IsAlternative = true; want false")
	}
}

func TestSeedDefToTemplate_WithLimits(t *testing.T) {
	t.Parallel()

	def := &seedDef{
		castleID:      2,
		cropID:        5067,
		seedID:        5024,
		matureID:      5097,
		reward1:       1867,
		reward2:       1894,
		level:         19,
		isAlternative: true,
		limitSeeds:    8100,
		limitCrops:    9000,
	}

	tpl := seedDefToTemplate(def)
	if tpl.LimitSeeds != 8100 {
		t.Errorf("LimitSeeds = %d; want 8100", tpl.LimitSeeds)
	}
	if tpl.LimitCrops != 9000 {
		t.Errorf("LimitCrops = %d; want 9000", tpl.LimitCrops)
	}
	if !tpl.IsAlternative {
		t.Error("IsAlternative = false; want true")
	}
}

func TestGetSeedTemplate_AfterLoad(t *testing.T) {
	t.Parallel()

	// SeedTable loaded in TestMain.
	tpl := GetSeedTemplate(5016)
	if tpl == nil {
		t.Fatal("GetSeedTemplate(5016) = nil; want non-nil")
	}
	if tpl.CastleID != 1 {
		t.Errorf("CastleID = %d; want 1", tpl.CastleID)
	}
}

func TestGetSeedTemplate_NotFound(t *testing.T) {
	t.Parallel()

	tpl := GetSeedTemplate(99999)
	if tpl != nil {
		t.Error("GetSeedTemplate(99999) = non-nil; want nil")
	}
}

func TestGetSeedByCropID(t *testing.T) {
	t.Parallel()

	tpl := GetSeedByCropID(5073)
	if tpl == nil {
		t.Fatal("GetSeedByCropID(5073) = nil; want non-nil")
	}
	if tpl.CropID != 5073 {
		t.Errorf("CropID = %d; want 5073", tpl.CropID)
	}
}

func TestGetSeedByCropAndCastle(t *testing.T) {
	t.Parallel()

	tpl := GetSeedByCropAndCastle(5073, 1)
	if tpl == nil {
		t.Fatal("GetSeedByCropAndCastle(5073, 1) = nil; want non-nil")
	}
	if tpl.CastleID != 1 {
		t.Errorf("CastleID = %d; want 1", tpl.CastleID)
	}

	// Несуществующая комбинация (castle 99 не существует).
	tpl = GetSeedByCropAndCastle(5073, 99)
	if tpl != nil {
		t.Error("GetSeedByCropAndCastle(5073, 99) = non-nil; want nil")
	}
}

func TestGetSeedsByCastle(t *testing.T) {
	t.Parallel()

	seeds := GetSeedsByCastle(1)
	if len(seeds) == 0 {
		t.Error("GetSeedsByCastle(1) returned empty; want non-empty")
	}

	for _, s := range seeds {
		if s.CastleID != 1 {
			t.Errorf("seed %d has CastleID = %d; want 1", s.SeedID, s.CastleID)
		}
	}
}

func TestGetAllCropIDs(t *testing.T) {
	t.Parallel()

	cropIDs := GetAllCropIDs()
	if len(cropIDs) == 0 {
		t.Error("GetAllCropIDs() returned empty; want non-empty")
	}
}

func TestGetAllSeedIDs(t *testing.T) {
	t.Parallel()

	seedIDs := GetAllSeedIDs()
	if len(seedIDs) == 0 {
		t.Error("GetAllSeedIDs() returned empty; want non-empty")
	}
}

func TestSeedReferencePrice_NoItemData(t *testing.T) {
	t.Parallel()

	// Без загрузки ItemTable — возвращает 1.
	price := SeedReferencePrice(5016)
	if price < 1 {
		t.Errorf("SeedReferencePrice(5016) = %d; want >= 1", price)
	}
}

func TestSeedMinMaxPrice(t *testing.T) {
	t.Parallel()

	// Min = ref * 0.6, Max = ref * 10.
	min := SeedMinPrice(5016)
	max := SeedMaxPrice(5016)

	if min > max {
		t.Errorf("SeedMinPrice(%d) > SeedMaxPrice(%d)", min, max)
	}
}

func TestCropMinMaxPrice(t *testing.T) {
	t.Parallel()

	min := CropMinPrice(5073)
	max := CropMaxPrice(5073)

	if min > max {
		t.Errorf("CropMinPrice(%d) > CropMaxPrice(%d)", min, max)
	}
}
