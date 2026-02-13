package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// ExEnchantSkillResult sends enchant skill outcome.
// S2C opcode 0xFE, sub-opcode 0x19.
type ExEnchantSkillResult struct {
	Result int32 // 0=fail, 1=success
}

// Write serializes ExEnchantSkillResult packet to binary format.
func (p *ExEnchantSkillResult) Write() ([]byte, error) {
	w := packet.NewWriter(8)
	w.WriteByte(0xFE)
	w.WriteShort(0x19)
	w.WriteInt(p.Result)
	return w.Bytes(), nil
}
