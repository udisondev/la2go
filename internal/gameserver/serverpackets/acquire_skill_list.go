package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeAcquireSkillList is the S2C opcode 0x8A.
const OpcodeAcquireSkillList byte = 0x8A

// AcquireSkillEntry is a single learnable skill entry.
type AcquireSkillEntry struct {
	SkillID      int32
	NextLevel    int32
	MaxLevel     int32
	SpCost       int32
	Requirements int32 // bitmask, always 0 for CLASS
}

// AcquireSkillList sends the list of skills available for learning.
type AcquireSkillList struct {
	SkillType int32 // 0=CLASS, 1=FISHING, 2=PLEDGE
	Skills    []AcquireSkillEntry
}

// Write serializes the packet.
func (p *AcquireSkillList) Write() ([]byte, error) {
	// 1 + 4 + 4 + count*20
	w := packet.NewWriter(9 + len(p.Skills)*20)
	w.WriteByte(OpcodeAcquireSkillList)
	w.WriteInt(p.SkillType)
	w.WriteInt(int32(len(p.Skills)))

	for _, s := range p.Skills {
		w.WriteInt(s.SkillID)
		w.WriteInt(s.NextLevel)
		w.WriteInt(s.MaxLevel)
		w.WriteInt(s.SpCost)
		w.WriteInt(s.Requirements)
	}

	return w.Bytes(), nil
}
