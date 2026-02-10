package clientpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAttackRequest is the client packet opcode for attack request.
// Client sends this when player clicks on enemy (shift+click or simple click if aggressive).
//
// Packet structure (C2S 0x0A):
//   - objectID (int32) — target objectID
//   - originX (int32) — player X coordinate (unused in MVP)
//   - originY (int32) — player Y coordinate (unused)
//   - originZ (int32) — player Z coordinate (unused)
//   - attackID (byte) — 0 = simple click, 1 = shift-click
//
// Phase 5.3: Basic Combat System.
// Java reference: AttackRequest.java (opcode 0x0A, line 44-50).
const OpcodeAttackRequest = 0x0A

// AttackRequest represents client attack request packet.
// Sent when player clicks on enemy to initiate auto-attack.
//
// Phase 5.3: Basic Combat System.
type AttackRequest struct {
	ObjectID uint32 // Target objectID
	OriginX  int32  // Player X coordinate (unused in MVP)
	OriginY  int32  // Player Y coordinate (unused)
	OriginZ  int32  // Player Z coordinate (unused)
	AttackID byte   // 0 = simple click, 1 = shift-click (attack intent)
}

// ParseAttackRequest parses AttackRequest packet from raw bytes.
// Opcode already stripped by HandlePacket.
//
// Returns error if packet parsing fails.
//
// Phase 5.3: Basic Combat System.
func ParseAttackRequest(data []byte) (*AttackRequest, error) {
	r := packet.NewReader(data)

	objectID, err := r.ReadInt()
	if err != nil {
		return nil, err
	}

	originX, err := r.ReadInt()
	if err != nil {
		return nil, err
	}

	originY, err := r.ReadInt()
	if err != nil {
		return nil, err
	}

	originZ, err := r.ReadInt()
	if err != nil {
		return nil, err
	}

	attackID, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	return &AttackRequest{
		ObjectID: uint32(objectID),
		OriginX:  originX,
		OriginY:  originY,
		OriginZ:  originZ,
		AttackID: attackID,
	}, nil
}

// IsShiftClick returns true if attack was initiated with shift+click.
// Shift+click explicitly indicates attack intent (vs simple click on aggressive NPC).
//
// Phase 5.3: Basic Combat System.
func (p *AttackRequest) IsShiftClick() bool {
	return p.AttackID == 1
}
