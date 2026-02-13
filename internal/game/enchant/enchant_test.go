package enchant

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

// --- helpers ---

func makeWeapon(t *testing.T, grade model.CrystalType, enchant int32) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID:      100,
		Name:        "Test Sword",
		Type:        model.ItemTypeWeapon,
		CrystalType: grade,
		BodyPartStr: "rhand",
	}
	item, err := model.NewItem(1, 100, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	if enchant > 0 {
		if err := item.SetEnchant(enchant); err != nil {
			t.Fatalf("SetEnchant(%d): %v", enchant, err)
		}
	}
	return item
}

func makeArmor(t *testing.T, grade model.CrystalType, bodyPart string, enchant int32) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID:      200,
		Name:        "Test Armor",
		Type:        model.ItemTypeArmor,
		CrystalType: grade,
		BodyPartStr: bodyPart,
	}
	item, err := model.NewItem(2, 200, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	if enchant > 0 {
		if err := item.SetEnchant(enchant); err != nil {
			t.Fatalf("SetEnchant(%d): %v", enchant, err)
		}
	}
	return item
}

func makeConsumable(t *testing.T) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID: 300,
		Name:   "Healing Potion",
		Type:   model.ItemTypeConsumable,
	}
	item, err := model.NewItem(3, 300, 1, 10, tmpl)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	return item
}

// --- IsScroll tests ---

func TestIsScroll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		itemID     int32
		wantFound  bool
		wantWeapon bool
		wantGrade  model.CrystalType
		wantType   ScrollType
	}{
		{"normal weapon A-grade", 729, true, true, model.CrystalA, ScrollNormal},
		{"normal weapon S-grade", 959, true, true, model.CrystalS, ScrollNormal},
		{"normal armor D-grade", 956, true, false, model.CrystalD, ScrollNormal},
		{"blessed weapon B-grade", 6571, true, true, model.CrystalB, ScrollBlessed},
		{"blessed armor C-grade", 6574, true, false, model.CrystalC, ScrollBlessed},
		{"crystal weapon A-grade", 731, true, true, model.CrystalA, ScrollCrystal},
		{"crystal armor S-grade", 962, true, false, model.CrystalS, ScrollCrystal},
		{"not a scroll", 57, false, false, model.CrystalNone, ScrollNormal},
		{"adena", 0, false, false, model.CrystalNone, ScrollNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info, found := IsScroll(tt.itemID)
			if found != tt.wantFound {
				t.Errorf("IsScroll(%d) found = %v; want %v", tt.itemID, found, tt.wantFound)
			}
			if !found {
				return
			}
			if info.IsWeapon != tt.wantWeapon {
				t.Errorf("IsScroll(%d).IsWeapon = %v; want %v", tt.itemID, info.IsWeapon, tt.wantWeapon)
			}
			if info.Grade != tt.wantGrade {
				t.Errorf("IsScroll(%d).Grade = %v; want %v", tt.itemID, info.Grade, tt.wantGrade)
			}
			if info.ScrollType != tt.wantType {
				t.Errorf("IsScroll(%d).ScrollType = %v; want %v", tt.itemID, info.ScrollType, tt.wantType)
			}
		})
	}
}

// --- IsEnchantable tests ---

func TestIsEnchantable(t *testing.T) {
	t.Parallel()

	weapon := makeWeapon(t, model.CrystalA, 0)
	armor := makeArmor(t, model.CrystalB, "chest", 0)
	consumable := makeConsumable(t)

	tests := []struct {
		name string
		item *model.Item
		want bool
	}{
		{"weapon", weapon, true},
		{"armor", armor, true},
		{"consumable", consumable, false},
		{"nil item", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsEnchantable(tt.item)
			if got != tt.want {
				t.Errorf("IsEnchantable(%s) = %v; want %v", tt.name, got, tt.want)
			}
		})
	}
}

// --- IsFullArmor tests ---

