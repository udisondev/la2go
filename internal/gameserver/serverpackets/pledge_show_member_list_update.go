package serverpackets

import (
	"github.com/udisondev/la2go/internal/gameserver/packet"
)

// OpcodePledgeShowMemberListUpdate updates a single member in the member list (S2C 0x54).
//
// Java reference: PledgeShowMemberListUpdate.java (opcode 0x54).
const OpcodePledgeShowMemberListUpdate = 0x54

// PledgeShowMemberListUpdate updates a single clan member entry.
//
// Packet structure (S2C 0x54):
//   - opcode     byte
//   - name       string
//   - level      int32
//   - classID    int32
//   - gender     int32
//   - raceID     int32
//   - online     int32
//   - pledgeType int32
type PledgeShowMemberListUpdate struct {
	Name       string
	Level      int32
	ClassID    int32
	Gender     int32
	RaceID     int32
	Online     int32
	PledgeType int32
}

// Write serializes the packet.
func (p *PledgeShowMemberListUpdate) Write() ([]byte, error) {
	w := packet.NewWriter(64)

	w.WriteByte(OpcodePledgeShowMemberListUpdate)
	w.WriteString(p.Name)
	w.WriteInt(p.Level)
	w.WriteInt(p.ClassID)
	w.WriteInt(p.Gender)
	w.WriteInt(p.RaceID)
	w.WriteInt(p.Online)
	w.WriteInt(p.PledgeType)

	return w.Bytes(), nil
}
