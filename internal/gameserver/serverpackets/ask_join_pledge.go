package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAskJoinPledge is the clan invite dialog prompt (S2C 0x32).
// Sent to target player: "Clan [name] invites you. Accept?"
//
// Java reference: AskJoinPledge.java (opcode 0x32).
const OpcodeAskJoinPledge = 0x32

// AskJoinPledge asks a player to join a clan.
//
// Packet structure (S2C 0x32):
//   - opcode      byte
//   - requestorID int32  objectID of the player who sent the invite
//   - clanName    string name of the clan
type AskJoinPledge struct {
	RequestorID int32
	ClanName    string
}

// NewAskJoinPledge creates a new AskJoinPledge packet.
func NewAskJoinPledge(requestorID int32, clanName string) AskJoinPledge {
	return AskJoinPledge{
		RequestorID: requestorID,
		ClanName:    clanName,
	}
}

// Write serializes the packet.
func (p *AskJoinPledge) Write() ([]byte, error) {
	w := packet.NewWriter(64)

	w.WriteByte(OpcodeAskJoinPledge)
	w.WriteInt(p.RequestorID)
	w.WriteString(p.ClanName)

	return w.Bytes(), nil
}
