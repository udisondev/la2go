package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExDuelReady is the sub-opcode for ExDuelReady (S2C 0xFE:0x4C).
const SubOpcodeExDuelReady int16 = 0x4C

// ExDuelReady packet (S2C 0xFE:0x4C) notifies clients that duel is about to start.
// Sent when both players accepted and duel countdown begins.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x4C
//   - partyDuel (int32) — 0 = 1v1, 1 = party duel
//
// Phase 20: Duel System.
type ExDuelReady struct {
	PartyDuel bool
}

// Write serializes ExDuelReady packet to binary format.
func (p ExDuelReady) Write() ([]byte, error) {
	// 1 opcode + 2 subop + 4 partyDuel = 7
	w := packet.NewWriter(7)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExDuelReady)

	var partyDuel int32
	if p.PartyDuel {
		partyDuel = 1
	}
	w.WriteInt(partyDuel)

	return w.Bytes(), nil
}
