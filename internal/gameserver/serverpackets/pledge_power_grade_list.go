package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgePowerGradeList sends the rank privilege list (S2C 0xFE sub 0x3B).
// Note: Extended packet 0xFE with sub-opcode.
//
// Java reference: PledgePowerGradeList.java.
const OpcodePledgePowerGradeList = 0xFE

// SubOpcodePledgePowerGradeList is the sub-opcode for extended packet.
// Java: PLEDGE_POWER_GRADE_LIST(0xFE, 0x3B). Note: 0x3C is PLEDGE_RECEIVE_POWER_INFO.
const SubOpcodePledgePowerGradeList int32 = 0x3B

// PledgePowerGradeEntry is a single rank entry.
type PledgePowerGradeEntry struct {
	PowerGrade int32
	Privileges int32
	MemberCount int32 // Number of members with this rank
}

// PledgePowerGradeList sends the full rank list.
type PledgePowerGradeList struct {
	Entries []PledgePowerGradeEntry
}

// Write serializes the packet.
func (p *PledgePowerGradeList) Write() ([]byte, error) {
	w := packet.NewWriter(128)

	w.WriteByte(OpcodePledgePowerGradeList)
	w.WriteShort(int16(SubOpcodePledgePowerGradeList))
	w.WriteInt(int32(len(p.Entries)))

	for i := range p.Entries {
		e := &p.Entries[i]
		w.WriteInt(e.PowerGrade)
		w.WriteInt(e.Privileges)
		w.WriteInt(e.MemberCount)
	}

	return w.Bytes(), nil
}
