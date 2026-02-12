package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestSellItem is the client packet opcode for RequestSellItem.
// Client sends this when player confirms selling items to NPC.
//
// Phase 8.3: NPC Shops.
// Java reference: RequestSellItem.java
const OpcodeRequestSellItem = 0x1E

// SellItemEntry represents a single item in a sell request.
type SellItemEntry struct {
	ObjectID int32 // Unique instance ID
	ItemID   int32 // Template ID
	Count    int32 // Quantity to sell
}

// RequestSellItem represents the client's sell request.
//
// Packet structure:
//   - count (int32): number of items
//   - for each item:
//   - objectID (int32): item instance ID
//   - itemID (int32): item template ID
//   - count (int32): quantity to sell
//
// Phase 8.3: NPC Shops.
type RequestSellItem struct {
	Items []SellItemEntry
}

// ParseRequestSellItem parses a RequestSellItem packet from raw bytes.
func ParseRequestSellItem(data []byte) (*RequestSellItem, error) {
	r := packet.NewReader(data)

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading item count: %w", err)
	}

	if count < 0 || count > 100 {
		return nil, fmt.Errorf("invalid item count: %d", count)
	}

	items := make([]SellItemEntry, count)
	for i := range count {
		objectID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] objectID: %w", i, err)
		}

		itemID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] itemID: %w", i, err)
		}

		qty, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] count: %w", i, err)
		}

		if qty <= 0 {
			return nil, fmt.Errorf("invalid quantity for item[%d]: %d", i, qty)
		}

		items[i] = SellItemEntry{
			ObjectID: objectID,
			ItemID:   itemID,
			Count:    qty,
		}
	}

	return &RequestSellItem{
		Items: items,
	}, nil
}
