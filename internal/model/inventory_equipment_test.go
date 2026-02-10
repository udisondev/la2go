package model

import (
	"testing"
)

// TestNewInventory verifies inventory creation.
func TestNewInventory(t *testing.T) {
	inv := NewInventory(100)

	if inv.OwnerID() != 100 {
		t.Errorf("OwnerID() = %d, want 100", inv.OwnerID())
	}
	if inv.TotalCount() != 0 {
		t.Errorf("TotalCount() = %d, want 0 (empty inventory)", inv.TotalCount())
	}
}

// TestInventory_AddRemoveItem verifies item add/remove operations.
func TestInventory_AddRemoveItem(t *testing.T) {
	inv := NewInventory(100)

	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item, _ := NewItem(1000, 1, 100, 1, template)

	// Add item
	err := inv.AddItem(item)
	if err != nil {
		t.Fatalf("AddItem() unexpected error: %v", err)
	}

	if inv.TotalCount() != 1 {
		t.Errorf("TotalCount() = %d, want 1", inv.TotalCount())
	}

	// Get item
	retrieved := inv.GetItem(1000)
	if retrieved != item {
		t.Errorf("GetItem(1000) mismatch")
	}

	// Add duplicate should fail
	err = inv.AddItem(item)
	if err == nil {
		t.Errorf("AddItem(duplicate) expected error, got nil")
	}

	// Remove item
	removed := inv.RemoveItem(1000)
	if removed != item {
		t.Errorf("RemoveItem(1000) mismatch")
	}

	if inv.TotalCount() != 0 {
		t.Errorf("TotalCount() = %d, want 0 (after removal)", inv.TotalCount())
	}

	// Remove non-existent should return nil
	removed = inv.RemoveItem(9999)
	if removed != nil {
		t.Errorf("RemoveItem(9999) = %v, want nil", removed)
	}
}

// TestInventory_EquipUnequip verifies equip/unequip operations.
func TestInventory_EquipUnequip(t *testing.T) {
	inv := NewInventory(100)

	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item, _ := NewItem(1000, 1, 100, 1, template)

	// Add item to inventory first
	inv.AddItem(item)

	// Equip item
	err := inv.EquipItem(item, PaperdollRHand)
	if err != nil {
		t.Fatalf("EquipItem() unexpected error: %v", err)
	}

	// Check paperdoll slot
	equipped := inv.GetPaperdollItem(PaperdollRHand)
	if equipped != item {
		t.Errorf("GetPaperdollItem(RHAND) mismatch")
	}

	// Check item state
	if !item.IsEquipped() {
		t.Errorf("IsEquipped() = false, want true")
	}
	if item.Slot() != PaperdollRHand {
		t.Errorf("Slot() = %d, want %d", item.Slot(), PaperdollRHand)
	}
	if item.Location() != ItemLocationPaperdoll {
		t.Errorf("Location() = %v, want %v", item.Location(), ItemLocationPaperdoll)
	}

	// Unequip item
	unequipped := inv.UnequipItem(PaperdollRHand)
	if unequipped != item {
		t.Errorf("UnequipItem(RHAND) mismatch")
	}

	// Check paperdoll slot empty
	equipped = inv.GetPaperdollItem(PaperdollRHand)
	if equipped != nil {
		t.Errorf("GetPaperdollItem(RHAND) = %v, want nil (after unequip)", equipped)
	}

	// Check item state
	if item.IsEquipped() {
		t.Errorf("IsEquipped() = true, want false (after unequip)")
	}
	if item.Slot() != -1 {
		t.Errorf("Slot() = %d, want -1 (after unequip)", item.Slot())
	}
	if item.Location() != ItemLocationInventory {
		t.Errorf("Location() = %v, want %v", item.Location(), ItemLocationInventory)
	}
}

// TestInventory_EquipItem_Validation verifies equip validation.
func TestInventory_EquipItem_Validation(t *testing.T) {
	inv := NewInventory(100)

	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item, _ := NewItem(1000, 1, 100, 1, template)

	// Equip item not in inventory should fail
	err := inv.EquipItem(item, PaperdollRHand)
	if err == nil {
		t.Errorf("EquipItem(not in inventory) expected error, got nil")
	}

	// Add item to inventory
	inv.AddItem(item)

	// Equip to invalid slot should fail
	err = inv.EquipItem(item, -1)
	if err == nil {
		t.Errorf("EquipItem(slot=-1) expected error, got nil")
	}

	err = inv.EquipItem(item, PaperdollTotalSlots)
	if err == nil {
		t.Errorf("EquipItem(slot=%d) expected error, got nil", PaperdollTotalSlots)
	}

	// Equip nil item should fail
	err = inv.EquipItem(nil, PaperdollRHand)
	if err == nil {
		t.Errorf("EquipItem(nil) expected error, got nil")
	}
}

