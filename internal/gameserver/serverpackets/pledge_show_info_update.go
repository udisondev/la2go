package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgeShowInfoUpdate shows updated clan info to all online members (S2C 0x88).
//
// Java reference: PledgeShowInfoUpdate.java (opcode 0x88).
const OpcodePledgeShowInfoUpdate = 0x88

// PledgeShowInfoUpdate sends updated clan info to the client.
//
// Packet structure (S2C 0x88):
//   - opcode       byte   0x88
//   - clanID       int32
//   - crestID      int32
//   - clanLevel    int32
//   - hasCastle    int32  (castle ID, 0=none)
//   - hasHideout   int32  (clan hall ID, 0=none)
//   - allyID       int32
//   - allyName     string
//   - allyCrestID  int32
//   - atWar        int32  (1 if at war, 0 otherwise)
type PledgeShowInfoUpdate struct {
	ClanID      int32
	CrestID     int32
	ClanLevel   int32
	HasCastle   int32
	HasHideout  int32
	AllyID      int32
	AllyName    string
	AllyCrestID int32
	AtWar       int32
}

// NewPledgeShowInfoUpdate creates a new packet.
func NewPledgeShowInfoUpdate(clanID, crestID, level, castle, hideout, allyID int32, allyName string, allyCrestID, atWar int32) PledgeShowInfoUpdate {
	return PledgeShowInfoUpdate{
		ClanID:      clanID,
		CrestID:     crestID,
		ClanLevel:   level,
		HasCastle:   castle,
		HasHideout:  hideout,
		AllyID:      allyID,
		AllyName:    allyName,
		AllyCrestID: allyCrestID,
		AtWar:       atWar,
	}
}

// Write serializes the packet.
func (p *PledgeShowInfoUpdate) Write() ([]byte, error) {
	w := packet.NewWriter(64)

	w.WriteByte(OpcodePledgeShowInfoUpdate)
	w.WriteInt(p.ClanID)
	w.WriteInt(p.CrestID)
	w.WriteInt(p.ClanLevel)
	w.WriteInt(p.HasCastle)
	w.WriteInt(p.HasHideout)
	w.WriteInt(p.AllyID)
	w.WriteString(p.AllyName)
	w.WriteInt(p.AllyCrestID)
	w.WriteInt(p.AtWar)

	return w.Bytes(), nil
}
