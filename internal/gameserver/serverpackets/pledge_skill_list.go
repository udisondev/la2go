package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgeSkillList sends the clan skill list (S2C 0xFE sub 0x39).
//
// Java reference: PledgeSkillList.java.
const OpcodePledgeSkillList = 0xFE

// SubOpcodePledgeSkillList is the sub-opcode.
// Java: PLEDGE_SKILL_LIST(0xFE, 0x39). Note: 0x3A is PLEDGE_SKILL_LIST_ADD.
const SubOpcodePledgeSkillList int32 = 0x39

// PledgeSkillEntry is a single clan skill.
type PledgeSkillEntry struct {
	SkillID    int32
	SkillLevel int32
}

// PledgeSkillList sends all clan skills.
type PledgeSkillList struct {
	Skills []PledgeSkillEntry
}

// Write serializes the packet.
func (p *PledgeSkillList) Write() ([]byte, error) {
	w := packet.NewWriter(64)

	w.WriteByte(OpcodePledgeSkillList)
	w.WriteShort(int16(SubOpcodePledgeSkillList))
	w.WriteInt(int32(len(p.Skills)))

	for i := range p.Skills {
		w.WriteInt(p.Skills[i].SkillID)
		w.WriteInt(p.Skills[i].SkillLevel)
	}

	return w.Bytes(), nil
}
