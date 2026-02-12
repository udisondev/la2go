package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeMagicSkillUse is the server packet opcode for skill cast animation start.
// Sent when a character begins casting a skill.
//
// Phase 5.9.4: Cast Flow & Packets.
// Java reference: MagicSkillUse.java (opcode 0x48)
const OpcodeMagicSkillUse = 0x48

// MagicSkillUse packet (S2C 0x48) — broadcast when a skill cast begins.
// Triggers cast animation on all nearby clients.
type MagicSkillUse struct {
	CasterObjectID int32
	TargetObjectID int32
	SkillID        int32
	SkillLevel     int32
	HitTime        int32 // ms — cast time
	ReuseDelay     int32 // ms — cooldown
	CasterX        int32
	CasterY        int32
	CasterZ        int32
}

// NewMagicSkillUse creates a MagicSkillUse packet.
func NewMagicSkillUse(casterID, targetID, skillID, skillLevel, hitTime, reuseDelay, x, y, z int32) MagicSkillUse {
	return MagicSkillUse{
		CasterObjectID: casterID,
		TargetObjectID: targetID,
		SkillID:        skillID,
		SkillLevel:     skillLevel,
		HitTime:        hitTime,
		ReuseDelay:     reuseDelay,
		CasterX:        x,
		CasterY:        y,
		CasterZ:        z,
	}
}

// Write serializes the MagicSkillUse packet.
func (p *MagicSkillUse) Write() ([]byte, error) {
	// opcode(1) + 9*int32(36)
	w := packet.NewWriter(37)

	w.WriteByte(OpcodeMagicSkillUse)
	w.WriteInt(p.CasterObjectID)
	w.WriteInt(p.TargetObjectID)
	w.WriteInt(p.SkillID)
	w.WriteInt(p.SkillLevel)
	w.WriteInt(p.HitTime)
	w.WriteInt(p.ReuseDelay)
	w.WriteInt(p.CasterX)
	w.WriteInt(p.CasterY)
	w.WriteInt(p.CasterZ)

	return w.Bytes(), nil
}
