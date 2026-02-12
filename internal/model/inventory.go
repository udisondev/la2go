package model

import (
	"fmt"
	"sync"
)

// Inventory — хранилище предметов персонажа (inventory + paperdoll + warehouse).
//
// Phase 5.5: Weapon & Equipment System.
// Java reference: PcInventory.java, Inventory.java
type Inventory struct {
	ownerID int64 // Character ID владельца

	items           map[uint32]*Item           // objectID → Item (все items)
	paperdoll       [PaperdollTotalSlots]*Item  // Equipped items (17 slots)
	unequippedCount int                         // O(1) counter for Count(), maintained by Add/Remove/Equip/Unequip

	mu sync.RWMutex
}

// Paperdoll slots (Java Inventory.java:80-96).
// These constants match L2J Mobius Inventory.PAPERDOLL_* values.
const (
	PaperdollUnder      = 0  // Underwear
	PaperdollLEar       = 1  // Left Ear
	PaperdollREar       = 2  // Right Ear
	PaperdollNeck       = 3  // Necklace
	PaperdollLFinger    = 4  // Left Ring
	PaperdollRFinger    = 5  // Right Ring
	PaperdollHead       = 6  // Helmet
	PaperdollRHand      = 7  // Right Hand (weapon)
	PaperdollLHand      = 8  // Left Hand (shield/dual weapon)
	PaperdollGloves     = 9  // Gloves
	PaperdollChest      = 10 // Chest Armor
	PaperdollLegs       = 11 // Legs Armor
	PaperdollFeet       = 12 // Boots
	PaperdollCloak      = 13 // Cloak
	PaperdollFace       = 14 // Face accessory
	PaperdollHair       = 15 // Hair accessory
	PaperdollHair2      = 16 // Hair accessory 2
	PaperdollTotalSlots = 17
)

// NewInventory создаёт новый инвентарь для персонажа.
func NewInventory(ownerID int64) *Inventory {
	return &Inventory{
		ownerID: ownerID,
		items:   make(map[uint32]*Item),
	}
}

// OwnerID возвращает character ID владельца.
func (inv *Inventory) OwnerID() int64 {
	return inv.ownerID
}

// GetPaperdollItem возвращает equipped item для указанного slot (может быть nil).
//
// Parameters:
//   - slot: paperdoll slot index (0..16)
//
// Returns:
//   - *Item: equipped item или nil если slot пустой
func (inv *Inventory) GetPaperdollItem(slot int32) *Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	if slot < 0 || slot >= PaperdollTotalSlots {
		return nil
	}
	return inv.paperdoll[slot]
}

// EquipItem надевает item в указанный slot.
//
// Parameters:
//   - item: предмет для надевания
//   - slot: paperdoll slot index (0..16)
//
// Returns:
//   - error: если валидация провалилась
//
// Validation:
//   - slot bounds check
//   - item must exist in inventory
//   - TODO Phase 5.6: unequip old item if slot occupied
//
// Phase 5.5: MVP без auto-unequip, без validation compatibility armor/weapon slots.
func (inv *Inventory) EquipItem(item *Item, slot int32) error {
	if item == nil {
		return fmt.Errorf("item cannot be nil")
	}
	if slot < 0 || slot >= PaperdollTotalSlots {
		return fmt.Errorf("invalid slot: %d (must be 0..%d)", slot, PaperdollTotalSlots-1)
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	// Check item exists in inventory
	if _, exists := inv.items[item.ObjectID()]; !exists {
		return fmt.Errorf("item objectID=%d not found in inventory", item.ObjectID())
	}

	// TODO Phase 5.6: Auto-unequip old item if slot occupied
	// TODO Phase 5.6: Validate item type matches slot (weapon → RHAND, armor bodypart → correct slot)

	// Equip item
	wasEquipped := item.IsEquipped()
	inv.paperdoll[slot] = item
	item.SetSlot(slot)
	item.SetLocation(ItemLocationPaperdoll)
	if !wasEquipped {
		inv.unequippedCount--
	}

	return nil
}

// UnequipItem снимает item из указанного slot.
//
// Parameters:
//   - slot: paperdoll slot index (0..16)
//
// Returns:
//   - *Item: unequipped item или nil если slot был пустой
func (inv *Inventory) UnequipItem(slot int32) *Item {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	if slot < 0 || slot >= PaperdollTotalSlots {
		return nil
	}

	item := inv.paperdoll[slot]
	if item != nil {
		item.SetSlot(-1)
		item.SetLocation(ItemLocationInventory)
		inv.paperdoll[slot] = nil
		inv.unequippedCount++
	}

	return item
}

// AddItem добавляет item в inventory.
//
// Parameters:
//   - item: предмет для добавления
//
// Returns:
//   - error: если item уже существует
//
// Phase 5.5: MVP без stacking, без capacity check.
func (inv *Inventory) AddItem(item *Item) error {
	if item == nil {
		return fmt.Errorf("item cannot be nil")
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	objectID := item.ObjectID()
	if _, exists := inv.items[objectID]; exists {
		return fmt.Errorf("item objectID=%d already exists in inventory", objectID)
	}

	inv.items[objectID] = item
	item.SetLocation(ItemLocationInventory)
	item.SetSlot(-1)
	inv.unequippedCount++

	return nil
}

// RemoveItem удаляет item из inventory.
//
// Parameters:
//   - objectID: unique ID предмета
//
// Returns:
//   - *Item: удалённый item или nil если не найден
//
// Phase 5.5: MVP — item должен быть unequipped before removal.
func (inv *Inventory) RemoveItem(objectID uint32) *Item {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	item, exists := inv.items[objectID]
	if !exists {
		return nil
	}

	// Unequip if equipped
	if item.IsEquipped() {
		slot := item.Slot()
		if slot >= 0 && slot < PaperdollTotalSlots {
			inv.paperdoll[slot] = nil
		}
		item.SetSlot(-1)
	} else {
		inv.unequippedCount--
	}

	delete(inv.items, objectID)
	item.SetLocation(ItemLocationVoid)

	return item
}

// GetItem возвращает item по objectID (может быть nil).
func (inv *Inventory) GetItem(objectID uint32) *Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return inv.items[objectID]
}

// GetItems возвращает slice всех items в inventory (копия для безопасности).
//
// Returns:
//   - []*Item: копия slice с items (изменения не влияют на inventory)
func (inv *Inventory) GetItems() []*Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	items := make([]*Item, 0, len(inv.items))
	for _, item := range inv.items {
		items = append(items, item)
	}
	return items
}

