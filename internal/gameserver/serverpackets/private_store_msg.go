package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePrivateStoreMsgSell is the opcode for PrivateStoreMsgSell (S2C 0x9C).
// Broadcasts the sell store message to nearby players.
//
// Phase 8.1: Private Store System.
// Java reference: PrivateStoreMsgSell.java
const OpcodePrivateStoreMsgSell = 0x9C

// OpcodePrivateStoreMsgBuy is the opcode for PrivateStoreMsgBuy (S2C 0xB9).
// Broadcasts the buy store message to nearby players.
//
// Phase 8.1: Private Store System.
// Java reference: PrivateStoreMsgBuy.java
const OpcodePrivateStoreMsgBuy = 0xB9

// PrivateStoreMsgSell broadcasts the sell store title to nearby players.
//
// Packet structure:
//   - opcode (byte) — 0x9C
//   - objectID (int32) — store owner's ObjectID
//   - message (string) — store title (UTF-16LE null-terminated)
type PrivateStoreMsgSell struct {
	ObjectID uint32
	Message  string
}

// Write serializes PrivateStoreMsgSell packet.
func (p *PrivateStoreMsgSell) Write() ([]byte, error) {
	w := packet.NewWriter(1 + 4 + len(p.Message)*2 + 2)

	w.WriteByte(OpcodePrivateStoreMsgSell)
	w.WriteInt(int32(p.ObjectID))
	w.WriteString(p.Message)

	return w.Bytes(), nil
}

// PrivateStoreMsgBuy broadcasts the buy store title to nearby players.
//
// Packet structure:
//   - opcode (byte) — 0xB9
//   - objectID (int32) — store owner's ObjectID
//   - message (string) — store title (UTF-16LE null-terminated)
type PrivateStoreMsgBuy struct {
	ObjectID uint32
	Message  string
}

// Write serializes PrivateStoreMsgBuy packet.
func (p *PrivateStoreMsgBuy) Write() ([]byte, error) {
	w := packet.NewWriter(1 + 4 + len(p.Message)*2 + 2)

	w.WriteByte(OpcodePrivateStoreMsgBuy)
	w.WriteInt(int32(p.ObjectID))
	w.WriteString(p.Message)

	return w.Bytes(), nil
}
