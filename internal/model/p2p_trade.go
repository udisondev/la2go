package model

import (
	"fmt"
	"sync"
)

// P2PTradeItem represents an item in a player-to-player trade window.
type P2PTradeItem struct {
	ObjectID int32
	ItemID   int32
	Count    int32
	Enchant  int32
}

// P2PTradeList manages a player-to-player trade session.
// Each player in the trade has their own P2PTradeList instance
// linked to their partner's list via owner/partner references.
// Thread-safe: all mutable state protected by mu.
type P2PTradeList struct {
	mu        sync.Mutex
	owner     *Player
	partner   *Player
	items     []*P2PTradeItem
	confirmed bool
	locked    bool
}

// NewP2PTradeList creates a new P2P trade list for owner trading with partner.
func NewP2PTradeList(owner, partner *Player) *P2PTradeList {
	return &P2PTradeList{
		owner:   owner,
		partner: partner,
		items:   make([]*P2PTradeItem, 0, 8),
	}
}

// Owner returns the trade list owner.
func (tl *P2PTradeList) Owner() *Player {
	return tl.owner
}

// Partner returns the trade partner.
func (tl *P2PTradeList) Partner() *Player {
	return tl.partner
}

// AddItem adds an item to the trade list.
// Returns the created trade item, or error if validation fails.
func (tl *P2PTradeList) AddItem(objectID, count int32) (*P2PTradeItem, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be > 0, got %d", count)
	}

	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.locked {
		return nil, fmt.Errorf("trade list is locked")
	}
	if tl.confirmed {
		return nil, fmt.Errorf("trade list already confirmed")
	}

	// Проверим, не добавлен ли уже этот предмет
	for _, existing := range tl.items {
		if existing.ObjectID == objectID {
			return nil, fmt.Errorf("item objectID=%d already in trade", objectID)
		}
	}

	// Находим предмет в инвентаре владельца для получения метаданных
	inv := tl.owner.Inventory()
	item := inv.GetItem(uint32(objectID))
	if item == nil {
		return nil, fmt.Errorf("item objectID=%d not found in inventory", objectID)
	}

	if count > item.Count() {
		return nil, fmt.Errorf("not enough items: have %d, want %d", item.Count(), count)
	}

	tradeItem := &P2PTradeItem{
		ObjectID: objectID,
		ItemID:   item.ItemID(),
		Count:    count,
		Enchant:  item.Enchant(),
	}

	tl.items = append(tl.items, tradeItem)
	return tradeItem, nil
}

// Confirm marks the trade list as confirmed by the owner.
// Returns true if this is the first confirmation.
func (tl *P2PTradeList) Confirm() bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.confirmed {
		return false
	}
	tl.confirmed = true
	return true
}

// IsConfirmed returns true if the owner has confirmed the trade.
func (tl *P2PTradeList) IsConfirmed() bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	return tl.confirmed
}

// Lock prevents further modifications to the trade list.
func (tl *P2PTradeList) Lock() {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.locked = true
}

// Items returns a copy of all trade items.
func (tl *P2PTradeList) Items() []*P2PTradeItem {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	result := make([]*P2PTradeItem, len(tl.items))
	copy(result, tl.items)
	return result
}
