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

// TestGetPAtk_WithWeapon verifies weapon pAtk integration.
// Phase 5.5: Weapon & Equipment System.
func TestGetPAtk_WithWeapon(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// No weapon: base pAtk only
	// Formula: basePAtk × STRbonus × levelMod
	// Human Fighter level 10: basePAtk=4, STR=40 → STRbonus=1.20, levelMod=0.99
	// Expected: 4 × 1.20 × 0.99 = 4.752 ≈ 4
	pAtkNoWeapon := player.GetPAtk()

	t.Logf("No weapon: pAtk=%d", pAtkNoWeapon)

	// Equip Short Sword (pAtk=8)
	swordTemplate := &ItemTemplate{
		ItemID:      1,
		Name:        "Short Sword",
		Type:        ItemTypeWeapon,
		PAtk:        8,
		AttackRange: 40,
	}

	sword, _ := NewItem(1000, 1, 100, 1, swordTemplate)
	player.Inventory().AddItem(sword)
	_, _ = player.Inventory().EquipItem(sword, PaperdollRHand)

	pAtkWithWeapon := player.GetPAtk()

	t.Logf("With Short Sword (pAtk=8): pAtk=%d", pAtkWithWeapon)

	// Formula: (basePAtk + weaponPAtk) × STRbonus × levelMod
	// Expected: (4 + 8) × 1.20 × 0.99 = 14.256 ≈ 14

	// Weapon should increase pAtk significantly
	if pAtkWithWeapon <= pAtkNoWeapon {
		t.Errorf("Weapon should increase pAtk: with=%d vs without=%d", pAtkWithWeapon, pAtkNoWeapon)
	}

	// Check expected range (14 ± 1)
	if pAtkWithWeapon < 13 || pAtkWithWeapon > 15 {
		t.Errorf("pAtk with weapon = %d, expected 13-15 (formula: (4+8) × 1.20 × 0.99 = 14.256)", pAtkWithWeapon)
	}
}

// TestGetPAtk_DifferentWeapons verifies different weapons give different pAtk.
// Phase 5.5: Weapon & Equipment System.
func TestGetPAtk_DifferentWeapons(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "Test", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Create two weapons with different pAtk
	daggerTemplate := &ItemTemplate{
		ItemID:      1,
		Name:        "Dagger",
		Type:        ItemTypeWeapon,
		PAtk:        5,
		AttackRange: 40,
	}

	greatswordTemplate := &ItemTemplate{
		ItemID:      2,
		Name:        "Greatsword",
		Type:        ItemTypeWeapon,
		PAtk:        15,
		AttackRange: 40,
	}

	dagger, _ := NewItem(1000, 1, 100, 1, daggerTemplate)
	greatsword, _ := NewItem(1001, 2, 100, 1, greatswordTemplate)

	// Equip dagger
	player.Inventory().AddItem(dagger)
	_, _ = player.Inventory().EquipItem(dagger, PaperdollRHand)
	pAtkDagger := player.GetPAtk()

	// Unequip dagger, equip greatsword
	player.Inventory().UnequipItem(PaperdollRHand)
	player.Inventory().AddItem(greatsword)
	_, _ = player.Inventory().EquipItem(greatsword, PaperdollRHand)
	pAtkGreatsword := player.GetPAtk()

	t.Logf("Dagger (pAtk=5): pAtk=%d", pAtkDagger)
	t.Logf("Greatsword (pAtk=15): pAtk=%d", pAtkGreatsword)

	// Greatsword should have significantly higher pAtk
	if pAtkGreatsword <= pAtkDagger {
		t.Errorf("Greatsword pAtk=%d should be > Dagger pAtk=%d", pAtkGreatsword, pAtkDagger)
	}

	// Difference should be proportional to weapon pAtk difference (10)
	// Formula: diff ≈ (15 - 5) × STRbonus × levelMod = 10 × 1.20 × 0.99 = 11.88
	diff := pAtkGreatsword - pAtkDagger
	if diff < 10 || diff > 13 {
		t.Errorf("pAtk difference = %d, expected 10-13 (proportional to weapon pAtk diff)", diff)
	}
}

// TestGetPAtk_NoInventory verifies GetPAtk handles nil inventory gracefully.
// Edge case: inventory not initialized.
func TestGetPAtk_NoInventory(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "Test", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Set inventory to nil (edge case)
	player.inventory = nil

	// Should not panic
	pAtk := player.GetPAtk()

	// Should return base pAtk (no weapon)
	if pAtk < 1 {
		t.Errorf("GetPAtk() with nil inventory = %d, expected > 0", pAtk)
	}
}

