package serverpackets

import (
	"testing"

	"github.com/udisondev/la2go/internal/model"
)

func TestNewInventoryItemList_Empty(t *testing.T) {
	pkt := NewInventoryItemList(nil)
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// opcode(1) + showWindow(2) + itemCount(2) = 5 bytes
	if len(data) != 5 {
		t.Fatalf("expected 5 bytes, got %d", len(data))
	}

	if data[0] != OpcodeInventoryItemList {
		t.Errorf("opcode: got 0x%02X, want 0x%02X", data[0], OpcodeInventoryItemList)
	}

	// showWindow = 0 (LE short)
	if data[1] != 0 || data[2] != 0 {
		t.Errorf("showWindow: got [%d,%d], want [0,0]", data[1], data[2])
	}

	// itemCount = 0 (LE short)
	if data[3] != 0 || data[4] != 0 {
		t.Errorf("itemCount: got [%d,%d], want [0,0]", data[3], data[4])
	}
}

func TestNewInventoryItemList_WithItems(t *testing.T) {
	tmpl := &model.ItemTemplate{
		ItemID: 57,
		Name:   "Adena",
		Type:   model.ItemTypeEtcItem,
	}

	item, err := model.NewItem(1000, 57, 1, 100, tmpl)
	if err != nil {
		t.Fatalf("NewItem error: %v", err)
	}

	pkt := NewInventoryItemList([]*model.Item{item})
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// 5 header bytes + 36 bytes per item = 41 bytes
	if len(data) != 41 {
		t.Fatalf("expected 41 bytes, got %d", len(data))
	}

	// opcode
	if data[0] != OpcodeInventoryItemList {
		t.Errorf("opcode: got 0x%02X, want 0x%02X", data[0], OpcodeInventoryItemList)
	}

	// itemCount = 1 (LE short at offset 3)
	itemCount := int16(data[3]) | int16(data[4])<<8
	if itemCount != 1 {
		t.Errorf("itemCount: got %d, want 1", itemCount)
	}
}

func TestNewInventoryItemList_WeaponType(t *testing.T) {
	tmpl := &model.ItemTemplate{
		ItemID: 1,
		Name:   "Short Sword",
		Type:   model.ItemTypeWeapon,
	}

	item, err := model.NewItem(2000, 1, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem error: %v", err)
	}

	pkt := NewInventoryItemList([]*model.Item{item})
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// type1 at offset 5 (LE short): weapon = 0
	type1 := int16(data[5]) | int16(data[6])<<8
	if type1 != 0 {
		t.Errorf("type1 for weapon: got %d, want 0", type1)
	}
}

func TestNewInventoryItemList_ShowWindow(t *testing.T) {
	pkt := NewInventoryItemList(nil)
	pkt.ShowWindow = true
	data, err := pkt.Write()
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// showWindow = 1 (LE short at offset 1)
	showWindow := int16(data[1]) | int16(data[2])<<8
	if showWindow != 1 {
		t.Errorf("showWindow: got %d, want 1", showWindow)
	}
}

func TestBodyPartMask(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     *model.ItemTemplate
		expected int32
	}{
		{"weapon", &model.ItemTemplate{Type: model.ItemTypeWeapon, BodyPartMask: model.BodyPartRHand}, model.BodyPartRHand},
		{"chest", &model.ItemTemplate{Type: model.ItemTypeArmor, BodyPartMask: model.BodyPartChest}, model.BodyPartChest},
		{"legs", &model.ItemTemplate{Type: model.ItemTypeArmor, BodyPartMask: model.BodyPartLegs}, model.BodyPartLegs},
		{"head", &model.ItemTemplate{Type: model.ItemTypeArmor, BodyPartMask: model.BodyPartHead}, model.BodyPartHead},
		{"feet", &model.ItemTemplate{Type: model.ItemTypeArmor, BodyPartMask: model.BodyPartFeet}, model.BodyPartFeet},
		{"neck", &model.ItemTemplate{Type: model.ItemTypeArmor, BodyPartMask: model.BodyPartNeck}, model.BodyPartNeck},
		{"etc", &model.ItemTemplate{Type: model.ItemTypeEtcItem}, 0},
		{"nil", nil, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bodyPartMask(tc.tmpl)
			if got != tc.expected {
				t.Errorf("bodyPartMask(%s): got 0x%04X, want 0x%04X", tc.name, got, tc.expected)
			}
		})
	}
}

func TestItemType1(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     *model.ItemTemplate
		expected int16
	}{
		{"weapon", &model.ItemTemplate{Type1: model.Type1WeaponRingEarringNecklace}, model.Type1WeaponRingEarringNecklace},
		{"armor", &model.ItemTemplate{Type1: model.Type1ShieldArmor}, model.Type1ShieldArmor},
		{"etc", &model.ItemTemplate{Type1: model.Type1ItemQuestItemAdena}, model.Type1ItemQuestItemAdena},
		{"nil", nil, model.Type1ItemQuestItemAdena},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := itemType1(tc.tmpl)
			if got != tc.expected {
				t.Errorf("itemType1(%s): got %d, want %d", tc.name, got, tc.expected)
			}
		})
	}
}

func TestItemType2(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     *model.ItemTemplate
		expected int16
	}{
		{"weapon", &model.ItemTemplate{Type2: model.Type2Weapon}, model.Type2Weapon},
		{"armor", &model.ItemTemplate{Type2: model.Type2ShieldArmor}, model.Type2ShieldArmor},
		{"accessory", &model.ItemTemplate{Type2: model.Type2Accessory}, model.Type2Accessory},
		{"quest", &model.ItemTemplate{Type2: model.Type2Quest}, model.Type2Quest},
		{"nil", nil, model.Type2Other},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := itemType2(tc.tmpl)
			if got != tc.expected {
				t.Errorf("itemType2(%s): got %d, want %d", tc.name, got, tc.expected)
			}
		})
	}
}
