package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

const (
	// OpcodeSkillList is the opcode for SkillList packet (S2C 0x58)
	OpcodeSkillList = 0x58
)

// SkillList packet (S2C 0x58) sends list of character's skills.
// Sent after UserInfo during spawn.
type SkillList struct {
	// Skills []Skill // TODO Phase 5.0: implement skill system
}

// NewSkillList creates empty SkillList packet.
// TODO Phase 5.0: Load skills from database based on class/level.
func NewSkillList() *SkillList {
	return &SkillList{}
}

// Write serializes SkillList packet to binary format.
func (p *SkillList) Write() ([]byte, error) {
	// Empty skill list for now
	w := packet.NewWriter(16)

	w.WriteByte(OpcodeSkillList)
	w.WriteInt(0) // Skill count = 0

	return w.Bytes(), nil
}
