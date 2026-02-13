package model

import (
	"sync"
	"testing"
)

func TestPrivateStoreType_String(t *testing.T) {
	tests := []struct {
		storeType PrivateStoreType
		want      string
	}{
		{StoreNone, "None"},
		{StoreSell, "Sell"},
		{StoreSellManage, "SellManage"},
		{StoreBuy, "Buy"},
		{StoreBuyManage, "BuyManage"},
		{StoreManufacture, "Manufacture"},
		{StorePackageSell, "PackageSell"},
		{PrivateStoreType(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.storeType.String(); got != tt.want {
			t.Errorf("PrivateStoreType(%d).String() = %q, want %q", tt.storeType, got, tt.want)
		}
	}
}

func TestPrivateStoreType_IsInStoreMode(t *testing.T) {
	tests := []struct {
		storeType PrivateStoreType
		want      bool
	}{
		{StoreNone, false},
		{StoreSell, true},
		{StoreSellManage, false},
		{StoreBuy, true},
		{StoreBuyManage, false},
		{StoreManufacture, true},
		{StorePackageSell, true},
	}

	for _, tt := range tests {
		if got := tt.storeType.IsInStoreMode(); got != tt.want {
			t.Errorf("PrivateStoreType(%d).IsInStoreMode() = %v, want %v", tt.storeType, got, tt.want)
		}
	}
}

func TestTradeList_AddItem(t *testing.T) {
	tl := NewTradeList()

	// Add valid item
	ti := &TradeItem{
		ObjectID: 1001,
		ItemID:   57,
		Count:    10,
		Price:    100,
	}
	if err := tl.AddItem(ti); err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	if tl.ItemCount() != 1 {
		t.Errorf("ItemCount = %d, want 1", tl.ItemCount())
	}

	// Nil item
	if err := tl.AddItem(nil); err == nil {
		t.Error("expected error for nil item")
	}

	// Invalid count
	badCount := &TradeItem{ObjectID: 1002, ItemID: 58, Count: 0, Price: 100}
	if err := tl.AddItem(badCount); err == nil {
		t.Error("expected error for count=0")
	}

	// Negative price
	badPrice := &TradeItem{ObjectID: 1003, ItemID: 59, Count: 1, Price: -1}
	if err := tl.AddItem(badPrice); err == nil {
		t.Error("expected error for negative price")
	}

	// Overflow protection
	overflow := &TradeItem{ObjectID: 1004, ItemID: 60, Count: 2_000_000_000, Price: 2_000_000_000}
	if err := tl.AddItem(overflow); err == nil {
		t.Error("expected error for price overflow")
	}
}

func TestTradeList_AddItem_Locked(t *testing.T) {
	tl := NewTradeList()
	tl.Lock()

	ti := &TradeItem{ObjectID: 1001, ItemID: 57, Count: 1, Price: 100}
	if err := tl.AddItem(ti); err == nil {
		t.Error("expected error adding to locked list")
	}
}

func TestTradeList_FindItem(t *testing.T) {
	tl := NewTradeList()

	ti1 := &TradeItem{ObjectID: 1001, ItemID: 57, Count: 10, Price: 100}
	ti2 := &TradeItem{ObjectID: 1002, ItemID: 58, Count: 5, Price: 200}

	_ = tl.AddItem(ti1)
	_ = tl.AddItem(ti2)

	// Find by ObjectID
	found := tl.FindItem(1001)
	if found == nil || found.ObjectID != 1001 {
		t.Error("FindItem(1001) failed")
	}

	notFound := tl.FindItem(9999)
	if notFound != nil {
		t.Error("FindItem(9999) should return nil")
	}

	// Find by ItemID
	found2 := tl.FindItemByID(58)
	if found2 == nil || found2.ItemID != 58 {
		t.Error("FindItemByID(58) failed")
	}
}

func TestTradeList_RemoveItem(t *testing.T) {
	tl := NewTradeList()

	ti := &TradeItem{ObjectID: 1001, ItemID: 57, Count: 10, Price: 100}
	_ = tl.AddItem(ti)

	if !tl.RemoveItem(1001) {
		t.Error("RemoveItem(1001) should return true")
	}

	if tl.ItemCount() != 0 {
		t.Errorf("ItemCount = %d, want 0", tl.ItemCount())
	}

	if tl.RemoveItem(9999) {
		t.Error("RemoveItem(9999) should return false")
	}
}

func TestTradeList_UpdateItemCount(t *testing.T) {
	tl := NewTradeList()

	ti := &TradeItem{ObjectID: 1001, ItemID: 57, Count: 10, Price: 100}
	_ = tl.AddItem(ti)

	// Partial sell
	if !tl.UpdateItemCount(1001, 3) {
		t.Error("UpdateItemCount should return true")
	}

	found := tl.FindItem(1001)
	if found == nil || found.Count != 7 {
		t.Errorf("count after partial sell = %d, want 7", found.Count)
	}

	// Sell all remaining
	if !tl.UpdateItemCount(1001, 7) {
		t.Error("UpdateItemCount should return true")
	}

	if tl.ItemCount() != 0 {
		t.Error("item should be removed when count reaches 0")
	}
}

func TestTradeList_Clear(t *testing.T) {
	tl := NewTradeList()

	_ = tl.AddItem(&TradeItem{ObjectID: 1, ItemID: 1, Count: 1, Price: 1})
	_ = tl.AddItem(&TradeItem{ObjectID: 2, ItemID: 2, Count: 1, Price: 1})

	tl.Lock()
	tl.Clear()

	if tl.ItemCount() != 0 {
		t.Errorf("ItemCount after Clear = %d, want 0", tl.ItemCount())
	}

	if tl.IsLocked() {
		t.Error("should be unlocked after Clear")
	}
}

func TestTradeList_TitlePackaged(t *testing.T) {
	tl := NewTradeList()

	tl.SetTitle("My Store")
	if tl.Title() != "My Store" {
		t.Errorf("Title = %q, want %q", tl.Title(), "My Store")
	}

	// Max 29 characters
	longTitle := "This is a very long store title that exceeds limit"
	tl.SetTitle(longTitle)
	if len(tl.Title()) != 29 {
		t.Errorf("Title length = %d, want 29", len(tl.Title()))
	}

	tl.SetPackaged(true)
	if !tl.IsPackaged() {
		t.Error("should be packaged")
	}
}

func TestTradeList_Items_ReturnsCopy(t *testing.T) {
	tl := NewTradeList()
	_ = tl.AddItem(&TradeItem{ObjectID: 1, ItemID: 1, Count: 1, Price: 1})

	items := tl.Items()
	items[0] = nil // modify copy

	if tl.FindItem(1) == nil {
		t.Error("modifying Items() copy should not affect original")
	}
}

func TestTradeList_Concurrent(t *testing.T) {
	tl := NewTradeList()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := range 50 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = tl.AddItem(&TradeItem{
				ObjectID: uint32(idx),
				ItemID:   int32(idx),
				Count:    1,
				Price:    100,
			})
		}(i)
	}

	// Concurrent reads
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tl.Items()
			_ = tl.ItemCount()
			_ = tl.FindItem(25)
		}()
	}

	wg.Wait()

	if tl.ItemCount() != 50 {
		t.Errorf("ItemCount = %d, want 50", tl.ItemCount())
	}
}