func TestIsFullArmor(t *testing.T) {
	t.Parallel()

	fullArmor := makeArmor(t, model.CrystalA, "fullarmor", 0)
	chest := makeArmor(t, model.CrystalA, "chest", 0)
	weapon := makeWeapon(t, model.CrystalA, 0)

	tests := []struct {
		name string
		item *model.Item
		want bool
	}{
		{"full armor", fullArmor, true},
		{"chest armor", chest, false},
		{"weapon", weapon, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsFullArmor(tt.item)
			if got != tt.want {
				t.Errorf("IsFullArmor(%s) = %v; want %v", tt.name, got, tt.want)
			}
		})
	}
}

// --- Validate tests ---

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		scrollID   int32
		item       *model.Item
		wantReason string
	}{
		{
			name:     "valid weapon scroll on weapon",
			scrollID: 729,
			item:     makeWeapon(t, model.CrystalA, 0),
		},
		{
			name:     "valid armor scroll on armor",
			scrollID: 730,
			item:     makeArmor(t, model.CrystalA, "chest", 0),
		},
		{
			name:       "weapon scroll on armor",
			scrollID:   729,
			item:       makeArmor(t, model.CrystalA, "chest", 0),
			wantReason: "weapon scroll on non-weapon item",
		},
		{
			name:       "armor scroll on weapon",
			scrollID:   730,
			item:       makeWeapon(t, model.CrystalA, 0),
			wantReason: "armor scroll on non-armor item",
		},
		{
			name:       "grade mismatch: A scroll on D weapon",
			scrollID:   729,
			item:       makeWeapon(t, model.CrystalD, 0),
			wantReason: "scroll grade mismatch",
		},
		{
			name:       "grade mismatch: D scroll on S armor",
			scrollID:   956,
			item:       makeArmor(t, model.CrystalS, "chest", 0),
			wantReason: "scroll grade mismatch",
		},
		{
			name:       "consumable item",
			scrollID:   729,
			item:       makeConsumable(t),
			wantReason: "item not enchantable",
		},
		{
			name:       "max enchant reached",
			scrollID:   729,
			item:       makeWeapon(t, model.CrystalA, 16),
			wantReason: "max enchant level reached",
		},
		{
			name:     "enchant 15 is allowed",
			scrollID: 729,
			item:     makeWeapon(t, model.CrystalA, 15),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scroll, ok := IsScroll(tt.scrollID)
			if !ok {
				t.Fatalf("scroll %d not found", tt.scrollID)
			}
			reason := Validate(scroll, tt.item)
			if reason != tt.wantReason {
				t.Errorf("Validate() = %q; want %q", reason, tt.wantReason)
			}
		})
	}
}

// --- SuccessChance tests ---

func TestSuccessChance(t *testing.T) {
	t.Parallel()

	normalWeaponScroll := ScrollInfo{IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: 16}
	normalArmorScroll := ScrollInfo{IsWeapon: false, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: 16}
	crystalScroll := ScrollInfo{IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollCrystal}

	tests := []struct {
		name       string
		scroll     ScrollInfo
		item       *model.Item
		wantChance int
	}{
		// Weapon safe zone: enchant 0, 1, 2 → 100%
		{"weapon +0 → +1", normalWeaponScroll, makeWeapon(t, model.CrystalA, 0), 100},
		{"weapon +1 → +2", normalWeaponScroll, makeWeapon(t, model.CrystalA, 1), 100},
		{"weapon +2 → +3", normalWeaponScroll, makeWeapon(t, model.CrystalA, 2), 100},
		// Weapon risky zone: enchant 3+ → 66%
		{"weapon +3 → +4", normalWeaponScroll, makeWeapon(t, model.CrystalA, 3), 66},
		{"weapon +10 → +11", normalWeaponScroll, makeWeapon(t, model.CrystalA, 10), 66},
		{"weapon +15 → +16", normalWeaponScroll, makeWeapon(t, model.CrystalA, 15), 66},

		// Regular armor safe zone: enchant 0, 1, 2 → 100%
		{"armor chest +0 → +1", normalArmorScroll, makeArmor(t, model.CrystalA, "chest", 0), 100},
		{"armor chest +2 → +3", normalArmorScroll, makeArmor(t, model.CrystalA, "chest", 2), 100},
		// Regular armor risky zone: enchant 3+ → 66%
		{"armor chest +3 → +4", normalArmorScroll, makeArmor(t, model.CrystalA, "chest", 3), 66},

		// Full armor extended safe zone: enchant 0, 1, 2, 3 → 100%
		{"fullarmor +0 → +1", normalArmorScroll, makeArmor(t, model.CrystalA, "fullarmor", 0), 100},
		{"fullarmor +3 → +4", normalArmorScroll, makeArmor(t, model.CrystalA, "fullarmor", 3), 100},
		// Full armor risky zone: enchant 4+ → 66%
		{"fullarmor +4 → +5", normalArmorScroll, makeArmor(t, model.CrystalA, "fullarmor", 4), 66},

		// Crystal scroll — always 100%
		{"crystal +0", crystalScroll, makeWeapon(t, model.CrystalA, 0), 100},
		{"crystal +10", crystalScroll, makeWeapon(t, model.CrystalA, 10), 100},
		{"crystal +15", crystalScroll, makeWeapon(t, model.CrystalA, 15), 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SuccessChance(tt.scroll, tt.item)
			if got != tt.wantChance {
				t.Errorf("SuccessChance() = %d; want %d", got, tt.wantChance)
			}
		})
	}
}