// TestGetBasePAtk_Unchanged verifies GetBasePAtk still works after Phase 5.5.
// GetBasePAtk should return base pAtk WITHOUT weapon bonus.
func TestGetBasePAtk_Unchanged(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	basePAtk := player.GetBasePAtk()

	// Equip weapon
	swordTemplate := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
		PAtk:   10,
	}

	sword, _ := NewItem(1000, 1, 100, 1, swordTemplate)
	player.Inventory().AddItem(sword)
	_, _ = player.Inventory().EquipItem(sword, PaperdollRHand)

	// GetBasePAtk should be UNCHANGED (does not include weapon)
	basePAtkAfter := player.GetBasePAtk()

	if basePAtk != basePAtkAfter {
		t.Errorf("GetBasePAtk() changed after equipping weapon: before=%d, after=%d (should be same)", basePAtk, basePAtkAfter)
	}

	// GetPAtk should be HIGHER (includes weapon)
	finalPAtk := player.GetPAtk()

	if finalPAtk <= basePAtk {
		t.Errorf("GetPAtk()=%d should be > GetBasePAtk()=%d (weapon bonus)", finalPAtk, basePAtk)
	}
}

// TestGetPDef_WithArmor verifies armor pDef integration + slot subtraction.
// Phase 5.5: Weapon & Equipment System.
func TestGetPDef_WithArmor(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Nude pDef: basePDef × levelMod
	// Human Fighter level 10: basePDef=80, levelMod=0.99
	// Expected: 80 × 0.99 = 79.2 ≈ 79
	pDefNude := player.GetPDef()

	t.Logf("Nude: pDef=%d", pDefNude)

	// Equip Leather Shirt (chest pDef=43)
	leatherShirtTemplate := &ItemTemplate{
		ItemID:   1,
		Name:     "Leather Shirt",
		Type:     ItemTypeArmor,
		PDef:     43,
		BodyPart: ArmorSlotChest,
	}

	leatherShirt, _ := NewItem(1000, 1, 100, 1, leatherShirtTemplate)
	player.Inventory().AddItem(leatherShirt)
	_, _ = player.Inventory().EquipItem(leatherShirt, PaperdollChest)

	pDefWithArmor := player.GetPDef()

	t.Logf("With Leather Shirt (chest pDef=43): pDef=%d", pDefWithArmor)

	// Formula: (basePDef - slotDef[chest] + armorPDef) × levelMod
	// Human Fighter: chest slot def = 31
	// Expected: (80 - 31 + 43) × 0.99 = 92 × 0.99 = 91.08 ≈ 91

	// Armor should increase pDef
	if pDefWithArmor <= pDefNude {
		t.Errorf("Armor should increase pDef: with=%d vs nude=%d", pDefWithArmor, pDefNude)
	}

	// Check expected range (91 ± 2)
	if pDefWithArmor < 89 || pDefWithArmor > 93 {
		t.Errorf("pDef with armor = %d, expected 89-93 (formula: (80-31+43) × 0.99 = 91.08)", pDefWithArmor)
	}
}

// TestGetPDef_FullPlate verifies multiple armor pieces with slot subtraction.
// Phase 5.5: Weapon & Equipment System.
func TestGetPDef_FullPlate(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pDefNude := player.GetPDef()
	t.Logf("Nude: pDef=%d", pDefNude)

	// Equip Full Plate (chest pDef=100, legs pDef=50)
	chestTemplate := &ItemTemplate{
		ItemID:   1,
		Name:     "Full Plate Armor",
		Type:     ItemTypeArmor,
		PDef:     100,
		BodyPart: ArmorSlotChest,
	}

	legsTemplate := &ItemTemplate{
		ItemID:   2,
		Name:     "Full Plate Legs",
		Type:     ItemTypeArmor,
		PDef:     50,
		BodyPart: ArmorSlotLegs,
	}

	chest, _ := NewItem(1000, 1, 100, 1, chestTemplate)
	legs, _ := NewItem(1001, 2, 100, 1, legsTemplate)

	player.Inventory().AddItem(chest)
	player.Inventory().AddItem(legs)
	_, _ = player.Inventory().EquipItem(chest, PaperdollChest)
	_, _ = player.Inventory().EquipItem(legs, PaperdollLegs)

	pDefFullPlate := player.GetPDef()

	t.Logf("With Full Plate (chest=100, legs=50): pDef=%d", pDefFullPlate)

	// Formula: (basePDef - chest_slot - legs_slot + chest_pDef + legs_pDef) × levelMod
	// Human Fighter: chest=31, legs=18
	// Expected: (80 - 31 - 18 + 100 + 50) × 0.99 = 181 × 0.99 = 179.19 ≈ 179

	// Check expected range (179 ± 3)
	if pDefFullPlate < 176 || pDefFullPlate > 182 {
		t.Errorf("pDef with Full Plate = %d, expected 176-182 (formula: (80-31-18+100+50) × 0.99 = 179.19)", pDefFullPlate)
	}

	// Full Plate should be significantly stronger than nude
	if pDefFullPlate < pDefNude*2 {
		t.Errorf("Full Plate pDef=%d should be > 2× nude pDef=%d", pDefFullPlate, pDefNude)
	}
}

