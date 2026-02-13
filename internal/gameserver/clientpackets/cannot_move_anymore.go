package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCannotMoveAnymore is the opcode for CannotMoveAnymore packet (C2S 0x36).
// Client sends this when movement is blocked (reached wall or destination).
//
// Java reference: CannotMoveAnymore.java — reads x, y, z, heading.
const OpcodeCannotMoveAnymore = 0x36

// CannotMoveAnymore represents the parsed client packet.
type CannotMoveAnymore struct {
	X       int32
	Y       int32
	Z       int32
	Heading int32
}

// ParseCannotMoveAnymore parses the CannotMoveAnymore packet body (opcode already stripped).
func ParseCannotMoveAnymore(data []byte) (*CannotMoveAnymore, error) {
	r := packet.NewReader(data)

	x, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading x: %w", err)
	}

	y, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading y: %w", err)
	}

	z, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading z: %w", err)
	}

	heading, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading heading: %w", err)
	}

	return &CannotMoveAnymore{
		X:       x,
		Y:       y,
		Z:       z,
		Heading: heading,
	}, nil
}
