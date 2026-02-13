package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePartySmallWindowDeleteAll is the server packet opcode for disbanding party (S2C 0x50).
// Sent to all members when the party is disbanded.
//
// Java reference: PartySmallWindowDeleteAll.java (opcode 0x50).
const OpcodePartySmallWindowDeleteAll = 0x50

// PartySmallWindowDeleteAll represents the party disband packet.
// Clears the party window on the client side.
//
// Packet structure (S2C 0x50):
//   - opcode byte 0x50
//   - (empty body)
type PartySmallWindowDeleteAll struct{}

// NewPartySmallWindowDeleteAll creates a PartySmallWindowDeleteAll packet.
func NewPartySmallWindowDeleteAll() PartySmallWindowDeleteAll {
	return PartySmallWindowDeleteAll{}
}

// Write serializes the PartySmallWindowDeleteAll packet to bytes.
func (p *PartySmallWindowDeleteAll) Write() ([]byte, error) {
	w := packet.NewWriter(4) // opcode(1) + minimal padding

	w.WriteByte(OpcodePartySmallWindowDeleteAll)

	return w.Bytes(), nil
}