// GetEquippedItems возвращает slice всех equipped items (копия для безопасности).
//
// Returns:
//   - []*Item: копия slice с equipped items (non-nil slots)
func (inv *Inventory) GetEquippedItems() []*Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	equipped := make([]*Item, 0, PaperdollTotalSlots)
	for _, item := range inv.paperdoll {
		if item != nil {
			equipped = append(equipped, item)
		}
	}
	return equipped
}

// Count возвращает количество items в inventory (excluding equipped).
// O(1) via pre-maintained counter (updated by AddItem/RemoveItem/EquipItem/UnequipItem).
func (inv *Inventory) Count() int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return inv.unequippedCount
}

// TotalCount возвращает total количество items (including equipped).
func (inv *Inventory) TotalCount() int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return len(inv.items)
}

// AdenaItemID — item template ID for Adena (основная валюта L2).
// Phase 8.3: NPC Shops.
const AdenaItemID int32 = 57

// FindItemByItemID находит первый item с указанным template ID.
// Returns nil если не найден.
//
// Phase 8.3: NPC Shops.
func (inv *Inventory) FindItemByItemID(itemID int32) *Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	for _, item := range inv.items {
		if item.itemID == itemID {
			return item
		}
	}
	return nil
}

// GetAdena возвращает текущее количество Adena у игрока.
// Returns 0 если Adena нет в инвентаре.
//
// Phase 8.3: NPC Shops.
func (inv *Inventory) GetAdena() int64 {
	adena := inv.FindItemByItemID(AdenaItemID)
	if adena == nil {
		return 0
	}
	return int64(adena.Count())
}

// AddAdena увеличивает количество Adena в инвентаре.
// Если Adena ещё нет, создаёт новый Item (требует objectID и template).
// Returns error если item не найден (нужен CreateAdena для первого раза).
//
// Phase 8.3: NPC Shops.
func (inv *Inventory) AddAdena(amount int32) error {
	if amount <= 0 {
		return fmt.Errorf("adena amount must be > 0, got %d", amount)
	}

	adena := inv.FindItemByItemID(AdenaItemID)
	if adena == nil {
		return fmt.Errorf("no adena item in inventory, use AddItem first")
	}

	newCount := adena.Count() + amount
	return adena.SetCount(newCount)
}

// RemoveAdena уменьшает количество Adena в инвентаре.
// Returns error если недостаточно Adena.
//
// Phase 8.3: NPC Shops.
func (inv *Inventory) RemoveAdena(amount int32) error {
	if amount <= 0 {
		return fmt.Errorf("adena amount must be > 0, got %d", amount)
	}

	adena := inv.FindItemByItemID(AdenaItemID)
	if adena == nil {
		return fmt.Errorf("no adena in inventory")
	}

	current := adena.Count()
	if current < amount {
		return fmt.Errorf("not enough adena: have %d, need %d", current, amount)
	}

	return adena.SetCount(current - amount)
}

// GetSellableItems возвращает все неэкипированные, продаваемые предметы (кроме Adena).
//
// Phase 8.3: NPC Shops.
func (inv *Inventory) GetSellableItems() []*Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	var items []*Item
	for _, item := range inv.items {
		// Skip equipped items, Adena, and non-tradeable items
		if item.IsEquipped() {
			continue
		}
		if item.itemID == AdenaItemID {
			continue
		}
		if item.template != nil && !item.template.Tradeable {
			continue
		}
		items = append(items, item)
	}
	return items
}
