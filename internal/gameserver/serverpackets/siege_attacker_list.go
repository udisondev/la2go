package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSiegeAttackerList is the server packet for siege attacker list (S2C 0xCA).
//
// Java reference: SiegeAttackerList.java.
const OpcodeSiegeAttackerList = 0xCA

// SiegeAttackerEntry represents one attacking clan.
type SiegeAttackerEntry struct {
	ClanID      int32
	ClanName    string
	LeaderName  string
	CrestID     int32
	AllyID      int32
	AllyName    string
	AllyCrestID int32
}

// SiegeAttackerList sends the list of attacking clans for a castle siege.
//
// Packet structure (S2C 0xCA):
//   - opcode      byte   0xCA
//   - castleID    int32
//   - 00 00 00 00 int32  unknown
//   - 01 00 00 00 int32  unknown
//   - 00 00 00 00 int32  unknown
//   - count       int32  attacker count
//   - count       int32  attacker count (duplicate)
//   - for each attacker:
//     - clanID      int32
//     - clanName    string
//     - leaderName  string
//     - crestID     int32
//     - signedTime  int32  always 0
//     - allyID      int32
//     - allyName    string
//     - allyLeader  string always ""
//     - allyCrestID int32
type SiegeAttackerList struct {
	CastleID  int32
	Attackers []SiegeAttackerEntry
}

// Write serializes the SiegeAttackerList packet.
func (p *SiegeAttackerList) Write() ([]byte, error) {
	w := packet.NewWriter(128 + len(p.Attackers)*128)

	w.WriteByte(OpcodeSiegeAttackerList)
	w.WriteInt(p.CastleID)
	w.WriteInt(0) // unknown
	w.WriteInt(1) // unknown
	w.WriteInt(0) // unknown

	count := int32(len(p.Attackers))
	w.WriteInt(count)
	w.WriteInt(count)

	for _, a := range p.Attackers {
		w.WriteInt(a.ClanID)
		w.WriteString(a.ClanName)
		w.WriteString(a.LeaderName)
		w.WriteInt(a.CrestID)
		w.WriteInt(0) // signedTime (not stored)
		w.WriteInt(a.AllyID)
		w.WriteString(a.AllyName)
		w.WriteString("") // AllyLeaderName
		w.WriteInt(a.AllyCrestID)
	}

	return w.Bytes(), nil
}
