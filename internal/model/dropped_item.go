package model

import (
	"sync"
	"time"
)

// DroppedItem represents an item lying on the ground in the world.
// Phase 4.10 Part 3: Dropped items are visible to all players in range.
type DroppedItem struct {
	*WorldObject // embedded for position and ObjectID

	item      *Item     // Item data (itemType, count, enchant)
	dropTime  time.Time // When item was dropped
	dropperID uint32    // ObjectID of character who dropped (0=monster drop)

	mu sync.RWMutex
}

// NewDroppedItem creates a new dropped item at the given location.
// objectID should be in Item range (0x00000001-0x0FFFFFFF).
// dropper is the ObjectID of character who dropped (0 for monster drops).
func NewDroppedItem(objectID uint32, item *Item, location Location, dropperID uint32) *DroppedItem {
	if item == nil {
		panic("NewDroppedItem: item cannot be nil")
	}

	worldObj := NewWorldObject(objectID, "", location) // Items don't have names

	droppedItem := &DroppedItem{
		WorldObject: worldObj,
		item:        item,
		dropTime:    time.Now(),
		dropperID:   dropperID,
	}

	// Phase 5.7: Set WorldObject.Data reference for type assertion in pickup
	worldObj.Data = droppedItem

	return droppedItem
}

// Item returns the item data (read-only).
func (d *DroppedItem) Item() *Item {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.item
}

// DropTime returns when the item was dropped.
func (d *DroppedItem) DropTime() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.dropTime
}

// DropperID returns ObjectID of character who dropped the item.
// Returns 0 for monster drops.
func (d *DroppedItem) DropperID() uint32 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.dropperID
}

// IsProtected checks if item is protected for specific player.
// Protected items can only be picked up by dropper for first N seconds.
// Phase 4.10 Part 3: Basic implementation, always returns false.
// Protection time for PvP drops requires dropTime tracking and combat kill context;
// will be added when PvP drop system is implemented.
func (d *DroppedItem) IsProtected(playerObjectID uint32) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Monster drops (dropperID=0) are never protected
	if d.dropperID == 0 {
		return false
	}

	// No protection implemented yet — requires dropTime field + PvP kill context.
	// When implemented: check time.Since(d.dropTime) > 30s, then unprotect.
	return false
}
