package model

import (
	"fmt"
	"sync"
)

// PrivateStoreType определяет тип приватного магазина игрока.
//
// Phase 8.1: Private Store System.
// Java reference: PrivateStoreType.java
type PrivateStoreType int8

const (
	StoreNone        PrivateStoreType = 0
	StoreSell        PrivateStoreType = 1
	StoreSellManage  PrivateStoreType = 2
	StoreBuy         PrivateStoreType = 3
	StoreBuyManage   PrivateStoreType = 4
	StoreManufacture       PrivateStoreType = 5
	StoreManufactureManage PrivateStoreType = 6
	StorePackageSell       PrivateStoreType = 8
)

// String returns human-readable store type name.
func (t PrivateStoreType) String() string {
	switch t {
	case StoreNone:
		return "None"
	case StoreSell:
		return "Sell"
	case StoreSellManage:
		return "SellManage"
	case StoreBuy:
		return "Buy"
	case StoreBuyManage:
		return "BuyManage"
	case StoreManufacture:
		return "Manufacture"
	case StoreManufactureManage:
		return "ManufactureManage"
	case StorePackageSell:
		return "PackageSell"
	default:
		return "Unknown"
	}
}

// IsInStoreMode returns true if player has an active store (not manage mode).
func (t PrivateStoreType) IsInStoreMode() bool {
	return t == StoreSell || t == StoreBuy || t == StoreManufacture || t == StorePackageSell
}

// MaxAdena — максимальное количество адены (L2 Interlude limit).
const MaxAdena int64 = 2_147_483_647

// TradeItem представляет предмет в торговом списке (private store sell/buy).
//
// Phase 8.1: Private Store System.
// Java reference: TradeItem.java
type TradeItem struct {
	ObjectID   uint32 // Unique world ID предмета (для sell) или 0 (для buy)
	ItemID     int32  // Template ID
	Count      int32  // Количество на продажу/покупку
	StoreCount int32  // Начальное количество (для buy — макс. запрос)
	Price      int64  // Цена за единицу в Adena
	Enchant    int32  // Enchant level

	// Item display properties (из ItemTemplate)
	Type1    int16
	Type2    int16
	BodyPart int32
	Weight   int32
}

// ItemRequest представляет запрос на покупку/продажу из private store.
//
// Phase 8.1: Private Store System.
// Java reference: ItemRequest.java
type ItemRequest struct {
	ObjectID int32
	ItemID   int32
	Count    int32
	Price    int64
}

// TradeList — список предметов для private store (sell или buy).
// Потокобезопасный: защищён RWMutex.
//
// Phase 8.1: Private Store System.
// Java reference: TradeList.java
type TradeList struct {
	mu       sync.RWMutex
	items    []*TradeItem
	title    string
	packaged bool // package sell — всё или ничего
	locked   bool
}

// NewTradeList создаёт новый пустой торговый список.
func NewTradeList() *TradeList {
	return &TradeList{
		items: make([]*TradeItem, 0, 8),
	}
}

// AddItem добавляет предмет в торговый список.
// Возвращает ошибку при невалидных параметрах.
func (tl *TradeList) AddItem(item *TradeItem) error {
	if item == nil {
		return fmt.Errorf("trade item cannot be nil")
	}
	if item.Count <= 0 {
		return fmt.Errorf("trade item count must be > 0, got %d", item.Count)
	}
	if item.Price < 0 {
		return fmt.Errorf("trade item price cannot be negative, got %d", item.Price)
	}
	// Overflow protection: count * price не должно превышать MaxAdena
	if item.Price > 0 && int64(item.Count) > MaxAdena/item.Price {
		return fmt.Errorf("price overflow: count=%d * price=%d exceeds MaxAdena", item.Count, item.Price)
	}

	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.locked {
		return fmt.Errorf("trade list is locked")
	}

	tl.items = append(tl.items, item)
	return nil
}

// Items возвращает копию списка предметов.
func (tl *TradeList) Items() []*TradeItem {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	result := make([]*TradeItem, len(tl.items))
	copy(result, tl.items)
	return result
}

// ItemCount возвращает количество предметов в списке.
func (tl *TradeList) ItemCount() int {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	return len(tl.items)
}

// FindItem ищет TradeItem по ObjectID (для sell-магазинов).
func (tl *TradeList) FindItem(objectID uint32) *TradeItem {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	for _, item := range tl.items {
		if item.ObjectID == objectID {
			return item
		}
	}
	return nil
}

// FindItemByID ищет TradeItem по ItemID (для buy-магазинов).
func (tl *TradeList) FindItemByID(itemID int32) *TradeItem {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	for _, item := range tl.items {
		if item.ItemID == itemID {
			return item
		}
	}
	return nil
}

// RemoveItem удаляет предмет из списка по ObjectID.
// Возвращает true если предмет был найден и удалён.
func (tl *TradeList) RemoveItem(objectID uint32) bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	for i, item := range tl.items {
		if item.ObjectID == objectID {
			tl.items = append(tl.items[:i], tl.items[i+1:]...)
			return true
		}
	}
	return false
}

// UpdateItemCount обновляет count предмета по ObjectID.
// Если count становится 0, предмет удаляется.
// Возвращает true если предмет был найден.
func (tl *TradeList) UpdateItemCount(objectID uint32, soldCount int32) bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	for i, item := range tl.items {
		if item.ObjectID == objectID {
			item.Count -= soldCount
			if item.Count <= 0 {
				tl.items = append(tl.items[:i], tl.items[i+1:]...)
			}
			return true
		}
	}
	return false
}

// Clear очищает торговый список.
func (tl *TradeList) Clear() {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	tl.items = tl.items[:0]
	tl.locked = false
}

// Title возвращает название магазина.
func (tl *TradeList) Title() string {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	return tl.title
}

// SetTitle устанавливает название магазина (макс 29 символов, как в L2).
func (tl *TradeList) SetTitle(title string) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if len(title) > 29 {
		title = title[:29]
	}
	tl.title = title
}

// IsPackaged возвращает true если это package sell (всё или ничего).
func (tl *TradeList) IsPackaged() bool {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	return tl.packaged
}

// SetPackaged устанавливает режим package sell.
func (tl *TradeList) SetPackaged(packaged bool) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.packaged = packaged
}

// Lock блокирует список от изменений (во время транзакции).
func (tl *TradeList) Lock() {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.locked = true
}

// IsLocked возвращает true если список заблокирован.
func (tl *TradeList) IsLocked() bool {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	return tl.locked
}
