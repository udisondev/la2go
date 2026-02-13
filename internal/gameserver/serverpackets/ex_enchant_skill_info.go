package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// ExEnchantSkillInfo sends skill enchant cost/rate info.
// S2C opcode 0xFE, sub-opcode 0x18.
type ExEnchantSkillInfo struct {
	SkillID    int32
	SkillLevel int32
	SpCost     int32
	ExpCost    int64
	Rate       int32 // success chance 0-100
	HasBookReq bool  // whether enchant book (item 6622) is required
}

// Write serializes ExEnchantSkillInfo packet to binary format.
func (p *ExEnchantSkillInfo) Write() ([]byte, error) {
	reqCount := int32(0)
	if p.HasBookReq {
		reqCount = 1
	}

	w := packet.NewWriter(32)
	w.WriteByte(0xFE)
	w.WriteShort(0x18)
	w.WriteInt(p.SkillID)
	w.WriteInt(p.SkillLevel)
	w.WriteInt(p.SpCost)
	w.WriteLong(p.ExpCost)
	w.WriteInt(p.Rate)
	w.WriteInt(reqCount)
	if p.HasBookReq {
		w.WriteInt(99)   // type = item requirement
		w.WriteInt(6622) // Ancient Book of Enchant Skill
		w.WriteLong(1)   // count = 1
		w.WriteInt(50)   // unk (Java constant)
	}
	return w.Bytes(), nil
}
