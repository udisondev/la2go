package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeAcquireSkillInfo is the S2C opcode 0x8B.
const OpcodeAcquireSkillInfo byte = 0x8B

// AcquireSkillReq is a single item requirement for learning a skill.
type AcquireSkillReq struct {
	Type   int32 // 99 = item requirement
	ItemID int32
	Count  int64
	Unk    int32 // always 50 in Java
}

// AcquireSkillInfo sends details about a specific skill's learning cost.
type AcquireSkillInfo struct {
	SkillID   int32
	Level     int32
	SpCost    int32
	SkillType int32 // 0=CLASS, 1=FISHING, 2=PLEDGE
	Reqs      []AcquireSkillReq
}

// Write serializes the packet.
func (p *AcquireSkillInfo) Write() ([]byte, error) {
	// 1 + 4*5 + count*(4+4+8+4)
	w := packet.NewWriter(21 + len(p.Reqs)*20)
	w.WriteByte(OpcodeAcquireSkillInfo)
	w.WriteInt(p.SkillID)
	w.WriteInt(p.Level)
	w.WriteInt(p.SpCost)
	w.WriteInt(p.SkillType)
	w.WriteInt(int32(len(p.Reqs)))

	for _, r := range p.Reqs {
		w.WriteInt(r.Type)
		w.WriteInt(r.ItemID)
		w.WriteLong(r.Count)
		w.WriteInt(r.Unk)
	}

	return w.Bytes(), nil
}
