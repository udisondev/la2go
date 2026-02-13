package model

import (
	"fmt"
	"sync"
)

// Inventory — хранилище предметов персонажа (inventory + paperdoll + warehouse).
//
// Phase 5.5: Weapon & Equipment System.
// Phase 8: Added warehouse storage.
// Java reference: PcInventory.java, Inventory.java
type Inventory struct {
	ownerID int64 // Character ID владельца

	items           map[uint32]*Item           // objectID → Item (все items в инвентаре)
	paperdoll       [PaperdollTotalSlots]*Item  // Equipped items (17 slots)
	warehouse       map[uint32]*Item           // objectID → Item (warehouse items)
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
		ownerID:   ownerID,
		items:     make(map[uint32]*Item),
		warehouse: make(map[uint32]*Item),
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
// If the slot is occupied, auto-unequips the old item first.
//
// Parameters:
//   - item: предмет для надевания
//   - slot: paperdoll slot index (0..16)
//
// Returns:
//   - *Item: previously equipped item (nil if slot was empty)
//   - error: если валидация провалилась
//
// Phase 8: auto-unequip old item if slot occupied.
func (inv *Inventory) EquipItem(item *Item, slot int32) (*Item, error) {
	if item == nil {
		return nil, fmt.Errorf("item cannot be nil")
	}
	if slot < 0 || slot >= PaperdollTotalSlots {
		return nil, fmt.Errorf("invalid slot: %d (must be 0..%d)", slot, PaperdollTotalSlots-1)
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	// Check item exists in inventory
	if _, exists := inv.items[item.ObjectID()]; !exists {
		return nil, fmt.Errorf("item objectID=%d not found in inventory", item.ObjectID())
	}

	// Auto-unequip old item if slot occupied
	var oldItem *Item
	if inv.paperdoll[slot] != nil {
		oldItem = inv.paperdoll[slot]
		oldItem.SetSlot(-1)
		oldItem.SetLocation(ItemLocationInventory)
		inv.paperdoll[slot] = nil
		inv.unequippedCount++
	}

	// Equip new item
	wasEquipped := item.IsEquipped()
	inv.paperdoll[slot] = item
	item.SetSlot(slot)
	item.SetLocation(ItemLocationPaperdoll)
	if !wasEquipped {
		inv.unequippedCount--
	}

	return oldItem, nil
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

// GetSellableItems возвращает все неэкипированные, продаваемые предметы (кроме Adena и quest items).
//
// Phase 8.3: NPC Shops.
func (inv *Inventory) GetSellableItems() []*Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	var items []*Item
	for _, item := range inv.items {
		// Skip equipped items, Adena, quest items, and non-tradeable items
		if item.IsEquipped() {
			continue
		}
		if item.itemID == AdenaItemID {
			continue
		}
		if item.template != nil && item.template.Type == ItemTypeQuestItem {
			continue
		}
		if item.template != nil && !item.template.Tradeable {
			continue
		}
		items = append(items, item)
	}
	return items
}

// GetDepositableItems returns all unequipped items that can be deposited to warehouse.
// Excludes equipped items, Adena, quest items, and non-tradeable items.
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) GetDepositableItems() []*Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	var items []*Item
	for _, item := range inv.items {
		if item.IsEquipped() {
			continue
		}
		if item.itemID == AdenaItemID {
			continue
		}
		if item.template != nil && item.template.Type == ItemTypeQuestItem {
			continue
		}
		if item.template != nil && !item.template.Tradeable {
			continue
		}
		items = append(items, item)
	}
	return items
}

// GetWarehouseItems returns all items stored in warehouse.
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) GetWarehouseItems() []*Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	items := make([]*Item, 0, len(inv.warehouse))
	for _, item := range inv.warehouse {
		items = append(items, item)
	}
	return items
}

// GetWarehouseItem returns a warehouse item by objectID (nil if not found).
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) GetWarehouseItem(objectID uint32) *Item {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return inv.warehouse[objectID]
}

// WarehouseCount returns the number of items in warehouse.
func (inv *Inventory) WarehouseCount() int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return len(inv.warehouse)
}

// warehouseDepositFee is the fee per item slot deposited (30 Adena per Java reference).
const warehouseDepositFee int32 = 30

