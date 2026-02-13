package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestJoinAlly is the C2S opcode 0x82.
// Client sends this when alliance leader invites another clan leader.
const OpcodeRequestJoinAlly byte = 0x82

// RequestJoinAlly represents a request to invite a clan to the alliance.
//
// Packet structure:
//   - ObjectID (int32): target player object ID (clan leader to invite)
type RequestJoinAlly struct {
	ObjectID int32
}

// ParseRequestJoinAlly parses the C2S RequestJoinAlly packet.
func ParseRequestJoinAlly(data []byte) (*RequestJoinAlly, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading objectID: %w", err)
	}

	return &RequestJoinAlly{ObjectID: objectID}, nil
}
