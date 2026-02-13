package itemhandler

import (
	"testing"

	"github.com/udisondev/la2go/internal/data"
	"github.com/udisondev/la2go/internal/model"
)

func TestInit_RegistersAllHandlers(t *testing.T) {
	t.Parallel()
	Init()

	handlers := []string{
		"ItemSkills", "Elixir", "SoulShots", "SpiritShot",
		"BlessedSpiritShot", "FishShots", "Book", "Recipes",
		"RollingDice", "CharmOfCourage", "PetFood",
	}
	for _, name := range handlers {
		if Get(name) == nil {
			t.Errorf("handler %q not registered", name)
		}
	}
}

func TestGet_UnknownHandler(t *testing.T) {
	t.Parallel()
	if Get("NonExistent") != nil {
		t.Error("expected nil for unknown handler")
	}
}

// --- ItemSkills Handler ---

func TestItemSkills_UseItem_WithSkill(t *testing.T) {
	t.Parallel()

	// Load skill data for test
	if err := data.LoadSkills(); err != nil {
		t.Skipf("skill data not loaded: %v", err)
	}

	h := &itemSkillsHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 1060, 1) // Lesser Healing Potion

	// Use a skill that exists in data
	result := h.UseItem(player, item, 2031, 1) // Lesser Healing
	if result == nil {
		t.Skip("skill 2031 L1 not in skill data")
	}

	if result.ConsumeCount != 1 {
		t.Errorf("ConsumeCount = %d; want 1", result.ConsumeCount)
	}
	if result.SkillID != 2031 {
		t.Errorf("SkillID = %d; want 2031", result.SkillID)
	}
	if result.SkillLevel != 1 {
		t.Errorf("SkillLevel = %d; want 1", result.SkillLevel)
	}
	if !result.Broadcast {
		t.Error("Broadcast should be true")
	}
}

func TestItemSkills_UseItem_NoSkill(t *testing.T) {
	t.Parallel()
	h := &itemSkillsHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 57, 100) // Adena

	result := h.UseItem(player, item, 0, 0)
	if result != nil {
		t.Error("expected nil result for item with no skill")
	}
}

// --- SoulShots Handler ---

func TestSoulShots_NoWeapon(t *testing.T) {
	t.Parallel()
	h := &soulShotsHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 1835, 100) // Soul Shot no-grade

	result := h.UseItem(player, item, 0, 0)
	if result != nil {
		t.Error("expected nil result when no weapon equipped")
	}
}

// --- Book Handler ---

func TestBook_UseItem(t *testing.T) {
	t.Parallel()
	h := &bookHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 5588, 1) // Some book

	result := h.UseItem(player, item, 0, 0)
	if result == nil {
		t.Fatal("expected non-nil result for book")
	}
	if result.ConsumeCount != 0 {
		t.Errorf("ConsumeCount = %d; want 0 (books not consumed)", result.ConsumeCount)
	}
}

// --- Recipes Handler ---

func TestRecipes_UseItem(t *testing.T) {
	t.Parallel()
	h := &recipesHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 1900, 1) // Recipe item

	result := h.UseItem(player, item, 0, 0)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ConsumeCount != 1 {
		t.Errorf("ConsumeCount = %d; want 1", result.ConsumeCount)
	}
}

// --- CharmOfCourage ---

func TestCharmOfCourage_UseItem(t *testing.T) {
	t.Parallel()
	h := &charmOfCourageHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 5041, 1)

	result := h.UseItem(player, item, 2170, 1)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ConsumeCount != 1 {
		t.Errorf("ConsumeCount = %d; want 1", result.ConsumeCount)
	}
	if result.SkillID != 2170 {
		t.Errorf("SkillID = %d; want 2170", result.SkillID)
	}
}

func TestCharmOfCourage_NoSkill(t *testing.T) {
	t.Parallel()
	h := &charmOfCourageHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 5041, 1)

	result := h.UseItem(player, item, 0, 0)
	if result != nil {
		t.Error("expected nil result when no skill")
	}
}

// --- PetFood ---

func TestPetFood_NotImplemented(t *testing.T) {
	t.Parallel()
	h := &petFoodHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 2515, 10)

	result := h.UseItem(player, item, 0, 0)
	if result != nil {
		t.Error("expected nil for pet food (pets not implemented)")
	}
}

// --- RollingDice ---

func TestRollingDice_UseItem(t *testing.T) {
	t.Parallel()
	h := &rollingDiceHandler{}
	player := newTestPlayer(t)
	item := newTestItem(t, 4625, 1)

	result := h.UseItem(player, item, 0, 0)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ConsumeCount != 0 {
		t.Errorf("ConsumeCount = %d; want 0 (dice not consumed)", result.ConsumeCount)
	}
}

