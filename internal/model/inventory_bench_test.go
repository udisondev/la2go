package model

import (
	"sync"
	"testing"
)

// --- helpers ---

func benchItem(objectID uint32, itemID int32) *Item {
	tmpl := &ItemTemplate{
		ItemID:    itemID,
		Name:      "BenchItem",
		Type:      ItemTypeWeapon,
		PAtk:      100,
		PDef:      0,
		Weight:    120,
		Stackable: false,
		Tradeable: true,
	}
	item, err := NewItem(objectID, itemID, 1, 1, tmpl)
	if err != nil {
		panic(err)
	}
	return item
}

func benchArmorItem(objectID uint32, itemID int32, slot ArmorSlot) *Item {
	tmpl := &ItemTemplate{
		ItemID:   itemID,
		Name:     "BenchArmor",
		Type:     ItemTypeArmor,
		PDef:     50,
		BodyPart: slot,
		Weight:   200,
	}
	item, err := NewItem(objectID, itemID, 1, 1, tmpl)
	if err != nil {
		panic(err)
	}
	return item
}

func benchInventoryWithItems(count int) *Inventory {
	inv := NewInventory(1)
	for i := range count {
		item := benchItem(uint32(i+1), int32(i+100))
		if err := inv.AddItem(item); err != nil {
			panic(err)
		}
	}
	return inv
}

// --- AddItem benchmarks ---

// BenchmarkInventory_AddItem_Empty benchmarks adding item to empty inventory.
// Expected: ~50-100ns (map insert + RWMutex.Lock).
func BenchmarkInventory_AddItem_Empty(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)
	items := make([]*Item, b.N)
	for i := range b.N {
		items[i] = benchItem(uint32(i+1), int32(i+100))
	}

	b.ResetTimer()
	for i := range b.N {
		_ = inv.AddItem(items[i])
	}
}

// BenchmarkInventory_AddItem_50Items benchmarks adding to inventory with 50 items.
// Expected: ~50-100ns (map insert with pre-existing entries).
func BenchmarkInventory_AddItem_50Items(b *testing.B) {
	b.ReportAllocs()
	inv := benchInventoryWithItems(50)
	items := make([]*Item, b.N)
	for i := range b.N {
		items[i] = benchItem(uint32(i+1000), int32(i+1000))
	}

	b.ResetTimer()
	for i := range b.N {
		_ = inv.AddItem(items[i])
	}
}

// BenchmarkInventory_AddItem_Concurrent benchmarks concurrent item additions.
// Expected: measures RWMutex contention under parallel load.
func BenchmarkInventory_AddItem_Concurrent(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)
	var counter uint32

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var mu sync.Mutex
		for pb.Next() {
			mu.Lock()
			counter++
			id := counter
			mu.Unlock()
			item := benchItem(id, int32(id+100))
			_ = inv.AddItem(item)
		}
	})
}

// --- RemoveItem benchmarks ---

// BenchmarkInventory_RemoveItem benchmarks removing an existing item.
// Expected: ~50-100ns (map delete + RWMutex.Lock).
func BenchmarkInventory_RemoveItem(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)
	items := make([]*Item, b.N)
	for i := range b.N {
		item := benchItem(uint32(i+1), int32(i+100))
		items[i] = item
		if err := inv.AddItem(item); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := range b.N {
		_ = inv.RemoveItem(items[i].ObjectID())
	}
}

// BenchmarkInventory_RemoveItem_NotFound benchmarks removing non-existent item.
// Expected: ~30-50ns (map lookup miss + RWMutex.Lock).
func BenchmarkInventory_RemoveItem_NotFound(b *testing.B) {
	b.ReportAllocs()
	inv := benchInventoryWithItems(50)

	b.ResetTimer()
	for range b.N {
		_ = inv.RemoveItem(999999)
	}
}

// --- EquipItem / UnequipItem benchmarks ---

// BenchmarkInventory_EquipItem benchmarks equipping an item to paperdoll slot.
// Expected: ~30-50ns (array write + item.SetSlot + RWMutex.Lock).
func BenchmarkInventory_EquipItem(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)
	item := benchArmorItem(1, 100, ArmorSlotChest)
	if err := inv.AddItem(item); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		_ = inv.EquipItem(item, PaperdollChest)
	}
}

// BenchmarkInventory_UnequipItem benchmarks unequipping from paperdoll.
// Expected: ~30-50ns (array clear + item.SetSlot + RWMutex.Lock).
func BenchmarkInventory_UnequipItem(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)
	item := benchArmorItem(1, 100, ArmorSlotChest)
	if err := inv.AddItem(item); err != nil {
		b.Fatal(err)
	}
	if err := inv.EquipItem(item, PaperdollChest); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		_ = inv.UnequipItem(PaperdollChest)
		// Re-equip for next iteration
		b.StopTimer()
		_ = inv.EquipItem(item, PaperdollChest)
		b.StartTimer()
	}
}

