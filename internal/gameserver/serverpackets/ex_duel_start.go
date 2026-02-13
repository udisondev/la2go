package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExDuelStart is the sub-opcode for ExDuelStart (S2C 0xFE:0x4D).
const SubOpcodeExDuelStart int16 = 0x4D

// ExDuelStart packet (S2C 0xFE:0x4D) notifies clients that the duel has started.
// Sent after countdown reaches 0.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x4D
//   - partyDuel (int32) — 0 = 1v1, 1 = party duel
//
// Phase 20: Duel System.
type ExDuelStart struct {
	PartyDuel bool
}

// Write serializes ExDuelStart packet to binary format.
func (p ExDuelStart) Write() ([]byte, error) {
	// 1 opcode + 2 subop + 4 partyDuel = 7
	w := packet.NewWriter(7)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExDuelStart)

	var partyDuel int32
	if p.PartyDuel {
		partyDuel = 1
	}
	w.WriteInt(partyDuel)

	return w.Bytes(), nil
}
