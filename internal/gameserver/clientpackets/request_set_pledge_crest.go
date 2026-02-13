package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestSetPledgeCrest is the C2S opcode 0x53.
// Client uploads a clan crest image.
const OpcodeRequestSetPledgeCrest byte = 0x53

// RequestSetPledgeCrest represents a request to set/clear the clan crest.
//
// Packet structure:
//   - Length (int32): byte length of the crest data (0 = remove crest)
//   - Data   ([]byte): raw crest image bytes (max 256)
type RequestSetPledgeCrest struct {
	Length int32
	Data   []byte
}

// ParseRequestSetPledgeCrest parses the C2S RequestSetPledgeCrest packet.
func ParseRequestSetPledgeCrest(data []byte) (*RequestSetPledgeCrest, error) {
	r := packet.NewReader(data)

	length, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading length: %w", err)
	}

	if length <= 0 {
		return &RequestSetPledgeCrest{Length: length}, nil
	}

	crestData, err := r.ReadBytes(int(length))
	if err != nil {
		return nil, fmt.Errorf("reading crest data: %w", err)
	}

	return &RequestSetPledgeCrest{
		Length: length,
		Data:   crestData,
	}, nil
}
