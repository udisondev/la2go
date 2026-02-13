package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeCharCreateOk is the opcode for CharCreateOk (S2C 0x19).
// Java reference: ServerPackets.CHAR_CREATE_OK(0x19).
const OpcodeCharCreateOk = 0x19

// CharCreateOk confirms successful character creation.
//
// Packet structure:
//   - opcode (byte) — 0x19
//   - result (int32) — always 1
type CharCreateOk struct{}

// Write serializes CharCreateOk packet.
func (p *CharCreateOk) Write() ([]byte, error) {
	w := packet.NewWriter(8)
	w.WriteByte(OpcodeCharCreateOk)
	w.WriteInt(1) // success
	return w.Bytes(), nil
}
