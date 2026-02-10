package model

import (
	"testing"
)

// TestNewItem_Phase55 verifies item creation with ItemTemplate (Phase 5.5).
func TestNewItem_Phase55(t *testing.T) {
	template := &ItemTemplate{
		ItemID:      1,
		Name:        "Short Sword",
		Type:        ItemTypeWeapon,
		PAtk:        8,
		AttackRange: 40,
	}

	tests := []struct {
		name      string
		objectID  uint32
		itemID    int32
		ownerID   int64
		count     int32
		template  *ItemTemplate
		wantError bool
	}{
		{
			name:      "valid weapon",
			objectID:  1000,
			itemID:    1,
			ownerID:   100,
			count:     1,
			template:  template,
			wantError: false,
		},
		{
			name:      "nil template",
			objectID:  1001,
			itemID:    1,
			ownerID:   100,
			count:     1,
			template:  nil,
			wantError: true,
		},
		{
			name:      "zero count",
			objectID:  1002,
			itemID:    1,
			ownerID:   100,
			count:     0,
			template:  template,
			wantError: true,
		},
		{
			name:      "negative count",
			objectID:  1003,
			itemID:    1,
			ownerID:   100,
			count:     -1,
			template:  template,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewItem(tt.objectID, tt.itemID, tt.ownerID, tt.count, tt.template)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewItem() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewItem() unexpected error: %v", err)
			}

			if item.ObjectID() != tt.objectID {
				t.Errorf("ObjectID() = %d, want %d", item.ObjectID(), tt.objectID)
			}
			if item.ItemID() != tt.itemID {
				t.Errorf("ItemID() = %d, want %d", item.ItemID(), tt.itemID)
			}
			if item.OwnerID() != tt.ownerID {
				t.Errorf("OwnerID() = %d, want %d", item.OwnerID(), tt.ownerID)
			}
			if item.Count() != tt.count {
				t.Errorf("Count() = %d, want %d", item.Count(), tt.count)
			}
			if item.Enchant() != 0 {
				t.Errorf("Enchant() = %d, want 0", item.Enchant())
			}
			if item.Location() != ItemLocationInventory {
				t.Errorf("Location() = %v, want %v", item.Location(), ItemLocationInventory)
			}
			if item.Slot() != -1 {
				t.Errorf("Slot() = %d, want -1", item.Slot())
			}
			if item.Template() != tt.template {
				t.Errorf("Template() mismatch")
			}
		})
	}
}

// TestItem_IsEquipped_Phase55 verifies equipped state detection.
func TestItem_IsEquipped_Phase55(t *testing.T) {
	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item, _ := NewItem(1000, 1, 100, 1, template)

	// Not equipped initially
	if item.IsEquipped() {
		t.Errorf("IsEquipped() = true, want false (initial state)")
	}

	// Set to paperdoll location + slot
	item.SetLocation(ItemLocationPaperdoll)
	item.SetSlot(7) // PAPERDOLL_RHAND

	// Should be equipped now
	if !item.IsEquipped() {
		t.Errorf("IsEquipped() = false, want true (paperdoll + slot >= 0)")
	}

	// Unequip (set slot to -1)
	item.SetSlot(-1)

	// Should NOT be equipped (slot < 0)
	if item.IsEquipped() {
		t.Errorf("IsEquipped() = true, want false (slot < 0)")
	}
}

// TestItem_TypeChecks_Phase55 verifies IsWeapon/IsArmor/IsConsumable.
func TestItem_TypeChecks_Phase55(t *testing.T) {
	tests := []struct {
		name           string
		itemType       ItemType
		wantWeapon     bool
		wantArmor      bool
		wantConsumable bool
	}{
		{
			name:           "weapon",
			itemType:       ItemTypeWeapon,
			wantWeapon:     true,
			wantArmor:      false,
			wantConsumable: false,
		},
		{
			name:           "armor",
			itemType:       ItemTypeArmor,
			wantWeapon:     false,
			wantArmor:      true,
			wantConsumable: false,
		},
		{
			name:           "consumable",
			itemType:       ItemTypeConsumable,
			wantWeapon:     false,
			wantArmor:      false,
			wantConsumable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := &ItemTemplate{
				ItemID: 1,
				Name:   "Test Item",
				Type:   tt.itemType,
			}

			item, _ := NewItem(1000, 1, 100, 1, template)

			if item.IsWeapon() != tt.wantWeapon {
				t.Errorf("IsWeapon() = %v, want %v", item.IsWeapon(), tt.wantWeapon)
			}
			if item.IsArmor() != tt.wantArmor {
				t.Errorf("IsArmor() = %v, want %v", item.IsArmor(), tt.wantArmor)
			}
			if item.IsConsumable() != tt.wantConsumable {
				t.Errorf("IsConsumable() = %v, want %v", item.IsConsumable(), tt.wantConsumable)
			}
		})
	}
}

