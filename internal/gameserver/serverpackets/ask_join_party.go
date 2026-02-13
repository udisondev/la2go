package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeAskJoinParty is the server packet opcode for party invite dialog (S2C 0x39).
// Sent to the target player to display party invite confirmation dialog.
//
// Java reference: AskJoinParty.java (opcode 0x39).
const OpcodeAskJoinParty = 0x39

// AskJoinParty represents a party invitation dialog sent to the invited player.
//
// Packet structure (S2C 0x39):
//   - opcode           byte    0x39
//   - requesterName    string  inviting player's name (UTF-16LE null-terminated)
//   - itemDistribution int32   loot distribution rule (0-4)
type AskJoinParty struct {
	RequesterName    string
	ItemDistribution int32
}

// NewAskJoinParty creates a new AskJoinParty packet.
func NewAskJoinParty(requesterName string, itemDistribution int32) AskJoinParty {
	return AskJoinParty{
		RequesterName:    requesterName,
		ItemDistribution: itemDistribution,
	}
}

// Write serializes the AskJoinParty packet to bytes.
func (p *AskJoinParty) Write() ([]byte, error) {
	// opcode(1) + name(~32) + itemDistribution(4)
	w := packet.NewWriter(64)

	w.WriteByte(OpcodeAskJoinParty)
	w.WriteString(p.RequesterName)
	w.WriteInt(p.ItemDistribution)

	return w.Bytes(), nil
}
