package model

import (
	"testing"
)

func newTestInventoryWithAdena(t *testing.T, adenaCount int32) *Inventory {
	t.Helper()
	inv := NewInventory(1)
	adenaTmpl := &ItemTemplate{ItemID: AdenaItemID, Name: "Adena", Type: ItemTypeEtcItem, Stackable: true, Tradeable: true}
	adenaItem, err := NewItem(50000, AdenaItemID, 1, adenaCount, adenaTmpl)
	if err != nil {
		t.Fatalf("NewItem(adena) error: %v", err)
	}
	if err := inv.AddItem(adenaItem); err != nil {
		t.Fatalf("AddItem(adena) error: %v", err)
	}
	return inv
}

func TestWarehouse_DepositFullStack(t *testing.T) {
	inv := newTestInventoryWithAdena(t, 10000)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Sword", Type: ItemTypeWeapon, Tradeable: true}
	sword, err := NewItem(50001, 100, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(sword); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	if err := inv.DepositToWarehouse(50001, 1); err != nil {
		t.Fatalf("DepositToWarehouse() error: %v", err)
	}

	// Item should be in warehouse, not inventory
	if inv.GetItem(50001) != nil {
		t.Error("item should not be in inventory after deposit")
	}
	if inv.GetWarehouseItem(50001) == nil {
		t.Error("item should be in warehouse after deposit")
	}
	if inv.WarehouseCount() != 1 {
		t.Errorf("WarehouseCount() = %d, want 1", inv.WarehouseCount())
	}
}

func TestWarehouse_DepositEquippedFails(t *testing.T) {
	inv := newTestInventoryWithAdena(t, 10000)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Sword", Type: ItemTypeWeapon, Tradeable: true}
	sword, err := NewItem(50001, 100, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(sword); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}
	if _, err := inv.EquipItem(sword, PaperdollRHand); err != nil {
		t.Fatalf("EquipItem() error: %v", err)
	}

	err = inv.DepositToWarehouse(50001, 1)
	if err == nil {
		t.Error("DepositToWarehouse() should fail for equipped item")
	}
}

func TestWarehouse_DepositNotFound(t *testing.T) {
	inv := NewInventory(1)

	err := inv.DepositToWarehouse(99999, 1)
	if err == nil {
		t.Error("DepositToWarehouse() should fail for non-existent item")
	}
}

func TestWarehouse_WithdrawFullStack(t *testing.T) {
	inv := NewInventory(1)

	// Add item directly to warehouse
	tmpl := &ItemTemplate{ItemID: 100, Name: "Sword", Type: ItemTypeWeapon, Tradeable: true}
	sword, err := NewItem(50001, 100, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddWarehouseItem(sword); err != nil {
		t.Fatalf("AddWarehouseItem() error: %v", err)
	}

	if err := inv.WithdrawFromWarehouse(50001, 1, 60001); err != nil {
		t.Fatalf("WithdrawFromWarehouse() error: %v", err)
	}

	// Item should be in inventory, not warehouse
	if inv.GetWarehouseItem(50001) != nil {
		t.Error("item should not be in warehouse after withdraw")
	}
	if inv.GetItem(50001) == nil {
		t.Error("item should be in inventory after withdraw")
	}
}

func TestWarehouse_WithdrawNotFound(t *testing.T) {
	inv := NewInventory(1)

	err := inv.WithdrawFromWarehouse(99999, 1, 60001)
	if err == nil {
		t.Error("WithdrawFromWarehouse() should fail for non-existent item")
	}
}

func TestWarehouse_DepositSplitStack(t *testing.T) {
	inv := NewInventory(1)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Arrow", Type: ItemTypeEtcItem, Stackable: true, Tradeable: true}
	arrows, err := NewItem(50001, 100, 1, 500, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(arrows); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	if err := inv.DepositToWarehouseSplit(50001, 200, 60001); err != nil {
		t.Fatalf("DepositToWarehouseSplit() error: %v", err)
	}

	// Inventory should have 300 arrows
	invItem := inv.GetItem(50001)
	if invItem == nil {
		t.Fatal("item should still be in inventory")
	}
	if invItem.Count() != 300 {
		t.Errorf("inventory count = %d, want 300", invItem.Count())
	}

	// Warehouse should have 200 arrows
	whItems := inv.GetWarehouseItems()
	if len(whItems) != 1 {
		t.Fatalf("warehouse items count = %d, want 1", len(whItems))
	}
	if whItems[0].Count() != 200 {
		t.Errorf("warehouse count = %d, want 200", whItems[0].Count())
	}
}

func TestWarehouse_WithdrawSplitStack(t *testing.T) {
	inv := NewInventory(1)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Arrow", Type: ItemTypeEtcItem, Stackable: true, Tradeable: true}
	arrows, err := NewItem(50001, 100, 1, 500, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddWarehouseItem(arrows); err != nil {
		t.Fatalf("AddWarehouseItem() error: %v", err)
	}

	if err := inv.WithdrawFromWarehouse(50001, 200, 60001); err != nil {
		t.Fatalf("WithdrawFromWarehouse() error: %v", err)
	}

	// Warehouse should have 300 arrows
	whItem := inv.GetWarehouseItem(50001)
	if whItem == nil {
		t.Fatal("item should still be in warehouse")
	}
	if whItem.Count() != 300 {
		t.Errorf("warehouse count = %d, want 300", whItem.Count())
	}

	// Inventory should have 200 arrows
	invItem := inv.GetItem(60001)
	if invItem == nil {
		t.Fatal("split item should be in inventory")
	}
	if invItem.Count() != 200 {
		t.Errorf("inventory count = %d, want 200", invItem.Count())
	}
}

func TestWarehouse_DepositSplitMergesExisting(t *testing.T) {
	inv := NewInventory(1)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Arrow", Type: ItemTypeEtcItem, Stackable: true, Tradeable: true}

	// Already have 100 arrows in warehouse
	whArrows, err := NewItem(60001, 100, 1, 100, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddWarehouseItem(whArrows); err != nil {
		t.Fatalf("AddWarehouseItem() error: %v", err)
	}

	// 300 arrows in inventory
	invArrows, err := NewItem(50001, 100, 1, 300, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(invArrows); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	// Deposit 200 from inventory to warehouse → should merge with existing 100
	if err := inv.DepositToWarehouseSplit(50001, 200, 70001); err != nil {
		t.Fatalf("DepositToWarehouseSplit() error: %v", err)
	}

	// Inventory: 100 arrows
	if invArrows.Count() != 100 {
		t.Errorf("inventory count = %d, want 100", invArrows.Count())
	}

	// Warehouse: 300 arrows (100 existing + 200 deposited)
	if whArrows.Count() != 300 {
		t.Errorf("warehouse count = %d, want 300", whArrows.Count())
	}

	// Should still be only 1 warehouse item (merged, not created new)
	if inv.WarehouseCount() != 1 {
		t.Errorf("WarehouseCount() = %d, want 1 (should merge)", inv.WarehouseCount())
	}
}

func TestWarehouse_GetDepositableItems(t *testing.T) {
	inv := NewInventory(1)

	// Adena — not depositable
	adenaTmpl := &ItemTemplate{ItemID: AdenaItemID, Name: "Adena", Type: ItemTypeEtcItem, Stackable: true, Tradeable: true}
	adena, _ := NewItem(50000, AdenaItemID, 1, 10000, adenaTmpl)
	inv.AddItem(adena)

	// Sword — depositable
	swordTmpl := &ItemTemplate{ItemID: 100, Name: "Sword", Type: ItemTypeWeapon, Tradeable: true}
	sword, _ := NewItem(50001, 100, 1, 1, swordTmpl)
	inv.AddItem(sword)

	// Quest item — not depositable (non-tradeable)
	questTmpl := &ItemTemplate{ItemID: 200, Name: "Quest Item", Type: ItemTypeQuestItem, Tradeable: false}
	quest, _ := NewItem(50002, 200, 1, 1, questTmpl)
	inv.AddItem(quest)

	// Equipped armor — not depositable
	armorTmpl := &ItemTemplate{ItemID: 300, Name: "Armor", Type: ItemTypeArmor, Tradeable: true, BodyPart: ArmorSlotChest}
	armor, _ := NewItem(50003, 300, 1, 1, armorTmpl)
	inv.AddItem(armor)
	inv.EquipItem(armor, PaperdollChest)

	depositable := inv.GetDepositableItems()
	if len(depositable) != 1 {
		t.Fatalf("GetDepositableItems() count = %d, want 1", len(depositable))
	}
	if depositable[0].ItemID() != 100 {
		t.Errorf("depositable[0].ItemID() = %d, want 100 (Sword)", depositable[0].ItemID())
	}
}

func TestWarehouse_CountItemsByID(t *testing.T) {
	inv := NewInventory(1)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Arrow", Type: ItemTypeEtcItem, Stackable: true, Tradeable: true}

	// Two stacks of the same item
	stack1, _ := NewItem(50001, 100, 1, 200, tmpl)
	inv.AddItem(stack1)

	stack2, _ := NewItem(50002, 100, 1, 300, tmpl)
	inv.AddItem(stack2)

	count := inv.CountItemsByID(100)
	if count != 500 {
		t.Errorf("CountItemsByID(100) = %d, want 500", count)
	}

	// Non-existent item
	count = inv.CountItemsByID(999)
	if count != 0 {
		t.Errorf("CountItemsByID(999) = %d, want 0", count)
	}
}

func TestWarehouse_RemoveItemsByID(t *testing.T) {
	inv := NewInventory(1)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Arrow", Type: ItemTypeEtcItem, Stackable: true, Tradeable: true}
	arrows, _ := NewItem(50001, 100, 1, 500, tmpl)
	inv.AddItem(arrows)

	// Remove 200 arrows
	removed := inv.RemoveItemsByID(100, 200)
	if removed != 200 {
		t.Errorf("RemoveItemsByID() returned %d, want 200", removed)
	}
	if arrows.Count() != 300 {
		t.Errorf("item count = %d, want 300", arrows.Count())
	}

	// Remove all remaining
	removed = inv.RemoveItemsByID(100, 300)
	if removed != 300 {
		t.Errorf("RemoveItemsByID() returned %d, want 300", removed)
	}
	if inv.GetItem(50001) != nil {
		t.Error("item should be removed from inventory")
	}
}

func TestWarehouse_RemoveItemsByID_NotEnough(t *testing.T) {
	inv := NewInventory(1)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Arrow", Type: ItemTypeEtcItem, Stackable: true, Tradeable: true}
	arrows, _ := NewItem(50001, 100, 1, 100, tmpl)
	inv.AddItem(arrows)

	// Try to remove more than available
	removed := inv.RemoveItemsByID(100, 500)
	if removed != 100 {
		t.Errorf("RemoveItemsByID() returned %d, want 100 (all available)", removed)
	}
}

func TestWarehouse_EquipAutoUnequip(t *testing.T) {
	inv := NewInventory(1)

	swordTmpl := &ItemTemplate{ItemID: 100, Name: "Old Sword", Type: ItemTypeWeapon, Tradeable: true}
	oldSword, _ := NewItem(50001, 100, 1, 1, swordTmpl)
	inv.AddItem(oldSword)

	newSwordTmpl := &ItemTemplate{ItemID: 101, Name: "New Sword", Type: ItemTypeWeapon, Tradeable: true}
	newSword, _ := NewItem(50002, 101, 1, 1, newSwordTmpl)
	inv.AddItem(newSword)

	// Equip old sword
	prev, err := inv.EquipItem(oldSword, PaperdollRHand)
	if err != nil {
		t.Fatalf("EquipItem(old) error: %v", err)
	}
	if prev != nil {
		t.Error("first equip should have nil previous item")
	}

	// Equip new sword → should auto-unequip old
	prev, err = inv.EquipItem(newSword, PaperdollRHand)
	if err != nil {
		t.Fatalf("EquipItem(new) error: %v", err)
	}
	if prev == nil {
		t.Fatal("second equip should return previously equipped item")
	}
	if prev.ObjectID() != 50001 {
		t.Errorf("previous item objectID = %d, want 50001", prev.ObjectID())
	}

	// Old sword should be unequipped
	if oldSword.IsEquipped() {
		t.Error("old sword should be unequipped")
	}

	// New sword should be equipped
	if !newSword.IsEquipped() {
		t.Error("new sword should be equipped")
	}

	equipped := inv.GetPaperdollItem(PaperdollRHand)
	if equipped == nil || equipped.ObjectID() != 50002 {
		t.Error("paperdoll RHand should have new sword")
	}
}
