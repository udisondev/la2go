package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// SubOpcodeExAutoSoulShot is the S2C extended sub-opcode 0xFE:0x12.
const SubOpcodeExAutoSoulShot int16 = 0x12

// ExAutoSoulShot confirms auto-soulshot toggle to the client.
type ExAutoSoulShot struct {
	ItemID int32 // SoulShot/SpiritShot item ID
	Type   int32 // 1 = on, 0 = off
}

// Write serializes the packet.
func (p *ExAutoSoulShot) Write() ([]byte, error) {
	w := packet.NewWriter(11) // 1 + 2 + 4 + 4
	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExAutoSoulShot)
	w.WriteInt(p.ItemID)
	w.WriteInt(p.Type)
	return w.Bytes(), nil
}
