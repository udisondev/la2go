package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeStartRotating is the opcode for StartRotating (C2S 0x4A).
// Java reference: ClientPackets.START_ROTATING(0x4A).
const OpcodeStartRotating = 0x4A

// OpcodeFinishRotating is the opcode for FinishRotating (C2S 0x4B).
// Java reference: ClientPackets.FINISH_ROTATING(0x4B).
const OpcodeFinishRotating = 0x4B

// StartRotating represents a client notification that character started rotating.
//
// Packet structure (body after opcode):
//   - degree (int32) — rotation angle
//   - side   (int32) — rotation direction (1=left, -1=right)
type StartRotating struct {
	Degree int32
	Side   int32
}

// FinishRotating represents a client notification that character finished rotating.
//
// Packet structure (body after opcode):
//   - degree  (int32) — final rotation angle
//   - unknown (int32) — reserved
type FinishRotating struct {
	Degree  int32
	Unknown int32
}

// ParseStartRotating parses StartRotating packet.
func ParseStartRotating(data []byte) (*StartRotating, error) {
	r := packet.NewReader(data)

	degree, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading degree: %w", err)
	}

	side, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading side: %w", err)
	}

	return &StartRotating{Degree: degree, Side: side}, nil
}

// ParseFinishRotating parses FinishRotating packet.
func ParseFinishRotating(data []byte) (*FinishRotating, error) {
	r := packet.NewReader(data)

	degree, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading degree: %w", err)
	}

	unknown, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading unknown: %w", err)
	}

	return &FinishRotating{Degree: degree, Unknown: unknown}, nil
}
