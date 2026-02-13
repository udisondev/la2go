package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeJoinPledge is the clan join confirmation (S2C 0x33).
// Sent to the new member after they accept the invite.
//
// Java reference: JoinPledge.java (opcode 0x33).
const OpcodeJoinPledge = 0x33

// JoinPledge confirms a player has joined a clan.
//
// Packet structure (S2C 0x35):
//   - opcode     byte
//   - pledgeID   int32  clan ID
type JoinPledge struct {
	PledgeID int32
}

// NewJoinPledge creates a new JoinPledge packet.
func NewJoinPledge(pledgeID int32) JoinPledge {
	return JoinPledge{PledgeID: pledgeID}
}

// Write serializes the packet.
func (p *JoinPledge) Write() ([]byte, error) {
	w := packet.NewWriter(8)

	w.WriteByte(OpcodeJoinPledge)
	w.WriteInt(p.PledgeID)

	return w.Bytes(), nil
}
