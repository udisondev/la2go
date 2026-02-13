package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestSSQStatus is the opcode for Seven Signs status request (C2S 0xC7).
const OpcodeRequestSSQStatus byte = 0xC7

// RequestSSQStatus packet (C2S 0xC7) sent when player opens Seven Signs UI.
//
// Packet structure:
//   - page (byte) — 1-4, which status page to display
//
// Phase 25: Seven Signs.
type RequestSSQStatus struct {
	Page byte
}

// ParseRequestSSQStatus parses the Seven Signs status request.
func ParseRequestSSQStatus(data []byte) (*RequestSSQStatus, error) {
	r := packet.NewReader(data)

	page, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading page: %w", err)
	}

	return &RequestSSQStatus{
		Page: page,
	}, nil
}
