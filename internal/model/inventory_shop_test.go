package model

import (
	"testing"
)

func TestInventory_FindItemByItemID(t *testing.T) {
	inv := NewInventory(1)

	tmpl := &ItemTemplate{ItemID: 100, Name: "Sword", Type: ItemTypeWeapon}
	item, err := NewItem(50001, 100, 1, 1, tmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(item); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	// Should find by itemID
	found := inv.FindItemByItemID(100)
	if found == nil {
		t.Fatal("FindItemByItemID() = nil, want item")
	}
	if found.ObjectID() != 50001 {
		t.Errorf("found.ObjectID() = %d, want 50001", found.ObjectID())
	}

	// Should return nil for missing itemID
	notFound := inv.FindItemByItemID(999)
	if notFound != nil {
		t.Errorf("FindItemByItemID(999) = %v, want nil", notFound)
	}
}

func TestInventory_Adena(t *testing.T) {
	inv := NewInventory(1)

	// Initially no Adena
	if adena := inv.GetAdena(); adena != 0 {
		t.Errorf("GetAdena() = %d, want 0", adena)
	}

	// Create Adena item
	adenaTmpl := &ItemTemplate{ItemID: AdenaItemID, Name: "Adena", Type: ItemTypeEtcItem, Stackable: true}
	adenaItem, err := NewItem(50000, AdenaItemID, 1, 10000, adenaTmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(adenaItem); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	// Check Adena count
	if adena := inv.GetAdena(); adena != 10000 {
		t.Errorf("GetAdena() = %d, want 10000", adena)
	}

	// Add Adena
	if err := inv.AddAdena(5000); err != nil {
		t.Fatalf("AddAdena() error: %v", err)
	}
	if adena := inv.GetAdena(); adena != 15000 {
		t.Errorf("GetAdena() after add = %d, want 15000", adena)
	}

	// Remove Adena
	if err := inv.RemoveAdena(3000); err != nil {
		t.Fatalf("RemoveAdena() error: %v", err)
	}
	if adena := inv.GetAdena(); adena != 12000 {
		t.Errorf("GetAdena() after remove = %d, want 12000", adena)
	}
}

func TestInventory_RemoveAdena_NotEnough(t *testing.T) {
	inv := NewInventory(1)

	adenaTmpl := &ItemTemplate{ItemID: AdenaItemID, Name: "Adena", Type: ItemTypeEtcItem, Stackable: true}
	adenaItem, err := NewItem(50000, AdenaItemID, 1, 100, adenaTmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(adenaItem); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	err = inv.RemoveAdena(200)
	if err == nil {
		t.Error("RemoveAdena(200) with 100 adena should return error")
	}
}

func TestInventory_AddAdena_NoAdenaItem(t *testing.T) {
	inv := NewInventory(1)

	err := inv.AddAdena(1000)
	if err == nil {
		t.Error("AddAdena() without adena item should return error")
	}
}

func TestInventory_GetSellableItems(t *testing.T) {
	inv := NewInventory(1)

	// Add Adena (should be excluded)
	adenaTmpl := &ItemTemplate{ItemID: AdenaItemID, Name: "Adena", Type: ItemTypeEtcItem, Stackable: true}
	adenaItem, err := NewItem(50000, AdenaItemID, 1, 10000, adenaTmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(adenaItem); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	// Add sellable sword
	swordTmpl := &ItemTemplate{ItemID: 100, Name: "Sword", Type: ItemTypeWeapon, Tradeable: true}
	sword, err := NewItem(50001, 100, 1, 1, swordTmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(sword); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	// Add non-tradeable quest item (should be excluded)
	questTmpl := &ItemTemplate{ItemID: 200, Name: "Quest Item", Type: ItemTypeQuestItem, Tradeable: false}
	questItem, err := NewItem(50002, 200, 1, 1, questTmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(questItem); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}

	// Get sellable items
	sellable := inv.GetSellableItems()

	if len(sellable) != 1 {
		t.Fatalf("GetSellableItems() count = %d, want 1", len(sellable))
	}
	if sellable[0].ItemID() != 100 {
		t.Errorf("sellable[0].ItemID = %d, want 100 (Sword)", sellable[0].ItemID())
	}
}

func TestInventory_GetSellableItems_ExcludesEquipped(t *testing.T) {
	inv := NewInventory(1)

	// Add and equip sword
	swordTmpl := &ItemTemplate{ItemID: 100, Name: "Sword", Type: ItemTypeWeapon, Tradeable: true}
	sword, err := NewItem(50001, 100, 1, 1, swordTmpl)
	if err != nil {
		t.Fatalf("NewItem() error: %v", err)
	}
	if err := inv.AddItem(sword); err != nil {
		t.Fatalf("AddItem() error: %v", err)
	}
	if _, err := inv.EquipItem(sword, PaperdollRHand); err != nil {
		t.Fatalf("EquipItem() error: %v", err)
	}

	sellable := inv.GetSellableItems()
	if len(sellable) != 0 {
		t.Errorf("GetSellableItems() count = %d, want 0 (equipped items excluded)", len(sellable))
	}
}
