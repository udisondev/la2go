package clientpackets

import (
	"fmt"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeRequestExMagicSkillUseGround is the 0xD0 sub-opcode 0x2F.
const SubOpcodeRequestExMagicSkillUseGround int16 = 0x2F

// RequestExMagicSkillUseGround is a ground-targeted skill cast (AoE).
type RequestExMagicSkillUseGround struct {
	X            int32
	Y            int32
	Z            int32
	SkillID      int32
	CtrlPressed  bool
	ShiftPressed bool
}

// ParseRequestExMagicSkillUseGround parses the packet from raw bytes (after sub-opcode).
func ParseRequestExMagicSkillUseGround(data []byte) (*RequestExMagicSkillUseGround, error) {
	r := packet.NewReader(data)

	x, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading X: %w", err)
	}

	y, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Y: %w", err)
	}

	z, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading Z: %w", err)
	}

	skillID, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading SkillID: %w", err)
	}

	ctrl, err := r.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("reading CtrlPressed: %w", err)
	}

	shift, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("reading ShiftPressed: %w", err)
	}

	return &RequestExMagicSkillUseGround{
		X:            x,
		Y:            y,
		Z:            z,
		SkillID:      skillID,
		CtrlPressed:  ctrl != 0,
		ShiftPressed: shift != 0,
	}, nil
}
