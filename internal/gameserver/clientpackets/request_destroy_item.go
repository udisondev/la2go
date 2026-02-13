package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestDestroyItem is the opcode for RequestDestroyItem (C2S 0x59).
// Java reference: ClientPackets.REQUEST_DESTROY_ITEM(0x59).
const OpcodeRequestDestroyItem = 0x59

// RequestDestroyItem represents a client request to destroy an item.
//
// Packet structure (body after opcode):
//   - objectID (int32) — item object ID
//   - count    (int32) — number of items to destroy
type RequestDestroyItem struct {
	ObjectID int32
	Count    int32
}

// ParseRequestDestroyItem parses RequestDestroyItem packet.
func ParseRequestDestroyItem(data []byte) (*RequestDestroyItem, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading objectID: %w", err)
	}

	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading count: %w", err)
	}

	return &RequestDestroyItem{ObjectID: objectID, Count: count}, nil
}
