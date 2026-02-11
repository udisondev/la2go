package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeMagicSkillCanceled is the server packet opcode for cancelled cast.
// Sent when a skill cast is interrupted (stun, damage, etc.).
//
// Phase 5.9.4: Cast Flow & Packets.
// Java reference: MagicSkillCanceled.java (opcode 0x49)
const OpcodeMagicSkillCanceled = 0x49

// MagicSkillCanceled packet (S2C 0x49) — broadcast when skill cast is interrupted.
type MagicSkillCanceled struct {
	ObjectID int32
}

// NewMagicSkillCanceled creates a MagicSkillCanceled packet.
func NewMagicSkillCanceled(objectID int32) *MagicSkillCanceled {
	return &MagicSkillCanceled{ObjectID: objectID}
}

// Write serializes the MagicSkillCanceled packet.
func (p *MagicSkillCanceled) Write() ([]byte, error) {
	// opcode(1) + objectID(4)
	w := packet.NewWriter(5)

	w.WriteByte(OpcodeMagicSkillCanceled)
	w.WriteInt(p.ObjectID)

	return w.Bytes(), nil
}
