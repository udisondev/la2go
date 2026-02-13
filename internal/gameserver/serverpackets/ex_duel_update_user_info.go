package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExDuelUpdateUserInfo is the sub-opcode for ExDuelUpdateUserInfo (S2C 0xFE:0x4F).
const SubOpcodeExDuelUpdateUserInfo int16 = 0x4F

// ExDuelUpdateUserInfo packet (S2C 0xFE:0x4F) sends opponent's info in a duel.
// Sent to both participants at duel start and during HP/MP/CP updates.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x4F
//   - objectID (int32)
//   - name (string)
//   - currentHP (int32)
//   - maxHP (int32)
//   - currentMP (int32)
//   - maxMP (int32)
//   - currentCP (int32)
//   - maxCP (int32)
//
// Phase 20: Duel System.
type ExDuelUpdateUserInfo struct {
	ObjectID  uint32
	Name      string
	CurrentHP int32
	MaxHP     int32
	CurrentMP int32
	MaxMP     int32
	CurrentCP int32
	MaxCP     int32
}

// Write serializes ExDuelUpdateUserInfo packet to binary format.
func (p ExDuelUpdateUserInfo) Write() ([]byte, error) {
	// 1 opcode + 2 subop + 4 objectID + name + 6×4 stats = ~64
	w := packet.NewWriter(64 + len(p.Name)*2)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExDuelUpdateUserInfo)
	w.WriteInt(int32(p.ObjectID))
	w.WriteString(p.Name)
	w.WriteInt(p.CurrentHP)
	w.WriteInt(p.MaxHP)
	w.WriteInt(p.CurrentMP)
	w.WriteInt(p.MaxMP)
	w.WriteInt(p.CurrentCP)
	w.WriteInt(p.MaxCP)

	return w.Bytes(), nil
}
