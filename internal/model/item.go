package model

import (
	"fmt"
	"sync"
)

// Item — конкретный экземпляр предмета (weapon, armor, consumable, etc.).
// Может быть в инвентаре, на персонаже (equipped), в warehouse, и т.д.
//
// Phase 5.5: Weapon & Equipment System (refactored from Phase 4.6).
// Java reference: Item.java
type Item struct {
	objectID uint32 // Unique ID в world (Phase 4.15)
	itemID   int32  // Template ID (ссылка на ItemTemplate)
	ownerID  int64  // Character ID владельца
	location ItemLocation
	slot     int32 // Paperdoll slot (-1 если не equipped)
	count    int32 // Stack count (1 для weapons/armor)
	enchant  int32 // Enchant level (0 для Phase 5.5, +15 max в будущем)

	template *ItemTemplate // Stats template (pAtk, pDef, etc.)

	mu sync.RWMutex
}

// ItemLocation определяет где хранится предмет.
type ItemLocation int32

const (
	ItemLocationInventory ItemLocation = iota
	ItemLocationPaperdoll  // Equipped
	ItemLocationWarehouse
	ItemLocationVoid // Deleted/dropped
)

// String returns human-readable item location name.
func (il ItemLocation) String() string {
	switch il {
	case ItemLocationInventory:
		return "Inventory"
	case ItemLocationPaperdoll:
		return "Paperdoll"
	case ItemLocationWarehouse:
		return "Warehouse"
	case ItemLocationVoid:
		return "Void"
	default:
		return "Unknown"
	}
}

// NewItem создаёт новый предмет с валидацией.
//
// Parameters:
//   - objectID: unique ID в world (from world.IDGenerator().NextItemID())
//   - itemID: template ID (ссылка на ItemTemplate)
//   - ownerID: character ID владельца
//   - count: stack count (должен быть > 0)
//   - template: ItemTemplate со stats
//
// Returns:
//   - *Item: новый предмет
//   - error: если валидация провалилась
func NewItem(objectID uint32, itemID int32, ownerID int64, count int32, template *ItemTemplate) (*Item, error) {
	if template == nil {
		return nil, fmt.Errorf("template cannot be nil")
	}
	if count <= 0 {
		return nil, fmt.Errorf("count must be > 0, got %d", count)
	}

	return &Item{
		objectID: objectID,
		itemID:   itemID,
		ownerID:  ownerID,
		location: ItemLocationInventory,
		slot:     -1, // Not equipped
		count:    count,
		enchant:  0,
		template: template,
	}, nil
}

// ObjectID возвращает unique ID в world.
func (i *Item) ObjectID() uint32 {
	return i.objectID
}

// ItemID возвращает template ID.
func (i *Item) ItemID() int32 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.itemID
}

// OwnerID возвращает character ID владельца.
func (i *Item) OwnerID() int64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.ownerID
}

// Location возвращает текущее местоположение предмета.
func (i *Item) Location() ItemLocation {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.location
}

// SetLocation устанавливает местоположение предмета.
func (i *Item) SetLocation(location ItemLocation) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.location = location
}

// Slot возвращает paperdoll slot (-1 если не equipped).
func (i *Item) Slot() int32 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.slot
}

// SetSlot устанавливает paperdoll slot.
func (i *Item) SetSlot(slot int32) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.slot = slot
}

// Count возвращает stack count (количество предметов в стаке).
func (i *Item) Count() int32 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.count
}

// SetCount устанавливает stack count с валидацией.
func (i *Item) SetCount(count int32) error {
	if count < 0 {
		return fmt.Errorf("count cannot be negative, got %d", count)
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	i.count = count
	return nil
}

// Enchant возвращает enchant level (0 для non-enchanted).
func (i *Item) Enchant() int32 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.enchant
}

// SetEnchant устанавливает enchant level с валидацией.
func (i *Item) SetEnchant(enchant int32) error {
	if enchant < 0 {
		return fmt.Errorf("enchant cannot be negative, got %d", enchant)
	}
	// TODO Phase 5.6+: validate max enchant level (+15 for normal items)

	i.mu.Lock()
	defer i.mu.Unlock()
	i.enchant = enchant
	return nil
}

// Template возвращает ItemTemplate со stats (immutable).
func (i *Item) Template() *ItemTemplate {
	return i.template
}

// IsEquipped возвращает true если предмет надет (paperdoll).
func (i *Item) IsEquipped() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.slot >= 0 && i.location == ItemLocationPaperdoll
}

// IsWeapon возвращает true если это оружие.
func (i *Item) IsWeapon() bool {
	return i.template.IsWeapon()
}

// IsArmor возвращает true если это броня.
func (i *Item) IsArmor() bool {
	return i.template.IsArmor()
}

// IsConsumable возвращает true если это consumable (potion, scroll).
func (i *Item) IsConsumable() bool {
	return i.template.IsConsumable()
}

// Name возвращает название предмета из template.
func (i *Item) Name() string {
	return i.template.Name
}
