package augment

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestIsLifeStone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		itemID int32
		want   bool
	}{
		{"min life stone", LifeStoneMin, true},
		{"max life stone", LifeStoneMax, true},
		{"mid life stone", 8740, true},
		{"below range", LifeStoneMin - 1, false},
		{"above range", LifeStoneMax + 1, false},
		{"zero", 0, false},
		{"random item", 57, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsLifeStone(tt.itemID); got != tt.want {
				t.Errorf("IsLifeStone(%d) = %v, want %v", tt.itemID, got, tt.want)
			}
		})
	}
}

func TestLifeStoneGrade(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		itemID int32
		want   int32
	}{
		{"no-grade first", 8723, GradeNone},
		{"no-grade last", 8732, GradeNone},
		{"mid first", 8733, GradeMid},
		{"mid last", 8742, GradeMid},
		{"high first", 8743, GradeHigh},
		{"high last", 8752, GradeHigh},
		{"top first", 8753, GradeTop},
		{"top last", 8762, GradeTop},
		{"invalid below", 8722, -1},
		{"invalid above", 8763, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := LifeStoneGrade(tt.itemID); got != tt.want {
				t.Errorf("LifeStoneGrade(%d) = %d, want %d", tt.itemID, got, tt.want)
			}
		})
	}
}

func TestLifeStoneLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		itemID int32
		want   int32
	}{
		{"level 0", 8723, 0},
		{"level 5", 8728, 5},
		{"level 9", 8732, 9},
		{"mid level 0", 8733, 0},
		{"top level 9", 8762, 9},
		{"invalid", 8722, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := LifeStoneLevel(tt.itemID); got != tt.want {
				t.Errorf("LifeStoneLevel(%d) = %d, want %d", tt.itemID, got, tt.want)
			}
		})
	}
}

func TestGemstoneRequirement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		grade     model.CrystalType
		wantID    int32
		wantCount int32
	}{
		{model.CrystalC, GemstoneD, 20},
		{model.CrystalB, GemstoneD, 30},
		{model.CrystalA, GemstoneC, 20},
		{model.CrystalS, GemstoneC, 25},
		{model.CrystalNone, 0, 0},
		{model.CrystalD, 0, 0},
	}
	for _, tc := range tests {
		gemID, gemCount := GemstoneRequirement(tc.grade)
		if gemID != tc.wantID {
			t.Errorf("GemstoneRequirement(%v) gemID = %d, want %d", tc.grade, gemID, tc.wantID)
		}
		if gemCount != tc.wantCount {
			t.Errorf("GemstoneRequirement(%v) gemCount = %d, want %d", tc.grade, gemCount, tc.wantCount)
		}
	}
}

func makeTestWeapon(t *testing.T, augID int32) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID: 100,
		Name:   "Test Sword",
		Type:   model.ItemTypeWeapon,
		PAtk:   50,
	}
	item, err := model.NewItem(1, 100, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error = %v", err)
	}
	if augID > 0 {
		item.SetAugmentationID(augID)
	}
	return item
}

func makeTestArmor(t *testing.T) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID: 200,
		Name:   "Test Shield",
		Type:   model.ItemTypeArmor,
	}
	item, err := model.NewItem(2, 200, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error = %v", err)
	}
	return item
}

func makeTestLifeStone(t *testing.T) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID: LifeStoneMin,
		Name:   "Life Stone",
		Type:   model.ItemTypeEtcItem,
	}
	item, err := model.NewItem(10, LifeStoneMin, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error = %v", err)
	}
	return item
}

func TestService_ValidateTarget(t *testing.T) {
	t.Parallel()

	svc := NewService()

	tests := []struct {
		name    string
		item    *model.Item
		wantErr error
	}{
		{"valid weapon", makeTestWeapon(t, 0), nil},
		{"nil item", nil, nil}, // nil handled separately
		{"not weapon", makeTestArmor(t), ErrNotWeapon},
		{"already augmented", makeTestWeapon(t, 1234), ErrAlreadyAugmented},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := svc.ValidateTarget(tt.item)
			if tt.item == nil {
				if err == nil {
					t.Error("ValidateTarget(nil) = nil, want error")
				}
				return
			}
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ValidateTarget() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateTarget() unexpected error = %v", err)
			}
		})
	}
}

