package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSendWareHouseWithDrawList is the client packet opcode for warehouse withdraw.
// Client sends this when player confirms withdrawal from warehouse.
//
// Phase 8: Trade System Foundation.
// Java reference: SendWareHouseWithDrawList.java
const OpcodeSendWareHouseWithDrawList = 0x32

// WarehouseWithdrawEntry represents a single item to withdraw.
type WarehouseWithdrawEntry struct {
	ObjectID int32
	Count    int32
}

// SendWareHouseWithDrawList represents the client's warehouse withdraw request.
//
// Packet structure:
//   - count (int32): number of items
//   - for each item:
//   - objectId (int32): item instance ID
//   - count (int32): quantity to withdraw
//
// Phase 8: Trade System Foundation.
type SendWareHouseWithDrawList struct {
	Items []WarehouseWithdrawEntry
}

// ParseSendWareHouseWithDrawList parses a SendWareHouseWithDrawList packet from raw bytes.
func ParseSendWareHouseWithDrawList(data []byte) (*SendWareHouseWithDrawList, error) {
	r := packet.NewReader(data)

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading item count: %w", err)
	}

	if count <= 0 || count > 100 {
		return nil, fmt.Errorf("invalid item count: %d", count)
	}

	// Validate batch: count * 8 bytes (objectId + count)
	expectedBytes := count * 8
	if r.Remaining() < int(expectedBytes) {
		return nil, fmt.Errorf("insufficient data: need %d bytes, have %d", expectedBytes, r.Remaining())
	}

	items := make([]WarehouseWithdrawEntry, count)
	for i := range count {
		objectID, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] objectID: %w", i, err)
		}

		qty, err := r.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("reading item[%d] count: %w", i, err)
		}

		if qty <= 0 {
			return nil, fmt.Errorf("invalid quantity for item[%d]: %d", i, qty)
		}

		items[i] = WarehouseWithdrawEntry{
			ObjectID: objectID,
			Count:    qty,
		}
	}

	return &SendWareHouseWithDrawList{Items: items}, nil
}