// DepositToWarehouse moves an item from inventory to warehouse.
// If count < item.Count() for stackable items, splits the stack.
// Returns error if item not found, equipped, or insufficient count.
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) DepositToWarehouse(objectID uint32, count int32) error {
	if count <= 0 {
		return fmt.Errorf("deposit count must be > 0, got %d", count)
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	item, exists := inv.items[objectID]
	if !exists {
		return fmt.Errorf("item objectID=%d not found in inventory", objectID)
	}

	if item.IsEquipped() {
		return fmt.Errorf("cannot deposit equipped item objectID=%d", objectID)
	}

	if count > item.Count() {
		return fmt.Errorf("not enough items: have %d, want %d", item.Count(), count)
	}

	if count == item.Count() {
		// Move entire item to warehouse
		delete(inv.items, objectID)
		inv.unequippedCount--
		item.SetLocation(ItemLocationWarehouse)
		item.SetSlot(-1)
		inv.warehouse[objectID] = item
	} else {
		// Split stack: decrease count in inventory, create new item in warehouse
		if err := item.SetCount(item.Count() - count); err != nil {
			return fmt.Errorf("decreasing item count: %w", err)
		}
		// For split, we reuse the same objectID scheme — caller must provide new objectID
		// In practice, we just move count difference; the warehouse "slot" stores the partial item
		// Simplified: we move the whole item and adjust counts
		// Actually, for stack split we need a new Item. The caller (handler) should handle this.
		// For MVP: only full stack moves are supported via this method.
		// Restore the count and return error for partial
		if err := item.SetCount(item.Count() + count); err != nil {
			return fmt.Errorf("restoring item count: %w", err)
		}
		return fmt.Errorf("partial deposit not supported for objectID=%d, use DepositToWarehouseSplit", objectID)
	}

	return nil
}

// DepositToWarehouseSplit moves a partial stack from inventory to warehouse,
// creating a new warehouse item with the given objectID for the split portion.
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) DepositToWarehouseSplit(objectID uint32, count int32, newObjectID uint32) error {
	if count <= 0 {
		return fmt.Errorf("deposit count must be > 0, got %d", count)
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	item, exists := inv.items[objectID]
	if !exists {
		return fmt.Errorf("item objectID=%d not found in inventory", objectID)
	}

	if item.IsEquipped() {
		return fmt.Errorf("cannot deposit equipped item objectID=%d", objectID)
	}

	if count > item.Count() {
		return fmt.Errorf("not enough items: have %d, want %d", item.Count(), count)
	}

	if count == item.Count() {
		// Move entire item
		delete(inv.items, objectID)
		inv.unequippedCount--
		item.SetLocation(ItemLocationWarehouse)
		item.SetSlot(-1)
		inv.warehouse[objectID] = item
		return nil
	}

	// Split: decrease inventory count, check if item already in warehouse (stack merge)
	if err := item.SetCount(item.Count() - count); err != nil {
		return fmt.Errorf("decreasing item count: %w", err)
	}

	// Check if same itemID already in warehouse (merge stacks)
	for _, whItem := range inv.warehouse {
		if whItem.ItemID() == item.ItemID() {
			if err := whItem.SetCount(whItem.Count() + count); err != nil {
				return fmt.Errorf("merging warehouse stack: %w", err)
			}
			return nil
		}
	}

	// Create new warehouse item
	whItem := &Item{
		objectID: newObjectID,
		itemID:   item.itemID,
		ownerID:  inv.ownerID,
		location: ItemLocationWarehouse,
		slot:     -1,
		count:    count,
		enchant:  item.enchant,
		template: item.template,
	}
	inv.warehouse[newObjectID] = whItem

	return nil
}

// WithdrawFromWarehouse moves an item from warehouse to inventory.
// If count < item.Count() for stackable items, splits the stack.
// Returns error if item not found or insufficient count.
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) WithdrawFromWarehouse(objectID uint32, count int32, newObjectID uint32) error {
	if count <= 0 {
		return fmt.Errorf("withdraw count must be > 0, got %d", count)
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	item, exists := inv.warehouse[objectID]
	if !exists {
		return fmt.Errorf("item objectID=%d not found in warehouse", objectID)
	}

	if count > item.Count() {
		return fmt.Errorf("not enough items in warehouse: have %d, want %d", item.Count(), count)
	}

	if count == item.Count() {
		// Move entire item to inventory
		delete(inv.warehouse, objectID)
		item.SetLocation(ItemLocationInventory)
		item.SetSlot(-1)
		inv.items[objectID] = item
		inv.unequippedCount++
		return nil
	}

	// Split: decrease warehouse count, check if same itemID exists in inventory (merge)
	if err := item.SetCount(item.Count() - count); err != nil {
		return fmt.Errorf("decreasing warehouse item count: %w", err)
	}

	// Check if same itemID already in inventory (merge stacks)
	for _, invItem := range inv.items {
		if invItem.ItemID() == item.ItemID() && !invItem.IsEquipped() {
			if err := invItem.SetCount(invItem.Count() + count); err != nil {
				return fmt.Errorf("merging inventory stack: %w", err)
			}
			return nil
		}
	}

	// Create new inventory item
	invItem := &Item{
		objectID: newObjectID,
		itemID:   item.itemID,
		ownerID:  inv.ownerID,
		location: ItemLocationInventory,
		slot:     -1,
		count:    count,
		enchant:  item.enchant,
		template: item.template,
	}
	inv.items[newObjectID] = invItem
	inv.unequippedCount++

	return nil
}

// AddWarehouseItem adds an item directly to warehouse (used for loading from DB).
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) AddWarehouseItem(item *Item) error {
	if item == nil {
		return fmt.Errorf("item cannot be nil")
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	objectID := item.ObjectID()
	if _, exists := inv.warehouse[objectID]; exists {
		return fmt.Errorf("item objectID=%d already exists in warehouse", objectID)
	}

	inv.warehouse[objectID] = item
	item.SetLocation(ItemLocationWarehouse)
	item.SetSlot(-1)

	return nil
}

// CountItemsByID counts total quantity of items with given template ID across inventory.
// Used for multisell ingredient checks.
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) CountItemsByID(itemID int32) int64 {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	var total int64
	for _, item := range inv.items {
		if item.itemID == itemID {
			total += int64(item.Count())
		}
	}
	return total
}

