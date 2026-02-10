package model

import (
	"sync"
	"testing"
	"time"
)

func TestItemLocation_String(t *testing.T) {
	tests := []struct {
		location ItemLocation
		want     string
	}{
		{ItemLocationVoid, "VOID"},
		{ItemLocationInventory, "INVENTORY"},
		{ItemLocationPaperdoll, "PAPERDOLL"},
		{ItemLocationWarehouse, "WAREHOUSE"},
		{ItemLocation(999), "UNKNOWN(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.location.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewItem(t *testing.T) {
	tests := []struct {
		name     string
		ownerID  int64
		itemType int32
		count    int32
		wantErr  bool
	}{
		{
			name:     "valid item",
			ownerID:  100,
			itemType: 57, // Adena
			count:    1000,
			wantErr:  false,
		},
		{
			name:     "count = 1",
			ownerID:  100,
			itemType: 1,
			count:    1,
			wantErr:  false,
		},
		{
			name:     "count = 0 (invalid)",
			ownerID:  100,
			itemType: 1,
			count:    0,
			wantErr:  true,
		},
		{
			name:     "count negative (invalid)",
			ownerID:  100,
			itemType: 1,
			count:    -10,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewItem(tt.ownerID, tt.itemType, tt.count)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewItem() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Errorf("NewItem() unexpected error = %v", err)
				return
			}

			if item == nil {
				t.Fatal("NewItem() returned nil")
			}

			// Проверяем поля
			if item.OwnerID() != tt.ownerID {
				t.Errorf("OwnerID() = %d, want %d", item.OwnerID(), tt.ownerID)
			}
			if item.ItemType() != tt.itemType {
				t.Errorf("ItemType() = %d, want %d", item.ItemType(), tt.itemType)
			}
			if item.Count() != tt.count {
				t.Errorf("Count() = %d, want %d", item.Count(), tt.count)
			}

			// Enchant должен быть 0
			if item.Enchant() != 0 {
				t.Errorf("Enchant() = %d, want 0", item.Enchant())
			}

			// Location должна быть INVENTORY
			loc, slotID := item.Location()
			if loc != ItemLocationInventory {
				t.Errorf("Location() = %v, want INVENTORY", loc)
			}
			if slotID != -1 {
				t.Errorf("SlotID() = %d, want -1", slotID)
			}

			// CreatedAt должно быть недавно
			if time.Since(item.CreatedAt()) > time.Second {
				t.Errorf("CreatedAt() = %v, want recent time", item.CreatedAt())
			}

			// IsInInventory должно быть true
			if !item.IsInInventory() {
				t.Error("IsInInventory() = false, want true")
			}
			if item.IsEquipped() {
				t.Error("IsEquipped() = true, want false")
			}
		})
	}
}

func TestItem_ImmutableFields(t *testing.T) {
	item, _ := NewItem(100, 57, 1000)

	// ItemID и OwnerID должны быть immutable
	id1 := item.ItemID()
	id2 := item.ItemID()
	if id1 != id2 {
		t.Errorf("ItemID changed: %d != %d", id1, id2)
	}

	ownerID1 := item.OwnerID()
	ownerID2 := item.OwnerID()
	if ownerID1 != ownerID2 {
		t.Errorf("OwnerID changed: %d != %d", ownerID1, ownerID2)
	}
}

func TestItem_Count(t *testing.T) {
	item, _ := NewItem(100, 57, 1000)

	// SetCount valid
	if err := item.SetCount(500); err != nil {
		t.Errorf("SetCount(500) error = %v", err)
	}
	if item.Count() != 500 {
		t.Errorf("After SetCount(500), Count() = %d", item.Count())
	}

	// SetCount zero (invalid)
	if err := item.SetCount(0); err == nil {
		t.Error("SetCount(0) error = nil, want error")
	}

	// SetCount negative (invalid)
	if err := item.SetCount(-10); err == nil {
		t.Error("SetCount(-10) error = nil, want error")
	}

	// Count не должно измениться после invalid SetCount
	if item.Count() != 500 {
		t.Errorf("After invalid SetCount, Count() = %d, want 500", item.Count())
	}
}