func TestService_ValidateRefiner(t *testing.T) {
	t.Parallel()

	svc := NewService()

	t.Run("valid life stone", func(t *testing.T) {
		t.Parallel()
		ls := makeTestLifeStone(t)
		if err := svc.ValidateRefiner(ls); err != nil {
			t.Errorf("ValidateRefiner() = %v, want nil", err)
		}
	})

	t.Run("nil item", func(t *testing.T) {
		t.Parallel()
		if err := svc.ValidateRefiner(nil); err == nil {
			t.Error("ValidateRefiner(nil) = nil, want error")
		}
	})

	t.Run("not life stone", func(t *testing.T) {
		t.Parallel()
		armor := makeTestArmor(t)
		if err := svc.ValidateRefiner(armor); err != ErrInvalidLifeStone {
			t.Errorf("ValidateRefiner(armor) = %v, want %v", err, ErrInvalidLifeStone)
		}
	})
}

func TestService_Augment(t *testing.T) {
	t.Parallel()

	svc := NewService()
	weapon := makeTestWeapon(t, 0)

	augID, err := svc.Augment(weapon, LifeStoneMin)
	if err != nil {
		t.Fatalf("Augment() error = %v", err)
	}
	if augID <= 0 {
		t.Errorf("Augment() augID = %d, want > 0", augID)
	}
	if weapon.AugmentationID() != augID {
		t.Errorf("weapon.AugmentationID() = %d, want %d", weapon.AugmentationID(), augID)
	}
}

func TestService_Augment_AlreadyAugmented(t *testing.T) {
	t.Parallel()

	svc := NewService()
	weapon := makeTestWeapon(t, 5000)

	_, err := svc.Augment(weapon, LifeStoneMin)
	if err == nil {
		t.Error("Augment() on already-augmented weapon should fail")
	}
}

func TestService_Augment_InvalidLifeStone(t *testing.T) {
	t.Parallel()

	svc := NewService()
	weapon := makeTestWeapon(t, 0)

	_, err := svc.Augment(weapon, 999)
	if err != ErrInvalidLifeStone {
		t.Errorf("Augment(invalid LS) = %v, want %v", err, ErrInvalidLifeStone)
	}
}

func TestService_RemoveAugmentation(t *testing.T) {
	t.Parallel()

	svc := NewService()
	weapon := makeTestWeapon(t, 12345)

	oldID, err := svc.RemoveAugmentation(weapon)
	if err != nil {
		t.Fatalf("RemoveAugmentation() error = %v", err)
	}
	if oldID != 12345 {
		t.Errorf("RemoveAugmentation() oldID = %d, want 12345", oldID)
	}
	if weapon.AugmentationID() != 0 {
		t.Errorf("weapon.AugmentationID() after remove = %d, want 0", weapon.AugmentationID())
	}
}

func TestService_RemoveAugmentation_NotAugmented(t *testing.T) {
	t.Parallel()

	svc := NewService()
	weapon := makeTestWeapon(t, 0)

	_, err := svc.RemoveAugmentation(weapon)
	if err != ErrNotAugmented {
		t.Errorf("RemoveAugmentation(non-augmented) = %v, want %v", err, ErrNotAugmented)
	}
}

func TestService_RemoveAugmentation_NotWeapon(t *testing.T) {
	t.Parallel()

	svc := NewService()
	armor := makeTestArmor(t)

	_, err := svc.RemoveAugmentation(armor)
	if err != ErrNotWeapon {
		t.Errorf("RemoveAugmentation(armor) = %v, want %v", err, ErrNotWeapon)
	}
}

func TestGenerateAugmentID_Distribution(t *testing.T) {
	t.Parallel()

	// Verify generated IDs are within valid range
	for range 100 {
		id := generateAugmentID(GradeNone)
		if id <= 0 || id > 38440 {
			t.Errorf("generateAugmentID(None) = %d, want 1-38440", id)
		}
	}

	for range 100 {
		id := generateAugmentID(GradeTop)
		if id <= 0 || id > 38440 {
			t.Errorf("generateAugmentID(Top) = %d, want 1-38440", id)
		}
	}
}
