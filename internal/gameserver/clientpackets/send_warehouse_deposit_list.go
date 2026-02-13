package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSendWareHouseDepositList is the client packet opcode for warehouse deposit.
// Client sends this when player confirms deposit into warehouse.
//
// Phase 8: Trade System Foundation.
// Java reference: SendWareHouseDepositList.java
const OpcodeSendWareHouseDepositList = 0x31

// WarehouseDepositEntry represents a single item to deposit.
type WarehouseDepositEntry struct {
	ObjectID int32
	Count    int32
}

// SendWareHouseDepositList represents the client's warehouse deposit request.
//
// Packet structure:
//   - count (int32): number of items
//   - for each item:
//   - objectId (int32): item instance ID
//   - count (int32): quantity to deposit
//
// Phase 8: Trade System Foundation.
type SendWareHouseDepositList struct {
	Items []WarehouseDepositEntry
}

// ParseSendWareHouseDepositList parses a SendWareHouseDepositList packet from raw bytes.
func ParseSendWareHouseDepositList(data []byte) (*SendWareHouseDepositList, error) {
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

	items := make([]WarehouseDepositEntry, count)
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

		items[i] = WarehouseDepositEntry{
			ObjectID: objectID,
			Count:    qty,
		}
	}

	return &SendWareHouseDepositList{Items: items}, nil
}