func TestPlayer_PrivateStore(t *testing.T) {
	p, err := NewPlayer(1, 100, 1, "TestPlayer", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Default state
	if p.IsTrading() {
		t.Error("should not be trading by default")
	}
	if p.IsInStoreMode() {
		t.Error("should not be in store mode by default")
	}
	if p.PrivateStoreType() != StoreNone {
		t.Errorf("PrivateStoreType = %v, want StoreNone", p.PrivateStoreType())
	}

	// Open sell store
	p.SetPrivateStoreType(StoreSell)
	if !p.IsTrading() {
		t.Error("should be trading after SetPrivateStoreType(StoreSell)")
	}
	if !p.IsInStoreMode() {
		t.Error("should be in store mode")
	}

	// Set sell list
	sl := NewTradeList()
	_ = sl.AddItem(&TradeItem{ObjectID: 1, ItemID: 1, Count: 5, Price: 1000})
	p.SetSellList(sl)

	if p.SellList() == nil || p.SellList().ItemCount() != 1 {
		t.Error("SellList should have 1 item")
	}

	// Set store message
	p.SetStoreMessage("Cheap swords!")
	if p.StoreMessage() != "Cheap swords!" {
		t.Errorf("StoreMessage = %q, want %q", p.StoreMessage(), "Cheap swords!")
	}

	// Close store
	p.ClosePrivateStore()
	if p.IsTrading() {
		t.Error("should not be trading after ClosePrivateStore")
	}
	if p.SellList() != nil {
		t.Error("SellList should be nil after close")
	}
	if p.StoreMessage() != "" {
		t.Error("StoreMessage should be empty after close")
	}
}

func TestPlayer_ClosePrivateStore_BuyList(t *testing.T) {
	p, err := NewPlayer(2, 200, 1, "TestBuyer", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	// Open buy store
	p.SetPrivateStoreType(StoreBuy)
	bl := NewTradeList()
	_ = bl.AddItem(&TradeItem{ItemID: 57, Count: 100, StoreCount: 100, Price: 1})
	p.SetBuyList(bl)

	if p.BuyList() == nil {
		t.Error("BuyList should not be nil")
	}

	p.ClosePrivateStore()
	if p.BuyList() != nil {
		t.Error("BuyList should be nil after close")
	}
	if p.PrivateStoreType() != StoreNone {
		t.Errorf("PrivateStoreType = %v, want StoreNone", p.PrivateStoreType())
	}
}

func TestPlayer_SetStoreMessage_MaxLength(t *testing.T) {
	p, err := NewPlayer(3, 300, 1, "TestMsg", 1, 0, 0)
	if err != nil {
		t.Fatalf("NewPlayer failed: %v", err)
	}

	longMsg := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	p.SetStoreMessage(longMsg)
	if len(p.StoreMessage()) != 29 {
		t.Errorf("StoreMessage length = %d, want 29", len(p.StoreMessage()))
	}
}
