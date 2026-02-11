package clientpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeRequestMagicSkillUse is the client packet opcode for skill cast request.
// Client sends this when player uses a skill from the skill bar.
//
// Packet structure (C2S 0x2F):
//   - skillID (int32) — the skill to cast
//   - ctrlPressed (int32) — 1 if Ctrl key held (force attack/self-cast)
//   - shiftPressed (byte) — 1 if Shift key held
//
// Phase 5.9.4: Cast Flow & Packets.
// Java reference: RequestMagicSkillUse.java
const OpcodeRequestMagicSkillUse = 0x2F

// RequestMagicSkillUse represents client skill use request.
type RequestMagicSkillUse struct {
	SkillID      int32
	CtrlPressed  bool
	ShiftPressed bool
}

// ParseRequestMagicSkillUse parses RequestMagicSkillUse from raw bytes.
// Opcode already stripped by HandlePacket.
func ParseRequestMagicSkillUse(data []byte) (*RequestMagicSkillUse, error) {
	r := packet.NewReader(data)

	skillID, err := r.ReadInt()
	if err != nil {
		return nil, err
	}

	ctrl, err := r.ReadInt()
	if err != nil {
		return nil, err
	}

	shift, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	return &RequestMagicSkillUse{
		SkillID:      skillID,
		CtrlPressed:  ctrl != 0,
		ShiftPressed: shift != 0,
	}, nil
}
