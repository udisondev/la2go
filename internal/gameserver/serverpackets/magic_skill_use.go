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
//
// Java field order: casterObjID, targetObjID, skillID, skillLevel, hitTime, reuseDelay,
// casterX, casterY, casterZ, critical(int32), [if critical: short 0], targetX, targetY, targetZ.
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
	Critical       bool  // true = critical cast (skill crit)
	TargetX        int32
	TargetY        int32
	TargetZ        int32
}

// NewMagicSkillUse creates a MagicSkillUse packet.
func NewMagicSkillUse(casterID, targetID, skillID, skillLevel, hitTime, reuseDelay, cx, cy, cz, tx, ty, tz int32, critical bool) MagicSkillUse {
	return MagicSkillUse{
		CasterObjectID: casterID,
		TargetObjectID: targetID,
		SkillID:        skillID,
		SkillLevel:     skillLevel,
		HitTime:        hitTime,
		ReuseDelay:     reuseDelay,
		CasterX:        cx,
		CasterY:        cy,
		CasterZ:        cz,
		Critical:       critical,
		TargetX:        tx,
		TargetY:        ty,
		TargetZ:        tz,
	}
}

// Write serializes the MagicSkillUse packet.
// Java: casterID, targetID, skillID, level, hitTime, reuseDelay,
//
//	casterX, casterY, casterZ, critical(int32), [short 0 if critical],
//	targetX, targetY, targetZ.
func (p *MagicSkillUse) Write() ([]byte, error) {
	// opcode(1) + 9*int32(36) + critical(4) + optional short(2) + 3*int32(12) = max 55
	w := packet.NewWriter(55)

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

	if p.Critical {
		w.WriteInt(1)
		w.WriteShort(0) // padding for critical
	} else {
		w.WriteInt(0)
	}

	w.WriteInt(p.TargetX)
	w.WriteInt(p.TargetY)
	w.WriteInt(p.TargetZ)

	return w.Bytes(), nil
}
