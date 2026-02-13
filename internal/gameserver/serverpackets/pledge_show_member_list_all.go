package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgeShowMemberListAll shows full clan member list (S2C 0x53).
//
// Java reference: PledgeShowMemberListAll.java (opcode 0x53).
const OpcodePledgeShowMemberListAll = 0x53

// PledgeMemberEntry is a single member in the member list.
type PledgeMemberEntry struct {
	Name       string
	Level      int32
	ClassID    int32
	Gender     int32 // 0=male, 1=female
	RaceID     int32
	Online     int32 // 1=online, 0=offline
	PledgeType int32 // Sub-pledge type
}

// PledgeShowMemberListAll sends the full member list of a clan.
//
// Packet structure (S2C 0x53):
//   - opcode        byte
//   - pledgeType    int32  (0=main, etc.)
//   - clanName      string
//   - leaderName    string
//   - crestID       int32
//   - clanLevel     int32
//   - castleID      int32
//   - hallID        int32
//   - rank          int32  (0 = no castle rank)
//   - reputation    int32
//   - reserved1     int32  (0)
//   - reserved2     int32  (0)
//   - allyID        int32
//   - allyName      string
//   - allyCrestID   int32
//   - atWar         int32
//   - memberCount   int32
//   - [members]     repeated
type PledgeShowMemberListAll struct {
	PledgeType int32
	ClanName   string
	LeaderName string
	CrestID    int32
	ClanLevel  int32
	CastleID   int32
	HallID     int32
	Rank       int32
	Reputation int32
	AllyID     int32
	AllyName   string
	AllyCrest  int32
	AtWar      int32
	Members    []PledgeMemberEntry
}

// Write serializes the packet.
func (p *PledgeShowMemberListAll) Write() ([]byte, error) {
	w := packet.NewWriter(256)

	w.WriteByte(OpcodePledgeShowMemberListAll)
	w.WriteInt(p.PledgeType)
	w.WriteString(p.ClanName)
	w.WriteString(p.LeaderName)
	w.WriteInt(p.CrestID)
	w.WriteInt(p.ClanLevel)
	w.WriteInt(p.CastleID)
	w.WriteInt(p.HallID)
	w.WriteInt(p.Rank)
	w.WriteInt(p.Reputation)
	w.WriteInt(0) // reserved
	w.WriteInt(0) // reserved
	w.WriteInt(p.AllyID)
	w.WriteString(p.AllyName)
	w.WriteInt(p.AllyCrest)
	w.WriteInt(p.AtWar)
	w.WriteInt(int32(len(p.Members)))

	for i := range p.Members {
		m := &p.Members[i]
		w.WriteString(m.Name)
		w.WriteInt(m.Level)
		w.WriteInt(m.ClassID)
		w.WriteInt(m.Gender)
		w.WriteInt(m.RaceID)
		w.WriteInt(m.Online)
		w.WriteInt(m.PledgeType)
	}

	return w.Bytes(), nil
}