func TestItem_AddCount(t *testing.T) {
	item, _ := NewItem(100, 57, 1000)

	// AddCount positive
	if err := item.AddCount(500); err != nil {
		t.Errorf("AddCount(500) error = %v", err)
	}
	if item.Count() != 1500 {
		t.Errorf("After AddCount(500), Count() = %d, want 1500", item.Count())
	}

	// AddCount negative (valid if result > 0)
	if err := item.AddCount(-200); err != nil {
		t.Errorf("AddCount(-200) error = %v", err)
	}
	if item.Count() != 1300 {
		t.Errorf("After AddCount(-200), Count() = %d, want 1300", item.Count())
	}

	// AddCount negative resulting in 0 (invalid)
	if err := item.AddCount(-1300); err == nil {
		t.Error("AddCount(-1300) error = nil, want error (would result in 0)")
	}

	// AddCount negative resulting in negative (invalid)
	if err := item.AddCount(-2000); err == nil {
		t.Error("AddCount(-2000) error = nil, want error (would result in negative)")
	}

	// Count не должно измениться после invalid AddCount
	if item.Count() != 1300 {
		t.Errorf("After invalid AddCount, Count() = %d, want 1300", item.Count())
	}
}

func TestItem_Enchant(t *testing.T) {
	item, _ := NewItem(100, 1, 1)

	// Initial enchant = 0
	if item.Enchant() != 0 {
		t.Errorf("Initial Enchant() = %d, want 0", item.Enchant())
	}

	// SetEnchant valid
	if err := item.SetEnchant(5); err != nil {
		t.Errorf("SetEnchant(5) error = %v", err)
	}
	if item.Enchant() != 5 {
		t.Errorf("After SetEnchant(5), Enchant() = %d", item.Enchant())
	}

	// SetEnchant high value (valid)
	if err := item.SetEnchant(65); err != nil {
		t.Errorf("SetEnchant(65) error = %v", err)
	}
	if item.Enchant() != 65 {
		t.Errorf("After SetEnchant(65), Enchant() = %d", item.Enchant())
	}

	// SetEnchant negative (invalid)
	if err := item.SetEnchant(-1); err == nil {
		t.Error("SetEnchant(-1) error = nil, want error")
	}

	// Enchant не должно измениться после invalid SetEnchant
	if item.Enchant() != 65 {
		t.Errorf("After invalid SetEnchant, Enchant() = %d, want 65", item.Enchant())
	}
}

func TestItem_Location(t *testing.T) {
	item, _ := NewItem(100, 1, 1)

	// Initial location = INVENTORY, slot = -1
	loc, slot := item.Location()
	if loc != ItemLocationInventory {
		t.Errorf("Initial Location() = %v, want INVENTORY", loc)
	}
	if slot != -1 {
		t.Errorf("Initial SlotID = %d, want -1", slot)
	}

	// SetLocation PAPERDOLL
	item.SetLocation(ItemLocationPaperdoll, 5)
	loc, slot = item.Location()
	if loc != ItemLocationPaperdoll {
		t.Errorf("After SetLocation PAPERDOLL, Location() = %v", loc)
	}
	if slot != 5 {
		t.Errorf("After SetLocation slot 5, SlotID = %d", slot)
	}

	// IsEquipped должно быть true
	if !item.IsEquipped() {
		t.Error("IsEquipped() = false, want true")
	}
	if item.IsInInventory() {
		t.Error("IsInInventory() = true, want false")
	}

	// SetLocation WAREHOUSE
	item.SetLocation(ItemLocationWarehouse, 10)
	loc, slot = item.Location()
	if loc != ItemLocationWarehouse {
		t.Errorf("After SetLocation WAREHOUSE, Location() = %v", loc)
	}
	if slot != 10 {
		t.Errorf("After SetLocation slot 10, SlotID = %d", slot)
	}

	// Оба IsEquipped и IsInInventory должны быть false
	if item.IsEquipped() {
		t.Error("IsEquipped() = true, want false (in warehouse)")
	}
	if item.IsInInventory() {
		t.Error("IsInInventory() = true, want false (in warehouse)")
	}

	// SetLocation обратно в INVENTORY
	item.SetLocation(ItemLocationInventory, -1)
	if !item.IsInInventory() {
		t.Error("IsInInventory() = false, want true")
	}
}

