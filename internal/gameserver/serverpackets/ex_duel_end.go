package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExDuelEnd is the sub-opcode for ExDuelEnd (S2C 0xFE:0x4E).
const SubOpcodeExDuelEnd int16 = 0x4E

// ExDuelEnd packet (S2C 0xFE:0x4E) notifies clients that the duel has ended.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x4E
//   - partyDuel (int32) — 0 = 1v1, 1 = party duel
//
// Phase 20: Duel System.
type ExDuelEnd struct {
	PartyDuel bool
}

// Write serializes ExDuelEnd packet to binary format.
func (p ExDuelEnd) Write() ([]byte, error) {
	// 1 opcode + 2 subop + 4 partyDuel = 7
	w := packet.NewWriter(7)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExDuelEnd)

	var partyDuel int32
	if p.PartyDuel {
		partyDuel = 1
	}
	w.WriteInt(partyDuel)

	return w.Bytes(), nil
}
