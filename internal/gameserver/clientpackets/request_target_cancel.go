package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestTargetCanceld is the opcode for RequestTargetCanceld (C2S 0x37).
// Java reference: ClientPackets.TARGET_CANCEL(0x37).
const OpcodeRequestTargetCanceld = 0x37

// RequestTargetCanceld represents a client request to cancel the current target.
//
// Packet structure (body after opcode):
//   - unselect (int16) — 0 to cancel cast as well, 1 to just clear target
type RequestTargetCanceld struct {
	Unselect int16
}

// ParseRequestTargetCanceld parses RequestTargetCanceld packet.
func ParseRequestTargetCanceld(data []byte) (*RequestTargetCanceld, error) {
	r := packet.NewReader(data)

	unselect, err := r.ReadShort()
	if err != nil {
		return nil, fmt.Errorf("reading unselect: %w", err)
	}

	return &RequestTargetCanceld{Unselect: unselect}, nil
}