// TestGetPDef_ArmorWeakerThanNude verifies weak armor decreases pDef (slot subtraction).
// Edge case: armor pDef < slot base def.
// Phase 5.5: Weapon & Equipment System.
func TestGetPDef_ArmorWeakerThanNude(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	pDefNude := player.GetPDef()

	// Equip very weak armor (pDef=10, less than chest slot base def=31)
	weakArmorTemplate := &ItemTemplate{
		ItemID:   1,
		Name:     "Cloth Shirt",
		Type:     ItemTypeArmor,
		PDef:     10,
		BodyPart: ArmorSlotChest,
	}

	weakArmor, _ := NewItem(1000, 1, 100, 1, weakArmorTemplate)
	player.Inventory().AddItem(weakArmor)
	_, _ = player.Inventory().EquipItem(weakArmor, PaperdollChest)

	pDefWithWeakArmor := player.GetPDef()

	t.Logf("Nude: pDef=%d", pDefNude)
	t.Logf("With Weak Armor (pDef=10 < chest base=31): pDef=%d", pDefWithWeakArmor)

	// Formula: (80 - 31 + 10) × 0.99 = 59 × 0.99 = 58.41 ≈ 58
	// Weak armor should DECREASE pDef compared to nude

	if pDefWithWeakArmor >= pDefNude {
		t.Errorf("Weak armor should decrease pDef: with=%d vs nude=%d", pDefWithWeakArmor, pDefNude)
	}
}

// TestGetBasePDef_Unchanged verifies GetBasePDef still works after Phase 5.5.
// GetBasePDef should return nude pDef WITHOUT armor.
func TestGetBasePDef_Unchanged(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 10, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	basePDef := player.GetBasePDef()

	// Equip armor
	armorTemplate := &ItemTemplate{
		ItemID: 1,
		Name:   "Armor",
		Type:   ItemTypeArmor,
		PDef:   100,
	}

	armor, _ := NewItem(1000, 1, 100, 1, armorTemplate)
	player.Inventory().AddItem(armor)
	_, _ = player.Inventory().EquipItem(armor, PaperdollChest)

	// GetBasePDef should be UNCHANGED (does not include armor)
	basePDefAfter := player.GetBasePDef()

	if basePDef != basePDefAfter {
		t.Errorf("GetBasePDef() changed after equipping armor: before=%d, after=%d (should be same)", basePDef, basePDefAfter)
	}

	// GetPDef should be HIGHER (includes armor)
	finalPDef := player.GetPDef()

	if finalPDef <= basePDef {
		t.Errorf("GetPDef()=%d should be > GetBasePDef()=%d (armor bonus)", finalPDef, basePDef)
	}
}

// TestGetAttackRange_Fists verifies attack range without weapon (fists).
// Phase 5.5: Weapon & Equipment System.
func TestGetAttackRange_Fists(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "TestFighter", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// No weapon: template base range
	// Human Fighter: BaseAtkRange = 20
	rangeFists := player.GetAttackRange()

	t.Logf("Fists: attack range=%d", rangeFists)

	if rangeFists != 20 {
		t.Errorf("GetAttackRange() fists = %d, expected 20 (Human Fighter base)", rangeFists)
	}
}

// TestGetAttackRange_Sword verifies attack range with sword weapon.
// Phase 5.5: Weapon & Equipment System.
func TestGetAttackRange_Sword(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "Test", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	rangeFists := player.GetAttackRange()

	// Equip sword (range=40)
	swordTemplate := &ItemTemplate{
		ItemID:      1,
		Name:        "Short Sword",
		Type:        ItemTypeWeapon,
		PAtk:        8,
		AttackRange: 40,
	}

	sword, _ := NewItem(1000, 1, 100, 1, swordTemplate)
	player.Inventory().AddItem(sword)
	_, _ = player.Inventory().EquipItem(sword, PaperdollRHand)

	rangeSword := player.GetAttackRange()

	t.Logf("Fists: range=%d, Sword: range=%d", rangeFists, rangeSword)

	// Sword should have longer range than fists
	if rangeSword <= rangeFists {
		t.Errorf("Sword range=%d should be > fists range=%d", rangeSword, rangeFists)
	}

	if rangeSword != 40 {
		t.Errorf("Sword range = %d, expected 40", rangeSword)
	}
}

// TestGetAttackRange_Bow verifies attack range with bow weapon (long range).
// Phase 5.5: Weapon & Equipment System.
func TestGetAttackRange_Bow(t *testing.T) {
	player, err := NewPlayer(1, 100, 200, "Test", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Equip bow (range=500)
	bowTemplate := &ItemTemplate{
		ItemID:      2,
		Name:        "Long Bow",
		Type:        ItemTypeWeapon,
		PAtk:        12,
		AttackRange: 500,
	}

	bow, _ := NewItem(1001, 2, 100, 1, bowTemplate)
	player.Inventory().AddItem(bow)
	_, _ = player.Inventory().EquipItem(bow, PaperdollRHand)

	rangeBow := player.GetAttackRange()

	t.Logf("Bow: attack range=%d", rangeBow)

	if rangeBow != 500 {
		t.Errorf("Bow range = %d, expected 500", rangeBow)
	}

	// Unequip bow → should revert to fists range
	player.Inventory().UnequipItem(PaperdollRHand)

	rangeFistsAfter := player.GetAttackRange()

	if rangeFistsAfter != 20 {
		t.Errorf("After unequip: range=%d, expected 20 (fists)", rangeFistsAfter)
	}
}
