package serverpackets

import "github.com/udisondev/la2go/internal/gameserver/packet"

// OpcodeAllianceInfo is the S2C opcode 0xB4 for alliance information.
const OpcodeAllianceInfo byte = 0xB4

// AllianceClanInfo holds per-clan details within an alliance.
type AllianceClanInfo struct {
	ClanName          string
	ClanLevel         int32
	ClanLeaderName    string
	ClanTotalMembers  int32
	ClanOnlineMembers int32
}

// AllianceInfo sends full alliance information to the client.
//
// Packet structure (S2C 0xB4):
//   - opcode          byte    0xB4
//   - allyName        string  alliance name
//   - totalMembers    int32   total member count across all clans
//   - onlineMembers   int32   online member count across all clans
//   - leaderClanName  string  name of the leader clan
//   - leaderName      string  name of the alliance leader
//   - clanCount       int32   number of clans in the alliance
//   - per clan:
//   - clanName          string  clan name
//   - unknown           int32   reserved (always 0)
//   - clanLevel         int32   clan level
//   - clanLeaderName    string  clan leader name
//   - clanTotalMembers  int32   total members in clan
//   - clanOnlineMembers int32   online members in clan
type AllianceInfo struct {
	AllyName       string
	TotalMembers   int32
	OnlineMembers  int32
	LeaderClanName string
	LeaderName     string
	Clans          []AllianceClanInfo
}

// Write serializes the AllianceInfo packet.
func (p *AllianceInfo) Write() ([]byte, error) {
	w := packet.NewWriter(256)
	w.WriteByte(OpcodeAllianceInfo)
	w.WriteString(p.AllyName)
	w.WriteInt(p.TotalMembers)
	w.WriteInt(p.OnlineMembers)
	w.WriteString(p.LeaderClanName)
	w.WriteString(p.LeaderName)
	w.WriteInt(int32(len(p.Clans)))
	for _, c := range p.Clans {
		w.WriteString(c.ClanName)
		w.WriteInt(0) // reserved field (unknown in protocol)
		w.WriteInt(c.ClanLevel)
		w.WriteString(c.ClanLeaderName)
		w.WriteInt(c.ClanTotalMembers)
		w.WriteInt(c.ClanOnlineMembers)
	}
	return w.Bytes(), nil
}