func TestItem_SetItemID(t *testing.T) {
	item, _ := NewItem(100, 1, 1)

	if item.ItemID() != 0 {
		t.Errorf("Initial ItemID() = %d, want 0", item.ItemID())
	}

	// SetItemID (для repository.Create)
	item.SetItemID(999)

	if item.ItemID() != 999 {
		t.Errorf("After SetItemID, ItemID() = %d, want 999", item.ItemID())
	}
}

func TestItem_CreatedAt(t *testing.T) {
	item, _ := NewItem(100, 1, 1)

	// CreatedAt должно быть недавно
	if time.Since(item.CreatedAt()) > time.Second {
		t.Errorf("CreatedAt() = %v, want recent time", item.CreatedAt())
	}

	// SetCreatedAt (для загрузки из DB)
	customTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	item.SetCreatedAt(customTime)

	if item.CreatedAt() != customTime {
		t.Errorf("After SetCreatedAt, CreatedAt() = %v, want %v", item.CreatedAt(), customTime)
	}
}

func TestItem_ConcurrentCountUpdates(t *testing.T) {
	item, _ := NewItem(100, 57, 10000)

	const numUpdaters = 50
	var wg sync.WaitGroup
	wg.Add(numUpdaters)

	// Concurrent AddCount
	for range numUpdaters {
		go func() {
			defer wg.Done()

			for range 100 {
				_ = item.AddCount(1)
			}
		}()
	}

	wg.Wait()

	// Финальный count должен быть > 10000
	count := item.Count()
	expectedMin := int32(10000 + numUpdaters*100)
	if count < expectedMin {
		t.Errorf("After concurrent AddCount, Count() = %d, want >= %d", count, expectedMin)
	}
}

func TestItem_ConcurrentLocationUpdates(t *testing.T) {
	item, _ := NewItem(100, 1, 1)

	const numUpdaters = 50
	var wg sync.WaitGroup
	wg.Add(numUpdaters)

	// Concurrent SetLocation
	for i := range numUpdaters {
		go func(id int) {
			defer wg.Done()

			for j := range 100 {
				loc := ItemLocation(j % 4) // 0-3
				slot := int32(id*100 + j)
				item.SetLocation(loc, slot)
			}
		}(i)
	}

	wg.Wait()

	// Финальная location должна быть валидной
	loc, slot := item.Location()
	if loc < ItemLocationVoid || loc > ItemLocationWarehouse {
		t.Errorf("Invalid location after concurrent updates: %v", loc)
	}
	if slot < 0 {
		t.Errorf("Invalid slot after concurrent updates: %d", slot)
	}
}

func TestItem_MixedConcurrentAccess(t *testing.T) {
	item, _ := NewItem(100, 57, 1000)

	const numReaders = 50
	const numWriters = 10
	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	// Readers
	for range numReaders {
		go func() {
			defer wg.Done()

			for range 500 {
				_ = item.Count()
				_ = item.Enchant()
				_, _ = item.Location()
				_ = item.IsEquipped()
			}
		}()
	}

	// Writers
	for i := range numWriters {
		go func(id int) {
			defer wg.Done()

			for j := range 100 {
				_ = item.AddCount(1)
				_ = item.SetEnchant(int32(j % 10))
				loc := ItemLocation(j % 4)
				item.SetLocation(loc, int32(id))
			}
		}(i)
	}

	wg.Wait()

	// Финальные значения должны быть консистентными
	if item.Count() <= 1000 {
		t.Errorf("Count() = %d, want > 1000", item.Count())
	}
	if item.Enchant() < 0 {
		t.Errorf("Enchant() = %d, want >= 0", item.Enchant())
	}
}

// Benchmark для hot path methods
func BenchmarkItem_Count(b *testing.B) {
	item, _ := NewItem(100, 57, 1000)

	b.ResetTimer()
	for b.Loop() {
		_ = item.Count()
	}
}

func BenchmarkItem_AddCount(b *testing.B) {
	item, _ := NewItem(100, 57, 1000000000) // Large initial count

	b.ResetTimer()
	for b.Loop() {
		_ = item.AddCount(1)
	}
}

func BenchmarkItem_Location(b *testing.B) {
	item, _ := NewItem(100, 57, 1000)

	b.ResetTimer()
	for b.Loop() {
		_, _ = item.Location()
	}
}
