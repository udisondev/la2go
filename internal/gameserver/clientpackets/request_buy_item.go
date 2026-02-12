package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestBuyItem is the client packet opcode for RequestBuyItem.
// Client sends this when player confirms purchase from NPC shop.
//
// Phase 8.3: NPC Shops.
// Java reference: RequestBuyItem.java
const OpcodeRequestBuyItem = 0x1F

// BuyItemEntry represents a single item in a buy request.
type BuyItemEntry struct {
	ItemID int32
	Count  int32
}

// RequestBuyItem represents the client's buy request.
//
// Packet structure:
//   - listID (int32): BuyList ID
//   - count (int32): number of items
//   - for each item:
//   - itemID (int32): item template ID
//   - count (int32): quantity to buy
//
// Phase 8.3: NPC Shops.
type RequestBuyItem struct {
	ListID int32
	Items  []BuyItemEntry
}

// ParseRequestBuyItem parses a RequestBuyItem packet from raw bytes.
func ParseRequestBuyItem(data []byte) (*RequestBuyItem, error) {
	r := packet.NewReader(data)

	listID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading listID: %w", err)
	}

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading item count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid item count: %d", count)
	}

	items := make([]BuyItemEntry, count)
	for i := range count {
		itemID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] ID: %w", i, err)
		}

		qty, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] count: %w", i, err)
		}

		if qty <= 0 {
			return nil, fmt.Errorf("invalid quantity for item[%d]: %d", i, qty)
		}

		items[i] = BuyItemEntry{
			ItemID: itemID,
			Count:  qty,
		}
	}

	return &RequestBuyItem{
		ListID: listID,
		Items:  items,
	}, nil
}
