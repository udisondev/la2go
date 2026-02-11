package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestPickup is the client packet opcode for RequestPickup (C2S 0x14).
// Client sends this when player tries to pick up item from ground.
//
// Phase 5.7: Loot System.
const OpcodeRequestPickup = 0x14

// RequestPickup represents the client's item pickup request packet.
// Sent when player clicks on a dropped item on the ground.
//
// Packet structure:
//   - ObjectID (int32): DroppedItem object ID
//
// Reference: RequestGetItem.java (L2J Mobius)
//
// Phase 5.7: Loot System.
type RequestPickup struct {
	ObjectID int32 // DroppedItem object ID
}

// ParseRequestPickup parses a RequestPickup packet from raw bytes.
//
// The packet format is:
//   - objectID (int32): dropped item object ID
//
// Returns an error if parsing fails.
//
// Phase 5.7: Loot System.
func ParseRequestPickup(data []byte) (*RequestPickup, error) {
	r := packet.NewReader(data)

	// Read dropped item object ID
	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading ObjectID: %w", err)
	}

	return &RequestPickup{
		ObjectID: objectID,
	}, nil
}
