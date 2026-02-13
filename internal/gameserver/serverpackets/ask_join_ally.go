package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeAskJoinAlly is the S2C opcode 0xA8 for alliance join request dialog.
const OpcodeAskJoinAlly byte = 0xA8

// AskJoinAlly asks a clan leader to join an alliance.
//
// Packet structure (S2C 0xA8):
//   - opcode             byte    0xA8
//   - requestorObjectID  int32   object ID of the requestor
//   - allyName           string  name of the alliance to join
type AskJoinAlly struct {
	RequestorObjectID int32
	AllyName          string
}

// Write serializes the AskJoinAlly packet.
func (p *AskJoinAlly) Write() ([]byte, error) {
	w := packet.NewWriter(64)
	w.WriteByte(OpcodeAskJoinAlly)
	w.WriteInt(p.RequestorObjectID)
	w.WriteString(p.AllyName)
	return w.Bytes(), nil
}
