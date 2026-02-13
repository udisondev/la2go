package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestSetAllyCrest is the C2S opcode 0x87.
// Client uploads an alliance crest image.
const OpcodeRequestSetAllyCrest byte = 0x87

// RequestSetAllyCrest represents a request to set the alliance crest.
//
// Packet structure:
//   - Length (int32): byte length of the crest data
//   - Data ([]byte): raw crest image bytes
type RequestSetAllyCrest struct {
	Length int32
	Data   []byte
}

// ParseRequestSetAllyCrest parses the C2S RequestSetAllyCrest packet.
func ParseRequestSetAllyCrest(data []byte) (*RequestSetAllyCrest, error) {
	r := packet.NewReader(data)

	length, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading length: %w", err)
	}

	crestData, err := r.ReadBytes(int(length))
	if err != nil {
		return nil, fmt.Errorf("reading crest data: %w", err)
	}

	return &RequestSetAllyCrest{
		Length: length,
		Data:   crestData,
	}, nil
}