// --- Grade Match ---

func TestGradeMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		shot, weapon string
		want         bool
	}{
		{"", "D", true},         // no-grade shot → any weapon
		{"NONE", "D", true},     // NONE shot → any weapon
		{"D", "D", true},        // match
		{"D", "C", false},       // mismatch
		{"S", "S", true},        // match
		{"A", "S", false},       // mismatch
	}

	for _, tt := range tests {
		got := gradeMatch(tt.shot, tt.weapon)
		if got != tt.want {
			t.Errorf("gradeMatch(%q, %q) = %v; want %v",
				tt.shot, tt.weapon, got, tt.want)
		}
	}
}

// --- Shot Skill Grade Mapping ---

func TestSoulShotSkillForGrade(t *testing.T) {
	t.Parallel()

	if soulShotSkillForGrade("D") != SkillSoulShotD {
		t.Error("grade D mismatch")
	}
	if soulShotSkillForGrade("S") != SkillSoulShotS {
		t.Error("grade S mismatch")
	}
	if soulShotSkillForGrade("NONE") != SkillSoulShotNone {
		t.Error("grade NONE mismatch")
	}
}

func TestSpiritShotSkillForGrade(t *testing.T) {
	t.Parallel()

	if spiritShotSkillForGrade("C") != SkillSpiritShotC {
		t.Error("grade C mismatch")
	}
	if spiritShotSkillForGrade("") != SkillSpiritShotNone {
		t.Error("empty grade mismatch")
	}
}

func TestBlessedSpiritShotSkillForGrade(t *testing.T) {
	t.Parallel()

	if blessedSpiritShotSkillForGrade("A") != SkillBlessedSpiritShotA {
		t.Error("grade A mismatch")
	}
}

// --- Item Shot Charges ---

func TestItem_ShotCharges(t *testing.T) {
	t.Parallel()

	tmpl := &model.ItemTemplate{
		Name:  "Test Weapon",
		Type:  model.ItemTypeWeapon,
	}
	item, err := model.NewItem(1, 1, 1, 1, tmpl)
	if err != nil {
		t.Fatal(err)
	}

	// Initially not charged
	if item.IsChargedSoulShot() {
		t.Error("should not be charged initially")
	}

	// Charge soul shot
	item.SetChargedSoulShot(true)
	if !item.IsChargedSoulShot() {
		t.Error("should be charged after set")
	}

	// Charge spirit shot
	item.SetChargedSpiritShot(true)
	if !item.IsChargedSpiritShot() {
		t.Error("spirit shot should be charged")
	}

	// Charge blessed spirit shot
	item.SetChargedBlessedSpiritShot(true)
	if !item.IsChargedBlessedSpiritShot() {
		t.Error("blessed spirit shot should be charged")
	}

	// Clear all
	item.ClearAllShotCharges()
	if item.IsChargedSoulShot() || item.IsChargedSpiritShot() || item.IsChargedBlessedSpiritShot() {
		t.Error("all charges should be cleared")
	}
}

// --- Player Item Cooldowns ---

func TestPlayer_ItemCooldown(t *testing.T) {
	t.Parallel()

	player := newTestPlayer(t)

	// Not on cooldown initially
	if player.IsItemOnCooldown(1060) {
		t.Error("should not be on cooldown initially")
	}

	// Set cooldown
	player.SetItemCooldown(1060, 500_000_000) // 500ms

	// Should be on cooldown
	if !player.IsItemOnCooldown(1060) {
		t.Error("should be on cooldown after set")
	}

	// Other items unaffected
	if player.IsItemOnCooldown(1061) {
		t.Error("unrelated item should not be on cooldown")
	}
}

// --- Player Olympiad State ---

func TestPlayer_IsInOlympiad(t *testing.T) {
	t.Parallel()

	player := newTestPlayer(t)

	if player.IsInOlympiad() {
		t.Error("should not be in olympiad initially")
	}

	player.SetInOlympiad(true)
	if !player.IsInOlympiad() {
		t.Error("should be in olympiad after set")
	}

	player.SetInOlympiad(false)
	if player.IsInOlympiad() {
		t.Error("should not be in olympiad after clear")
	}
}

// --- Helpers ---

func newTestPlayer(t *testing.T) *model.Player {
	t.Helper()
	p, err := model.NewPlayer(1, 1, 1, "TestPlayer", 40, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func newTestItem(t *testing.T, itemID int32, count int32) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		Name: "TestItem",
		Type: model.ItemTypeEtcItem,
	}
	item, err := model.NewItem(uint32(itemID), itemID, 1, count, tmpl)
	if err != nil {
		t.Fatal(err)
	}
	return item
}
