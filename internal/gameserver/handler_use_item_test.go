package gameserver

import (
	"testing"

	"github.com/udisondev/la2go/internal/gameserver/serverpackets"
	"github.com/udisondev/la2go/internal/model"
)

// --- helpers ---

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	return NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

func newTestWeapon(t *testing.T, objectID uint32, itemID int32, bodyPart string, grade model.CrystalType) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID:      itemID,
		Name:        "TestSword",
		Type:        model.ItemTypeWeapon,
		PAtk:        100,
		AttackRange: 40,
		CrystalType: grade,
		BodyPartStr: bodyPart,
	}
	item, err := model.NewItem(objectID, itemID, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	return item
}

func newTestArmor(t *testing.T, objectID uint32, itemID int32, bodyPart string, grade model.CrystalType) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID:      itemID,
		Name:        "TestArmor",
		Type:        model.ItemTypeArmor,
		PDef:        50,
		CrystalType: grade,
		BodyPartStr: bodyPart,
	}
	item, err := model.NewItem(objectID, itemID, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	return item
}

func newTestEtcItem(t *testing.T, objectID uint32, itemID int32) *model.Item {
	t.Helper()
	tmpl := &model.ItemTemplate{
		ItemID: itemID,
		Name:   "TestEtcItem",
		Type:   model.ItemTypeEtcItem,
	}
	item, err := model.NewItem(objectID, itemID, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	return item
}

func newTestPlayerAtLevel(t *testing.T, objectID uint32, level int32) *model.Player {
	t.Helper()
	player, err := model.NewPlayer(objectID, 1, 1, "TestPlayer", level, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}
	player.SetLocation(model.NewLocation(0, 0, 0, 0))
	return player
}

// --- gradeMinLevel tests ---

func TestGradeMinLevel(t *testing.T) {
	tests := []struct {
		grade    model.CrystalType
		expected int32
	}{
		{model.CrystalNone, 1},
		{model.CrystalD, 20},
		{model.CrystalC, 40},
		{model.CrystalB, 52},
		{model.CrystalA, 61},
		{model.CrystalS, 76},
	}
	for _, tt := range tests {
		t.Run(tt.grade.String(), func(t *testing.T) {
			got := gradeMinLevel(tt.grade)
			if got != tt.expected {
				t.Errorf("gradeMinLevel(%v) = %d, want %d", tt.grade, got, tt.expected)
			}
		})
	}
}

// --- checkGradeRestriction tests ---

func TestCheckGradeRestriction(t *testing.T) {
	tests := []struct {
		name      string
		level     int32
		grade     model.CrystalType
		wantError bool
	}{
		{"no grade level 1", 1, model.CrystalNone, false},
		{"D-grade level 20", 20, model.CrystalD, false},
		{"D-grade level 19 fail", 19, model.CrystalD, true},
		{"C-grade level 40", 40, model.CrystalC, false},
		{"C-grade level 39 fail", 39, model.CrystalC, true},
		{"B-grade level 52", 52, model.CrystalB, false},
		{"B-grade level 51 fail", 51, model.CrystalB, true},
		{"A-grade level 61", 61, model.CrystalA, false},
		{"A-grade level 60 fail", 60, model.CrystalA, true},
		{"S-grade level 76", 76, model.CrystalS, false},
		{"S-grade level 75 fail", 75, model.CrystalS, true},
		{"S-grade level 80", 80, model.CrystalS, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := newTestPlayerAtLevel(t, 5000, tt.level)
			tmpl := &model.ItemTemplate{
				ItemID:      100,
				Name:        "TestItem",
				Type:        model.ItemTypeWeapon,
				CrystalType: tt.grade,
			}
			err := checkGradeRestriction(player, tmpl)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// --- resolveDoubleSlot tests ---

func TestResolveDoubleSlot_Earrings(t *testing.T) {
	player := newTestPlayerAtLevel(t, 6000, 50)
	inv := player.Inventory()
	rSlot, lSlot := model.EarringSlots()

	// Both empty → should return right
	got := resolveDoubleSlot(inv, "rear", rSlot)
	if got != rSlot {
		t.Errorf("both empty: got slot %d, want right=%d", got, rSlot)
	}

	// Equip right earring
	earR := newTestArmor(t, 7001, 200, "rear", model.CrystalNone)
	if err := inv.AddItem(earR); err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if _, err := inv.EquipItem(earR, rSlot); err != nil {
		t.Fatalf("EquipItem right: %v", err)
	}

	// Right occupied → should return left
	got = resolveDoubleSlot(inv, "rear", rSlot)
	if got != lSlot {
		t.Errorf("right occupied: got slot %d, want left=%d", got, lSlot)
	}

	// Equip left earring
	earL := newTestArmor(t, 7002, 201, "lear", model.CrystalNone)
	if err := inv.AddItem(earL); err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if _, err := inv.EquipItem(earL, lSlot); err != nil {
		t.Fatalf("EquipItem left: %v", err)
	}

	// Both occupied → should return right (replacement)
	got = resolveDoubleSlot(inv, "rear", rSlot)
	if got != rSlot {
		t.Errorf("both occupied: got slot %d, want right=%d", got, rSlot)
	}
}

func TestResolveDoubleSlot_Rings(t *testing.T) {
	player := newTestPlayerAtLevel(t, 6100, 50)
	inv := player.Inventory()
	rSlot, lSlot := model.RingSlots()

	// Both empty → right
	got := resolveDoubleSlot(inv, "rfinger", rSlot)
	if got != rSlot {
		t.Errorf("both empty: got slot %d, want right=%d", got, rSlot)
	}

	// Equip right ring
	ringR := newTestArmor(t, 7101, 300, "rfinger", model.CrystalNone)
	if err := inv.AddItem(ringR); err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if _, err := inv.EquipItem(ringR, rSlot); err != nil {
		t.Fatalf("EquipItem right: %v", err)
	}

	// Right occupied → left
	got = resolveDoubleSlot(inv, "rfinger", rSlot)
	if got != lSlot {
		t.Errorf("right occupied: got slot %d, want left=%d", got, lSlot)
	}
}

func TestResolveDoubleSlot_NonPaired(t *testing.T) {
	player := newTestPlayerAtLevel(t, 6200, 50)
	inv := player.Inventory()

	// Non-paired slot returns default
	got := resolveDoubleSlot(inv, "chest", model.PaperdollChest)
	if got != model.PaperdollChest {
		t.Errorf("chest: got slot %d, want %d", got, model.PaperdollChest)
	}
}

// --- BodyPartToPaperdollSlot tests ---

func TestBodyPartToPaperdollSlot(t *testing.T) {
	tests := []struct {
		bodyPart string
		expected int32
	}{
		{"rhand", model.PaperdollRHand},
		{"lhand", model.PaperdollLHand},
		{"lrhand", model.PaperdollRHand},
		{"chest", model.PaperdollChest},
		{"legs", model.PaperdollLegs},
		{"head", model.PaperdollHead},
		{"feet", model.PaperdollFeet},
		{"gloves", model.PaperdollGloves},
		{"neck", model.PaperdollNeck},
		{"under", model.PaperdollUnder},
		{"back", model.PaperdollCloak},
		{"unknown", int32(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.bodyPart, func(t *testing.T) {
			got := model.BodyPartToPaperdollSlot(tt.bodyPart)
			if got != tt.expected {
				t.Errorf("BodyPartToPaperdollSlot(%q) = %d, want %d", tt.bodyPart, got, tt.expected)
			}
		})
	}
}

// --- BodyPartToAdditionalSlot tests ---

func TestBodyPartToAdditionalSlot(t *testing.T) {
	tests := []struct {
		bodyPart string
		expected int32
	}{
		{"lrhand", model.PaperdollLHand},
		{"onepiece", model.PaperdollLegs},
		{"alldress", model.PaperdollLegs},
		{"rhand", int32(-1)},
		{"chest", int32(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.bodyPart, func(t *testing.T) {
			got := model.BodyPartToAdditionalSlot(tt.bodyPart)
			if got != tt.expected {
				t.Errorf("BodyPartToAdditionalSlot(%q) = %d, want %d", tt.bodyPart, got, tt.expected)
			}
		})
	}
}

// --- InventoryUpdate packet tests ---

func TestInventoryUpdate_Write(t *testing.T) {
	item := newTestWeapon(t, 8001, 100, "rhand", model.CrystalNone)
	entry := serverpackets.InvUpdateEntry{
		ChangeType: serverpackets.InvUpdateModify,
		Item:       item,
	}

	pkt := serverpackets.NewInventoryUpdate(entry)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty data")
	}

	// Opcode should be 0x37
	if data[0] != serverpackets.OpcodeInventoryUpdate {
		t.Errorf("opcode = 0x%02X, want 0x%02X", data[0], serverpackets.OpcodeInventoryUpdate)
	}
}

// --- handleEquipToggle integration tests ---

func TestHandleEquipToggle_EquipWeapon(t *testing.T) {
	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 9000, 50)
	buf := make([]byte, 65536)

	// Create weapon and add to inventory
	sword := newTestWeapon(t, 9001, 100, "rhand", model.CrystalNone)
	if err := player.Inventory().AddItem(sword); err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	tmpl := sword.Template()
	n, ok, err := h.handleEquipToggle(nil, player, sword, tmpl, buf)
	if err != nil {
		t.Fatalf("handleEquipToggle: %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Error("expected bytes written > 0")
	}

	// Item should be equipped
	if !sword.IsEquipped() {
		t.Error("sword should be equipped after equip toggle")
	}

	// Opcode should be InventoryUpdate (0x37)
	if buf[0] != serverpackets.OpcodeInventoryUpdate {
		t.Errorf("opcode = 0x%02X, want 0x%02X", buf[0], serverpackets.OpcodeInventoryUpdate)
	}
}

func TestHandleEquipToggle_UnequipWeapon(t *testing.T) {
	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 9100, 50)
	buf := make([]byte, 65536)

	// Create and equip weapon
	sword := newTestWeapon(t, 9101, 100, "rhand", model.CrystalNone)
	inv := player.Inventory()
	if err := inv.AddItem(sword); err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if _, err := inv.EquipItem(sword, model.PaperdollRHand); err != nil {
		t.Fatalf("EquipItem: %v", err)
	}

	if !sword.IsEquipped() {
		t.Fatal("sword should be equipped before test")
	}

	tmpl := sword.Template()
	n, ok, err := h.handleEquipToggle(nil, player, sword, tmpl, buf)
	if err != nil {
		t.Fatalf("handleEquipToggle (unequip): %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Error("expected bytes written > 0")
	}

	// Item should be unequipped
	if sword.IsEquipped() {
		t.Error("sword should be unequipped after toggle")
	}
}

func TestHandleEquipToggle_GradeRestriction(t *testing.T) {
	h := newTestHandler(t)
	// Level 10 player trying to equip D-grade (requires level 20)
	player := newTestPlayerAtLevel(t, 9200, 10)
	buf := make([]byte, 65536)

	dSword := newTestWeapon(t, 9201, 200, "rhand", model.CrystalD)
	if err := player.Inventory().AddItem(dSword); err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	tmpl := dSword.Template()
	n, ok, err := h.handleEquipToggle(nil, player, dSword, tmpl, buf)
	if err != nil {
		t.Fatalf("handleEquipToggle (grade): %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Error("expected bytes written > 0 (SystemMessage)")
	}

	// Item should NOT be equipped
	if dSword.IsEquipped() {
		t.Error("D-grade sword should NOT be equipped at level 10")
	}

	// Response should be SystemMessage (0x64)
	if buf[0] != serverpackets.OpcodeSystemMessage {
		t.Errorf("opcode = 0x%02X, want 0x%02X (SystemMessage)", buf[0], serverpackets.OpcodeSystemMessage)
	}
}

func TestHandleEquip_TwoHandedRemovesShield(t *testing.T) {
	h := newTestHandler(t)
	player := newTestPlayerAtLevel(t, 9300, 50)
	buf := make([]byte, 65536)
	inv := player.Inventory()

	// First equip a shield in left hand
	shield := newTestArmor(t, 9301, 300, "lhand", model.CrystalNone)
	if err := inv.AddItem(shield); err != nil {
		t.Fatalf("AddItem shield: %v", err)
	}
	if _, err := inv.EquipItem(shield, model.PaperdollLHand); err != nil {
		t.Fatalf("EquipItem shield: %v", err)
	}

	if !shield.IsEquipped() {
		t.Fatal("shield should be equipped before test")
	}

	// Now equip a two-handed weapon (lrhand)
	bow := newTestWeapon(t, 9302, 400, "lrhand", model.CrystalNone)
	if err := inv.AddItem(bow); err != nil {
		t.Fatalf("AddItem bow: %v", err)
	}

	tmpl := bow.Template()
	n, ok, err := h.handleEquip(nil, player, bow, tmpl, inv, buf)
	if err != nil {
		t.Fatalf("handleEquip (two-handed): %v", err)
	}
	if !ok {
		t.Error("expected keepAlive=true")
	}
	if n == 0 {
		t.Error("expected bytes written > 0")
	}

	// Bow should be equipped, shield should be unequipped
	if !bow.IsEquipped() {
		t.Error("bow should be equipped")
	}
	if shield.IsEquipped() {
		t.Error("shield should be unequipped when equipping two-handed weapon")
	}
}

// --- CrystalType tests ---

func TestCrystalTypeFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected model.CrystalType
	}{
		{"NONE", model.CrystalNone},
		{"D", model.CrystalD},
		{"C", model.CrystalC},
		{"B", model.CrystalB},
		{"A", model.CrystalA},
		{"S", model.CrystalS},
		{"", model.CrystalNone},
		{"X", model.CrystalNone},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := model.CrystalTypeFromString(tt.input)
			if got != tt.expected {
				t.Errorf("CrystalTypeFromString(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCrystalTypeString(t *testing.T) {
	tests := []struct {
		ct       model.CrystalType
		expected string
	}{
		{model.CrystalNone, "NONE"},
		{model.CrystalD, "D"},
		{model.CrystalC, "C"},
		{model.CrystalB, "B"},
		{model.CrystalA, "A"},
		{model.CrystalS, "S"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.ct.String()
			if got != tt.expected {
				t.Errorf("CrystalType(%d).String() = %q, want %q", tt.ct, got, tt.expected)
			}
		})
	}
}

// --- IsEquippable / IsTwoHanded tests ---

func TestItemTemplate_IsEquippable(t *testing.T) {
	weapon := &model.ItemTemplate{Type: model.ItemTypeWeapon}
	armor := &model.ItemTemplate{Type: model.ItemTypeArmor}
	etc := &model.ItemTemplate{Type: model.ItemTypeEtcItem}
	quest := &model.ItemTemplate{Type: model.ItemTypeQuestItem}

	if !weapon.IsEquippable() {
		t.Error("weapon should be equippable")
	}
	if !armor.IsEquippable() {
		t.Error("armor should be equippable")
	}
	if etc.IsEquippable() {
		t.Error("etc item should NOT be equippable")
	}
	if quest.IsEquippable() {
		t.Error("quest item should NOT be equippable")
	}
}

func TestItemTemplate_IsTwoHanded(t *testing.T) {
	bow := &model.ItemTemplate{BodyPartStr: "lrhand"}
	sword := &model.ItemTemplate{BodyPartStr: "rhand"}

	if !bow.IsTwoHanded() {
		t.Error("lrhand should be two-handed")
	}
	if sword.IsTwoHanded() {
		t.Error("rhand should NOT be two-handed")
	}
}