// RemoveItemsByID removes count items with given template ID from inventory.
// Removes from unequipped items first, consuming full stacks before partial.
// Returns the total count actually removed.
//
// Phase 8: Trade System Foundation.
func (inv *Inventory) RemoveItemsByID(itemID int32, count int64) int64 {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	var removed int64
	var toDelete []uint32

	for objID, item := range inv.items {
		if item.itemID != itemID || item.IsEquipped() {
			continue
		}
		if removed >= count {
			break
		}

		available := int64(item.Count())
		need := count - removed

		if available <= need {
			// Remove entire item
			toDelete = append(toDelete, objID)
			removed += available
		} else {
			// Partial removal
			if err := item.SetCount(int32(available - need)); err == nil {
				removed += need
			}
		}
	}

	for _, objID := range toDelete {
		item := inv.items[objID]
		delete(inv.items, objID)
		inv.unequippedCount--
		item.SetLocation(ItemLocationVoid)
	}

	return removed
}

// BodyPartToPaperdollSlot maps an item's bodyPart string to paperdoll slot index.
// Returns -1 if the body part doesn't correspond to a paperdoll slot.
//
// Java reference: Inventory.java getPaperdollIndex()
// Phase 19: UseItem handler equipment mapping.
func BodyPartToPaperdollSlot(bodyPart string) int32 {
	switch bodyPart {
	case "under":
		return PaperdollUnder
	case "rear":
		return PaperdollREar
	case "lear":
		return PaperdollLEar
	case "neck":
		return PaperdollNeck
	case "rfinger":
		return PaperdollRFinger
	case "lfinger":
		return PaperdollLFinger
	case "head":
		return PaperdollHead
	case "rhand":
		return PaperdollRHand
	case "lhand":
		return PaperdollLHand
	case "gloves":
		return PaperdollGloves
	case "chest":
		return PaperdollChest
	case "legs":
		return PaperdollLegs
	case "feet":
		return PaperdollFeet
	case "back":
		return PaperdollCloak
	case "face":
		return PaperdollFace
	case "hair":
		return PaperdollHair
	case "hairall":
		return PaperdollHair
	case "alldress", "onepiece":
		// Full body armor uses chest slot
		return PaperdollChest
	case "lrhand":
		// Two-handed weapons use right hand slot
		return PaperdollRHand
	default:
		return -1
	}
}

// BodyPartToAdditionalSlot returns a second slot that must be managed
// when equipping the given bodyPart, or -1 if none.
//
// Examples:
//   - "lrhand" (two-handed) → also unequip PaperdollLHand (shield)
//   - "onepiece"/"alldress" → also unequip PaperdollLegs
//
// Java reference: Inventory.java equipItem() two-hand/fullbody logic.
func BodyPartToAdditionalSlot(bodyPart string) int32 {
	switch bodyPart {
	case "lrhand":
		return PaperdollLHand // Two-handed → remove shield
	case "onepiece", "alldress":
		return PaperdollLegs // Full body → remove legs
	default:
		return -1
	}
}

// EarringSlots returns the two earring paperdoll slots.
// When equipping an earring, fill the first empty slot; if both occupied, replace right.
// Java reference: Inventory.java equipItem() earring logic.
func EarringSlots() (int32, int32) {
	return PaperdollREar, PaperdollLEar
}

// RingSlots returns the two ring paperdoll slots.
// When equipping a ring, fill the first empty slot; if both occupied, replace right.
func RingSlots() (int32, int32) {
	return PaperdollRFinger, PaperdollLFinger
}
