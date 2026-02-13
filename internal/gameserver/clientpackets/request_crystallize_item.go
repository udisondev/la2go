package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestCrystallizeItem is the C2S opcode 0x72.
// Client sends this to break an item into crystals.
const OpcodeRequestCrystallizeItem byte = 0x72

// RequestCrystallizeItem represents a crystallize request.
type RequestCrystallizeItem struct {
	ObjectID int32 // item object ID
	Count    int64 // how many to crystallize (for stackable, usually 1)
}

// ParseRequestCrystallizeItem parses the packet from raw bytes.
func ParseRequestCrystallizeItem(data []byte) (*RequestCrystallizeItem, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading ObjectID: %w", err)
	}

	// Count is int32 in Interlude (not int64)
	count, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Count: %w", err)
	}

	return &RequestCrystallizeItem{
		ObjectID: objectID,
		Count:    int64(count),
	}, nil
}
