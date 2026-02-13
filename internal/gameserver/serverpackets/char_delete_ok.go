package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCharDeleteOk is the opcode for CharDeleteOk (S2C 0x23).
// Java reference: ServerPackets.CHAR_DELETE_OK(0x23).
const OpcodeCharDeleteOk = 0x23

// CharDeleteOk confirms successful character deletion (or timer set).
//
// Packet structure:
//   - opcode (byte) — 0x23
type CharDeleteOk struct{}

// Write serializes CharDeleteOk packet.
func (p *CharDeleteOk) Write() ([]byte, error) {
	w := packet.NewWriter(4)
	w.WriteByte(OpcodeCharDeleteOk)
	return w.Bytes(), nil
}
