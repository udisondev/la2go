package clientpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestShowBoard is the client packet opcode for RequestShowBoard (0x57).
// Client sends this when player presses ALT+B to open Community Board.
//
// Java reference: RequestShowBoard.java
const OpcodeRequestShowBoard = 0x57

// RequestShowBoard represents the client's request to show community board.
//
// Packet structure:
//   - unknown (int32) — unused parameter (always ignored by server)
//
// Phase 30: Community Board.
type RequestShowBoard struct {
	Unknown int32
}

// ParseRequestShowBoard parses a RequestShowBoard packet from raw bytes.
func ParseRequestShowBoard(data []byte) (*RequestShowBoard, error) {
	r := packet.NewReader(data)

	// Java: readInt() — unused, always ignored
	unknown, _ := r.ReadInt()

	return &RequestShowBoard{
		Unknown: unknown,
	}, nil
}
