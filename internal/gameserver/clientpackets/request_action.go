package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestAction is the client packet opcode for RequestAction.
// Client sends this when player clicks on an object (target selection or action).
const OpcodeRequestAction = 0x04

// Action types for RequestAction packet.
const (
	ActionSimpleClick = 0 // Simple click (select target)
	ActionShiftClick  = 1 // Shift+click (attack intent)
)

// RequestAction represents the client's action request packet.
// Sent when player clicks on an object in the game world.
//
// Packet structure:
//   - ObjectID (int32): Target object ID (player, NPC, or item)
//   - OriginX (int32): Player's current X coordinate
//   - OriginY (int32): Player's current Y coordinate
//   - OriginZ (int32): Player's current Z coordinate
//   - ActionType (byte): 0=simple click, 1=shift+click (attack)
//
// Reference: RequestActionUse.java (L2J Mobius)
type RequestAction struct {
	ObjectID   int32 // Target object ID
	OriginX    int32 // Player's X coordinate
	OriginY    int32 // Player's Y coordinate
	OriginZ    int32 // Player's Z coordinate
	ActionType byte  // 0=select, 1=attack
}

// ParseRequestAction parses a RequestAction packet from raw bytes.
//
// The packet format is:
//   - objectID (int32): target object ID
//   - originX (int32): player X
//   - originY (int32): player Y
//   - originZ (int32): player Z
//   - actionType (byte): action type
//
// Returns an error if parsing fails.
func ParseRequestAction(data []byte) (*RequestAction, error) {
	r := packet.NewReader(data)

	// Read target object ID
	objectID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading ObjectID: %w", err)
	}

	// Read player origin coordinates
	originX, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading OriginX: %w", err)
	}

	originY, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading OriginY: %w", err)
	}

	originZ, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading OriginZ: %w", err)
	}

	// Read action type
	actionType, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading ActionType: %w", err)
	}

	return &RequestAction{
		ObjectID:   objectID,
		OriginX:    originX,
		OriginY:    originY,
		OriginZ:    originZ,
		ActionType: actionType,
	}, nil
}

// IsAttackIntent returns true if this is a shift+click (attack intent).
func (r *RequestAction) IsAttackIntent() bool {
	return r.ActionType == ActionShiftClick
}
