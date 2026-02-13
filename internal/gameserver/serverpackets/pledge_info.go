package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgeInfo is the server packet for clan info display (S2C 0x83).
// Sent when a player clicks on another player's clan name.
//
// Java reference: PledgeInfo.java (opcode 0x83).
const OpcodePledgeInfo = 0x83

// PledgeInfo sends basic clan information to the client.
//
// Packet structure (S2C 0x84):
//   - opcode   byte   0x84
//   - clanID   int32  clan ID
//   - clanName string clan name (UTF-16LE)
//   - allyName string alliance name (UTF-16LE)
type PledgeInfo struct {
	ClanID   int32
	ClanName string
	AllyName string
}

// NewPledgeInfo creates a new PledgeInfo packet.
func NewPledgeInfo(clanID int32, clanName, allyName string) PledgeInfo {
	return PledgeInfo{
		ClanID:   clanID,
		ClanName: clanName,
		AllyName: allyName,
	}
}

// Write serializes the PledgeInfo packet.
func (p *PledgeInfo) Write() ([]byte, error) {
	w := packet.NewWriter(64)

	w.WriteByte(OpcodePledgeInfo)
	w.WriteInt(p.ClanID)
	w.WriteString(p.ClanName)
	w.WriteString(p.AllyName)

	return w.Bytes(), nil
}
