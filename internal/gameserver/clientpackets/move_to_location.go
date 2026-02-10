package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeMoveToLocation is the opcode for MoveToLocation packet (C2S 0x01)
	OpcodeMoveToLocation = 0x01
)

// MoveToLocation represents the MoveToLocation packet sent by client.
// Client sends this when player clicks on ground to move (click-to-move).
type MoveToLocation struct {
	TargetX int32 // Target X coordinate
	TargetY int32 // Target Y coordinate
	TargetZ int32 // Target Z coordinate
	OriginX int32 // Origin X coordinate (current position)
	OriginY int32 // Origin Y coordinate (current position)
	OriginZ int32 // Origin Z coordinate (current position)
	MoveType int32 // Movement type (0=walk, 1=run)
}

// ParseMoveToLocation parses MoveToLocation packet from raw bytes.
// Packet structure: [opcode:1] [targetX:4] [targetY:4] [targetZ:4] [originX:4] [originY:4] [originZ:4] [moveType:4]
func ParseMoveToLocation(data []byte) (*MoveToLocation, error) {
	r := packet.NewReader(data)

	targetX, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading targetX: %w", err)
	}

	targetY, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading targetY: %w", err)
	}

	targetZ, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading targetZ: %w", err)
	}

	originX, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading originX: %w", err)
	}

	originY, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading originY: %w", err)
	}

	originZ, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading originZ: %w", err)
	}

	moveType, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading moveType: %w", err)
	}

	return &MoveToLocation{
		TargetX:  targetX,
		TargetY:  targetY,
		TargetZ:  targetZ,
		OriginX:  originX,
		OriginY:  originY,
		OriginZ:  originZ,
		MoveType: moveType,
	}, nil
}
