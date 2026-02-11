package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeMagicSkillLaunched is the server packet opcode for skill effect application.
// Sent when a skill finishes casting and effects are applied to targets.
//
// Phase 5.9.4: Cast Flow & Packets.
// Java reference: MagicSkillLaunched.java (opcode 0x76)
const OpcodeMagicSkillLaunched = 0x76

// MagicSkillLaunched packet (S2C 0x76) — broadcast when skill effects land.
type MagicSkillLaunched struct {
	CasterObjectID int32
	SkillID        int32
	SkillLevel     int32
	Targets        []int32 // target object IDs
}

// NewMagicSkillLaunched creates a MagicSkillLaunched packet.
func NewMagicSkillLaunched(casterID, skillID, skillLevel int32, targets []int32) *MagicSkillLaunched {
	return &MagicSkillLaunched{
		CasterObjectID: casterID,
		SkillID:        skillID,
		SkillLevel:     skillLevel,
		Targets:        targets,
	}
}

// Write serializes the MagicSkillLaunched packet.
func (p *MagicSkillLaunched) Write() ([]byte, error) {
	// opcode(1) + casterID(4) + skillID(4) + skillLevel(4) + targetCount(4) + targets(4*n)
	w := packet.NewWriter(17 + 4*len(p.Targets))

	w.WriteByte(OpcodeMagicSkillLaunched)
	w.WriteInt(p.CasterObjectID)
	w.WriteInt(p.SkillID)
	w.WriteInt(p.SkillLevel)
	w.WriteInt(int32(len(p.Targets)))
	for _, targetID := range p.Targets {
		w.WriteInt(targetID)
	}

	return w.Bytes(), nil
}