// --- TryEnchantWithRoll tests ---

func TestTryEnchant_SafeZone(t *testing.T) {
	t.Parallel()

	scroll := ScrollInfo{IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: 16}

	// В safe zone (enchant < 3) — всегда успех, даже с roll=99
	for enchant := int32(0); enchant < 3; enchant++ {
		item := makeWeapon(t, model.CrystalA, enchant)
		result := TryEnchantWithRoll(scroll, item, 99)
		if !result.Success {
			t.Errorf("TryEnchant(+%d, roll=99) failed; want success (safe zone)", enchant)
		}
		if result.NewEnchant != enchant+1 {
			t.Errorf("TryEnchant(+%d).NewEnchant = %d; want %d", enchant, result.NewEnchant, enchant+1)
		}
	}
}

func TestTryEnchant_FullArmorSafeZone(t *testing.T) {
	t.Parallel()

	scroll := ScrollInfo{IsWeapon: false, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: 16}

	// Full armor safe zone: enchant 0, 1, 2, 3 → всегда успех
	for enchant := int32(0); enchant < 4; enchant++ {
		item := makeArmor(t, model.CrystalA, "fullarmor", enchant)
		result := TryEnchantWithRoll(scroll, item, 99)
		if !result.Success {
			t.Errorf("TryEnchant(fullarmor +%d, roll=99) failed; want success (safe zone)", enchant)
		}
		if result.NewEnchant != enchant+1 {
			t.Errorf("TryEnchant(fullarmor +%d).NewEnchant = %d; want %d", enchant, result.NewEnchant, enchant+1)
		}
	}
}

func TestTryEnchant_NormalScroll_Success(t *testing.T) {
	t.Parallel()

	scroll := ScrollInfo{IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: 16}
	item := makeWeapon(t, model.CrystalA, 5)

	// roll=65 < 66 → успех
	result := TryEnchantWithRoll(scroll, item, 65)
	if !result.Success {
		t.Error("TryEnchant(+5, roll=65) failed; want success (roll < 66)")
	}
	if result.NewEnchant != 6 {
		t.Errorf("NewEnchant = %d; want 6", result.NewEnchant)
	}
	if result.Destroyed {
		t.Error("Destroyed = true; want false")
	}
}

func TestTryEnchant_NormalScroll_Fail(t *testing.T) {
	t.Parallel()

	scroll := ScrollInfo{IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: 16}
	item := makeWeapon(t, model.CrystalA, 5)

	// roll=66 >= 66 → провал, normal → destroyed
	result := TryEnchantWithRoll(scroll, item, 66)
	if result.Success {
		t.Error("TryEnchant(+5, roll=66) succeeded; want failure (roll >= 66)")
	}
	if !result.Destroyed {
		t.Error("Destroyed = false; want true (normal scroll)")
	}
}