// TestInventory_RemoveEquippedItem verifies removing equipped item auto-unequips.
func TestInventory_RemoveEquippedItem(t *testing.T) {
	inv := NewInventory(100)

	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item, _ := NewItem(1000, 1, 100, 1, template)

	// Add and equip item
	inv.AddItem(item)
	inv.EquipItem(item, PaperdollRHand)

	// Remove item (should auto-unequip)
	removed := inv.RemoveItem(1000)
	if removed != item {
		t.Errorf("RemoveItem() mismatch")
	}

	// Check paperdoll slot empty
	equipped := inv.GetPaperdollItem(PaperdollRHand)
	if equipped != nil {
		t.Errorf("GetPaperdollItem(RHAND) = %v, want nil (after removal)", equipped)
	}

	// Check item state
	if item.IsEquipped() {
		t.Errorf("IsEquipped() = true, want false (after removal)")
	}
	if item.Location() != ItemLocationVoid {
		t.Errorf("Location() = %v, want %v", item.Location(), ItemLocationVoid)
	}
}

// TestInventory_GetItems verifies GetItems returns copy.
func TestInventory_GetItems(t *testing.T) {
	inv := NewInventory(100)

	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item1, _ := NewItem(1000, 1, 100, 1, template)
	item2, _ := NewItem(1001, 1, 100, 1, template)

	inv.AddItem(item1)
	inv.AddItem(item2)

	// Get items
	items := inv.GetItems()

	if len(items) != 2 {
		t.Errorf("GetItems() len = %d, want 2", len(items))
	}

	// Modify slice should not affect inventory
	items[0] = nil

	if inv.GetItem(1000) == nil {
		t.Errorf("GetItem(1000) = nil, modification of GetItems() affected inventory")
	}
}

// TestInventory_GetEquippedItems verifies GetEquippedItems returns only equipped.
func TestInventory_GetEquippedItems(t *testing.T) {
	inv := NewInventory(100)

	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item1, _ := NewItem(1000, 1, 100, 1, template)
	item2, _ := NewItem(1001, 1, 100, 1, template)
	item3, _ := NewItem(1002, 1, 100, 1, template)

	inv.AddItem(item1)
	inv.AddItem(item2)
	inv.AddItem(item3)

	// Equip two items
	inv.EquipItem(item1, PaperdollRHand)
	inv.EquipItem(item2, PaperdollHead)

	// Get equipped items
	equipped := inv.GetEquippedItems()

	if len(equipped) != 2 {
		t.Errorf("GetEquippedItems() len = %d, want 2", len(equipped))
	}

	// Verify equipped items
	foundItem1 := false
	foundItem2 := false
	for _, item := range equipped {
		if item == item1 {
			foundItem1 = true
		}
		if item == item2 {
			foundItem2 = true
		}
		if item == item3 {
			t.Errorf("GetEquippedItems() contains item3 (not equipped)")
		}
	}

	if !foundItem1 {
		t.Errorf("GetEquippedItems() missing item1")
	}
	if !foundItem2 {
		t.Errorf("GetEquippedItems() missing item2")
	}
}

// TestInventory_Count verifies Count/TotalCount.
func TestInventory_Count(t *testing.T) {
	inv := NewInventory(100)

	template := &ItemTemplate{
		ItemID: 1,
		Name:   "Sword",
		Type:   ItemTypeWeapon,
	}

	item1, _ := NewItem(1000, 1, 100, 1, template)
	item2, _ := NewItem(1001, 1, 100, 1, template)

	inv.AddItem(item1)
	inv.AddItem(item2)

	// Both in inventory (not equipped)
	if inv.Count() != 2 {
		t.Errorf("Count() = %d, want 2 (both in inventory)", inv.Count())
	}
	if inv.TotalCount() != 2 {
		t.Errorf("TotalCount() = %d, want 2", inv.TotalCount())
	}

	// Equip one item
	inv.EquipItem(item1, PaperdollRHand)

	// Count excludes equipped, TotalCount includes
	if inv.Count() != 1 {
		t.Errorf("Count() = %d, want 1 (one equipped excluded)", inv.Count())
	}
	if inv.TotalCount() != 2 {
		t.Errorf("TotalCount() = %d, want 2 (includes equipped)", inv.TotalCount())
	}
}

// TestInventory_PaperdollSlotBounds verifies GetPaperdollItem bounds check.
func TestInventory_PaperdollSlotBounds(t *testing.T) {
	inv := NewInventory(100)

	// Negative slot
	item := inv.GetPaperdollItem(-1)
	if item != nil {
		t.Errorf("GetPaperdollItem(-1) = %v, want nil", item)
	}

	// Out of bounds slot
	item = inv.GetPaperdollItem(PaperdollTotalSlots)
	if item != nil {
		t.Errorf("GetPaperdollItem(%d) = %v, want nil", PaperdollTotalSlots, item)
	}

	// Valid slot (empty)
	item = inv.GetPaperdollItem(PaperdollRHand)
	if item != nil {
		t.Errorf("GetPaperdollItem(RHAND) = %v, want nil (empty slot)", item)
	}
}
