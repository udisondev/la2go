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
	enchant        int32 // Enchant level (0 для Phase 5.5, +15 max в будущем)
	augmentationID int32 // Phase 28: Augmentation ID (0 = none)

	// Shot charge state (Phase 52: Item Handlers)
	chargedSoulShot          bool
	chargedSpiritShot        bool
	chargedBlessedSpiritShot bool

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
// Max enchant is 65535 (int16 protocol limit).
func (i *Item) SetEnchant(enchant int32) error {
	if enchant < 0 {
		return fmt.Errorf("enchant cannot be negative, got %d", enchant)
	}
	if enchant > 65535 {
		return fmt.Errorf("enchant exceeds max (65535), got %d", enchant)
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	i.enchant = enchant
	return nil
}

// AugmentationID возвращает ID аугментации (0 = нет аугментации).
// Phase 28: Augmentation System.
func (i *Item) AugmentationID() int32 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.augmentationID
}

// SetAugmentationID устанавливает ID аугментации (0 = удалить).
// Phase 28: Augmentation System.
func (i *Item) SetAugmentationID(id int32) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.augmentationID = id
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

// IsQuestItem возвращает true если это квестовый предмет.
// Делегирует к ItemTemplate.Type == ItemTypeQuestItem.
// Java reference: Item.isQuestItem() → ItemTemplate.isQuestItem()
func (i *Item) IsQuestItem() bool {
	return i.template != nil && i.template.Type == ItemTypeQuestItem
}

// Name возвращает название предмета из template.
func (i *Item) Name() string {
	return i.template.Name
}

// --- Shot Charge Methods (Phase 51: Item Handler System) ---

// IsChargedSoulShot returns true if weapon is charged with Soul Shot.
func (i *Item) IsChargedSoulShot() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.chargedSoulShot
}

// SetChargedSoulShot sets or clears Soul Shot charge on weapon.
func (i *Item) SetChargedSoulShot(charged bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.chargedSoulShot = charged
}

// IsChargedSpiritShot returns true if weapon is charged with Spirit Shot.
func (i *Item) IsChargedSpiritShot() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.chargedSpiritShot
}

// SetChargedSpiritShot sets or clears Spirit Shot charge on weapon.
func (i *Item) SetChargedSpiritShot(charged bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.chargedSpiritShot = charged
}

// IsChargedBlessedSpiritShot returns true if weapon is charged with Blessed Spirit Shot.
func (i *Item) IsChargedBlessedSpiritShot() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.chargedBlessedSpiritShot
}

// SetChargedBlessedSpiritShot sets or clears Blessed Spirit Shot charge.
func (i *Item) SetChargedBlessedSpiritShot(charged bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.chargedBlessedSpiritShot = charged
}

// ClearAllShotCharges clears all shot charges (called after attack/cast).
func (i *Item) ClearAllShotCharges() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.chargedSoulShot = false
	i.chargedSpiritShot = false
	i.chargedBlessedSpiritShot = false
}
