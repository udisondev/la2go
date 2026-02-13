package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestAllyCrest is the C2S opcode 0x88.
// Client requests alliance crest image by crestId.
const OpcodeRequestAllyCrest byte = 0x88

// RequestAllyCrest contains the alliance crest ID to retrieve.
type RequestAllyCrest struct {
	CrestID int32
}

// ParseRequestAllyCrest parses the C2S RequestAllyCrest packet.
func ParseRequestAllyCrest(data []byte) (*RequestAllyCrest, error) {
	r := packet.NewReader(data)
	crestID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading crestID: %w", err)
	}
	return &RequestAllyCrest{CrestID: crestID}, nil
}