// --- GetItems benchmarks ---

// BenchmarkInventory_GetItems_Empty benchmarks GetItems on empty inventory.
// Expected: ~10-20ns (empty slice creation).
func BenchmarkInventory_GetItems_Empty(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)

	b.ResetTimer()
	for range b.N {
		_ = inv.GetItems()
	}
}

// BenchmarkInventory_GetItems_50Items benchmarks defensive copy of 50 items.
// Expected: ~500ns-1us (slice alloc + 50 append iterations).
func BenchmarkInventory_GetItems_50Items(b *testing.B) {
	b.ReportAllocs()
	inv := benchInventoryWithItems(50)

	b.ResetTimer()
	for range b.N {
		_ = inv.GetItems()
	}
}

// BenchmarkInventory_GetItems_200Items benchmarks defensive copy of max inventory.
// Expected: ~2-4us (slice alloc + 200 append iterations).
func BenchmarkInventory_GetItems_200Items(b *testing.B) {
	b.ReportAllocs()
	inv := benchInventoryWithItems(200)

	b.ResetTimer()
	for range b.N {
		_ = inv.GetItems()
	}
}

// --- GetEquippedItems benchmarks ---

// BenchmarkInventory_GetEquippedItems_Empty benchmarks with no items equipped.
// Expected: ~50-100ns (17-slot iteration, all nil).
func BenchmarkInventory_GetEquippedItems_Empty(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)

	b.ResetTimer()
	for range b.N {
		_ = inv.GetEquippedItems()
	}
}

// BenchmarkInventory_GetEquippedItems_FullSet benchmarks with all 17 slots equipped.
// Expected: ~100-500ns (17-slot iteration + 17 appends).
func BenchmarkInventory_GetEquippedItems_FullSet(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)

	// Equip items in all paperdoll slots
	slots := []int32{
		PaperdollUnder, PaperdollLEar, PaperdollREar, PaperdollNeck,
		PaperdollLFinger, PaperdollRFinger, PaperdollHead, PaperdollRHand,
		PaperdollLHand, PaperdollGloves, PaperdollChest, PaperdollLegs,
		PaperdollFeet, PaperdollCloak, PaperdollFace, PaperdollHair, PaperdollHair2,
	}
	for i, slot := range slots {
		item := benchArmorItem(uint32(i+1), int32(i+100), ArmorSlotNone)
		if err := inv.AddItem(item); err != nil {
			b.Fatal(err)
		}
		if err := inv.EquipItem(item, slot); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for range b.N {
		_ = inv.GetEquippedItems()
	}
}

// --- GetItem / GetPaperdollItem benchmarks ---

// BenchmarkInventory_GetItem_Hit benchmarks map lookup for existing item.
// Expected: ~30-50ns (map lookup + RWMutex.RLock).
func BenchmarkInventory_GetItem_Hit(b *testing.B) {
	b.ReportAllocs()
	inv := benchInventoryWithItems(50)

	b.ResetTimer()
	for range b.N {
		_ = inv.GetItem(25)
	}
}

// BenchmarkInventory_GetItem_Miss benchmarks map lookup for non-existent item.
// Expected: ~30-50ns (map lookup miss + RWMutex.RLock).
func BenchmarkInventory_GetItem_Miss(b *testing.B) {
	b.ReportAllocs()
	inv := benchInventoryWithItems(50)

	b.ResetTimer()
	for range b.N {
		_ = inv.GetItem(999999)
	}
}

// BenchmarkInventory_GetPaperdollItem benchmarks paperdoll slot access.
// Expected: ~20-30ns (array index + RWMutex.RLock).
func BenchmarkInventory_GetPaperdollItem(b *testing.B) {
	b.ReportAllocs()
	inv := NewInventory(1)
	item := benchArmorItem(1, 100, ArmorSlotChest)
	if err := inv.AddItem(item); err != nil {
		b.Fatal(err)
	}
	if err := inv.EquipItem(item, PaperdollChest); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		_ = inv.GetPaperdollItem(PaperdollChest)
	}
}

// BenchmarkInventory_Count benchmarks item count retrieval.
// Expected: ~20-30ns (map len + RWMutex.RLock).
func BenchmarkInventory_Count(b *testing.B) {
	b.ReportAllocs()
	inv := benchInventoryWithItems(50)

	b.ResetTimer()
	for range b.N {
		_ = inv.Count()
	}
}
