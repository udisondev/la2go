package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
	"github.com/udisondev/la2go/internal/model"
)

const (
	// OpcodeSkillList is the opcode for SkillList packet (S2C 0x58)
	OpcodeSkillList = 0x58
)

// SkillList packet (S2C 0x58) sends list of character's skills.
// Sent after UserInfo during spawn and after level-up when new skills are granted.
//
// Phase 5.9.2: Skill Trees & Player Skills.
type SkillList struct {
	Skills []*model.SkillInfo
}

// NewSkillList creates SkillList packet with player's learned skills.
// Pass nil or empty slice for characters with no skills.
func NewSkillList(skills []*model.SkillInfo) *SkillList {
	return &SkillList{Skills: skills}
}

// Write serializes SkillList packet to binary format.
// Format: opcode(1) + count(4) + [passive(4) + level(4) + skillID(4) + disabled(1)] per skill.
func (p *SkillList) Write() ([]byte, error) {
	count := len(p.Skills)
	// 1 (opcode) + 4 (count) + 13*count (passive+level+skillID+disabled per skill)
	w := packet.NewWriter(5 + 13*count)

	w.WriteByte(OpcodeSkillList)
	w.WriteInt(int32(count))

	for _, s := range p.Skills {
		passive := int32(0)
		if s.Passive {
			passive = 1
		}
		w.WriteInt(passive)    // 1=passive, 0=active
		w.WriteInt(s.Level)    // skill level
		w.WriteInt(s.SkillID)  // skill ID
		w.WriteByte(0)         // disabled: 0=enabled
	}

	return w.Bytes(), nil
}
