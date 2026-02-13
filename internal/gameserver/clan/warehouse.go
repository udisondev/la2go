package clan

import (
	"errors"
	"sync"
)

// Warehouse capacity.
const MaxWarehouseSlots = 150

// Warehouse errors.
var (
	ErrWarehouseFull    = errors.New("clan warehouse is full")
	ErrItemNotFound     = errors.New("item not found in warehouse")
	ErrInsufficientItem = errors.New("insufficient item count")
)

// WarehouseItem represents an item stored in the clan warehouse.
type WarehouseItem struct {
	ObjectID int64
	ItemID   int32
	Count    int64
	EnchantLevel int32
}

// Warehouse is a thread-safe clan warehouse (shared storage).
type Warehouse struct {
	mu    sync.RWMutex
	items map[int64]*WarehouseItem // objectID -> item
}

// NewWarehouse creates a new empty warehouse.
func NewWarehouse() *Warehouse {
	return &Warehouse{
		items: make(map[int64]*WarehouseItem, 32),
	}
}

// AddItem adds or stacks an item in the warehouse.
// For stackable items, increases count if item already exists.
// Returns ErrWarehouseFull if the warehouse has reached max slots.
func (w *Warehouse) AddItem(item *WarehouseItem) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Попытка стакнуть с существующим предметом
	for _, existing := range w.items {
		if existing.ItemID == item.ItemID && existing.EnchantLevel == item.EnchantLevel {
			existing.Count += item.Count
			return nil
		}
	}

	if int32(len(w.items)) >= MaxWarehouseSlots {
		return ErrWarehouseFull
	}

	w.items[item.ObjectID] = item
	return nil
}

// RemoveItem removes count units of an item.
// Returns ErrItemNotFound if item doesn't exist.
// Returns ErrInsufficientItem if count exceeds available.
func (w *Warehouse) RemoveItem(objectID int64, count int64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	item, ok := w.items[objectID]
	if !ok {
		return ErrItemNotFound
	}

	if item.Count < count {
		return ErrInsufficientItem
	}

	item.Count -= count
	if item.Count <= 0 {
		delete(w.items, objectID)
	}
	return nil
}

// Item returns a warehouse item by object ID, or nil if not found.
func (w *Warehouse) Item(objectID int64) *WarehouseItem {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.items[objectID]
}

// Items returns a snapshot of all warehouse items.
func (w *Warehouse) Items() []*WarehouseItem {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]*WarehouseItem, 0, len(w.items))
	for _, item := range w.items {
		result = append(result, item)
	}
	return result
}

// Count returns the number of items in the warehouse.
func (w *Warehouse) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.items)
}

// Clear removes all items from the warehouse.
func (w *Warehouse) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.items = make(map[int64]*WarehouseItem, 32)
}
