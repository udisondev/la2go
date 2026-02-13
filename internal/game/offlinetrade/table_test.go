package offlinetrade

import (
	"sync"
	"testing"
	"time"

	"github.com/udisondev/la2go/internal/model"
)

func newTestTrader(objectID uint32, charID int64, storeType model.PrivateStoreType) *Trader {
	return &Trader{
		CharacterID: charID,
		ObjectID:    objectID,
		AccountName: "testaccount",
		StoreType:   storeType,
		Title:       "Test Store",
		Items: []TradeEntry{
			{ItemIdentifier: 57, Count: 100, Price: 1000},
		},
		StartedAt: time.Now(),
	}
}

func TestNewTable(t *testing.T) {
	tbl := NewTable(48 * time.Hour)
	if tbl == nil {
		t.Fatal("NewTable returned nil")
	}
	if tbl.Count() != 0 {
		t.Errorf("Count() = %d, want 0", tbl.Count())
	}
}

func TestTable_Add(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 1, model.StoreSell)

	if err := tbl.Add(trader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if tbl.Count() != 1 {
		t.Errorf("Count() = %d, want 1", tbl.Count())
	}
	if !tbl.IsOfflineTrader(1001) {
		t.Error("IsOfflineTrader(1001) = false, want true")
	}
}

func TestTable_Add_Duplicate(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 1, model.StoreSell)

	if err := tbl.Add(trader); err != nil {
		t.Fatalf("first Add failed: %v", err)
	}
	if err := tbl.Add(trader); err == nil {
		t.Error("duplicate Add should return error")
	}
}

func TestTable_Add_InvalidStoreType(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 1, model.StoreNone)

	if err := tbl.Add(trader); err == nil {
		t.Error("Add with StoreNone should return error")
	}
}

func TestTable_Add_NilTrader(t *testing.T) {
	tbl := NewTable(0)
	if err := tbl.Add(nil); err == nil {
		t.Error("Add(nil) should return error")
	}
}

func TestTable_Add_ZeroObjectID(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(0, 1, model.StoreSell)
	if err := tbl.Add(trader); err == nil {
		t.Error("Add with objectID=0 should return error")
	}
}

func TestTable_Remove(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 1, model.StoreSell)
	if err := tbl.Add(trader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	removed := tbl.Remove(1001)
	if removed == nil {
		t.Fatal("Remove returned nil")
	}
	if removed.ObjectID != 1001 {
		t.Errorf("removed.ObjectID = %d, want 1001", removed.ObjectID)
	}
	if tbl.Count() != 0 {
		t.Errorf("Count() = %d, want 0", tbl.Count())
	}
	if tbl.IsOfflineTrader(1001) {
		t.Error("IsOfflineTrader(1001) should be false after Remove")
	}
}

func TestTable_Remove_NotFound(t *testing.T) {
	tbl := NewTable(0)
	if removed := tbl.Remove(9999); removed != nil {
		t.Error("Remove non-existent should return nil")
	}
}

func TestTable_RemoveByCharacter(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 42, model.StoreBuy)
	if err := tbl.Add(trader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	removed := tbl.RemoveByCharacter(42)
	if removed == nil {
		t.Fatal("RemoveByCharacter returned nil")
	}
	if removed.CharacterID != 42 {
		t.Errorf("removed.CharacterID = %d, want 42", removed.CharacterID)
	}
	if tbl.Count() != 0 {
		t.Errorf("Count() = %d, want 0", tbl.Count())
	}
}

func TestTable_RemoveByCharacter_NotFound(t *testing.T) {
	tbl := NewTable(0)
	if removed := tbl.RemoveByCharacter(9999); removed != nil {
		t.Error("RemoveByCharacter non-existent should return nil")
	}
}

func TestTable_Get(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 1, model.StoreSell)
	if err := tbl.Add(trader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	got := tbl.Get(1001)
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if got.Title != "Test Store" {
		t.Errorf("Title = %q, want %q", got.Title, "Test Store")
	}
}

func TestTable_Get_NotFound(t *testing.T) {
	tbl := NewTable(0)
	if got := tbl.Get(9999); got != nil {
		t.Error("Get non-existent should return nil")
	}
}

func TestTable_IsCharacterOffline(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 42, model.StoreSell)
	if err := tbl.Add(trader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if !tbl.IsCharacterOffline(42) {
		t.Error("IsCharacterOffline(42) = false, want true")
	}
	if tbl.IsCharacterOffline(99) {
		t.Error("IsCharacterOffline(99) = true, want false")
	}
}

