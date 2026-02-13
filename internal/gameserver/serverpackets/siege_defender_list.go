package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSiegeDefenderList is the server packet for siege defender list (S2C 0xCB).
//
// Java reference: SiegeDefenderList.java.
const OpcodeSiegeDefenderList = 0xCB

// Defender type constants for the client.
const (
	DefenderTypeOwner    = 1 // Castle owner
	DefenderTypePending  = 2 // Waiting for approval
	DefenderTypeApproved = 3 // Approved defender
)

// SiegeDefenderEntry represents one defending clan.
type SiegeDefenderEntry struct {
	ClanID      int32
	ClanName    string
	LeaderName  string
	CrestID     int32
	Type        int32 // DefenderTypeOwner/DefenderTypePending/DefenderTypeApproved
	AllyID      int32
	AllyName    string
	AllyCrestID int32
}

// SiegeDefenderList sends the list of defending clans for a castle siege.
//
// Packet structure (S2C 0xCB):
//   - opcode      byte   0xCB
//   - castleID    int32
//   - 00 00 00 00 int32  unknown
//   - 01 00 00 00 int32  unknown
//   - 00 00 00 00 int32  unknown
//   - count       int32  total defender count
//   - count       int32  total defender count (duplicate)
//   - for each defender:
//     - clanID      int32
//     - clanName    string
//     - leaderName  string
//     - crestID     int32
//     - signedTime  int32  always 0
//     - type        int32  1=owner, 2=pending, 3=approved
//     - allyID      int32
//     - allyName    string
//     - allyLeader  string always ""
//     - allyCrestID int32
type SiegeDefenderList struct {
	CastleID  int32
	Defenders []SiegeDefenderEntry
}

// Write serializes the SiegeDefenderList packet.
func (p *SiegeDefenderList) Write() ([]byte, error) {
	w := packet.NewWriter(128 + len(p.Defenders)*128)

	w.WriteByte(OpcodeSiegeDefenderList)
	w.WriteInt(p.CastleID)
	w.WriteInt(0) // unknown
	w.WriteInt(1) // unknown
	w.WriteInt(0) // unknown

	count := int32(len(p.Defenders))
	w.WriteInt(count)
	w.WriteInt(count)

	for _, d := range p.Defenders {
		w.WriteInt(d.ClanID)
		w.WriteString(d.ClanName)
		w.WriteString(d.LeaderName)
		w.WriteInt(d.CrestID)
		w.WriteInt(0)      // signedTime (not stored)
		w.WriteInt(d.Type) // 1=owner, 2=pending, 3=approved
		w.WriteInt(d.AllyID)
		w.WriteString(d.AllyName)
		w.WriteString("") // AllyLeaderName
		w.WriteInt(d.AllyCrestID)
	}

	return w.Bytes(), nil
}
