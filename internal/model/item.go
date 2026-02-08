package model

import (
	"fmt"
	"sync"
	"time"
)

// ItemLocation представляет местоположение предмета.
type ItemLocation int32

const (
	ItemLocationVoid      ItemLocation = 0 // Удалённый/несуществующий
	ItemLocationInventory ItemLocation = 1 // В инвентаре
	ItemLocationPaperdoll ItemLocation = 2 // Экипировано
	ItemLocationWarehouse ItemLocation = 3 // В складе
)

// String возвращает строковое представление ItemLocation.
func (l ItemLocation) String() string {
	switch l {
	case ItemLocationVoid:
		return "VOID"
	case ItemLocationInventory:
		return "INVENTORY"
	case ItemLocationPaperdoll:
		return "PAPERDOLL"
	case ItemLocationWarehouse:
		return "WAREHOUSE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", l)
	}
}

// Item представляет предмет в игре.
type Item struct {
	itemID    int64
	ownerID   int64
	itemType  int32
	count     int32
	enchant   int32
	location  ItemLocation
	slotID    int32
	createdAt time.Time

	mu sync.RWMutex
}

// NewItem создаёт новый предмет с валидацией.
func NewItem(ownerID int64, itemType int32, count int32) (*Item, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive, got %d", count)
	}

	return &Item{
		ownerID:   ownerID,
		itemType:  itemType,
		count:     count,
		enchant:   0,
		location:  ItemLocationInventory,
		slotID:    -1,
		createdAt: time.Now(),
	}, nil
}

// ItemID возвращает DB ID предмета (immutable).
func (i *Item) ItemID() int64 {
	return i.itemID
}

// OwnerID возвращает ID владельца (immutable).
func (i *Item) OwnerID() int64 {
	return i.ownerID
}

// ItemType возвращает тип предмета.
func (i *Item) ItemType() int32 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.itemType
}

// Count возвращает количество предметов (для stackable).
func (i *Item) Count() int32 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.count
}

// SetCount устанавливает количество с валидацией.
func (i *Item) SetCount(count int32) error {
	if count <= 0 {
		return fmt.Errorf("count must be positive, got %d", count)
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	i.count = count
	return nil
}

// AddCount добавляет количество (для stackable items).
// Может быть отрицательным для уменьшения количества.
func (i *Item) AddCount(delta int32) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	newCount := i.count + delta
	if newCount <= 0 {
		return fmt.Errorf("count would become %d (non-positive)", newCount)
	}

	i.count = newCount
	return nil
}

// Enchant возвращает уровень заточки.
func (i *Item) Enchant() int32 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.enchant
}

// SetEnchant устанавливает уровень заточки с валидацией.
func (i *Item) SetEnchant(enchant int32) error {
	if enchant < 0 {
		return fmt.Errorf("enchant cannot be negative, got %d", enchant)
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	i.enchant = enchant
	return nil
}

// Location возвращает местоположение предмета и слот.
func (i *Item) Location() (ItemLocation, int32) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.location, i.slotID
}

// SetLocation устанавливает местоположение и слот предмета.
func (i *Item) SetLocation(loc ItemLocation, slotID int32) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.location = loc
	i.slotID = slotID
}

// CreatedAt возвращает время создания предмета.
func (i *Item) CreatedAt() time.Time {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.createdAt
}

// SetItemID устанавливает DB ID после создания в БД.
func (i *Item) SetItemID(id int64) {
	// itemID immutable, но setter нужен для repository.Create
	i.itemID = id
}

// SetCreatedAt устанавливает время создания (для загрузки из DB).
func (i *Item) SetCreatedAt(t time.Time) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.createdAt = t
}

// IsEquipped проверяет экипирован ли предмет.
func (i *Item) IsEquipped() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.location == ItemLocationPaperdoll
}

// IsInInventory проверяет находится ли предмет в инвентаре.
func (i *Item) IsInInventory() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.location == ItemLocationInventory
}
