package serverpackets

import (
	"time"

	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodeSiegeInfo is the server packet for siege information (S2C 0xC9).
//
// Java reference: SiegeInfo.java.
const OpcodeSiegeInfo = 0xC9

// SiegeInfo sends siege info for a castle.
//
// Packet structure (S2C 0xC9):
//   - opcode       byte   0xC9
//   - castleID     int32
//   - canManage    int32  1 if caller is castle owner's leader, 0 otherwise
//   - ownerClanID  int32
//   - ownerName    string clan name
//   - leaderName   string clan leader name
//   - allyID       int32
//   - allyName     string
//   - currentTime  int32  unix timestamp (seconds)
//   - siegeTime    int32  unix timestamp (seconds), 0 if owner can choose time
//   - hourCount    int32  number of available hour choices (0 if time is set)
//   - hours[]      int32  available timestamps for choosing
type SiegeInfo struct {
	CastleID    int32
	CanManage   bool
	OwnerClanID int32
	OwnerName   string
	LeaderName  string
	AllyID      int32
	AllyName    string
	SiegeDate   time.Time
	TimeRegOver bool
}

// Write serializes the SiegeInfo packet.
func (p *SiegeInfo) Write() ([]byte, error) {
	w := packet.NewWriter(128)

	w.WriteByte(OpcodeSiegeInfo)
	w.WriteInt(p.CastleID)

	manage := int32(0)
	if p.CanManage {
		manage = 1
	}
	w.WriteInt(manage)
	w.WriteInt(p.OwnerClanID)
	w.WriteString(p.OwnerName)
	w.WriteString(p.LeaderName)
	w.WriteInt(p.AllyID)
	w.WriteString(p.AllyName)

	w.WriteInt(int32(time.Now().Unix()))

	if p.TimeRegOver {
		w.WriteInt(int32(p.SiegeDate.Unix()))
		w.WriteInt(0) // no hour choices — time already set
	} else {
		w.WriteInt(0) // 0 = owner can choose time
		w.WriteInt(0) // no hour choices for now
	}

	return w.Bytes(), nil
}
