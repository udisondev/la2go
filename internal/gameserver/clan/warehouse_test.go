package clan

import "testing"

func TestWarehouse_AddAndRetrieve(t *testing.T) {
	wh := NewWarehouse()

	item := &WarehouseItem{ObjectID: 1, ItemID: 57, Count: 1000, EnchantLevel: 0}
	if err := wh.AddItem(item); err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	if wh.Count() != 1 {
		t.Errorf("Count = %d, want 1", wh.Count())
	}

	got := wh.Item(1)
	if got == nil {
		t.Fatal("Item(1) = nil, want item")
	}
	if got.Count != 1000 {
		t.Errorf("Item.Count = %d, want 1000", got.Count)
	}
}

func TestWarehouse_StackableItems(t *testing.T) {
	wh := NewWarehouse()

	// Add Adena (itemID=57).
	item1 := &WarehouseItem{ObjectID: 1, ItemID: 57, Count: 500}
	if err := wh.AddItem(item1); err != nil {
		t.Fatalf("AddItem first: %v", err)
	}

	// Add more Adena — should stack.
	item2 := &WarehouseItem{ObjectID: 2, ItemID: 57, Count: 300}
	if err := wh.AddItem(item2); err != nil {
		t.Fatalf("AddItem stack: %v", err)
	}

	// Should still be 1 item (stacked).
	if wh.Count() != 1 {
		t.Errorf("Count = %d, want 1 (stacked)", wh.Count())
	}

	got := wh.Item(1)
	if got.Count != 800 {
		t.Errorf("Stacked Count = %d, want 800", got.Count)
	}
}

func TestWarehouse_DifferentEnchantNoStack(t *testing.T) {
	wh := NewWarehouse()

	// Same item ID but different enchant levels should NOT stack.
	item1 := &WarehouseItem{ObjectID: 1, ItemID: 100, Count: 1, EnchantLevel: 0}
	item2 := &WarehouseItem{ObjectID: 2, ItemID: 100, Count: 1, EnchantLevel: 5}

	wh.AddItem(item1) //nolint:errcheck
	wh.AddItem(item2) //nolint:errcheck

	if wh.Count() != 2 {
		t.Errorf("Count = %d, want 2 (different enchant)", wh.Count())
	}
}

func TestWarehouse_RemoveItem(t *testing.T) {
	wh := NewWarehouse()

	item := &WarehouseItem{ObjectID: 1, ItemID: 57, Count: 1000}
	wh.AddItem(item) //nolint:errcheck

	// Remove some.
	if err := wh.RemoveItem(1, 400); err != nil {
		t.Fatalf("RemoveItem: %v", err)
	}
	if wh.Item(1).Count != 600 {
		t.Errorf("Count after remove = %d, want 600", wh.Item(1).Count)
	}

	// Remove all — item should be deleted.
	if err := wh.RemoveItem(1, 600); err != nil {
		t.Fatalf("RemoveItem all: %v", err)
	}
	if wh.Count() != 0 {
		t.Errorf("Count = %d after full remove, want 0", wh.Count())
	}
}

func TestWarehouse_RemoveItem_Errors(t *testing.T) {
	wh := NewWarehouse()

	// Not found.
	if err := wh.RemoveItem(999, 1); err != ErrItemNotFound {
		t.Errorf("RemoveItem(999) = %v, want ErrItemNotFound", err)
	}

	item := &WarehouseItem{ObjectID: 1, ItemID: 57, Count: 10}
	wh.AddItem(item) //nolint:errcheck

	// Insufficient.
	if err := wh.RemoveItem(1, 100); err != ErrInsufficientItem {
		t.Errorf("RemoveItem(1, 100) = %v, want ErrInsufficientItem", err)
	}
}

func TestWarehouse_Full(t *testing.T) {
	wh := NewWarehouse()

	for i := int64(1); i <= MaxWarehouseSlots; i++ {
		item := &WarehouseItem{ObjectID: i, ItemID: int32(i), Count: 1}
		if err := wh.AddItem(item); err != nil {
			t.Fatalf("AddItem(%d): %v", i, err)
		}
	}

	// Next item should fail (different item ID so no stacking).
	item := &WarehouseItem{ObjectID: MaxWarehouseSlots + 1, ItemID: MaxWarehouseSlots + 1, Count: 1}
	if err := wh.AddItem(item); err != ErrWarehouseFull {
		t.Errorf("AddItem when full = %v, want ErrWarehouseFull", err)
	}
}

func TestWarehouse_Items_Snapshot(t *testing.T) {
	wh := NewWarehouse()

	for i := int64(1); i <= 3; i++ {
		wh.AddItem(&WarehouseItem{ObjectID: i, ItemID: int32(i), Count: 1}) //nolint:errcheck
	}

	items := wh.Items()
	if len(items) != 3 {
		t.Errorf("Items() = %d items, want 3", len(items))
	}
}

func TestWarehouse_Clear(t *testing.T) {
	wh := NewWarehouse()
	wh.AddItem(&WarehouseItem{ObjectID: 1, ItemID: 57, Count: 100}) //nolint:errcheck

	wh.Clear()
	if wh.Count() != 0 {
		t.Errorf("Count after Clear = %d, want 0", wh.Count())
	}
}
