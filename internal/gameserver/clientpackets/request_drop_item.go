package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestDropItem is the opcode for RequestDropItem (C2S 0x12).
// Java reference: ClientPackets.REQUEST_DROP_ITEM(0x12).
const OpcodeRequestDropItem = 0x12

// RequestDropItem represents a client request to drop an item on the ground.
//
// Packet structure (body after opcode):
//   - objectID (int32) — item object ID
//   - count    (int64) — number of items to drop (for stackables)
//   - x        (int32) — drop target X coordinate
//   - y        (int32) — drop target Y coordinate
//   - z        (int32) — drop target Z coordinate
type RequestDropItem struct {
	ObjectID int32
	Count    int64
	X        int32
	Y        int32
	Z        int32
}

// ParseRequestDropItem parses RequestDropItem packet.
func ParseRequestDropItem(data []byte) (*RequestDropItem, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading objectID: %w", err)
	}

	count, err := r.ReadLong()
	if err != nil {
		return nil, fmt.Errorf("reading count: %w", err)
	}

	x, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading x: %w", err)
	}

	y, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading y: %w", err)
	}

	z, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading z: %w", err)
	}

	return &RequestDropItem{
		ObjectID: objectID,
		Count:    count,
		X:        x,
		Y:        y,
		Z:        z,
	}, nil
}
