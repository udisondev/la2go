package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeJoinParty is the server packet opcode for party join response (S2C 0x3A).
// Sent to the requester to indicate whether the invite was accepted.
//
// Java reference: JoinParty.java (opcode 0x3A).
const OpcodeJoinParty = 0x3A

// JoinParty represents the server response to a party invite attempt.
//
// Packet structure (S2C 0x3A):
//   - opcode   byte   0x3A
//   - response int32  0 = declined, 1 = accepted
type JoinParty struct {
	Response int32
}

// NewJoinParty creates a new JoinParty packet.
func NewJoinParty(response int32) JoinParty {
	return JoinParty{Response: response}
}

// Write serializes the JoinParty packet to bytes.
func (p *JoinParty) Write() ([]byte, error) {
	w := packet.NewWriter(8) // opcode(1) + response(4)

	w.WriteByte(OpcodeJoinParty)
	w.WriteInt(p.Response)

	return w.Bytes(), nil
}
