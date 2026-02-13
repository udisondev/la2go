package data

import "testing"

// TestMain is in recipe_accessors_test.go — loads both recipes and henna templates.

func TestLoadHennaTemplates(t *testing.T) {
	t.Parallel()

	if len(HennaTable) == 0 {
		t.Fatal("HennaTable is empty after load")
	}

	if len(HennaTable) != 180 {
		t.Errorf("HennaTable length = %d; want 180", len(HennaTable))
	}

	if len(hennaByClass) == 0 {
		t.Fatal("hennaByClass is empty after load")
	}
}

func TestGetHennaDef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dyeID   int32
		wantNil bool
		wantSTR int32
	}{
		{
			name:    "dyeID=1 (STR+1, CON-3)",
			dyeID:   1,
			wantSTR: 1,
		},
		{
			name:    "dyeID=3 (STR-3, CON+1)",
			dyeID:   3,
			wantSTR: -3,
		},
		{
			name:    "nonexistent dyeID",
			dyeID:   99999,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			def := GetHennaDef(tt.dyeID)
			if tt.wantNil {
				if def != nil {
					t.Errorf("GetHennaDef(%d) = %v; want nil", tt.dyeID, def)
				}
				return
			}

			if def == nil {
				t.Fatalf("GetHennaDef(%d) = nil; want non-nil", tt.dyeID)
			}

			if def.StatSTR() != tt.wantSTR {
				t.Errorf("GetHennaDef(%d).StatSTR() = %d; want %d", tt.dyeID, def.StatSTR(), tt.wantSTR)
			}
		})
	}
}

func TestGetHennaListForClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		classID int32
		wantNil bool
		wantMin int
	}{
		{
			name:    "classID=11 has hennas",
			classID: 11,
			wantMin: 6,
		},
		{
			name:    "classID=999 has no hennas",
			classID: 999,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			list := GetHennaListForClass(tt.classID)
			if tt.wantNil {
				if list != nil {
					t.Errorf("GetHennaListForClass(%d) = %d items; want nil", tt.classID, len(list))
				}
				return
			}

			if list == nil {
				t.Fatalf("GetHennaListForClass(%d) = nil; want non-nil", tt.classID)
			}

			if len(list) < tt.wantMin {
				t.Errorf("GetHennaListForClass(%d) = %d items; want >= %d", tt.classID, len(list), tt.wantMin)
			}
		})
	}
}

func TestHennaDefAccessors(t *testing.T) {
	t.Parallel()

	def := GetHennaDef(1)
	if def == nil {
		t.Fatal("GetHennaDef(1) = nil")
	}

	if def.DyeID() != 1 {
		t.Errorf("DyeID() = %d; want 1", def.DyeID())
	}
	if def.DyeName() != "dye_s1c3_d" {
		t.Errorf("DyeName() = %q; want %q", def.DyeName(), "dye_s1c3_d")
	}
	if def.DyeItemID() != 4445 {
		t.Errorf("DyeItemID() = %d; want 4445", def.DyeItemID())
	}
	if def.StatSTR() != 1 {
		t.Errorf("StatSTR() = %d; want 1", def.StatSTR())
	}
	if def.StatCON() != -3 {
		t.Errorf("StatCON() = %d; want -3", def.StatCON())
	}
	if def.StatDEX() != 0 {
		t.Errorf("StatDEX() = %d; want 0", def.StatDEX())
	}
	if def.StatINT() != 0 {
		t.Errorf("StatINT() = %d; want 0", def.StatINT())
	}
	if def.StatMEN() != 0 {
		t.Errorf("StatMEN() = %d; want 0", def.StatMEN())
	}
	if def.StatWIT() != 0 {
		t.Errorf("StatWIT() = %d; want 0", def.StatWIT())
	}
	if def.WearCount() != 10 {
		t.Errorf("WearCount() = %d; want 10", def.WearCount())
	}
	if def.WearFee() != 37000 {
		t.Errorf("WearFee() = %d; want 37000", def.WearFee())
	}
	if def.CancelCount() != 5 {
		t.Errorf("CancelCount() = %d; want 5", def.CancelCount())
	}
	if def.CancelFee() != 7400 {
		t.Errorf("CancelFee() = %d; want 7400", def.CancelFee())
	}
}

func TestHennaDefIsAllowedClass(t *testing.T) {
	t.Parallel()

	def := GetHennaDef(7) // classIDs: [11, 26, 39]
	if def == nil {
		t.Fatal("GetHennaDef(7) = nil")
	}

	if !def.IsAllowedClass(11) {
		t.Error("IsAllowedClass(11) = false; want true")
	}
	if !def.IsAllowedClass(26) {
		t.Error("IsAllowedClass(26) = false; want true")
	}
	if def.IsAllowedClass(1) {
		t.Error("IsAllowedClass(1) = true; want false (not in list)")
	}
}

func TestGetHennaInfo(t *testing.T) {
	t.Parallel()

	info := GetHennaInfo(1)
	if info == nil {
		t.Fatal("GetHennaInfo(1) = nil")
	}

	if info.DyeID != 1 {
		t.Errorf("DyeID = %d; want 1", info.DyeID)
	}
	if info.DyeItemID != 4445 {
		t.Errorf("DyeItemID = %d; want 4445", info.DyeItemID)
	}
	if info.StatSTR != 1 {
		t.Errorf("StatSTR = %d; want 1", info.StatSTR)
	}
	if info.WearFee != 37000 {
		t.Errorf("WearFee = %d; want 37000", info.WearFee)
	}

	if GetHennaInfo(99999) != nil {
		t.Error("GetHennaInfo(99999) should be nil")
	}
}

func TestGetHennaInfoListForClass(t *testing.T) {
	t.Parallel()

	list := GetHennaInfoListForClass(11)
	if len(list) == 0 {
		t.Fatal("GetHennaInfoListForClass(11) returned empty list")
	}

	for _, h := range list {
		if h.DyeID == 0 {
			t.Error("found HennaInfo with DyeID=0")
		}
	}

	if GetHennaInfoListForClass(999) != nil {
		t.Error("GetHennaInfoListForClass(999) should be nil")
	}
}

func TestHennaInfoIsAllowedClass(t *testing.T) {
	t.Parallel()

	info := GetHennaInfo(7)
	if info == nil {
		t.Fatal("GetHennaInfo(7) = nil")
	}

	if !info.IsAllowedClass(11) {
		t.Error("HennaInfo.IsAllowedClass(11) = false; want true")
	}
	if info.IsAllowedClass(1) {
		t.Error("HennaInfo.IsAllowedClass(1) = true; want false")
	}
}
