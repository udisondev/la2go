package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// RequestExSetPledgeCrestLarge represents a request to set/clear the large clan crest.
// Extended packet 0xD0:0x11.
//
// Packet structure (after sub-opcode):
//   - Length (int32): byte length of the crest data (0 = remove crest)
//   - Data   ([]byte): raw crest image bytes (max 2176)
type RequestExSetPledgeCrestLarge struct {
	Length int32
	Data   []byte
}

// ParseRequestExSetPledgeCrestLarge parses the C2S extended packet 0xD0:0x11.
func ParseRequestExSetPledgeCrestLarge(data []byte) (*RequestExSetPledgeCrestLarge, error) {
	r := packet.NewReader(data)

	length, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading length: %w", err)
	}

	if length <= 0 {
		return &RequestExSetPledgeCrestLarge{Length: length}, nil
	}

	crestData, err := r.ReadBytes(int(length))
	if err != nil {
		return nil, fmt.Errorf("reading crest data: %w", err)
	}

	return &RequestExSetPledgeCrestLarge{
		Length: length,
		Data:   crestData,
	}, nil
}
