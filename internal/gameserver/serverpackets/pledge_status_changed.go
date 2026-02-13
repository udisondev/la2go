package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgeStatusChanged notifies about clan status change (S2C 0xCD).
// Sent to update the client when clan level, crest, or ally changes.
//
// Java reference: PledgeStatusChanged.java — regular packet (NOT extended).
const OpcodePledgeStatusChanged = 0xCD

// PledgeStatusChanged updates clan status fields.
//
// Packet structure:
//
//	opcode (byte) = 0xCD
//	leaderID (int32) — leader ObjectID
//	clanID (int32)
//	crestID (int32) — clan crest
//	allyID (int32) — alliance ID (0 if none)
//	allyCrestID (int32) — alliance crest (0 if none)
//	unknown1 (int32) — always 0
//	unknown2 (int32) — always 0
type PledgeStatusChanged struct {
	LeaderID    int32
	ClanID      int32
	CrestID     int32
	AllyID      int32
	AllyCrestID int32
}

// Write serializes the packet.
func (p *PledgeStatusChanged) Write() ([]byte, error) {
	w := packet.NewWriter(32)

	w.WriteByte(OpcodePledgeStatusChanged)
	w.WriteInt(p.LeaderID)
	w.WriteInt(p.ClanID)
	w.WriteInt(p.CrestID)
	w.WriteInt(p.AllyID)
	w.WriteInt(p.AllyCrestID)
	w.WriteInt(0) // unknown1
	w.WriteInt(0) // unknown2

	return w.Bytes(), nil
}