func TestTryEnchant_BlessedScroll_Fail(t *testing.T) {
	t.Parallel()

	scroll := ScrollInfo{IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollBlessed, MaxEnchant: 16}
	item := makeWeapon(t, model.CrystalA, 10)

	// roll=99 >= 66 → провал, blessed → enchant reset to 0
	result := TryEnchantWithRoll(scroll, item, 99)
	if result.Success {
		t.Error("TryEnchant(blessed +10, roll=99) succeeded; want failure")
	}
	if result.NewEnchant != 0 {
		t.Errorf("NewEnchant = %d; want 0 (blessed fail)", result.NewEnchant)
	}
	if result.Destroyed {
		t.Error("Destroyed = true; want false (blessed scroll preserves item)")
	}
}

func TestTryEnchant_CrystalScroll_AlwaysSuccess(t *testing.T) {
	t.Parallel()

	scroll := ScrollInfo{IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollCrystal}

	// Crystal scroll всегда 100% даже с roll=99
	for _, enchant := range []int32{0, 3, 5, 10, 15} {
		item := makeWeapon(t, model.CrystalA, enchant)
		result := TryEnchantWithRoll(scroll, item, 99)
		if !result.Success {
			t.Errorf("TryEnchant(crystal +%d, roll=99) failed; want success", enchant)
		}
		if result.NewEnchant != enchant+1 {
			t.Errorf("TryEnchant(crystal +%d).NewEnchant = %d; want %d", enchant, result.NewEnchant, enchant+1)
		}
	}
}

func TestTryEnchant_NormalScroll_BoundaryRoll(t *testing.T) {
	t.Parallel()

	scroll := ScrollInfo{IsWeapon: true, Grade: model.CrystalA, ScrollType: ScrollNormal, MaxEnchant: 16}
	item := makeWeapon(t, model.CrystalA, 5)

	// Граничные значения
	tests := []struct {
		roll        int
		wantSuccess bool
	}{
		{0, true},   // roll=0 < 66 → success
		{65, true},  // roll=65 < 66 → success
		{66, false}, // roll=66 >= 66 → fail
		{99, false}, // roll=99 >= 66 → fail
	}

	for _, tt := range tests {
		result := TryEnchantWithRoll(scroll, item, tt.roll)
		if result.Success != tt.wantSuccess {
			t.Errorf("TryEnchant(+5, roll=%d).Success = %v; want %v", tt.roll, result.Success, tt.wantSuccess)
		}
	}
}

// --- Scroll coverage ---

func TestAllScrollsHaveGrade(t *testing.T) {
	t.Parallel()

	for itemID, info := range scrollTable {
		if info.Grade == model.CrystalNone {
			t.Errorf("scroll %d has CrystalNone grade (all enchant scrolls should have a grade)", itemID)
		}
	}
}

func TestScrollTableCompleteness(t *testing.T) {
	t.Parallel()

	// Проверяем что для каждого грейда (D, C, B, A, S) есть все 6 типов скроллов:
	// normal weapon, normal armor, blessed weapon, blessed armor, crystal weapon, crystal armor
	grades := []model.CrystalType{model.CrystalD, model.CrystalC, model.CrystalB, model.CrystalA, model.CrystalS}

	type scrollKey struct {
		grade    model.CrystalType
		isWeapon bool
		sType    ScrollType
	}

	found := make(map[scrollKey]bool, len(scrollTable))
	for _, info := range scrollTable {
		found[scrollKey{info.Grade, info.IsWeapon, info.ScrollType}] = true
	}

	for _, grade := range grades {
		for _, isWeapon := range []bool{true, false} {
			for _, sType := range []ScrollType{ScrollNormal, ScrollBlessed, ScrollCrystal} {
				key := scrollKey{grade, isWeapon, sType}
				if !found[key] {
					wpn := "armor"
					if isWeapon {
						wpn = "weapon"
					}
					t.Errorf("missing scroll: grade=%s, %s, type=%d", grade, wpn, sType)
				}
			}
		}
	}
}

func TestScrollCount(t *testing.T) {
	t.Parallel()

	// 5 grades * 2 (weapon/armor) * 3 (normal/blessed/crystal) = 30 scrolls
	want := 30
	got := len(scrollTable)
	if got != want {
		t.Errorf("len(scrollTable) = %d; want %d", got, want)
	}
}