func TestTable_ForEach(t *testing.T) {
	tbl := NewTable(0)
	for i := range 5 {
		trader := newTestTrader(uint32(1000+i), int64(i+1), model.StoreSell)
		if err := tbl.Add(trader); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	var count int
	tbl.ForEach(func(trader *Trader) bool {
		count++
		return true
	})
	if count != 5 {
		t.Errorf("ForEach visited %d traders, want 5", count)
	}
}

func TestTable_ForEach_EarlyStop(t *testing.T) {
	tbl := NewTable(0)
	for i := range 5 {
		trader := newTestTrader(uint32(1000+i), int64(i+1), model.StoreSell)
		if err := tbl.Add(trader); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	var count int
	tbl.ForEach(func(trader *Trader) bool {
		count++
		return count < 2 // stop after 2
	})
	if count != 2 {
		t.Errorf("ForEach visited %d traders, want 2", count)
	}
}

func TestTable_ExportAll(t *testing.T) {
	tbl := NewTable(0)
	for i := range 3 {
		trader := newTestTrader(uint32(1000+i), int64(i+1), model.StoreSell)
		if err := tbl.Add(trader); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	exported := tbl.ExportAll()
	if len(exported) != 3 {
		t.Errorf("ExportAll returned %d traders, want 3", len(exported))
	}
}

func TestTable_UpdateTraderItems(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 1, model.StoreSell)
	if err := tbl.Add(trader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	newItems := []TradeEntry{
		{ItemIdentifier: 100, Count: 50, Price: 500},
	}
	tbl.UpdateTraderItems(1001, newItems)

	got := tbl.Get(1001)
	if len(got.Items) != 1 {
		t.Fatalf("Items count = %d, want 1", len(got.Items))
	}
	if got.Items[0].ItemIdentifier != 100 {
		t.Errorf("Items[0].ItemIdentifier = %d, want 100", got.Items[0].ItemIdentifier)
	}
}

func TestTable_UpdateTraderItems_NotFound(t *testing.T) {
	tbl := NewTable(0)
	// Не должно паниковать
	tbl.UpdateTraderItems(9999, []TradeEntry{})
}

func TestTable_RemoveIfEmpty(t *testing.T) {
	tbl := NewTable(0)
	trader := newTestTrader(1001, 1, model.StoreSell)
	if err := tbl.Add(trader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Не пусто — не удаляем
	if tbl.RemoveIfEmpty(1001) {
		t.Error("RemoveIfEmpty should return false when items exist")
	}

	// Очищаем items
	tbl.UpdateTraderItems(1001, nil)

	// Теперь пусто — удаляем
	if !tbl.RemoveIfEmpty(1001) {
		t.Error("RemoveIfEmpty should return true when items empty")
	}
	if tbl.Count() != 0 {
		t.Errorf("Count() = %d, want 0", tbl.Count())
	}
}

func TestTable_RemoveIfEmpty_NotFound(t *testing.T) {
	tbl := NewTable(0)
	if tbl.RemoveIfEmpty(9999) {
		t.Error("RemoveIfEmpty non-existent should return false")
	}
}

func TestTable_Expiration(t *testing.T) {
	tbl := NewTable(100 * time.Millisecond)

	expired := make(chan uint32, 1)
	tbl.SetExpireCallback(func(objectID uint32) {
		expired <- objectID
	})

	trader := newTestTrader(1001, 1, model.StoreSell)
	trader.StartedAt = time.Now()
	if err := tbl.Add(trader); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	select {
	case id := <-expired:
		if id != 1001 {
			t.Errorf("expired objectID = %d, want 1001", id)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("trader did not expire within timeout")
	}

	if tbl.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after expiration", tbl.Count())
	}
}

func TestTable_Expiration_AlreadyExpired(t *testing.T) {
	tbl := NewTable(100 * time.Millisecond)

	trader := newTestTrader(1001, 1, model.StoreSell)
	trader.StartedAt = time.Now().Add(-200 * time.Millisecond) // уже истёк

	err := tbl.Add(trader)
	if err == nil {
		t.Error("Add already-expired trader should return error")
	}
	if tbl.Count() != 0 {
		t.Errorf("Count() = %d, want 0", tbl.Count())
	}
}

func TestTable_StopAll(t *testing.T) {
	tbl := NewTable(time.Hour) // долгий таймер
	for i := range 3 {
		trader := newTestTrader(uint32(1000+i), int64(i+1), model.StoreSell)
		if err := tbl.Add(trader); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	tbl.StopAll()

	if tbl.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after StopAll", tbl.Count())
	}
}

func TestTable_MultipleStoreTypes(t *testing.T) {
	tbl := NewTable(0)

	types := []model.PrivateStoreType{
		model.StoreSell,
		model.StoreBuy,
		model.StorePackageSell,
	}

	for i, st := range types {
		trader := newTestTrader(uint32(1000+i), int64(i+1), st)
		if err := tbl.Add(trader); err != nil {
			t.Fatalf("Add(%v) failed: %v", st, err)
		}
	}

	if tbl.Count() != 3 {
		t.Errorf("Count() = %d, want 3", tbl.Count())
	}

	// Проверяем store types
	for i, st := range types {
		got := tbl.Get(uint32(1000 + i))
		if got == nil {
			t.Fatalf("Get(%d) returned nil", 1000+i)
		}
		if got.StoreType != st {
			t.Errorf("trader %d StoreType = %v, want %v", 1000+i, got.StoreType, st)
		}
	}
}

func TestTable_ConcurrentAccess(t *testing.T) {
	tbl := NewTable(0)

	var wg sync.WaitGroup
	// Параллельное добавление
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			trader := newTestTrader(uint32(10000+i), int64(i+1), model.StoreSell)
			_ = tbl.Add(trader)
		}()
	}
	wg.Wait()

	if tbl.Count() != 100 {
		t.Errorf("Count() = %d, want 100", tbl.Count())
	}

	// Параллельное чтение + удаление
	for i := range 100 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			tbl.IsOfflineTrader(uint32(10000 + i))
		}()
		go func() {
			defer wg.Done()
			tbl.Remove(uint32(10000 + i))
		}()
	}
	wg.Wait()

	if tbl.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after removing all", tbl.Count())
	}
}

func TestTable_ManageStoreType_Rejected(t *testing.T) {
	tbl := NewTable(0)

	// Manage-типы не должны быть разрешены для offline trade
	manageTypes := []model.PrivateStoreType{
		model.StoreSellManage,
		model.StoreBuyManage,
	}

	for _, st := range manageTypes {
		trader := newTestTrader(1001, 1, st)
		if err := tbl.Add(trader); err == nil {
			t.Errorf("Add(%v) should return error for manage type", st)
		}
	}
}
