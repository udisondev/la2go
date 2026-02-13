package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// SubOpcodeExDuelAskStart is the sub-opcode for ExDuelAskStart (S2C 0xFE:0x4B).
const SubOpcodeExDuelAskStart int16 = 0x4B

// ExDuelAskStart packet (S2C 0xFE:0x4B) asks a player if they accept a duel challenge.
//
// Packet structure:
//   - opcode (byte) — 0xFE
//   - subOpcode (short) — 0x4B
//   - requestorName (string) — challenger's name
//   - partyDuel (int32) — 0 = 1v1, 1 = party duel
//
// Phase 20: Duel System.
type ExDuelAskStart struct {
	RequestorName string
	PartyDuel     bool
}

// Write serializes ExDuelAskStart packet to binary format.
func (p ExDuelAskStart) Write() ([]byte, error) {
	// 1 opcode + 2 subop + name (2*len + 2 null-term) + 4 partyDuel
	w := packet.NewWriter(32 + len(p.RequestorName)*2)

	w.WriteByte(0xFE)
	w.WriteShort(SubOpcodeExDuelAskStart)
	w.WriteString(p.RequestorName)

	var partyDuel int32
	if p.PartyDuel {
		partyDuel = 1
	}
	w.WriteInt(partyDuel)

	return w.Bytes(), nil
}