// TestItem_SetCount_Phase55 verifies count validation.
func TestItem_SetCount_Phase55(t *testing.T) {
	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Arrow",
		Type:   ItemTypeConsumable,
	}

	item, _ := NewItem(1000, 1, 100, 10, template)

	// Valid count
	err := item.SetCount(20)
	if err != nil {
		t.Errorf("SetCount(20) unexpected error: %v", err)
	}
	if item.Count() != 20 {
		t.Errorf("Count() = %d, want 20", item.Count())
	}

	// Negative count should fail
	err = item.SetCount(-1)
	if err == nil {
		t.Errorf("SetCount(-1) expected error, got nil")
	}

	// Zero count is allowed (item destroyed)
	err = item.SetCount(0)
	if err != nil {
		t.Errorf("SetCount(0) unexpected error: %v", err)
	}
	if item.Count() != 0 {
		t.Errorf("Count() = %d, want 0", item.Count())
	}
}

// TestItem_SetEnchant_Phase55 verifies enchant validation.
func TestItem_SetEnchant_Phase55(t *testing.T) {
	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item, _ := NewItem(1000, 1, 100, 1, template)

	// Valid enchant
	err := item.SetEnchant(5)
	if err != nil {
		t.Errorf("SetEnchant(5) unexpected error: %v", err)
	}
	if item.Enchant() != 5 {
		t.Errorf("Enchant() = %d, want 5", item.Enchant())
	}

	// Negative enchant should fail
	err = item.SetEnchant(-1)
	if err == nil {
		t.Errorf("SetEnchant(-1) expected error, got nil")
	}

	// Enchant unchanged after failed set
	if item.Enchant() != 5 {
		t.Errorf("Enchant() = %d, want 5 (unchanged)", item.Enchant())
	}
}

// TestItem_Name_Phase55 verifies name retrieval from template.
func TestItem_Name_Phase55(t *testing.T) {
	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Legendary Sword",
		Type:   ItemTypeWeapon,
	}

	item, _ := NewItem(1000, 1, 100, 1, template)

	if item.Name() != "Legendary Sword" {
		t.Errorf("Name() = %q, want %q", item.Name(), "Legendary Sword")
	}
}

// TestItemTemplate_TypeChecks verifies ItemTemplate helper methods.
func TestItemTemplate_TypeChecks(t *testing.T) {
	tests := []struct {
		name           string
		itemType       ItemType
		wantWeapon     bool
		wantArmor      bool
		wantConsumable bool
	}{
		{
			name:           "weapon",
			itemType:       ItemTypeWeapon,
			wantWeapon:     true,
			wantArmor:      false,
			wantConsumable: false,
		},
		{
			name:           "armor",
			itemType:       ItemTypeArmor,
			wantWeapon:     false,
			wantArmor:      true,
			wantConsumable: false,
		},
		{
			name:           "consumable",
			itemType:       ItemTypeConsumable,
			wantWeapon:     false,
			wantArmor:      false,
			wantConsumable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := &ItemTemplate{
				ItemID: 1,
				Name:   "Test",
				Type:   tt.itemType,
			}

			if template.IsWeapon() != tt.wantWeapon {
				t.Errorf("IsWeapon() = %v, want %v", template.IsWeapon(), tt.wantWeapon)
			}
			if template.IsArmor() != tt.wantArmor {
				t.Errorf("IsArmor() = %v, want %v", template.IsArmor(), tt.wantArmor)
			}
			if template.IsConsumable() != tt.wantConsumable {
				t.Errorf("IsConsumable() = %v, want %v", template.IsConsumable(), tt.wantConsumable)
			}
		})
	}
}

// TestItemType_String verifies item type string representation.
func TestItemType_String(t *testing.T) {
	tests := []struct {
		itemType ItemType
		want     string
	}{
		{ItemTypeWeapon, "Weapon"},
		{ItemTypeArmor, "Armor"},
		{ItemTypeConsumable, "Consumable"},
		{ItemTypeQuestItem, "QuestItem"},
		{ItemTypeEtcItem, "EtcItem"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.itemType.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestArmorSlot_String verifies armor slot string representation.
func TestArmorSlot_String(t *testing.T) {
	tests := []struct {
		slot ArmorSlot
		want string
	}{
		{ArmorSlotNone, "None"},
		{ArmorSlotChest, "Chest"},
		{ArmorSlotLegs, "Legs"},
		{ArmorSlotHead, "Head"},
		{ArmorSlotFeet, "Feet"},
		{ArmorSlotGloves, "Gloves"},
		{ArmorSlotUnder, "Underwear"},
		{ArmorSlotCloak, "Cloak"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.slot.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
